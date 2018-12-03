/*
 *    Copyright 2018 Insolar
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package pulsemanager

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/core/message"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/jetdrop"
	"github.com/insolar/insolar/ledger/storage"
	"github.com/insolar/insolar/utils/backoff"
)

// PulseManager implements core.PulseManager.
type PulseManager struct {
	LR             core.LogicRunner       `inject:""`
	Bus            core.MessageBus        `inject:""`
	NodeNet        core.NodeNetwork       `inject:""`
	JetCoordinator core.JetCoordinator    `inject:""`
	GIL            core.GlobalInsolarLock `inject:""`
	currentPulse   core.Pulse

	// internal stuff
	db *storage.DB
	// setLock locks Set method call.
	setLock sync.RWMutex
	stopped bool
	stop    chan struct{}
	// gotpulse signals if there is something to sync to Heavy
	gotpulse chan struct{}
	// syncdone closes when sync is over
	syncdone chan struct{}
	// sync backoff instance
	syncbackoff *backoff.Backoff
	// stores pulse manager options
	options pmOptions
}

type pmOptions struct {
	enableSync       bool
	syncMessageLimit int
	pulsesDeltaLimit core.PulseNumber
}

func backoffFromConfig(bconf configuration.Backoff) *backoff.Backoff {
	return &backoff.Backoff{
		Jitter: bconf.Jitter,
		Min:    bconf.Min,
		Max:    bconf.Max,
		Factor: bconf.Factor,
	}
}

// NewPulseManager creates PulseManager instance.
func NewPulseManager(db *storage.DB, conf configuration.Ledger) *PulseManager {
	pm := &PulseManager{
		db:           db,
		gotpulse:     make(chan struct{}, 1),
		currentPulse: *core.GenesisPulse,
	}
	pmconf := conf.PulseManager
	pm.options.enableSync = pmconf.HeavySyncEnabled
	pm.options.syncMessageLimit = pmconf.HeavySyncMessageLimit
	pm.options.pulsesDeltaLimit = conf.LightChainLimit
	pm.syncbackoff = backoffFromConfig(pmconf.HeavyBackoff)
	return pm
}

// Current returns copy (for concurrency safety) of current pulse structure.
func (m *PulseManager) Current(ctx context.Context) (*core.Pulse, error) {
	m.setLock.RLock()
	defer m.setLock.RUnlock()

	p := m.currentPulse
	return &p, nil
}

func (m *PulseManager) processDrop(ctx context.Context, latestPulseNumber core.PulseNumber) error {
	latestPulse, err := m.db.GetPulse(ctx, latestPulseNumber)
	if err != nil {
		return err
	}
	prevDrop, err := m.db.GetDrop(ctx, *latestPulse.Prev)
	if err != nil {
		return err
	}
	drop, messages, err := m.db.CreateDrop(ctx, latestPulseNumber, prevDrop.Hash)
	if err != nil {
		return err
	}
	err = m.db.SetDrop(ctx, drop)
	if err != nil {
		return err
	}

	dropSerialized, err := jetdrop.Encode(drop)
	if err != nil {
		return err
	}

	msg := &message.JetDrop{
		Drop:        dropSerialized,
		Messages:    messages,
		PulseNumber: latestPulseNumber,
	}
	_, err = m.Bus.Send(ctx, msg, nil)
	if err != nil {
		return err
	}
	return nil
}

// Set set's new pulse and closes current jet drop.
func (m *PulseManager) Set(ctx context.Context, pulse core.Pulse, dry bool) error {
	// Ensure this does not execute in parallel.
	m.setLock.Lock()
	defer m.setLock.Unlock()
	if m.stopped {
		return errors.New("can't call Set method on PulseManager after stop")
	}

	var latestPulseNumber core.PulseNumber
	var err error
	m.GIL.Acquire(ctx)

	// swap pulse
	m.currentPulse = pulse

	// TODO: swap active nodes and set prev pulse state to network

	if !dry {
		latestPulseNumber, err = m.db.GetLatestPulseNumber(ctx)
		if err != nil {
			return errors.Wrap(err, "call of GetLatestPulseNumber failed")
		}

		if err := m.db.AddPulse(ctx, pulse); err != nil {
			return errors.Wrap(err, "call of AddPulse failed")
		}
		err = m.db.SetActiveNodes(pulse.PulseNumber, m.NodeNet.GetActiveNodes())
		if err != nil {
			return errors.Wrap(err, "call of SetActiveNodes failed")
		}
	}

	m.GIL.Release(ctx)

	if dry {
		return nil
	}

	// Run only on material executor.
	// execute only on material executor
	// TODO: do as much as possible async.
	if m.NodeNet.GetOrigin().Role() == core.StaticRoleLightMaterial {
		if err = m.processDrop(ctx, latestPulseNumber); err != nil {
			return errors.Wrap(err, "processDrop failed")
		}

		m.SyncToHeavy()
	}

	return m.LR.OnPulse(ctx, pulse)
}

// SyncToHeavy signals to sync loop there is something to sync.
//
// Should never be called after Stop.
func (m *PulseManager) SyncToHeavy() {
	if !m.options.enableSync {
		return
	}
	// TODO: save current pulse as last should be processed
	if len(m.gotpulse) == 0 {
		m.gotpulse <- struct{}{}
		return
	}
}

// Start starts pulse manager, spawns replication goroutine under a hood.
func (m *PulseManager) Start(ctx context.Context) error {
	m.syncdone = make(chan struct{})
	m.stop = make(chan struct{})
	if m.options.enableSync {
		synclist, err := m.NextSyncPulses(ctx)
		if err != nil {
			return err
		}
		go m.syncloop(ctx, synclist)
	}
	return nil
}

// Stop stops PulseManager. Waits replication goroutine is done.
func (m *PulseManager) Stop(ctx context.Context) error {
	// There should not to be any Set call after Stop call
	m.setLock.Lock()
	m.stopped = true
	m.setLock.Unlock()
	close(m.stop)

	if m.options.enableSync {
		close(m.gotpulse)
		inslogger.FromContext(ctx).Info("waiting finish of replication to heavy node...")
		<-m.syncdone
	}
	return nil
}

func (m *PulseManager) syncloop(ctx context.Context, pulses []core.PulseNumber) {
	defer close(m.syncdone)

	var err error
	inslog := inslogger.FromContext(ctx)
	var retrydelay time.Duration
	attempt := 0
	// shift synced pulse
	finishpulse := func() {
		pulses = pulses[1:]
		// reset retry variables
		// TODO: use jitter value for zero 'retrydelay'
		retrydelay = 0
		attempt = 0
	}

	for {
		select {
		case <-time.After(retrydelay):
		case <-m.stop:
			if len(pulses) == 0 {
				// fmt.Println("Got stop signal and have nothing to do")
				return
			}
		}
		for {
			if len(pulses) != 0 {
				// TODO: drop too outdated pulses
				// if (current - start > N) { start = current - N }
				break
			}
			inslog.Info("syncronization waiting next chunk of work")
			_, ok := <-m.gotpulse
			if !ok {
				inslog.Debug("stop is called, so we are should just stop syncronization loop")
				return
			}
			inslog.Infof("syncronization got next chunk of work")
			// get latest RP
			pulses, err = m.NextSyncPulses(ctx)
			if err != nil {
				err = errors.Wrap(err,
					"PulseManager syncloop failed on NextSyncPulseNumber call")
				inslog.Error(err)
				panic(err)
			}
		}

		tosyncPN := pulses[0]
		if m.pulseIsOutdated(ctx, tosyncPN) {
			finishpulse()
			continue
		}
		inslog.Infof("start syncronization to heavy for pulse %v", tosyncPN)

		sholdretry := false
		syncerr := m.HeavySync(ctx, tosyncPN, attempt > 0)
		if syncerr != nil {

			if heavyerr, ok := syncerr.(HeavyErr); ok {
				sholdretry = heavyerr.IsRetryable()
			}

			syncerr = errors.Wrap(syncerr, "HeavySync failed")
			inslog.Errorf("%v (on attempt=%v, sholdretry=%v)", syncerr.Error(), attempt, sholdretry)

			if sholdretry {
				retrydelay = m.syncbackoff.ForAttempt(attempt)
				attempt++
				continue
			}
			// TODO: write some info in dust?
		}

		err = m.db.SetReplicatedPulse(ctx, tosyncPN)
		if err != nil {
			err = errors.Wrap(err, "SetReplicatedPulse failed")
			inslog.Error(err)
			panic(err)
		}

		finishpulse()
	}
}

func (m *PulseManager) pulseIsOutdated(ctx context.Context, pn core.PulseNumber) bool {
	current, err := m.Current(ctx)
	if err != nil {
		panic(err)
	}
	return current.PulseNumber-pn > m.options.pulsesDeltaLimit
}

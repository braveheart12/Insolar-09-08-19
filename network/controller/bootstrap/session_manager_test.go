//
// Modified BSD 3-Clause Clear License
//
// Copyright (c) 2019 Insolar Technologies GmbH
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted (subject to the limitations in the disclaimer below) provided that
// the following conditions are met:
//  * Redistributions of source code must retain the above copyright notice, this list
//    of conditions and the following disclaimer.
//  * Redistributions in binary form must reproduce the above copyright notice, this list
//    of conditions and the following disclaimer in the documentation and/or other materials
//    provided with the distribution.
//  * Neither the name of Insolar Technologies GmbH nor the names of its contributors
//    may be used to endorse or promote products derived from this software without
//    specific prior written permission.
//
// NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED
// BY THIS LICENSE. THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS
// AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
// INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS
// OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Notwithstanding any other provisions of this license, it is prohibited to:
//    (a) use this software,
//
//    (b) prepare modifications and derivative works of this software,
//
//    (c) distribute this software (including without limitation in source code, binary or
//        object code form), and
//
//    (d) reproduce copies of this software
//
//    for any commercial purposes, and/or
//
//    for the purposes of making available this software to third parties as a service,
//    including, without limitation, any software-as-a-service, platform-as-a-service,
//    infrastructure-as-a-service or other similar online service, irrespective of
//    whether it competes with the products or services of Insolar Technologies GmbH.
//

package bootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sessionMapLen(sm SessionManager) int {
	s := sm.(*sessionManager)

	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.sessions)
}

func sessionMapDelete(sm SessionManager, id SessionID) {
	s := sm.(*sessionManager)

	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.sessions, id)
}

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager().(*sessionManager)

	assert.Equal(t, sm.state, stateIdle)
}

func TestSessionManager_CleanupSimple(t *testing.T) {
	sm := NewSessionManager()

	err := sm.Start(context.Background())
	require.NoError(t, err)

	sm.NewSession(insolar.NewEmptyReference(), nil, time.Second)
	require.Equal(t, sessionMapLen(sm), 1)

	time.Sleep(1500 * time.Millisecond)
	assert.Equal(t, sessionMapLen(sm), 0)

	err = sm.Stop(context.Background())
	require.NoError(t, err)
}

func TestSessionManager_CleanupConcurrent(t *testing.T) {
	sm := NewSessionManager()

	err := sm.Start(context.Background())
	require.NoError(t, err)

	id := sm.NewSession(insolar.NewEmptyReference(), nil, time.Second)
	require.Equal(t, sessionMapLen(sm), 1)

	// delete session here and check nothing happened
	sessionMapDelete(sm, id)

	time.Sleep(1500 * time.Millisecond)
	assert.Equal(t, sessionMapLen(sm), 0)

	err = sm.Stop(context.Background())
	require.NoError(t, err)
}

func TestSessionManager_CleanupOrder(t *testing.T) {
	sm := NewSessionManager()

	err := sm.Start(context.Background())
	require.NoError(t, err)

	sm.NewSession(insolar.NewEmptyReference(), nil, 2*time.Second)
	sm.NewSession(insolar.NewEmptyReference(), nil, 2*time.Second)
	sm.NewSession(insolar.NewEmptyReference(), nil, time.Second)
	require.Equal(t, sessionMapLen(sm), 3)

	time.Sleep(1500 * time.Millisecond)
	assert.Equal(t, sessionMapLen(sm), 2)

	err = sm.Stop(context.Background())
	require.NoError(t, err)
}

func TestSessionManager_ImmediatelyStop(t *testing.T) {
	sm := NewSessionManager()

	err := sm.Start(context.Background())
	require.NoError(t, err)

	err = sm.Stop(context.Background())
	require.NoError(t, err)
}

func TestSessionManager_DoubleStart(t *testing.T) {
	sm := NewSessionManager()

	err := sm.Start(context.Background())
	require.NoError(t, err)

	err = sm.Start(context.Background())
	require.NoError(t, err)
}

func TestSessionManager_DoubleStop(t *testing.T) {
	sm := NewSessionManager()

	err := sm.Start(context.Background())
	require.NoError(t, err)

	err = sm.Stop(context.Background())
	require.NoError(t, err)

	err = sm.Stop(context.Background())
	require.NoError(t, err)
}

func TestSessionManager_ProlongateSession(t *testing.T) {
	sm := NewSessionManager()

	err := sm.Start(context.Background())
	require.NoError(t, err)

	ref := testutils.RandomRef()
	id := sm.NewSession(ref, nil, 2*time.Second)

	session, err := sm.ReleaseSession(id)
	require.NoError(t, err)
	assert.Equal(t, ref, session.NodeID)
	_, err = sm.ReleaseSession(id)
	assert.Error(t, err)

	sm.ProlongateSession(id, session)

	session, err = sm.ReleaseSession(id)
	require.NoError(t, err)
	assert.Equal(t, ref, session.NodeID)
}

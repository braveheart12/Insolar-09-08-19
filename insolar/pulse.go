//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package insolar

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/insolar/insolar/insolar/utils"
	"github.com/insolar/insolar/network/consensus/common/pulse"
	"github.com/pkg/errors"
)

const (
	// PulseNumberSize declares the number of bytes in the pulse number
	PulseNumberSize = 4
	// EntropySize declares the number of bytes in the pulse entropy
	EntropySize = 64
	// OriginIDSize declares the number of bytes in the origin id
	OriginIDSize = 16
)

// Entropy is 64 random bytes used in every pseudo-random calculations.
type Entropy [EntropySize]byte

func (entropy Entropy) Marshal() ([]byte, error) { return entropy[:], nil }
func (entropy Entropy) MarshalTo(data []byte) (int, error) {
	copy(data, entropy[:])
	return EntropySize, nil
}
func (entropy *Entropy) Unmarshal(data []byte) error {
	if len(data) != EntropySize {
		return errors.New("Not enough bytes to unpack Entropy")
	}
	copy(entropy[:], data)
	return nil
}
func (entropy *Entropy) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, entropy)
}
func (entropy Entropy) Size() int { return EntropySize }
func (entropy Entropy) Compare(other Entropy) int {
	return bytes.Compare(entropy[:], other[:])
}
func (entropy Entropy) Equal(other Entropy) bool {
	return entropy.Compare(other) == 0
}

// PulseNumber is a sequential number of Pulse.
// Upper 2 bits are reserved for use in references (scope), must be zero otherwise.
// Valid Absolute PulseNumber must be >65536.
// If PulseNumber <65536 it is a relative PulseNumber
type PulseNumber pulse.Number

// NewPulseNumber creates pulse number from bytes.
func NewPulseNumber(buf []byte) PulseNumber {
	return PulseNumber(binary.BigEndian.Uint32(buf))
}

func NewPulseNumberFromStr(pn string) (PulseNumber, error) {
	i, err := strconv.ParseUint(pn, 10, 32)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse pulse number")
	}
	return PulseNumber(i), nil
}

// Bytes serializes pulse number.
func (pn PulseNumber) Bytes() []byte {
	return utils.UInt32ToBytes(uint32(pn))
}

func (pn PulseNumber) String() string {
	return fmt.Sprintf("%d", pn)
}

func (pn *PulseNumber) MarshalTo(data []byte) (int, error) {
	buf := pn.Bytes()
	if len(data) < len(buf) {
		return 0, errors.New("Not enough bytes to marshal PulseNumber")
	}
	copy(data, buf)
	return len(buf), nil
}

func (pn *PulseNumber) Unmarshal(data []byte) error {
	*pn = PulseNumber(binary.BigEndian.Uint32(data))
	return nil
}

func (pn PulseNumber) Equal(other PulseNumber) bool {
	return pn == other
}

func (pn PulseNumber) Size() int {
	return len(pn.Bytes())
}

//go:generate minimock -i github.com/insolar/insolar/insolar.PulseManager -o ../testutils -s _mock.go -g

// PulseManager provides Ledger's methods related to Pulse.
type PulseManager interface {
	// Set set's new pulse and closes current jet drop. If dry is true, nothing will be saved to storage.
	Set(ctx context.Context, pulse Pulse) error
}

// PulseRange represents range of pulses.
type PulseRange struct {
	Begin PulseNumber
	End   PulseNumber
}

func (pr *PulseRange) String() string {
	return fmt.Sprintf("[%v:%v]", pr.Begin, pr.End)
}

// Pulse is base data structure for a pulse.
type Pulse struct {
	PulseNumber     PulseNumber
	PrevPulseNumber PulseNumber
	NextPulseNumber PulseNumber

	PulseTimestamp   int64
	EpochPulseNumber int
	OriginID         [OriginIDSize]byte

	Entropy Entropy
	Signs   map[string]PulseSenderConfirmation
}

// PulseSenderConfirmation contains confirmations of the pulse from other pulsars
// Because the system is using BFT for consensus between pulsars, because of it
// All pulsar send to the chosen pulsar their confirmations
// Every node in the network can verify the signatures
type PulseSenderConfirmation struct {
	PulseNumber     PulseNumber
	ChosenPublicKey string
	Entropy         Entropy
	Signature       []byte
}

const (
	// FirstPulseNumber is the hardcoded first pulse number. Because first 65536 numbers are saved for the system's needs
	FirstPulseNumber = pulse.MinTimePulse
	// PulseNumberJet is a special pulse number value that signifies jet ID.
	PulseNumberJet = PulseNumber(1)
	// BuiltinContractPulseNumber declares special pulse number that creates namespace for builtin contracts
	BuiltinContractPulseNumber = PulseNumber(200)

	InvalidPulseEpoch   int = 0
	EphemeralPulseEpoch     = InvalidPulseEpoch + 1
)

// GenesisPulse is a first pulse for the system
// because first 2 bits of pulse number and first 65536 pulses a are used by system needs and pulse numbers are related to the seconds of Unix time
// for calculation pulse numbers we use the formula = unix.Now() - firstPulseDate + 65536
var GenesisPulse = &Pulse{
	PulseNumber:      FirstPulseNumber,
	Entropy:          [EntropySize]byte{},
	EpochPulseNumber: FirstPulseNumber,
	PulseTimestamp:   pulse.UnixTimeOfMinTimePulse,
}

// EphemeralPulse is used for discovery network bootstrap
var EphemeralPulse = &Pulse{
	PulseNumber:      FirstPulseNumber,
	Entropy:          [EntropySize]byte{},
	EpochPulseNumber: EphemeralPulseEpoch,
	PulseTimestamp:   pulse.UnixTimeOfMinTimePulse,
}

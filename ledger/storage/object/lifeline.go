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

package object

import (
	"bytes"
	"context"
	"sync"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/ledger/storage/db"
	"github.com/ugorji/go/codec"
	"go.opencensus.io/stats"
)

//go:generate minimock -i github.com/insolar/insolar/ledger/storage/object.IndexAccessor -o ./ -s _mock.go

// IndexAccessor provides info about Index-values from storage.
type IndexAccessor interface {
	// ForID returns Index for provided id.
	ForID(ctx context.Context, id insolar.ID) (Lifeline, error)
}

//go:generate minimock -i github.com/insolar/insolar/ledger/storage/object.IndexModifier -o ./ -s _mock.go

// IndexModifier provides methods for setting Index-values to storage.
type IndexModifier interface {
	// Set saves new Index-value in storage.
	Set(ctx context.Context, id insolar.ID, index Lifeline) error
}

// Lifeline represents meta information for record object.
type Lifeline struct {
	LatestState         *insolar.ID // Amend or activate record.
	LatestStateApproved *insolar.ID // State approved by VM.
	ChildPointer        *insolar.ID // Meta record about child activation.
	Parent              insolar.Reference
	Delegates           map[insolar.Reference]insolar.Reference
	State               State
	LatestUpdate        insolar.PulseNumber
	JetID               insolar.JetID
}

// Encode converts lifeline index into binary format.
func Encode(index Lifeline) []byte {
	buff := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buff, &codec.CborHandle{})
	enc.MustEncode(index)

	return buff.Bytes()
}

// Decode converts byte array into lifeline index struct.
func Decode(buff []byte) (index Lifeline) {
	dec := codec.NewDecoderBytes(buff, &codec.CborHandle{})
	dec.MustDecode(&index)

	return
}

// Clone returns copy of argument idx value.
func Clone(idx Lifeline) Lifeline {
	if idx.LatestState != nil {
		tmp := *idx.LatestState
		idx.LatestState = &tmp
	}

	if idx.LatestStateApproved != nil {
		tmp := *idx.LatestStateApproved
		idx.LatestStateApproved = &tmp
	}

	if idx.ChildPointer != nil {
		tmp := *idx.ChildPointer
		idx.ChildPointer = &tmp
	}

	if idx.Delegates != nil {
		cp := make(map[insolar.Reference]insolar.Reference)
		for k, v := range idx.Delegates {
			cp[k] = v
		}
		idx.Delegates = cp
	}

	return idx
}

// IndexMemory is an in-memory struct for index-storage.
type IndexMemory struct {
	jetIndex db.JetIndexModifier

	lock   sync.RWMutex
	memory map[insolar.ID]Lifeline
}

// NewIndexMemory creates a new instance of Storage.
func NewIndexMemory() *IndexMemory {
	return &IndexMemory{
		memory:   map[insolar.ID]Lifeline{},
		jetIndex: db.NewJetIndex(),
	}
}

// Set saves new Index-value in storage.
func (s *IndexMemory) Set(ctx context.Context, id insolar.ID, index Lifeline) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	idx := Clone(index)

	s.memory[id] = idx
	s.jetIndex.Add(id, idx.JetID)

	stats.Record(ctx,
		statIndexInMemoryCount.M(1),
	)

	return nil
}

// ForID returns Index for provided id.
func (s *IndexMemory) ForID(ctx context.Context, id insolar.ID) (index Lifeline, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	idx, ok := s.memory[id]
	if !ok {
		err = ErrNotFound
		return
	}

	index = Clone(idx)

	return
}
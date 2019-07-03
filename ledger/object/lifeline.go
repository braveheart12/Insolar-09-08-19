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
	"github.com/insolar/insolar/insolar"
)

// EncodeLifeline converts lifeline index into binary format.
func EncodeLifeline(index Lifeline) []byte {
	res, err := index.Marshal()
	if err != nil {
		panic(err)
	}

	return res
}

// MustDecodeLifeline converts byte array into lifeline index struct.
func MustDecodeLifeline(buff []byte) (index Lifeline) {
	idx, err := DecodeLifeline(buff)
	if err != nil {
		panic(err)
	}

	return idx
}

// DecodeLifeline converts byte array into lifeline index struct.
func DecodeLifeline(buff []byte) (Lifeline, error) {
	lfl := Lifeline{}
	err := lfl.Unmarshal(buff)
	return lfl, err
}

// CloneLifeline returns copy of argument idx value.
func CloneLifeline(idx Lifeline) Lifeline {
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
		cp := make([]LifelineDelegate, len(idx.Delegates))
		copy(cp, idx.Delegates)
		idx.Delegates = cp
	} else {
		idx.Delegates = []LifelineDelegate{}
	}

	if idx.EarliestOpenRequest != nil {
		tmp := *idx.EarliestOpenRequest
		idx.EarliestOpenRequest = &tmp
	}

	if idx.PendingPointer != nil {
		tmp := *idx.PendingPointer
		idx.PendingPointer = &tmp
	}

	return idx
}

func (m *Lifeline) SetDelegate(key insolar.Reference, value insolar.Reference) {
	for _, d := range m.Delegates {
		if d.Key == key {
			d.Value = value
			return
		}
	}

	m.Delegates = append(m.Delegates, LifelineDelegate{Key: key, Value: value})
}

func (m *Lifeline) DelegateByKey(key insolar.Reference) (insolar.Reference, bool) {
	for _, d := range m.Delegates {
		if d.Key == key {
			return d.Value, true
		}
	}

	return [64]byte{}, false
}

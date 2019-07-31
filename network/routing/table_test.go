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

package routing

import (
	"encoding/binary"
	"strconv"
	"testing"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/network/hostnetwork/host"
	"github.com/insolar/insolar/network/node"
	"github.com/insolar/insolar/network/nodenetwork"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newNode(id int) insolar.NetworkNode {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(id))
	nodeId := insolar.NewIDFromBytes(bs)
	ref := insolar.NewReference(*nodeId)
	address := "127.0.0.1:" + strconv.Itoa(id)
	result := node.NewNode(*ref, insolar.StaticRoleUnknown, nil, address, "")
	result.(node.MutableNode).SetShortID(insolar.ShortNodeID(id))
	return result
}

func newTable() *Table {
	return &Table{NodeKeeper: nodenetwork.NewNodeKeeper(newNode(1))}
}

func TestTable_Resolve(t *testing.T) {
	table := newTable()
	table.NodeKeeper.SetInitialSnapshot([]insolar.NetworkNode{
		newNode(2),
	})
	host, err := table.Resolve(insolar.Reference{2})
	require.NoError(t, err)
	assert.EqualValues(t, 2, host.ShortID)
	assert.Equal(t, "127.0.0.1:2", host.Address.String())

	_, err = table.Resolve(insolar.Reference{4})
	assert.Error(t, err)
}

func TestTable_AddToKnownHosts(t *testing.T) {
	table := newTable()
	h, err := host.NewHostN("127.0.0.1:234", testutils.RandomRef())
	require.NoError(t, err)
	table.AddToKnownHosts(h)
}

func TestTable_ResolveConsensus_equal(t *testing.T) {
	table := newTable()
	table.NodeKeeper.SetInitialSnapshot([]insolar.NetworkNode{
		newNode(2),
	})
	h, err := table.ResolveConsensusRef(insolar.Reference{2})
	require.NoError(t, err)
	h2, err := table.Resolve(insolar.Reference{2})
	require.NoError(t, err)
	assert.True(t, h.Equal(*h2))
}

func TestTable_ResolveConsensus_equal2(t *testing.T) {
	table := newTable()
	table.NodeKeeper.SetInitialSnapshot([]insolar.NetworkNode{
		newNode(2),
	})
	h, err := table.ResolveConsensusRef(insolar.Reference{2})
	require.NoError(t, err)
	h2, err := table.ResolveConsensus(2)
	require.NoError(t, err)
	assert.True(t, h.Equal(*h2))
}

func TestTable_ResolveConsensus(t *testing.T) {
	table := newTable()
	table.NodeKeeper.SetInitialSnapshot([]insolar.NetworkNode{
		newNode(2),
	})
	table.NodeKeeper.GetConsensusInfo().AddTemporaryMapping(insolar.Reference{3}, 3, "127.0.0.1:3")
	h, err := table.ResolveConsensusRef(insolar.Reference{2})
	require.NoError(t, err)
	h2, err := table.ResolveConsensus(2)
	require.NoError(t, err)
	assert.True(t, h.Equal(*h2))
	assert.EqualValues(t, 2, h.ShortID)
	assert.Equal(t, "127.0.0.1:2", h.Address.String())

	h, err = table.ResolveConsensusRef(insolar.Reference{3})
	require.NoError(t, err)
	h2, err = table.ResolveConsensus(3)
	require.NoError(t, err)
	assert.True(t, h.Equal(*h2))
	assert.EqualValues(t, 3, h.ShortID)
	assert.Equal(t, "127.0.0.1:3", h.Address.String())

	_, err = table.ResolveConsensusRef(insolar.Reference{4})
	assert.Error(t, err)
	_, err = table.ResolveConsensus(4)
	assert.Error(t, err)
}

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

package nodenetwork

import (
	"crypto"
	"crypto/rand"
	"testing"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/cryptography"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/network"
	"github.com/insolar/insolar/platformpolicy"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNodeNetwork(t *testing.T) {
	cfg := configuration.Transport{Address: "invalid"}
	certMock := testutils.NewCertificateMock(t)
	certMock.GetRoleMock.Set(func() insolar.StaticRole { return insolar.StaticRoleUnknown })
	certMock.GetPublicKeyMock.Set(func() crypto.PublicKey { return nil })
	certMock.GetNodeRefMock.Set(func() *insolar.Reference { return &insolar.Reference{0} })
	certMock.GetDiscoveryNodesMock.Set(func() []insolar.DiscoveryNode { return nil })
	_, err := NewNodeNetwork(cfg, certMock)
	assert.Error(t, err)
	cfg.Address = "127.0.0.1:3355"
	_, err = NewNodeNetwork(cfg, certMock)
	assert.NoError(t, err)
}

func newNodeKeeper(t *testing.T, service insolar.CryptographyService) network.NodeKeeper {
	cfg := configuration.Transport{Address: "127.0.0.1:3355"}
	certMock := testutils.NewCertificateMock(t)
	keyProcessor := platformpolicy.NewKeyProcessor()
	secret, err := keyProcessor.GeneratePrivateKey()
	require.NoError(t, err)
	pk := keyProcessor.ExtractPublicKey(secret)
	if service == nil {
		service = cryptography.NewKeyBoundCryptographyService(secret)
	}
	require.NoError(t, err)
	certMock.GetRoleMock.Set(func() insolar.StaticRole { return insolar.StaticRoleUnknown })
	certMock.GetPublicKeyMock.Set(func() crypto.PublicKey { return pk })
	certMock.GetNodeRefMock.Set(func() *insolar.Reference { return &insolar.Reference{137} })
	certMock.GetDiscoveryNodesMock.Set(func() []insolar.DiscoveryNode { return nil })
	nw, err := NewNodeNetwork(cfg, certMock)
	require.NoError(t, err)
	return nw.(network.NodeKeeper)
}

func TestNewNodeKeeper(t *testing.T) {
	nk := newNodeKeeper(t, nil)
	origin := nk.GetOrigin()
	assert.NotNil(t, origin)
	nk.SetInitialSnapshot([]insolar.NetworkNode{origin})
	assert.NotNil(t, nk.GetAccessor(insolar.GenesisPulse.PulseNumber))
}

func TestNodekeeper_GetCloudHash(t *testing.T) {
	nk := newNodeKeeper(t, nil)
	cloudHash := make([]byte, 64)
	_, _ = rand.Read(cloudHash)
	nk.SetCloudHash(0, cloudHash)
	assert.Equal(t, cloudHash, nk.GetCloudHash(0))
}

// func TestNodekeeper_GetWorkingNodes(t *testing.T) {
//	nk := newNodeKeeper(t, nil)
//	assert.Empty(t, nk.GetAccessor().GetActiveNodes())
//	assert.Empty(t, nk.GetWorkingNodes())
//	origin, node1, node2, node3, node4 :=
//		newTestNodeWithRole(insolar.Reference{137}, insolar.NodeReady, insolar.StaticRoleUnknown),
//		newTestNode(insolar.Reference{1}, insolar.NodePending),
//		newTestNodeWithRole(insolar.Reference{2}, insolar.NodeReady, insolar.StaticRoleLightMaterial),
//		newTestNodeWithRole(insolar.Reference{3}, insolar.NodeReady, insolar.StaticRoleVirtual),
//		newTestNode(insolar.Reference{4}, insolar.NodeLeaving)
//	nk.SetInitialSnapshot([]insolar.NetworkNode{origin, node1, node2, node3, node4})
//	assert.Equal(t, 5, len(nk.GetAccessor().GetActiveNodes()))
//	assert.Equal(t, 3, len(nk.GetWorkingNodes()))
//	assert.NotNil(t, nk.GetWorkingNode(node2.ID()))
//	assert.Nil(t, nk.GetWorkingNode(node1.ID()))
//
//	assert.Nil(t, nk.GetWorkingNode(node4.ID()))
//	assert.NotNil(t, nk.GetAccessor().GetActiveNode(node4.ID()))
//
//	nodes := []insolar.NetworkNode{origin, node1, node2, node3}
//	claims := []packets.ReferendumClaim{newTestJoinClaim(insolar.Reference{5})}
//	err := nk.Sync(context.Background(), nodes, claims)
//	assert.NoError(t, err)
//	err = nk.MoveSyncToActive(context.Background(), 0)
//	assert.NoError(t, err)
//
//	assert.Nil(t, nk.GetAccessor().GetActiveNode(node4.ID()))
//	assert.Equal(t, insolar.NodeReady, nk.GetAccessor().GetActiveNode(node1.ID()).GetState())
//	node5 := nk.GetAccessor().GetActiveNode(insolar.Reference{5})
//	assert.NotNil(t, node5)
//	assert.Nil(t, nk.GetWorkingNode(node5.ID()))
//
//	nodes = []insolar.NetworkNode{nk.GetOrigin(), node1, node2, node3, node5}
//	err = nk.Sync(context.Background(), nodes, nil)
//	assert.NoError(t, err)
//	err = nk.MoveSyncToActive(context.Background(), 0)
//	assert.NoError(t, err)
//
//	assert.Equal(t, insolar.NodeReady, nk.GetAccessor().GetActiveNode(node5.ID()).GetState())
//
//	nodes = []insolar.NetworkNode{node1, node2, node3, node5}
//	err = nk.Sync(context.Background(), nodes, nil)
//	assert.Error(t, err)
// }

// func TestNodekeeper_GracefulStop(t *testing.T) {
//	nk := newNodeKeeper(t, nil)
//	nodeLeaveTriggered := false
//	handler := testutils.NewTerminationHandlerMock(t)
//	handler.OnLeaveApprovedFunc = func(context.Context) {
//		nodeLeaveTriggered = true
//	}
//	nk.(*nodekeeper).TerminationHandler = handler
//	nodes := []insolar.NetworkNode{
//		nk.GetOrigin(),
//		newTestNode(insolar.Reference{1}, insolar.NodeReady),
//		newTestNode(insolar.Reference{2}, insolar.NodeReady),
//	}
//	nk.SetInitialSnapshot(nodes)
//
//	claims := []packets.ReferendumClaim{&packets.NodeLeaveClaim{NodeID: nk.GetOrigin().ID()}}
//	err := nk.Sync(context.Background(), nodes, claims)
//	assert.NoError(t, err)
//	err = nk.MoveSyncToActive(context.Background(), 0)
//	assert.NoError(t, err)
//
//	assert.True(t, nodeLeaveTriggered)
// }

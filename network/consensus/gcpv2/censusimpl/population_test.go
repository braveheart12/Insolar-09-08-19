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

package censusimpl

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/insolar/insolar/network/consensus/gcpv2/api/member"

	"github.com/insolar/insolar/insolar"

	"github.com/insolar/insolar/network/consensus/common/cryptkit"

	"github.com/insolar/insolar/network/consensus/gcpv2/api/profiles"
)

func TestNewManyNodePopulation(t *testing.T) {
	svf := cryptkit.NewSignatureVerifierFactoryMock(t)
	sv := cryptkit.NewSignatureVerifierMock(t)
	svf.GetSignatureVerifierWithPKSMock.Set(func(cryptkit.PublicKeyStore) cryptkit.SignatureVerifier { return sv })
	require.Panics(t, func() { NewManyNodePopulation(nil, 0, nil) })

	sp := profiles.NewStaticProfileMock(t)
	pks := cryptkit.NewPublicKeyStoreMock(t)
	sp.GetPublicKeyStoreMock.Set(func() cryptkit.PublicKeyStore { return pks })
	nodeID := insolar.ShortNodeID(2)
	sp.GetStaticNodeIDMock.Set(func() insolar.ShortNodeID { return nodeID })
	sp.GetPrimaryRoleMock.Set(func() member.PrimaryRole { return member.PrimaryRoleNeutral })
	require.Panics(t, func() { NewManyNodePopulation([]profiles.StaticProfile{sp}, 0, nil) })

	require.Panics(t, func() { NewManyNodePopulation([]profiles.StaticProfile{sp}, nodeID+1, svf) })

	mnp := NewManyNodePopulation([]profiles.StaticProfile{sp}, nodeID, svf)
	require.NotNil(t, mnp.local)
}

func TestMNPGetSuspendedCount(t *testing.T) {
	suspendedCount := uint16(1)
	mnp := ManyNodePopulation{suspendedCount: suspendedCount}
	require.Equal(t, int(suspendedCount), mnp.GetSuspendedCount())
}

func TestMNPGetMistrustedCount(t *testing.T) {
	mistrustedCount := uint16(1)
	mnp := ManyNodePopulation{mistrustedCount: mistrustedCount}
	require.Equal(t, int(mistrustedCount), mnp.GetMistrustedCount())
}

func TestMNPGetIdleProfiles(t *testing.T) {
	mnp := ManyNodePopulation{}
	require.Nil(t, mnp.GetIdleProfiles())

	role := roleRecord{}
	roleCount := uint16(1)
	mnp.roles = make([]roleRecord, roleCount)
	mnp.roles[member.PrimaryRoleInactive] = role
	require.Panics(t, func() { mnp.GetIdleProfiles() })

	mnp.roles[member.PrimaryRoleInactive].container = &ManyNodePopulation{slots: make([]updatableSlot, roleCount)}
	require.Nil(t, mnp.GetIdleProfiles())

	mnp.roles[member.PrimaryRoleInactive].roleCount = roleCount
	require.Len(t, mnp.GetIdleProfiles(), int(roleCount))
}

func TestMNPGetIdleCount(t *testing.T) {
	mnp := ManyNodePopulation{}
	require.Zero(t, mnp.GetIdleCount())

	roleCount := uint16(1)
	role := roleRecord{roleCount: roleCount}
	mnp.roles = make([]roleRecord, roleCount)
	mnp.roles[member.PrimaryRoleInactive] = role
	require.Equal(t, int(roleCount), mnp.GetIdleCount())
}

func TestMNPGetIndexedCount(t *testing.T) {
	assignedSlotCount := uint16(1)
	mnp := ManyNodePopulation{assignedSlotCount: assignedSlotCount}
	require.Equal(t, int(assignedSlotCount), mnp.GetIndexedCount())
}

func TestMNPGetIndexedCapacity(t *testing.T) {
	len := 1
	mnp := ManyNodePopulation{slots: make([]updatableSlot, len)}
	require.Equal(t, len, mnp.GetIndexedCapacity())
}

func TestMNPIsValid(t *testing.T) {
	mnp := ManyNodePopulation{isInvalid: true}
	require.False(t, mnp.IsValid())

	mnp.isInvalid = false
	require.True(t, mnp.IsValid())
}

func TestMNPGetRolePopulation(t *testing.T) {
	mnp := ManyNodePopulation{}
	rolesCount := 2
	mnp.workingRoles = make([]member.PrimaryRole, rolesCount)
	require.Nil(t, mnp.GetRolePopulation(member.PrimaryRoleInactive))

	role := member.PrimaryRoleNeutral
	mnp.workingRoles = nil
	require.Nil(t, mnp.GetRolePopulation(role))

	mnp.workingRoles = make([]member.PrimaryRole, rolesCount)
	mnp.roles = make([]roleRecord, rolesCount)
	require.Nil(t, mnp.GetRolePopulation(role))

	mnp.roles[role].container = &ManyNodePopulation{}
	require.NotNil(t, mnp.GetRolePopulation(role))

	mnp.roles[role].container = nil
	mnp.roles[role].idleCount = 1
	require.NotNil(t, mnp.GetRolePopulation(role))
}

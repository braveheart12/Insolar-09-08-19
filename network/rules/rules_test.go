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

package rules

import (
	"testing"

	"github.com/insolar/insolar/network/node"

	"github.com/insolar/insolar/certificate"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRules_CheckMinRole(t *testing.T) {
	cert := testutils.NewCertificateMock(t)
	nodes := []insolar.NetworkNode{
		node.NewNode(testutils.RandomRef(), insolar.StaticRoleHeavyMaterial, nil, "", ""),
		node.NewNode(testutils.RandomRef(), insolar.StaticRoleLightMaterial, nil, "", ""),
		node.NewNode(testutils.RandomRef(), insolar.StaticRoleLightMaterial, nil, "", ""),
		node.NewNode(testutils.RandomRef(), insolar.StaticRoleVirtual, nil, "", ""),
		node.NewNode(testutils.RandomRef(), insolar.StaticRoleVirtual, nil, "", ""),
	}

	cert.GetMinRolesMock.Set(func() (r uint, r1 uint, r2 uint) {
		return 1, 0, 0
	})

	result := CheckMinRole(cert, nodes)
	assert.True(t, result)
}

func TestRules_CheckMajorityRule(t *testing.T) {
	cert := testutils.NewCertificateMock(t)
	cert.GetDiscoveryNodesMock.Set(func() (r []insolar.DiscoveryNode) {
		_, nodes := getDiscoveryNodes(5)
		return nodes
	})

	cert.GetMajorityRuleMock.Set(func() (r int) {
		return 4
	})

	nodes, _ := getDiscoveryNodes(5)
	nodes = append(nodes, newNode(250))

	result, count := CheckMajorityRule(cert, nodes)
	assert.True(t, result)
	assert.Equal(t, 5, count)
}

func getDiscoveryNodes(count int) ([]insolar.NetworkNode, []insolar.DiscoveryNode) {
	result1 := make([]insolar.NetworkNode, count)
	result2 := make([]insolar.DiscoveryNode, count)

	for i := 0; i < count; i++ {
		n := newNode(i)
		d := certificate.NewBootstrapNode(nil, "", "127.0.0.1:3000", n.ID().String(), n.Role().String())
		result1[i] = n
		result2[i] = d
	}

	return result1, result2
}

func newNode(id int) insolar.NetworkNode {
	recordRef := insolar.Reference{byte(id)}
	n := node.NewNode(recordRef, insolar.StaticRoleVirtual, nil, "127.0.0.1:3000", "")
	return n
}

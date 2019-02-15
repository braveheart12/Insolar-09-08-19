/*
 * The Clear BSD License
 *
 * Copyright (c) 2019 Insolar Technologies
 *
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without modification, are permitted (subject to the limitations in the disclaimer below) provided that the following conditions are met:
 *
 *  Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
 *  Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
 *  Neither the name of Insolar Technologies nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.
 *
 * NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED BY THIS LICENSE. THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package claimhandler

import (
	"github.com/insolar/insolar/consensus/packets"
	"github.com/insolar/insolar/core"
)

type joinClaimHandler struct {
	next        ClaimHandler
	queue       Queue
	ref         core.RecordRef
	activeCount int
}

func NewJoinClaimHandler(activeNodesCount int, claims []*packets.NodeJoinClaim, pulse *core.Pulse, next ClaimHandler) ClaimHandler {
	handler := &joinClaimHandler{activeCount: activeNodesCount, next: next}
	for _, claim := range claims {
		handler.queue.PushClaim(claim, getPriority(claim.NodeRef, pulse.Entropy))
	}
	return handler
}

func (jch *joinClaimHandler) HandleClaim(claim packets.ReferendumClaim) packets.ReferendumClaim {
	_, ok := claim.(*packets.NodeJoinClaim)
	if !ok {
		if jch.next == nil {
			return claim
		}
		jch.next.HandleClaim(claim)
	}
	return jch.handle(claim)
}

func (jch *joinClaimHandler) handle(claim packets.ReferendumClaim) packets.ReferendumClaim {
	return jch.queue.PopClaim()
}

func getPriority(ref core.RecordRef, entropy core.Entropy) []byte {
	res := make([]byte, len(ref))
	for i := 0; i < len(ref); i++ {
		res[i] = ref[i] ^ entropy[i]
	}
	return res
}

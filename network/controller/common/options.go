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

package common

import (
	"time"

	"github.com/insolar/insolar/configuration"
)

// Options contains configuration options for the local host.
type Options struct {
	// The maximum time to wait for a response to ping request.
	PingTimeout time.Duration

	// The maximum time to wait for a response to any packet.
	PacketTimeout time.Duration

	// The maximum time to wait for a response to ack packet.
	AckPacketTimeout time.Duration

	// Bootstrap reconnect timeout
	BootstrapTimeout time.Duration

	// Min bootstrap retry timeout
	MinTimeout time.Duration

	// Max bootstrap retry timeout
	MaxTimeout time.Duration

	// Multiplier for boostrap retry time
	TimeoutMult time.Duration

	// HandshakeSession TTL
	HandshakeSessionTTL time.Duration
}

// ConfigureOptions convert daemon configuration to controller options
func ConfigureOptions(conf configuration.Configuration) *Options {
	config := conf.Host
	return &Options{
		TimeoutMult:         time.Duration(config.TimeoutMult) * time.Second,
		MinTimeout:          time.Duration(config.MinTimeout) * time.Second,
		MaxTimeout:          time.Duration(config.MaxTimeout) * time.Second,
		PingTimeout:         1 * time.Second,
		PacketTimeout:       15 * time.Second,
		AckPacketTimeout:    5 * time.Second,
		BootstrapTimeout:    10 * time.Second,
		HandshakeSessionTTL: time.Duration(config.HandshakeSessionTTL) * time.Millisecond,
	}
}

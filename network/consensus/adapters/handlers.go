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

package adapters

import (
	"bytes"
	"context"
	"io"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/network"
	"github.com/insolar/insolar/network/consensus/common/endpoints"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/transport"
)

type PacketProcessor interface {
	ProcessPacket(ctx context.Context, payload transport.PacketParser, from endpoints.Inbound) error
}

type PacketParserFactory interface {
	ParsePacket(ctx context.Context, reader io.Reader) (transport.PacketParser, error)
}

type DatagramHandler struct {
	packetProcessor     PacketProcessor
	packetParserFactory PacketParserFactory
}

func NewDatagramHandler() *DatagramHandler {
	return &DatagramHandler{}
}

func (dh *DatagramHandler) SetPacketProcessor(packetProcessor PacketProcessor) {
	dh.packetProcessor = packetProcessor
}

func (dh *DatagramHandler) SetPacketParserFactory(packetParserFactory PacketParserFactory) {
	dh.packetParserFactory = packetParserFactory
}

func (dh *DatagramHandler) HandleDatagram(ctx context.Context, address string, buf []byte) {
	ctx, logger := inslogger.WithFields(ctx, map[string]interface{}{
		"sender_address": address,
	})

	if dh.packetProcessor == nil || dh.packetParserFactory == nil {
		logger.Error("Datagram handler not initialized")
		return
	}

	packetParser, err := dh.packetParserFactory.ParsePacket(ctx, bytes.NewReader(buf))
	if err != nil {
		logger.Error("Failed to get PacketParser: ", err)
		return
	}

	ctx, logger = inslogger.WithFields(ctx, map[string]interface{}{
		"sender_id":    packetParser.GetSourceID(),
		"packet_type":  packetParser.GetPacketType().String(),
		"packet_pulse": packetParser.GetPulseNumber(),
	})

	hostIdentity := endpoints.InboundConnection{
		Addr: endpoints.Name(address),
	}
	err = dh.packetProcessor.ProcessPacket(ctx, packetParser, &hostIdentity)
	if err != nil {
		logger.Error("Failed to process packet: ", err)
		return
	}
}

type PulseHandler struct {
	packetProcessor PacketProcessor
}

func NewPulseHandler() *PulseHandler {
	return &PulseHandler{}
}

func (ph *PulseHandler) SetPacketProcessor(packetProcessor PacketProcessor) {
	ph.packetProcessor = packetProcessor
}

func (ph *PulseHandler) SetPacketParserFactory(PacketParserFactory) {}

func (ph *PulseHandler) HandlePulse(ctx context.Context, pulse insolar.Pulse, packet network.ReceivedPacket) {
	logger := inslogger.FromContext(ctx)

	if ph.packetProcessor == nil {
		logger.Error("Pulse handler not initialized")
		return
	}

	pulseData := NewPulseData(pulse)
	pulsePayload := NewPulsePacketParser(pulseData, packet.Bytes())

	err := ph.packetProcessor.ProcessPacket(ctx, pulsePayload, &endpoints.InboundConnection{
		Addr: "pulsar",
	})
	if err != nil {
		logger.Error(err)
	}
}

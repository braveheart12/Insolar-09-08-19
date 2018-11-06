/*
 *    Copyright 2018 Insolar
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package transport

import (
	"bytes"
	"fmt"
	"net"

	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/network/transport/packet"
	"github.com/insolar/insolar/network/transport/relay"
	"github.com/pkg/errors"
)

const udpMaxPacketSize = 1400

// GetUDPMaxPacketSize returns udp max packet size
func GetUDPMaxPacketSize() int {
	return udpMaxPacketSize
}

type udpTransport struct {
	baseTransport
	serverConn net.PacketConn
}

func newUDPTransport(conn net.PacketConn, proxy relay.Proxy, publicAddress string) (*udpTransport, error) {
	transport := &udpTransport{
		baseTransport: newBaseTransport(proxy, publicAddress),
		serverConn:    conn}
	transport.sendFunc = transport.send

	return transport, nil
}

func (udpT *udpTransport) send(recvAddress string, data []byte) error {
	log.Debug("Sending PURE_UDP request")
	if len(data) > udpMaxPacketSize {
		return errors.New(fmt.Sprintf("udpTransport.send: too big input data. Maximum: %d. Current: %d",
			udpMaxPacketSize, len(data)))
	}

	// TODO: may be try to send second time if error
	// TODO: skip resolving every time by caching result
	udpAddr, err := net.ResolveUDPAddr("udp", recvAddress)
	if err != nil {
		return errors.Wrap(err, "udpTransport.send")
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return errors.Wrap(err, "udpTransport.send")
	}
	defer udpConn.Close()

	log.Debug("udpTransport.send: len = ", len(data))
	_, err = udpConn.Write(data)
	return errors.Wrap(err, "Failed to write data")
}

// Start starts networking.
func (udpT *udpTransport) Start() error {
	log.Info("Start UDP transport")
	for {
		buf := make([]byte, udpMaxPacketSize)
		n, addr, err := udpT.serverConn.ReadFrom(buf)
		if err != nil {
			<-udpT.disconnectFinished
			return err
		}

		go udpT.handleAcceptedConnection(buf[:n], addr)
	}
}

// Stop stops networking.
func (udpT *udpTransport) Stop() {
	udpT.mutex.Lock()
	defer udpT.mutex.Unlock()

	log.Info("Stop UDP transport")
	udpT.prepareDisconnect()

	err := udpT.serverConn.Close()
	if err != nil {
		log.Errorln("Failed to close socket:", err.Error())
	}
}

func (udpT *udpTransport) handleAcceptedConnection(data []byte, addr net.Addr) {
	r := bytes.NewReader(data)
	msg, err := packet.DeserializePacket(r)
	if err != nil {
		log.Error("[ handleAcceptedConnection ] ", err)
		return
	}
	log.Debug("[ handleAcceptedConnection ] Packet processed. size: ", len(data), ". Address: ", addr)

	udpT.handlePacket(msg)
}

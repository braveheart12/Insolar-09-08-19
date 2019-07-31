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

package servicenetwork

import (
	"context"
	"testing"

	"github.com/insolar/insolar/network/rules"

	"github.com/insolar/insolar/network/gateway"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/insolar/insolar/certificate"
	"github.com/insolar/insolar/component"
	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/pulse"
	"github.com/insolar/insolar/network/node"
	"github.com/insolar/insolar/network/nodenetwork"
	"github.com/insolar/insolar/testutils"
	networkUtils "github.com/insolar/insolar/testutils/network"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PublisherMock struct{}

func (p *PublisherMock) Publish(topic string, messages ...*message.Message) error {
	return nil
}

func (p *PublisherMock) Close() error {
	return nil
}

func prepareNetwork(t *testing.T, cfg configuration.Configuration) *ServiceNetwork {
	serviceNetwork, err := NewServiceNetwork(cfg, &component.Manager{})
	require.NoError(t, err)

	nodeKeeper := networkUtils.NewNodeKeeperMock(t)
	nodeMock := networkUtils.NewNetworkNodeMock(t)
	nodeMock.IDMock.Return(testutils.RandomRef())
	nodeKeeper.GetOriginMock.Return(nodeMock)
	serviceNetwork.NodeKeeper = nodeKeeper

	return serviceNetwork
}

func TestSendMessageHandler_ReceiverNotSet(t *testing.T) {
	cfg := configuration.NewConfiguration()

	serviceNetwork := prepareNetwork(t, cfg)

	p := []byte{1, 2, 3, 4, 5}
	meta := payload.Meta{
		Payload: p,
	}
	data, err := meta.Marshal()
	require.NoError(t, err)

	inMsg := message.NewMessage(watermill.NewUUID(), data)

	_, err = serviceNetwork.SendMessageHandler(inMsg)
	require.Error(t, err)
}

func TestSendMessageHandler_SameNode(t *testing.T) {
	cfg := configuration.NewConfiguration()
	cfg.Service.Skip = 5
	serviceNetwork, err := NewServiceNetwork(cfg, &component.Manager{})
	nodeRef := testutils.RandomRef()
	nodeN := networkUtils.NewNodeKeeperMock(t)
	nodeN.GetOriginMock.Set(func() (r insolar.NetworkNode) {
		n := networkUtils.NewNetworkNodeMock(t)
		n.IDMock.Set(func() (r insolar.Reference) {
			return nodeRef
		})
		return n
	})
	pubMock := &PublisherMock{}
	pulseMock := pulse.NewAccessorMock(t)
	pulseMock.LatestMock.Return(*insolar.GenesisPulse, nil)
	serviceNetwork.PulseAccessor = pulseMock
	serviceNetwork.NodeKeeper = nodeN
	serviceNetwork.Pub = pubMock

	p := []byte{1, 2, 3, 4, 5}
	meta := payload.Meta{
		Payload:  p,
		Receiver: nodeRef,
	}
	data, err := meta.Marshal()
	require.NoError(t, err)

	inMsg := message.NewMessage(watermill.NewUUID(), data)

	outMsgs, err := serviceNetwork.SendMessageHandler(inMsg)
	require.NoError(t, err)
	require.Nil(t, outMsgs)
}

func TestSendMessageHandler_SendError(t *testing.T) {
	cfg := configuration.NewConfiguration()
	cfg.Service.Skip = 5
	pubMock := &PublisherMock{}
	serviceNetwork, err := NewServiceNetwork(cfg, &component.Manager{})
	serviceNetwork.Pub = pubMock
	nodeN := networkUtils.NewNodeKeeperMock(t)
	nodeN.GetOriginMock.Set(func() (r insolar.NetworkNode) {
		n := networkUtils.NewNetworkNodeMock(t)
		n.IDMock.Set(func() (r insolar.Reference) {
			return testutils.RandomRef()
		})
		return n
	})
	rpc := networkUtils.NewRPCControllerMock(t)
	rpc.SendBytesMock.Set(func(p context.Context, p1 insolar.Reference, p2 string, p3 []byte) (r []byte, r1 error) {
		return nil, errors.New("test error")
	})
	pulseMock := pulse.NewAccessorMock(t)
	pulseMock.LatestMock.Return(*insolar.GenesisPulse, nil)
	serviceNetwork.PulseAccessor = pulseMock
	serviceNetwork.RPC = rpc
	serviceNetwork.NodeKeeper = nodeN

	p := []byte{1, 2, 3, 4, 5}
	meta := payload.Meta{
		Payload:  p,
		Receiver: testutils.RandomRef(),
	}
	data, err := meta.Marshal()
	require.NoError(t, err)

	inMsg := message.NewMessage(watermill.NewUUID(), data)

	_, err = serviceNetwork.SendMessageHandler(inMsg)
	require.Error(t, err)
}

func TestSendMessageHandler_WrongReply(t *testing.T) {
	cfg := configuration.NewConfiguration()
	cfg.Service.Skip = 5
	pubMock := &PublisherMock{}
	serviceNetwork, err := NewServiceNetwork(cfg, &component.Manager{})
	serviceNetwork.Pub = pubMock
	nodeN := networkUtils.NewNodeKeeperMock(t)
	nodeN.GetOriginMock.Set(func() (r insolar.NetworkNode) {
		n := networkUtils.NewNetworkNodeMock(t)
		n.IDMock.Set(func() (r insolar.Reference) {
			return testutils.RandomRef()
		})
		return n
	})
	rpc := networkUtils.NewRPCControllerMock(t)
	rpc.SendBytesMock.Set(func(p context.Context, p1 insolar.Reference, p2 string, p3 []byte) (r []byte, r1 error) {
		return nil, nil
	})
	pulseMock := pulse.NewAccessorMock(t)
	pulseMock.LatestMock.Return(*insolar.GenesisPulse, nil)
	serviceNetwork.PulseAccessor = pulseMock
	serviceNetwork.RPC = rpc
	serviceNetwork.NodeKeeper = nodeN

	p := []byte{1, 2, 3, 4, 5}
	meta := payload.Meta{
		Payload:  p,
		Receiver: testutils.RandomRef(),
	}
	data, err := meta.Marshal()
	require.NoError(t, err)

	inMsg := message.NewMessage(watermill.NewUUID(), data)

	_, err = serviceNetwork.SendMessageHandler(inMsg)
	require.Error(t, err)
}

func TestSendMessageHandler(t *testing.T) {
	cfg := configuration.NewConfiguration()
	cfg.Service.Skip = 5
	serviceNetwork, err := NewServiceNetwork(cfg, &component.Manager{})
	nodeN := networkUtils.NewNodeKeeperMock(t)
	nodeN.GetOriginMock.Set(func() (r insolar.NetworkNode) {
		n := networkUtils.NewNetworkNodeMock(t)
		n.IDMock.Set(func() (r insolar.Reference) {
			return testutils.RandomRef()
		})
		return n
	})
	rpc := networkUtils.NewRPCControllerMock(t)
	rpc.SendBytesMock.Set(func(p context.Context, p1 insolar.Reference, p2 string, p3 []byte) (r []byte, r1 error) {
		return ack, nil
	})
	pulseMock := pulse.NewAccessorMock(t)
	pulseMock.LatestMock.Return(*insolar.GenesisPulse, nil)
	serviceNetwork.PulseAccessor = pulseMock
	serviceNetwork.RPC = rpc
	serviceNetwork.NodeKeeper = nodeN

	p := []byte{1, 2, 3, 4, 5}
	meta := payload.Meta{
		Payload:  p,
		Receiver: testutils.RandomRef(),
	}
	data, err := meta.Marshal()
	require.NoError(t, err)

	inMsg := message.NewMessage(watermill.NewUUID(), data)

	outMsgs, err := serviceNetwork.SendMessageHandler(inMsg)
	require.NoError(t, err)
	require.Nil(t, outMsgs)
}

type stater struct{}

func (s *stater) State() []byte {
	return []byte("123")
}

func TestServiceNetwork_StartStop(t *testing.T) {
	cm := &component.Manager{}
	origin := testutils.RandomRef()
	nk := nodenetwork.NewNodeKeeper(node.NewNode(origin, insolar.StaticRoleUnknown, nil, "127.0.0.1:0", ""))
	cert := &certificate.Certificate{}
	cert.Reference = origin.String()
	certManager := certificate.NewCertificateManager(cert)
	serviceNetwork, err := NewServiceNetwork(configuration.NewConfiguration(), cm)
	require.NoError(t, err)
	ctx := context.Background()
	defer serviceNetwork.Stop(ctx)
	serviceNetwork.SetOperableFunc(func(ctx context.Context, operable bool) {

	})

	cm.Inject(serviceNetwork, nk, certManager, testutils.NewCryptographyServiceMock(t), pulse.NewAccessorMock(t),
		testutils.NewTerminationHandlerMock(t), testutils.NewPulseManagerMock(t), &PublisherMock{},
		testutils.NewMessageBusMock(t), testutils.NewContractRequesterMock(t), rules.NewRules(),
		bus.NewSenderMock(t), &stater{}, testutils.NewPlatformCryptographyScheme(), testutils.NewKeyProcessorMock(t))
	err = serviceNetwork.Init(ctx)
	require.NoError(t, err)
	err = serviceNetwork.Start(ctx)
	require.NoError(t, err)
}

type publisherMock struct {
	Error error
}

func (pm *publisherMock) Publish(topic string, messages ...*message.Message) error { return pm.Error }
func (pm *publisherMock) Close() error                                             { return nil }

func TestServiceNetwork_processIncoming(t *testing.T) {
	serviceNetwork, err := NewServiceNetwork(configuration.NewConfiguration(), &component.Manager{})
	require.NoError(t, err)
	pub := &publisherMock{}
	serviceNetwork.Pub = pub
	ctx := context.Background()
	_, err = serviceNetwork.processIncoming(ctx, []byte("ololo"))
	assert.Error(t, err)
	msg := message.NewMessage("1", nil)
	data, err := messageToBytes(msg)
	require.NoError(t, err)
	_, err = serviceNetwork.processIncoming(ctx, data)
	assert.NoError(t, err)
	pub.Error = errors.New("Failed to publish message")
	_, err = serviceNetwork.processIncoming(ctx, data)
	assert.Error(t, err)
}

func TestServiceNetwork_SetGateway(t *testing.T) {
	sn, err := NewServiceNetwork(configuration.NewConfiguration(), &component.Manager{})
	require.NoError(t, err)
	op := false
	tick := 0
	sn.SetOperableFunc(func(ctx context.Context, operable bool) {
		op = operable
		tick++
	})

	hn := networkUtils.NewHostNetworkMock(t)
	hn.RegisterRequestHandlerMock.Return()
	sn.HostNetwork = hn

	// initial set
	sn.SetGateway(gateway.NewNoNetwork(sn, sn.PulseManager, sn.NodeKeeper, sn.ContractRequester,
		sn.CryptographyService, sn.HostNetwork, sn.CertificateManager))
	assert.Equal(t, 1, tick)
	assert.False(t, op)

	type Test struct {
		state insolar.NetworkState
		lock  bool
	}

	for i, T := range []Test{
		{insolar.NoNetworkState, false},
		{insolar.CompleteNetworkState, true},
		{insolar.NoNetworkState, false},
	} {
		sn.SetGateway(sn.Gateway().NewGateway(T.state))
		assert.Equal(t, i+1, tick)
		assert.Equal(t, T.lock, op)

	}
}

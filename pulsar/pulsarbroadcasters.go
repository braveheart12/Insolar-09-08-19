package pulsar

import (
	"time"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/network/hostnetwork/host"
	"github.com/insolar/insolar/network/hostnetwork/id"
	"github.com/insolar/insolar/network/hostnetwork/packet"
	"github.com/insolar/insolar/network/hostnetwork/relay"
	transport2 "github.com/insolar/insolar/network/hostnetwork/transport"
)

func (currentPulsar *Pulsar) broadcastSignatureOfEntropy() {
	log.Debug("[broadcastSignatureOfEntropy]")
	if currentPulsar.IsStateFailed() {
		return
	}

	payload, err := currentPulsar.preparePayload(EntropySignaturePayload{PulseNumber: currentPulsar.ProcessingPulseNumber, Signature: currentPulsar.GeneratedEntropySign})
	if err != nil {
		currentPulsar.StateSwitcher.SwitchToState(Failed, err)
		return
	}

	for _, neighbour := range currentPulsar.Neighbours {
		broadcastCall := neighbour.OutgoingClient.Go(ReceiveSignatureForEntropy.String(),
			payload,
			nil,
			nil)
		reply := <-broadcastCall.Done
		if reply.Error != nil {
			log.Warnf("Response to %v finished with error - %v", neighbour.ConnectionAddress, reply.Error)
			continue
		}
		log.Infof("Sign of entropy sent to %v", neighbour.ConnectionAddress)
	}
}

func (currentPulsar *Pulsar) broadcastVector() {
	log.Debug("[broadcastVector]")
	if currentPulsar.IsStateFailed() {
		return
	}
	payload, err := currentPulsar.preparePayload(VectorPayload{
		PulseNumber: currentPulsar.ProcessingPulseNumber,
		Vector:      currentPulsar.OwnedBftRow})

	if err != nil {
		currentPulsar.StateSwitcher.SwitchToState(Failed, err)
		return
	}

	for _, neighbour := range currentPulsar.Neighbours {
		broadcastCall := neighbour.OutgoingClient.Go(ReceiveVector.String(),
			payload,
			nil,
			nil)
		reply := <-broadcastCall.Done
		if reply.Error != nil {
			log.Warnf("Response to %v finished with error - %v", neighbour.ConnectionAddress, reply.Error)
		}
	}
}

func (currentPulsar *Pulsar) broadcastEntropy() {
	log.Debug("[broadcastEntropy]")
	if currentPulsar.IsStateFailed() {
		return
	}

	payload, err := currentPulsar.preparePayload(EntropyPayload{PulseNumber: currentPulsar.ProcessingPulseNumber, Entropy: currentPulsar.GeneratedEntropy})
	if err != nil {
		currentPulsar.StateSwitcher.SwitchToState(Failed, err)
		return
	}

	for _, neighbour := range currentPulsar.Neighbours {
		broadcastCall := neighbour.OutgoingClient.Go(ReceiveEntropy.String(),
			payload,
			nil,
			nil)
		reply := <-broadcastCall.Done
		if reply.Error != nil {
			log.Warnf("Response to %v finished with error - %v", neighbour.ConnectionAddress, reply.Error)
		}
	}
}

func (currentPulsar *Pulsar) sendPulseToPulsars() {
	log.Debug("[sendPulseToPulsars]")
	if currentPulsar.IsStateFailed() {
		return
	}

	payload, err := currentPulsar.preparePayload(PulsePayload{Pulse: core.Pulse{
		PulseNumber: currentPulsar.ProcessingPulseNumber,
		Entropy:     currentPulsar.CurrentSlotEntropy,
		Signs:       currentPulsar.CurrentSlotSenderConfirmations,
	}})
	if err != nil {
		currentPulsar.StateSwitcher.SwitchToState(Failed, err)
		return
	}

	for _, neighbour := range currentPulsar.Neighbours {
		broadcastCall := neighbour.OutgoingClient.Go(ReceivePulse.String(),
			payload,
			nil,
			nil)
		reply := <-broadcastCall.Done
		if reply.Error != nil {
			log.Warnf("Response to %v finished with error - %v", neighbour.ConnectionAddress, reply.Error)
		}
	}
}

func (currentPulsar *Pulsar) sendVector() {
	log.Debug("[sendVector]")
	if currentPulsar.IsStateFailed() {
		return
	}

	if currentPulsar.isStandalone() {
		currentPulsar.StateSwitcher.SwitchToState(Verifying, nil)
		return
	}

	currentPulsar.broadcastVector()

	currentPulsar.SetBftGridItem(currentPulsar.PublicKeyRaw, currentPulsar.OwnedBftRow)
	currentPulsar.StateSwitcher.SwitchToState(WaitingForVectors, nil)
}

func (currentPulsar *Pulsar) sendEntropy() {
	log.Debug("[sendEntropy]")
	if currentPulsar.IsStateFailed() {
		return
	}

	if currentPulsar.isStandalone() {
		currentPulsar.StateSwitcher.SwitchToState(Verifying, nil)
		return
	}

	currentPulsar.broadcastEntropy()

	currentPulsar.StateSwitcher.SwitchToState(WaitingForEntropy, nil)
}

func (currentPulsar *Pulsar) sendPulseSign() {
	log.Debug("[sendPulseSign]")
	if currentPulsar.IsStateFailed() {
		return
	}

	signature, err := signData(currentPulsar.PrivateKey, currentPulsar.CurrentSlotPulseSender)
	if err != nil {
		currentPulsar.StateSwitcher.SwitchToState(Failed, err)
		return
	}
	confirmation := SenderConfirmationPayload{
		PulseNumber:     currentPulsar.ProcessingPulseNumber,
		ChosenPublicKey: currentPulsar.CurrentSlotPulseSender,
		Signature:       signature,
	}

	payload, err := currentPulsar.preparePayload(confirmation)
	if err != nil {
		currentPulsar.StateSwitcher.SwitchToState(Failed, err)
		return
	}

	call := currentPulsar.Neighbours[currentPulsar.CurrentSlotPulseSender].OutgoingClient.Go(ReceiveChosenSignature.String(), payload, nil, nil)
	reply := <-call.Done
	if reply.Error != nil {
		//Here should be retry
		log.Error(reply.Error)
		currentPulsar.StateSwitcher.SwitchToState(Failed, log.Error)
	}

	currentPulsar.StateSwitcher.SwitchToState(WaitingForStart, nil)
}

func (currentPulsar *Pulsar) sendPulseToNodesAndPulsars() {
	log.Debug("[sendPulseToNodesAndPulsars]. Pulse - %v", time.Now())

	if currentPulsar.IsStateFailed() {
		return
	}

	pulseForSending := core.Pulse{
		PulseNumber:     currentPulsar.ProcessingPulseNumber,
		Entropy:         currentPulsar.CurrentSlotEntropy,
		Signs:           currentPulsar.CurrentSlotSenderConfirmations,
		NextPulseNumber: currentPulsar.ProcessingPulseNumber + core.PulseNumber(currentPulsar.Config.NumberDelta),
	}

	pulsarHost, t, err := currentPulsar.prepareForSendingPulse()
	if err != nil {
		currentPulsar.StateSwitcher.SwitchToState(Failed, err)
		return
	}

	currentPulsar.sendPulseToNetwork(pulsarHost, t, pulseForSending)
	currentPulsar.sendPulseToPulsars()

	err = currentPulsar.Storage.SavePulse(&pulseForSending)
	if err != nil {
		log.Error(err)
	}
	err = currentPulsar.Storage.SetLastPulse(&pulseForSending)
	if err != nil {
		log.Error(err)
	}
	currentPulsar.LastPulse = &pulseForSending

	currentPulsar.StateSwitcher.SwitchToState(WaitingForStart, nil)
	defer func() {
		go t.Stop()
		<-t.Stopped()
		t.Close()
	}()
}

func (currentPulsar *Pulsar) prepareForSendingPulse() (pulsarHost *host.Host, t transport2.Transport, err error) {

	t, err = transport2.NewTransport(currentPulsar.Config.BootstrapListener, relay.NewProxy())
	if err != nil {
		return
	}

	go func() {
		err = t.Start()
		if err != nil {
			log.Error(err)
		}
	}()

	if err != nil {
		return
	}

	pulsarHostAddress, err := host.NewAddress(currentPulsar.Config.BootstrapListener.Address)
	if err != nil {
		return
	}
	pulsarHostID, err := id.NewID()
	if err != nil {
		return
	}
	pulsarHost = host.NewHost(pulsarHostAddress)
	pulsarHost.ID = pulsarHostID

	return
}

func (currentPulsar *Pulsar) sendPulseToNetwork(pulsarHost *host.Host, t transport2.Transport, pulse core.Pulse) {
	for _, bootstrapNode := range currentPulsar.Config.BootstrapNodes {
		receiverAddress, err := host.NewAddress(bootstrapNode)
		if err != nil {
			log.Error(err)
			continue
		}
		receiverHost := host.NewHost(receiverAddress)

		b := packet.NewBuilder()
		pingPacket := packet.NewPingPacket(pulsarHost, receiverHost)
		pingCall, err := t.SendRequest(pingPacket)
		if err != nil {
			log.Error(err)
			continue
		}
		pingResult := <-pingCall.Result()
		receiverHost.ID = pingResult.Sender.ID

		b = packet.NewBuilder()
		request := b.Sender(pulsarHost).Receiver(receiverHost).Request(&packet.RequestGetRandomHosts{HostsNumber: 5}).Type(packet.TypeGetRandomHosts).Build()

		call, err := t.SendRequest(request)
		if err != nil {
			log.Error(err)
			continue
		}
		result := <-call.Result()
		if result.Error != nil {
			log.Error(result.Error)
			continue
		}
		body := result.Data.(*packet.ResponseGetRandomHosts)
		if len(body.Error) != 0 {
			log.Error(body.Error)
			continue
		}

		if body.Hosts == nil || len(body.Hosts) == 0 {
			err := sendPulseToHost(pulsarHost, t, receiverHost, &pulse)
			if err != nil {
				log.Error(err)
			}
			continue
		}

		sendPulseToHosts(pulsarHost, t, body.Hosts, &pulse)
	}
}

func sendPulseToHost(sender *host.Host, t transport2.Transport, pulseReceiver *host.Host, pulse *core.Pulse) error {
	pb := packet.NewBuilder()
	pulseRequest := pb.Sender(sender).Receiver(pulseReceiver).Request(&packet.RequestPulse{Pulse: *pulse}).Type(packet.TypePulse).Build()
	call, err := t.SendRequest(pulseRequest)
	if err != nil {
		return err
	}
	result := <-call.Result()
	if result.Error != nil {
		return err
	}

	return nil
}

func sendPulseToHosts(sender *host.Host, t transport2.Transport, hosts []host.Host, pulse *core.Pulse) {
	for _, pulseReceiver := range hosts {
		err := sendPulseToHost(sender, t, &pulseReceiver, pulse)
		if err != nil {
			log.Error(err)
		}
	}
}

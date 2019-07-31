//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package insolar

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
)

// Arguments is a dedicated type for arguments, that represented as binary cbored blob
type Arguments []byte

// MarshalJSON uncbor Arguments slice recursively
func (args *Arguments) MarshalJSON() ([]byte, error) {
	result := make([]interface{}, 0)

	err := convertArgs(*args, &result)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&result)
}

func convertArgs(args []byte, result *[]interface{}) error {
	var value interface{}
	err := Deserialize(args, &value)
	if err != nil {
		return errors.Wrap(err, "Can't deserialize record")
	}

	tmp, ok := value.([]interface{})
	if !ok {
		*result = append(*result, value)
		return nil
	}

	inner := make([]interface{}, 0)

	for _, slItem := range tmp {
		switch v := slItem.(type) {
		case []byte:
			err := convertArgs(v, result)
			if err != nil {
				return err
			}
		default:
			inner = append(inner, v)
		}
	}

	*result = append(*result, inner)

	return nil
}

// MessageType is an enum type of message.
type MessageType byte

// ReplyType is an enum type of message reply.
type ReplyType byte

// Message is a routable packet, ATM just a method call
type Message interface {
	// Type returns message type.
	Type() MessageType

	// GetCaller returns initiator of this event.
	GetCaller() *Reference

	// DefaultTarget returns of target of this event.
	DefaultTarget() *Reference

	// DefaultRole returns role for this event
	DefaultRole() DynamicRole

	// AllowedSenderObjectAndRole extracts information from message
	// verify sender required to 's "caller" for sender
	// verification purpose. If nil then check of sender's role is not
	// provided by the message bus
	AllowedSenderObjectAndRole() (*Reference, DynamicRole)
}

type MessageSignature interface {
	GetSign() []byte
	GetSender() Reference
	SetSender(Reference)
}

//go:generate minimock -i github.com/insolar/insolar/insolar.Parcel -o ../testutils -s _mock.go -g

// Parcel by senders private key.
type Parcel interface {
	Message
	MessageSignature

	Message() Message
	Context(context.Context) context.Context

	Pulse() PulseNumber

	DelegationToken() DelegationToken
}

// Reply for an `Message`
type Reply interface {
	// Type returns message type.
	Type() ReplyType
}

// RedirectReply is used to create redirected messages.
type RedirectReply interface {
	// Redirected creates redirected message from redirect data.
	Redirected(genericMsg Message) Message
	// GetReceiver returns node reference to send message to.
	GetReceiver() *Reference
	// GetToken returns delegation token.
	GetToken() DelegationToken
}

// MessageSendOptions represents options for message sending.
type MessageSendOptions struct {
	Receiver *Reference
	Token    DelegationToken
}

// Safe returns original options, falling back on defaults if nil.
func (o *MessageSendOptions) Safe() *MessageSendOptions {
	if o == nil {
		return &MessageSendOptions{}
	}
	return o
}

//go:generate minimock -i github.com/insolar/insolar/insolar.MessageBus -o ../testutils -s _mock.go -g

// MessageBus interface
type MessageBus interface {
	// Send an `Message` and get a `Reply` or error from remote host.
	Send(context.Context, Message, *MessageSendOptions) (Reply, error)
	// Register saves message handler in the registry. Only one handler can be registered for a message type.
	Register(p MessageType, handler MessageHandler) error
	// MustRegister is a Register wrapper that panics if an error was returned.
	MustRegister(p MessageType, handler MessageHandler)

	// Called each new pulse, cleans next pulse messages buffer
	OnPulse(context.Context, Pulse) error
}

//go:generate minimock -i github.com/insolar/insolar/insolar.GlobalInsolarLock -o ../testutils -s _mock.go -g

// GlobalInsolarLock is lock of all incoming and outcoming network calls.
// It's not intended to be used in multiple threads. And main use of it is `Set` method of `PulseManager`.
type GlobalInsolarLock interface {
	Acquire(ctx context.Context)
	Release(ctx context.Context)
}

// MessageHandler is a function for message handling. It should be registered via Register method.
type MessageHandler func(context.Context, Parcel) (Reply, error)

//go:generate stringer -type=MessageType
const (
	// Logicrunner

	// TypeCallMethod calls method and returns request
	TypeCallMethod MessageType = iota
	// TypePutResults when execution finishes, tell results to requester
	TypeReturnResults
	// TypeExecutorResults message that goes to new Executor to validate previous Executor actions through CaseBind
	TypeExecutorResults
	// TypeValidationResults sends from Validator to new Executor with results of validation actions of previous Executor
	TypeValidationResults
	// TypePendingFinished is sent by the old executor to the current executor when pending execution finishes
	TypePendingFinished
	// TypeAdditionalCallFromPreviousExecutor is sent by the old executor to the current executor when Flow
	// cancels after registering the request but before adding the request to the execution queue.
	TypeAdditionalCallFromPreviousExecutor
	// TypeStillExecuting is sent by an old executor on pulse switch if it wants to continue executing
	// to the current executor
	TypeStillExecuting

	// TypeGenesisRequest used for bootstrap object generation.
	TypeGenesisRequest
)

// DelegationTokenType is an enum type of delegation token
type DelegationTokenType byte

//go:generate stringer -type=DelegationTokenType
const (
	// DTTypePendingExecution allows to continue method calls
	DTTypePendingExecution DelegationTokenType = iota + 1
	DTTypeGetObjectRedirect
	DTTypeGetCodeRedirect
)

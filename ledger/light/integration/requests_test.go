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

package integration_test

import (
	"crypto/rand"
	"testing"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/gen"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/stretchr/testify/require"
)

func Test_IncomingRequest_Check(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	cfg := DefaultLightConfig()
	s, err := NewServer(ctx, cfg, nil)
	require.NoError(t, err)

	// First pulse goes in storage then interrupts.
	s.SetPulse(ctx)
	// Second pulse goes in storage and starts processing, including pulse change in flow dispatcher.
	s.SetPulse(ctx)

	t.Run("registered is older than reason returns error", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()+1), true, true)
		rep := SendMessage(ctx, s, &msg)
		RequireError(rep)
	})

	t.Run("detached returns error", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)
		// Faking detached request.
		record.Unwrap(&msg.Request).(*record.IncomingRequest).ReturnMode = record.ReturnSaga
		rep := SendMessage(ctx, s, &msg)
		RequireError(rep)
	})

	t.Run("registered API request appears in pendings", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)
		rep := SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		reqInfo := rep.(*payload.RequestInfo)
		rep = CallGetPendings(ctx, s, reqInfo.RequestID)
		RequireNotError(rep)

		ids := rep.(*payload.IDs)
		require.Equal(t, 1, len(ids.IDs))
		require.Equal(t, reqInfo.RequestID, ids.IDs[0])
	})

	t.Run("registered request appears in pendings", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)
		firstObjP := SendMessage(ctx, s, &msg)
		RequireNotError(firstObjP)
		reqInfo := firstObjP.(*payload.RequestInfo)
		firstObjP, _ = CallActivateObject(ctx, s, reqInfo.RequestID)
		RequireNotError(firstObjP)

		msg, _ = MakeSetIncomingRequest(gen.ID(), reqInfo.RequestID, true, false)
		secondObjP := SendMessage(ctx, s, &msg)
		RequireNotError(secondObjP)
		secondReqInfo := secondObjP.(*payload.RequestInfo)
		secondPendings := CallGetPendings(ctx, s, secondReqInfo.RequestID)
		RequireNotError(secondPendings)

		ids := secondPendings.(*payload.IDs)
		require.Equal(t, 1, len(ids.IDs))
		require.Equal(t, secondReqInfo.RequestID, ids.IDs[0])
	})

	t.Run("closed request does not appear in pendings", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)
		rep := SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		reqInfo := rep.(*payload.RequestInfo)

		p, _ := CallActivateObject(ctx, s, reqInfo.RequestID)
		RequireNotError(p)

		p = CallGetPendings(ctx, s, reqInfo.RequestID)

		err := p.(*payload.Error)
		require.Equal(t, insolar.ErrNoPendingRequest.Error(), err.Text)
	})
}

func Test_IncomingRequest_Duplicate(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	cfg := DefaultLightConfig()
	s, err := NewServer(ctx, cfg, nil)
	require.NoError(t, err)

	// First pulse goes in storage then interrupts.
	s.SetPulse(ctx)
	// Second pulse goes in storage and starts processing, including pulse change in flow dispatcher.
	s.SetPulse(ctx)

	t.Run("creation request duplicate found", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)

		// Set first request.
		rep := SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		require.Nil(t, rep.(*payload.RequestInfo).Request)
		require.Nil(t, rep.(*payload.RequestInfo).Result)

		// Try to set it again.
		rep = SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		require.NotNil(t, rep.(*payload.RequestInfo).Request)
		require.Nil(t, rep.(*payload.RequestInfo).Result)

		// Check for result.
		receivedDuplicate := record.Material{}
		err = receivedDuplicate.Unmarshal(rep.(*payload.RequestInfo).Request)
		require.NoError(t, err)
		require.Equal(t, msg.Request, receivedDuplicate.Virtual)
	})

	t.Run("method request duplicate found", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)
		rep := SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		reasonID := rep.(*payload.RequestInfo).RequestID
		objectID := reasonID

		msg, _ = MakeSetIncomingRequest(objectID, reasonID, false, false)

		// Set first request.
		rep = SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		require.Nil(t, rep.(*payload.RequestInfo).Request)
		require.Nil(t, rep.(*payload.RequestInfo).Result)

		// Try to set it again.
		rep = SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		require.NotNil(t, rep.(*payload.RequestInfo).Request)
		require.Nil(t, rep.(*payload.RequestInfo).Result)

		// Check for found duplicate.
		receivedDuplicate := record.Material{}
		err = receivedDuplicate.Unmarshal(rep.(*payload.RequestInfo).Request)
		require.NoError(t, err)
		require.Equal(t, msg.Request, receivedDuplicate.Virtual)
	})

	t.Run("method request duplicate with result found", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)
		rep := SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		reasonID := rep.(*payload.RequestInfo).RequestID
		objectID := reasonID

		requestMsg, _ := MakeSetIncomingRequest(objectID, reasonID, false, false)

		// Set first request.
		rep = SendMessage(ctx, s, &requestMsg)
		RequireNotError(rep)
		require.Nil(t, rep.(*payload.RequestInfo).Request)
		require.Nil(t, rep.(*payload.RequestInfo).Result)
		requestID := rep.(*payload.RequestInfo).RequestID

		// Set result.
		resMsg, resultVirtual := MakeSetResult(objectID, requestID)
		rep = SendMessage(ctx, s, &resMsg)
		RequireNotError(rep)

		// Try to set request again.
		rep = SendMessage(ctx, s, &requestMsg)
		RequireNotError(rep)
		requestInfo := rep.(*payload.RequestInfo)
		require.NotNil(t, requestInfo.Request)
		require.NotNil(t, requestInfo.Result)

		// Check for found duplicate.
		receivedDuplicate := record.Material{}
		err = receivedDuplicate.Unmarshal(requestInfo.Request)
		require.NoError(t, err)
		require.Equal(t, requestMsg.Request, receivedDuplicate.Virtual)

		// Check for result duplicate.
		receivedResult := record.Material{}
		err = receivedResult.Unmarshal(requestInfo.Result)
		require.NoError(t, err)
		require.Equal(t, resultVirtual, receivedResult.Virtual)
	})
}

func Test_OutgoingRequest_Duplicate(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	cfg := DefaultLightConfig()
	s, err := NewServer(ctx, cfg, nil)
	require.NoError(t, err)

	// First pulse goes in storage then interrupts.
	s.SetPulse(ctx)
	// Second pulse goes in storage and starts processing, including pulse change in flow dispatcher.
	s.SetPulse(ctx)

	t.Run("method request duplicate found", func(t *testing.T) {
		args := make([]byte, 100)
		_, err := rand.Read(args)
		initReq := record.IncomingRequest{
			Object:    insolar.NewReference(gen.ID()),
			Arguments: args,
			CallType:  record.CTSaveAsChild,
			Reason:    *insolar.NewReference(*insolar.NewID(s.pulse.PulseNumber, []byte{1, 2, 3})),
			APINode:   gen.Reference(),
		}
		initReqMsg := &payload.SetIncomingRequest{
			Request: record.Wrap(&initReq),
		}

		// Set first request
		p := SendMessage(ctx, s, initReqMsg)
		RequireNotError(p)
		reqInfo := p.(*payload.RequestInfo)
		require.Nil(t, reqInfo.Request)
		require.Nil(t, reqInfo.Result)

		outgoingReq := record.OutgoingRequest{
			Object:   insolar.NewReference(reqInfo.RequestID),
			Reason:   *insolar.NewReference(reqInfo.RequestID),
			CallType: record.CTMethod,
			Caller:   *insolar.NewReference(reqInfo.RequestID),
		}
		outgoingReqMsg := &payload.SetOutgoingRequest{
			Request: record.Wrap(&outgoingReq),
		}

		// Set outgoing request
		outP := SendMessage(ctx, s, outgoingReqMsg)
		RequireNotError(outP)
		outReqInfo := p.(*payload.RequestInfo)
		require.Nil(t, outReqInfo.Request)
		require.Nil(t, outReqInfo.Result)

		// Try to set an outgoing again
		outSecondP := SendMessage(ctx, s, outgoingReqMsg)
		RequireNotError(outSecondP)
		outReqSecondInfo := outSecondP.(*payload.RequestInfo)
		require.NotNil(t, outReqSecondInfo.Request)
		require.Nil(t, outReqSecondInfo.Result)

		// Check for the result
		receivedDuplicate := record.Material{}
		err = receivedDuplicate.Unmarshal(outReqSecondInfo.Request)
		require.NoError(t, err)
		require.Equal(t, &outgoingReq, record.Unwrap(&receivedDuplicate.Virtual))
	})
}

func Test_DetachedRequest_notification(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	cfg := DefaultLightConfig()

	received := make(chan payload.SagaCallAcceptNotification)
	s, err := NewServer(ctx, cfg, func(meta payload.Meta, pl payload.Payload) {
		if notification, ok := pl.(*payload.SagaCallAcceptNotification); ok {
			received <- *notification
		}
	})
	require.NoError(t, err)

	// First pulse goes in storage then interrupts.
	s.SetPulse(ctx)
	// Second pulse goes in storage and starts processing, including pulse change in flow dispatcher.
	s.SetPulse(ctx)

	t.Run("detached notification sent on detached reason close", func(t *testing.T) {
		msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)
		rep := SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		objectID := rep.(*payload.RequestInfo).ObjectID

		msg, _ = MakeSetIncomingRequest(objectID, gen.IDWithPulse(s.Pulse()), false, true)
		rep = SendMessage(ctx, s, &msg)
		RequireNotError(rep)
		reasonID := rep.(*payload.RequestInfo).RequestID

		p, detachedRec := CallSetOutgoingRequest(ctx, s, objectID, reasonID, true)
		RequireNotError(p)
		detachedID := p.(*payload.RequestInfo).RequestID

		resMsg, _ := MakeSetResult(objectID, reasonID)
		rep = SendMessage(ctx, s, &resMsg)
		RequireNotError(rep)

		notification := <-received
		require.Equal(t, objectID, notification.ObjectID)
		require.Equal(t, detachedID, notification.DetachedRequestID)

		receivedRec := record.Virtual{}
		err := receivedRec.Unmarshal(notification.Request)
		require.NoError(t, err)
		require.Equal(t, detachedRec, receivedRec)
	})
}

func Test_Result_Duplicate(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	cfg := DefaultLightConfig()
	s, err := NewServer(ctx, cfg, nil)
	require.NoError(t, err)

	// First pulse goes in storage then interrupts.
	s.SetPulse(ctx)
	// Second pulse goes in storage and starts processing, including pulse change in flow dispatcher.
	s.SetPulse(ctx)

	msg, _ := MakeSetIncomingRequest(gen.ID(), gen.IDWithPulse(s.Pulse()), true, true)

	// Set request.
	rep := SendMessage(ctx, s, &msg)
	RequireNotError(rep)
	require.Nil(t, rep.(*payload.RequestInfo).Request)
	require.Nil(t, rep.(*payload.RequestInfo).Result)
	requestID := rep.(*payload.RequestInfo).RequestID
	objectID := requestID

	resMsg, resultVirtual := MakeSetResult(objectID, requestID)
	// Set result.
	rep = SendMessage(ctx, s, &resMsg)
	RequireNotError(rep)

	// Try to set it again.
	rep = SendMessage(ctx, s, &resMsg)
	RequireNotError(rep)

	resultInfo := rep.(*payload.ResultInfo)
	require.NotNil(t, resultInfo.Result)

	// Check duplicate.
	receivedResult := record.Material{}
	err = receivedResult.Unmarshal(resultInfo.Result)
	require.NoError(t, err)
	require.Equal(t, resultVirtual, receivedResult.Virtual)
}

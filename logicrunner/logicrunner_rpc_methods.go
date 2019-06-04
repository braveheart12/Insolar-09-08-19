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

package logicrunner

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/instracer"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/logicrunner/artifacts"
	"github.com/insolar/insolar/logicrunner/goplugin/rpctypes"
)

type RPCMethods struct {
	lr *LogicRunner
}

func NewRPCMethods(lr *LogicRunner) *RPCMethods {
	return &RPCMethods{lr: lr}
}

func recoverRPC(err *error) {
	if r := recover(); r != nil {
		// Global logger is used because there is no access to context here
		log.Errorf("Recovered panic:\n%s", string(debug.Stack()))
		if err != nil {
			if *err == nil {
				*err = errors.New(fmt.Sprint(r))
			} else {
				*err = errors.New(fmt.Sprint(*err, r))
			}
		}
	}
}

// GetCode is an RPC retrieving a code by its reference
func (m *RPCMethods) GetCode(req rpctypes.UpGetCodeReq, reply *rpctypes.UpGetCodeResp) (err error) {
	defer recoverRPC(&err)

	os := m.lr.GetObjectState(req.Callee)
	if os == nil {
		return errors.New("Failed to find requested object state. ref: " + req.Callee.String())
	}
	es, err := os.GetModeState(req.Mode)
	if err != nil {
		return errors.Wrap(err, "Failed to find needed execution state")
	}
	ctx := es.Current.Context

	inslogger.FromContext(ctx).Debug("In RPC.GetCode ....")

	ctx, span := instracer.StartSpan(ctx, "service.GetCode")
	defer span.End()

	codeDescriptor, err := m.lr.ArtifactManager.GetCode(ctx, req.Code)
	if err != nil {
		return err
	}
	reply.Code, err = codeDescriptor.Code()
	if err != nil {
		return err
	}
	return nil
}

// RouteCall routes call from a contract to a contract through event bus.
func (m *RPCMethods) RouteCall(req rpctypes.UpRouteReq, rep *rpctypes.UpRouteResp) (err error) {
	defer recoverRPC(&err)

	os := m.lr.GetObjectState(req.Callee)
	if os == nil {
		return errors.New("Failed to find requested object state. ref: " + req.Callee.String())
	}
	es, err := os.GetModeState(req.Mode)
	if err != nil {
		return errors.Wrap(err, "Failed to find needed execution state")
	}
	ctx := es.Current.Context

	if os.ExecutionState.Current.LogicContext.Immutable {
		return errors.New("Try to call route from immutable method")
	}

	// TODO: delegation token

	es.Current.Nonce++

	msg := &message.CallMethod{
		Request: record.Request{
			Caller:          req.Callee,
			CallerPrototype: req.CalleePrototype,
			Nonce:           es.Current.Nonce,

			Immutable: req.Immutable,

			Object:    &req.Object,
			Prototype: &req.Prototype,
			Method:    req.Method,
			Arguments: req.Arguments,
		},
	}

	if !req.Wait {
		msg.ReturnMode = record.ReturnNoWait
	}

	res, err := m.lr.ContractRequester.CallMethod(ctx, msg)
	if err != nil {
		return err
	}

	if req.Wait {
		rep.Result = res.(*reply.CallMethod).Result
	}

	return nil
}

// SaveAsChild is an RPC saving data as memory of a contract as child a parent
func (m *RPCMethods) SaveAsChild(req rpctypes.UpSaveAsChildReq, rep *rpctypes.UpSaveAsChildResp) (err error) {
	defer recoverRPC(&err)

	os := m.lr.GetObjectState(req.Callee)
	if os == nil {
		return errors.New("Failed to find requested object state. ref: " + req.Callee.String())
	}
	es, err := os.GetModeState(req.Mode)
	if err != nil {
		return errors.Wrap(err, "Failed to find needed execution state")
	}
	ctx := es.Current.Context

	es.Current.Nonce++

	msg := &message.CallMethod{
		Request: record.Request{
			Caller:          req.Callee,
			CallerPrototype: req.CalleePrototype,
			Nonce:           es.Current.Nonce,

			CallType:  record.CTSaveAsChild,
			Base:      &req.Parent,
			Prototype: &req.Prototype,
			Method:    req.ConstructorName,
			Arguments: req.ArgsSerialized,
		},
	}

	ref, err := m.lr.ContractRequester.CallConstructor(ctx, msg)

	rep.Reference = ref

	return err
}

// SaveAsDelegate is an RPC saving data as memory of a contract as child a parent
func (m *RPCMethods) SaveAsDelegate(req rpctypes.UpSaveAsDelegateReq, rep *rpctypes.UpSaveAsDelegateResp) (err error) {
	defer recoverRPC(&err)

	os := m.lr.GetObjectState(req.Callee)
	if os == nil {
		return errors.New("Failed to find requested object state. ref: " + req.Callee.String())
	}
	es, err := os.GetModeState(req.Mode)
	if err != nil {
		return errors.Wrap(err, "Failed to find needed execution state")
	}
	ctx := es.Current.Context

	es.Current.Nonce++

	msg := &message.CallMethod{
		Request: record.Request{
			Caller:          req.Callee,
			CallerPrototype: req.CalleePrototype,
			Nonce:           es.Current.Nonce,

			CallType:  record.CTSaveAsDelegate,
			Base:      &req.Into,
			Prototype: &req.Prototype,
			Method:    req.ConstructorName,
			Arguments: req.ArgsSerialized,
		},
	}

	ref, err := m.lr.ContractRequester.CallConstructor(ctx, msg)

	rep.Reference = ref
	return err
}

var iteratorBuffSize = 1000
var iteratorMap = make(map[string]artifacts.RefIterator)
var iteratorMapLock = sync.RWMutex{}

// GetObjChildrenIterator is an RPC returns an iterator over object children with specified prototype
func (m *RPCMethods) GetObjChildrenIterator(
	req rpctypes.UpGetObjChildrenIteratorReq,
	rep *rpctypes.UpGetObjChildrenIteratorResp,
) (
	err error,
) {

	defer recoverRPC(&err)

	os := m.lr.GetObjectState(req.Callee)
	if os == nil {
		return errors.New("Failed to find requested object state. ref: " + req.Callee.String())
	}
	es, err := os.GetModeState(req.Mode)
	if err != nil {
		return errors.Wrap(err, "Failed to find needed execution state")
	}
	ctx := es.Current.Context
	am := m.lr.ArtifactManager
	iteratorID := req.IteratorID

	iteratorMapLock.RLock()
	iterator, ok := iteratorMap[iteratorID]
	iteratorMapLock.RUnlock()

	if !ok {
		newIterator, err := am.GetChildren(ctx, req.Object, nil)
		if err != nil {
			return errors.Wrap(err, "[ GetObjChildrenIterator ] Can't get children")
		}

		id, err := uuid.NewV4()
		if err != nil {
			return errors.Wrap(err, "[ GetObjChildrenIterator ] Can't generate UUID")
		}

		iteratorID = id.String()

		iteratorMapLock.Lock()
		iterator, ok = iteratorMap[iteratorID]
		if !ok {
			iteratorMap[iteratorID] = newIterator
			iterator = newIterator
		}
		iteratorMapLock.Unlock()
	}

	iter := iterator

	rep.Iterator.ID = iteratorID
	rep.Iterator.CanFetch = iter.HasNext()
	for len(rep.Iterator.Buff) < iteratorBuffSize && iter.HasNext() {
		r, err := iter.Next()
		if err != nil {
			return errors.Wrap(err, "[ GetObjChildrenIterator ] Can't get Next")
		}
		rep.Iterator.CanFetch = iter.HasNext()

		o, err := am.GetObject(ctx, *r)

		if err != nil {
			if err == insolar.ErrDeactivated {
				continue
			}
			return errors.Wrap(err, "[ GetObjChildrenIterator ] Can't call GetObject on Next")
		}
		protoRef, err := o.Prototype()
		if err != nil {
			return errors.Wrap(err, "[ GetObjChildrenIterator ] Can't get prototype reference")
		}

		if protoRef.Equal(req.Prototype) {
			rep.Iterator.Buff = append(rep.Iterator.Buff, *r)
		}
	}

	if !iter.HasNext() {
		iteratorMapLock.Lock()
		delete(iteratorMap, rep.Iterator.ID)
		iteratorMapLock.Unlock()
	}

	return nil
}

// GetDelegate is an RPC saving data as memory of a contract as child a parent
func (m *RPCMethods) GetDelegate(req rpctypes.UpGetDelegateReq, rep *rpctypes.UpGetDelegateResp) (err error) {
	defer recoverRPC(&err)

	os := m.lr.GetObjectState(req.Callee)
	if os == nil {
		return errors.New("Failed to find requested object state. ref: " + req.Callee.String())
	}
	es, err := os.GetModeState(req.Mode)
	if err != nil {
		return errors.Wrap(err, "Failed to find needed execution state")
	}
	ctx := es.Current.Context

	ref, err := m.lr.ArtifactManager.GetDelegate(ctx, req.Object, req.OfType)
	if err != nil {
		return err
	}
	rep.Object = *ref
	return nil
}

// DeactivateObject is an RPC saving data as memory of a contract as child a parent
func (m *RPCMethods) DeactivateObject(req rpctypes.UpDeactivateObjectReq, rep *rpctypes.UpDeactivateObjectResp) (err error) {
	defer recoverRPC(&err)

	os := m.lr.GetObjectState(req.Callee)
	if os == nil {
		return errors.New("Failed to find requested object state. ref: " + req.Callee.String())
	}
	es, err := os.GetModeState(req.Mode)
	if err != nil {
		return errors.Wrap(err, "Failed to find needed execution state")
	}

	es.Current.Deactivate = true
	return nil
}
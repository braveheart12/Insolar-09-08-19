/*
 *    Copyright 2018 INS Ecosystem
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

package core

import (
	"fmt"

	"github.com/insolar/insolar/genesis/model/class"
	"github.com/insolar/insolar/genesis/model/domain"
	"github.com/insolar/insolar/genesis/model/factory"
	"github.com/insolar/insolar/genesis/model/object"
)

// ReferenceDomainName is a name for reference domain.
const ReferenceDomainName = "ReferenceDomain"

// ReferenceDomain is a contract that allows to publish and resolve global references.
type ReferenceDomain interface {
	// Base domain implementation.
	domain.Domain
	// RegisterReference is used to publish new global references.
	RegisterReference(*object.Reference, string) (string, error)
	// ResolveReference provides reference instance from record.
	ResolveReference(string) (*object.Reference, error)
	// InitGlobalMap sets globalResolverMap for references register/resolving.
	InitGlobalMap(globalInstanceMap *map[string]object.Proxy)
}

type referenceDomain struct {
	domain.BaseDomain
	globalResolverMap *map[string]object.Proxy
}

// newReferenceDomain creates new instance of ReferenceDomain.
func newReferenceDomain(parent object.Parent) *referenceDomain {
	refDomain := &referenceDomain{
		BaseDomain: *domain.NewBaseDomain(parent, ReferenceDomainName),
	}
	// Bootstrap case
	if parent == nil {
		refDomain.Parent = refDomain
	}
	return refDomain
}

// GetClassID returns string representation of ReferenceDomain's class.
func (rd *referenceDomain) GetClassID() string {
	return class.ReferenceDomainID
}

// InitGlobalMap sets globalResolverMap for register/resolve references.
func (rd *referenceDomain) InitGlobalMap(globalInstanceMap *map[string]object.Proxy) {
	if rd.globalResolverMap != nil {
		return
	}
	rd.globalResolverMap = globalInstanceMap
}

// RegisterReference sets new reference as a child to domain storage.
func (rd *referenceDomain) RegisterReference(ref *object.Reference, classID string) (string, error) {
	record, err := rd.ChildStorage.Set(ref)
	if err != nil {
		return "", err
	}
	resolver := rd.GetResolver()
	obj, err := resolver.GetObject(ref, classID)
	if err != nil {
		return "", err
	}
	proxy, ok := obj.(object.Proxy)
	if !ok {
		return "", fmt.Errorf("object with reference `%s` is not `Proxy` instance", ref)
	}
	(*rd.globalResolverMap)[record] = proxy

	return record, nil
}

// ResolveReference returns reference from its record in domain storage.
func (rd *referenceDomain) ResolveReference(record string) (*object.Reference, error) {
	reference, err := rd.ChildStorage.Get(record)
	if err != nil {
		return nil, err
	}

	result, ok := reference.(*object.Reference)
	if !ok {
		return nil, fmt.Errorf("object with record `%s` is not `Reference` instance", record)
	}

	return result, nil
}

type referenceDomainProxy struct {
	//instance *referenceDomain
	object.BaseProxy
}

// newReferenceDomainProxy creates new proxy and associate it with new instance of ReferenceDomain.
func newReferenceDomainProxy(parent object.Parent) *referenceDomainProxy {
	return &referenceDomainProxy{
		BaseProxy: object.BaseProxy{
			Instance: newReferenceDomain(parent),
		},
	}
}

// RegisterReference proxy call for instance method.
func (rdp *referenceDomainProxy) RegisterReference(address *object.Reference, classID string) (string, error) {
	return rdp.Instance.(*referenceDomain).RegisterReference(address, classID)
}

// ResolveReference is a proxy call for instance method.
func (rdp *referenceDomainProxy) ResolveReference(record string) (*object.Reference, error) {
	return rdp.Instance.(*referenceDomain).ResolveReference(record)
}

// GetReference is a proxy call for instance method.
/*func (rdp *referenceDomainProxy) GetReference() *object.Reference {
	return rdp.Instance.GetReference()
}

// GetParent is a proxy call for instance method.
func (rdp *referenceDomainProxy) GetParent() object.Parent {
	return rdp.Instance.GetParent()
}

// GetClassID is a proxy call for instance method.
func (rdp *referenceDomainProxy) GetClassID() string {
	return class.ReferenceDomainID
}*/

// InitGlobalMap is a proxy call for instance method.
func (rdp *referenceDomainProxy) InitGlobalMap(globalInstanceMap *map[string]object.Proxy) {
	rdp.Instance.(*referenceDomain).InitGlobalMap(globalInstanceMap)
}

type referenceDomainFactory struct{}

// NewReferenceDomainFactory creates new factory for ReferenceDomain.
func NewReferenceDomainFactory() factory.Factory {
	return &referenceDomainFactory{}
}

// GetClassID returns string representation of ReferenceDomain's class.
func (adf *referenceDomainFactory) GetClassID() string {
	return class.ReferenceDomainID
}

// GetReference returns nil for not published factory.
func (adf *referenceDomainFactory) GetReference() *object.Reference {
	return nil
}

func (adf *referenceDomainFactory) SetReference(reference *object.Reference) {
}

// Create factory method for new ReferenceDomain instances.
func (adf *referenceDomainFactory) Create(parent object.Parent) (object.Proxy, error) {
	proxy := newReferenceDomainProxy(parent)
	_, err := parent.AddChild(proxy)
	if err != nil {
		return nil, err
	}
	return proxy, nil
}

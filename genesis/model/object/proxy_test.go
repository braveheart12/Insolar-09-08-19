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

package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockInstance struct {
	ref *Reference
}

func (p *mockInstance) GetClassID() string {
	return "mockChild"
}

func (p *mockInstance) GetReference() *Reference {
	return p.ref
}

func (p *mockInstance) SetReference(reference *Reference) {
	p.ref = reference
}

func (p *mockInstance) GetParent() Parent {
	return nil
}

func TestBaseProxy_GetClassID(t *testing.T) {
	proxy := &BaseProxy{
		Instance: &mockInstance{},
	}
	assert.Equal(t, "mockChild", proxy.GetClassID())
}

func TestBaseProxy_SetGetReference(t *testing.T) {
	ref, _ := NewReference("1", "2", GlobalScope)
	proxy := &BaseProxy{
		Instance: &mockInstance{},
	}
	proxy.SetReference(ref)
	assert.Equal(t, ref, proxy.GetReference())
}

func TestBaseProxy_GetParent(t *testing.T) {
	proxy := &BaseProxy{
		Instance: &mockInstance{},
	}
	assert.Nil(t, proxy.GetParent())
}

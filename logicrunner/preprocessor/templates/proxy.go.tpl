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

package {{ .PackageName }}

import (
{{- range $import, $i := .Imports }}
	{{ $import }}
{{- end }}
)

{{ range $typeStruct := .Types }}
	{{- $typeStruct }}
{{ end }}

// PrototypeReference to prototype of this contract
// error checking hides in generator
var PrototypeReference, _ = insolar.NewReferenceFromBase58("{{ .ClassReference }}")


// {{ .ContractType }} holds proxy type
type {{ .ContractType }} struct {
	Reference insolar.Reference
	Prototype insolar.Reference
	Code insolar.Reference
}

// ContractConstructorHolder holds logic with object construction
type ContractConstructorHolder struct {
	constructorName string
	argsSerialized []byte
}

// AsChild saves object as child
func (r *ContractConstructorHolder) AsChild(objRef insolar.Reference) (*{{ .ContractType }}, error) {
	ref, ret, err := common.CurrentProxyCtx.SaveAsChild(objRef, *PrototypeReference, r.constructorName, r.argsSerialized)
	if err != nil {
		return nil, err
	}

	var constructorError *foundation.Error
	err = common.CurrentProxyCtx.Deserialize(ret, []interface{}{ &constructorError })
	if err != nil {
		return nil, err
	}

	if constructorError != nil {
		return nil, constructorError
	}

	return &{{ .ContractType }}{Reference: *ref}, nil
}

// GetObject returns proxy object
func GetObject(ref insolar.Reference) (r *{{ .ContractType }}) {
	return &{{ .ContractType }}{Reference: ref}
}

// GetPrototype returns reference to the prototype
func GetPrototype() insolar.Reference {
	return *PrototypeReference
}

{{ range $func := .ConstructorsProxies }}
// {{ $func.Name }} is constructor
func {{ $func.Name }}( {{ $func.Arguments }} ) *ContractConstructorHolder {
	{{ $func.InitArgs }}

	var argsSerialized []byte
	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		panic(err)
	}

	return &ContractConstructorHolder{constructorName: "{{ $func.Name }}", argsSerialized: argsSerialized}
}
{{ end }}

// GetReference returns reference of the object
func (r *{{ $.ContractType }}) GetReference() insolar.Reference {
	return r.Reference
}

// GetPrototype returns reference to the code
func (r *{{ $.ContractType }}) GetPrototype() (insolar.Reference, error) {
	if r.Prototype.IsEmpty() {
		ret := [2]interface{}{}
		var ret0 insolar.Reference
		ret[0] = &ret0
		var ret1 *foundation.Error
		ret[1] = &ret1

		res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, false, "GetPrototype", make([]byte, 0), *PrototypeReference)
		if err != nil {
			return ret0, err
		}

		err = common.CurrentProxyCtx.Deserialize(res, &ret)
		if err != nil {
			return ret0, err
		}

		if ret1 != nil {
			return ret0, ret1
		}

		r.Prototype = ret0
	}

	return r.Prototype, nil

}

// GetCode returns reference to the code
func (r *{{ $.ContractType }}) GetCode() (insolar.Reference, error) {
	if r.Code.IsEmpty() {
		ret := [2]interface{}{}
		var ret0 insolar.Reference
		ret[0] = &ret0
		var ret1 *foundation.Error
		ret[1] = &ret1

		res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, false, "GetCode", make([]byte, 0), *PrototypeReference)
		if err != nil {
			return ret0, err
		}

		err = common.CurrentProxyCtx.Deserialize(res, &ret)
		if err != nil {
			return ret0, err
		}

		if ret1 != nil {
			return ret0, ret1
		}

		r.Code = ret0
	}

	return r.Code, nil
}

{{ range $method := .MethodsProxies }}
// {{ $method.Name }} is proxy generated method
func (r *{{ $.ContractType }}) {{ $method.Name }}{{if $method.Immutable}}AsMutable{{end}}( {{ $method.Arguments }} ) ( {{ $method.ResultsTypes }} ) {
	{{ $method.InitArgs }}
	var argsSerialized []byte

	{{ $method.ResultZeroList }}

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return {{ $method.ResultsWithErr }}
	}

	{{/* Saga call doesn't has a reply (it's `nil`), thus we shouldn't try to deserialize it. */}}
	{{if $method.SagaInfo.IsSaga }}
	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, true, false, {{ $method.SagaInfo.IsSaga }}, "{{ $method.Name }}", argsSerialized, *PrototypeReference)
	if err != nil {
		return {{ $method.ResultsWithErr }}
	}
	{{else}}
	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, {{ $method.SagaInfo.IsSaga }}, "{{ $method.Name }}", argsSerialized, *PrototypeReference)
	if err != nil {
		return {{ $method.ResultsWithErr }}
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return {{ $method.ResultsWithErr }}
	}

	if {{ $method.ErrorVar }} != nil {
		return {{ $method.Results }}
	}
	{{end -}}

	return {{ $method.ResultsNilError }}
}

{{if not $method.SagaInfo.IsSaga}}
// {{ $method.Name }}NoWait is proxy generated method
func (r *{{ $.ContractType }}) {{ $method.Name }}NoWait( {{ $method.Arguments }} ) error {
	{{ $method.InitArgs }}
	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, false, "{{ $method.Name }}", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// {{ $method.Name }}AsImmutable is proxy generated method
func (r *{{ $.ContractType }}) {{ $method.Name }}{{if not $method.Immutable}}AsImmutable{{end}}( {{ $method.Arguments }} ) ( {{ $method.ResultsTypes }} ) {
	{{ $method.InitArgs }}
	var argsSerialized []byte

	{{ $method.ResultZeroList }}

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return {{ $method.ResultsWithErr }}
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, false, "{{ $method.Name }}", argsSerialized, *PrototypeReference)
	if err != nil {
		return {{ $method.ResultsWithErr }}
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return {{ $method.ResultsWithErr }}
	}

	if {{ $method.ErrorVar }} != nil {
		return {{ $method.Results }}
	}
	return {{ $method.ResultsNilError }}
}
{{ end }}
{{ end }}

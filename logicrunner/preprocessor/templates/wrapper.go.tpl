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

package {{ .Package }}

import (
{{- range $import, $i := .Imports }}
	{{ $import }}
{{- end }}
{{ if $.GenerateInitialize -}}
    XXX_insolar "github.com/insolar/insolar/insolar"
{{- end }}
// TODO: this is a part of horrible hack for making "index not found" error NOT system error. You MUST remove it in INS-3099
{{ if .Methods }}
    "strings"
{{ end }}
// TODO: this is the end of a horrible hack, please remove it
)

type ExtendableError struct{
	S string
}

func ( e *ExtendableError ) Error() string{
	return e.S
}

func INS_META_INFO() ([] map[string]string) {
	result := make([]map[string] string, 0)
	{{ range $method := .Methods }}
		{{ if $method.SagaInfo.IsSaga }}
        {
		info := make(map[string] string, 3)
		info["Type"] = "SagaInfo"
		info["MethodName"] = "{{ $method.Name }}"
		info["RollbackMethodName"] = "{{ $method.SagaInfo.RollbackMethodName }}"
        result = append(result, info)
        }
		{{end}}
	{{end}}
	return result
}

func INSMETHOD_GetCode(object []byte, data []byte) ([]byte, []byte, error) {
	ph := common.CurrentProxyCtx
	self := new({{ $.ContractType }})

	if len(object) == 0 {
		return nil, nil, &ExtendableError{ S: "[ Fake GetCode ] ( Generated Method ) Object is nil"}
	}

	err := ph.Deserialize(object, self)
	if err != nil {
		e := &ExtendableError{ S: "[ Fake GetCode ] ( Generated Method ) Can't deserialize args.Data: " + err.Error() }
		return nil, nil, e
	}

	state := []byte{}
	err = ph.Serialize(self, &state)
	if err != nil {
		return nil, nil, err
	}

	ret := []byte{}
	err = ph.Serialize([]interface{} { self.GetCode().Bytes() }, &ret)

	return state, ret, err
}

func INSMETHOD_GetPrototype(object []byte, data []byte) ([]byte, []byte, error) {
	ph := common.CurrentProxyCtx
	self := new({{ $.ContractType }})

	if len(object) == 0 {
		return nil, nil, &ExtendableError{ S: "[ Fake GetPrototype ] ( Generated Method ) Object is nil"}
	}

	err := ph.Deserialize(object, self)
	if err != nil {
		e := &ExtendableError{ S: "[ Fake GetPrototype ] ( Generated Method ) Can't deserialize args.Data: " + err.Error() }
		return nil, nil, e
	}

	state := []byte{}
	err = ph.Serialize(self, &state)
	if err != nil {
		return nil, nil, err
	}

	ret := []byte{}
	err = ph.Serialize([]interface{} { self.GetPrototype().Bytes() }, &ret)

	return state, ret, err
}

{{ range $method := .Methods }}
func INSMETHOD_{{ $method.Name }}(object []byte, data []byte) ([]byte, []byte, error) {
	ph := common.CurrentProxyCtx
	ph.SetSystemError(nil)
	self := new({{ $.ContractType }})

	if len(object) == 0 {
		return nil, nil, &ExtendableError{ S: "[ Fake{{ $method.Name }} ] ( INSMETHOD_* ) ( Generated Method ) Object is nil"}
	}

	err := ph.Deserialize(object, self)
	if err != nil {
		e := &ExtendableError{ S: "[ Fake{{ $method.Name }} ] ( INSMETHOD_* ) ( Generated Method ) Can't deserialize args.Data: " + err.Error() }
		return nil, nil, e
	}

	{{ $method.ArgumentsZeroList }}
	err = ph.Deserialize(data, &args)
	if err != nil {
		e := &ExtendableError{ S: "[ Fake{{ $method.Name }} ] ( INSMETHOD_* ) ( Generated Method ) Can't deserialize args.Arguments: " + err.Error() }
		return nil, nil, e
	}

{{ if $method.Results }}
	{{ $method.Results }} := self.{{ $method.Name }}( {{ $method.Arguments }} )
{{ else }}
	self.{{ $method.Name }}( {{ $method.Arguments }} )
{{ end }}

// TODO: this is a part of horrible hack for making "index not found" error NOT system error. You MUST remove it in INS-3099
	systemErr := ph.GetSystemError()

	if systemErr != nil && strings.Contains(systemErr.Error(), "index not found") {
		systemErr = nil
	}
// TODO: this is the end of a horrible hack, please remove it

	if systemErr != nil {
		return nil, nil, ph.GetSystemError()
	}

	state := []byte{}
	err = ph.Serialize(self, &state)
	if err != nil {
		return nil, nil, err
	}

{{ range $i := $method.ErrorInterfaceInRes }}
	ret{{ $i }} = ph.MakeErrorSerializable(ret{{ $i }})
{{ end }}

	ret := []byte{}
	err = ph.Serialize([]interface{} { {{ $method.Results }} }, &ret)

	return state, ret, err
}
{{ end }}


{{ range $f := .Functions }}
func INSCONSTRUCTOR_{{ $f.Name }}(data []byte) ([]byte, []byte, error) {
	ph := common.CurrentProxyCtx
	ph.SetSystemError(nil)
	{{ $f.ArgumentsZeroList }}
	err := ph.Deserialize(data, &args)
	if err != nil {
		e := &ExtendableError{ S: "[ Fake{{ $f.Name }} ] ( INSCONSTRUCTOR_* ) ( Generated Method ) Can't deserialize args.Arguments: " + err.Error() }
		return nil, nil, e
	}

	{{ $f.Results }} := {{ $f.Name }}( {{ $f.Arguments }} )
	ret1 = ph.MakeErrorSerializable(ret1)
	if ret0 == nil && ret1 == nil {
		ret1 = &ExtendableError{ S: "constructor returned nil" }
	}

	result := []byte{}
	err = ph.Serialize([]interface{} { ret1 }, &result)
	if err != nil {
		return nil, nil, err
	}

	if ret1 != nil {
		// logical error, the result should be registered with type RequestSideEffectNone
		return nil, result, nil
	}

	state := []byte{}
	err = ph.Serialize(ret0, &state)
	if err != nil {
		return nil, nil, err
	}

	return state, result, nil
}
{{ end }}

{{ if $.GenerateInitialize -}}
func Initialize() XXX_insolar.ContractWrapper {
    return XXX_insolar.ContractWrapper{
        GetCode: INSMETHOD_GetCode,
        GetPrototype: INSMETHOD_GetPrototype,
        Methods: XXX_insolar.ContractMethods{
            {{ range $method := .Methods -}}
                    "{{ $method.Name }}": INSMETHOD_{{ $method.Name }},
            {{ end }}
        },
        Constructors: XXX_insolar.ContractConstructors{
            {{ range $f := .Functions -}}
                    "{{ $f.Name }}": INSCONSTRUCTOR_{{ $f.Name }},
            {{ end }}
        },
    }
}
{{- end }}

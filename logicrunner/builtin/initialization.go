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

// THIS CODE IS AUTOGENERATED

package builtin

import (
	"github.com/pkg/errors"

	costcenter "github.com/insolar/insolar/logicrunner/builtin/contract/costcenter"
	deposit "github.com/insolar/insolar/logicrunner/builtin/contract/deposit"
	helloworld "github.com/insolar/insolar/logicrunner/builtin/contract/helloworld"
	member "github.com/insolar/insolar/logicrunner/builtin/contract/member"
	migrationadmin "github.com/insolar/insolar/logicrunner/builtin/contract/migrationadmin"
	migrationshard "github.com/insolar/insolar/logicrunner/builtin/contract/migrationshard"
	nodedomain "github.com/insolar/insolar/logicrunner/builtin/contract/nodedomain"
	noderecord "github.com/insolar/insolar/logicrunner/builtin/contract/noderecord"
	pkshard "github.com/insolar/insolar/logicrunner/builtin/contract/pkshard"
	rootdomain "github.com/insolar/insolar/logicrunner/builtin/contract/rootdomain"
	wallet "github.com/insolar/insolar/logicrunner/builtin/contract/wallet"

	XXX_insolar "github.com/insolar/insolar/insolar"
	XXX_rootdomain "github.com/insolar/insolar/insolar/rootdomain"
	XXX_artifacts "github.com/insolar/insolar/logicrunner/artifacts"
)

func InitializeContractMethods() map[string]XXX_insolar.ContractWrapper {
	return map[string]XXX_insolar.ContractWrapper{
		"costcenter":     costcenter.Initialize(),
		"deposit":        deposit.Initialize(),
		"helloworld":     helloworld.Initialize(),
		"member":         member.Initialize(),
		"migrationadmin": migrationadmin.Initialize(),
		"migrationshard": migrationshard.Initialize(),
		"nodedomain":     nodedomain.Initialize(),
		"noderecord":     noderecord.Initialize(),
		"pkshard":        pkshard.Initialize(),
		"rootdomain":     rootdomain.Initialize(),
		"wallet":         wallet.Initialize(),
	}
}

func shouldLoadRef(strRef string) XXX_insolar.Reference {
	ref, err := XXX_insolar.NewReferenceFromBase58(strRef)
	if err != nil {
		panic(errors.Wrap(err, "Unexpected error, bailing out"))
	}
	return *ref
}

func InitializeCodeRefs() map[XXX_insolar.Reference]string {
	rv := make(map[XXX_insolar.Reference]string, 0)

	rv[shouldLoadRef("111A7tUo1FeZ5DSoroiinMCKwzLacaYBAAcwAaNj6bc.11111111111111111111111111111111")] = "costcenter"
	rv[shouldLoadRef("111A79KGpeDUjYhRJP1n1AwYgwU9KEWmc2TNNc3KQjV.11111111111111111111111111111111")] = "deposit"
	rv[shouldLoadRef("111A5w1GcnTsht82duVrnWdVHVNyrxCUVcSPLtgQCPR.11111111111111111111111111111111")] = "helloworld"
	rv[shouldLoadRef("111A72gPKWyrF9c7yzDoccRoPQ62g1uQQDBecWJwAYr.11111111111111111111111111111111")] = "member"
	rv[shouldLoadRef("111A6516TVnMLh8DAzTWbtEJrgZkESeCpdn2viV6D61.11111111111111111111111111111111")] = "migrationadmin"
	rv[shouldLoadRef("111A66L3aoDPf2wedyRo2gyns8ghV9vdeJdJntVaGEf.11111111111111111111111111111111")] = "migrationshard"
	rv[shouldLoadRef("111A7Q5FK2ebPG9WnSiUc4iqF45w9oYkJkRjEtBohGe.11111111111111111111111111111111")] = "nodedomain"
	rv[shouldLoadRef("111A86xPKUQ1ZxSscgv5brbw93LkwiVhUWgGrYYsMar.11111111111111111111111111111111")] = "noderecord"
	rv[shouldLoadRef("111A5tzn16hnKGCZCyYA8Dv9FALvPYYQu4VA41SVx6s.11111111111111111111111111111111")] = "pkshard"
	rv[shouldLoadRef("111A63R5cAgGHC5DJffqF16vUkCuSVj3GExbMLy56cS.11111111111111111111111111111111")] = "rootdomain"
	rv[shouldLoadRef("111A5e49cJW6GKGegWBhtgrJs7nFh1kSWhBtT2VgK4t.11111111111111111111111111111111")] = "wallet"

	return rv
}

func InitializeCodeDescriptors() []XXX_artifacts.CodeDescriptor {
	rv := make([]XXX_artifacts.CodeDescriptor, 0)

	// costcenter
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A7tUo1FeZ5DSoroiinMCKwzLacaYBAAcwAaNj6bc.11111111111111111111111111111111"),
	))
	// deposit
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A79KGpeDUjYhRJP1n1AwYgwU9KEWmc2TNNc3KQjV.11111111111111111111111111111111"),
	))
	// helloworld
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A5w1GcnTsht82duVrnWdVHVNyrxCUVcSPLtgQCPR.11111111111111111111111111111111"),
	))
	// member
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A72gPKWyrF9c7yzDoccRoPQ62g1uQQDBecWJwAYr.11111111111111111111111111111111"),
	))
	// migrationadmin
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A6516TVnMLh8DAzTWbtEJrgZkESeCpdn2viV6D61.11111111111111111111111111111111"),
	))
	// migrationshard
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A66L3aoDPf2wedyRo2gyns8ghV9vdeJdJntVaGEf.11111111111111111111111111111111"),
	))
	// nodedomain
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A7Q5FK2ebPG9WnSiUc4iqF45w9oYkJkRjEtBohGe.11111111111111111111111111111111"),
	))
	// noderecord
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A86xPKUQ1ZxSscgv5brbw93LkwiVhUWgGrYYsMar.11111111111111111111111111111111"),
	))
	// pkshard
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A5tzn16hnKGCZCyYA8Dv9FALvPYYQu4VA41SVx6s.11111111111111111111111111111111"),
	))
	// rootdomain
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A63R5cAgGHC5DJffqF16vUkCuSVj3GExbMLy56cS.11111111111111111111111111111111"),
	))
	// wallet
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A5e49cJW6GKGegWBhtgrJs7nFh1kSWhBtT2VgK4t.11111111111111111111111111111111"),
	))

	return rv
}

func InitializePrototypeDescriptors() []XXX_artifacts.ObjectDescriptor {
	rv := make([]XXX_artifacts.ObjectDescriptor, 0)

	{ // costcenter
		pRef := shouldLoadRef("111A62HrJvAimG7M1r8XdeBVMw4X6ge8hGzVStfnn4e.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A7tUo1FeZ5DSoroiinMCKwzLacaYBAAcwAaNj6bc.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // deposit
		pRef := shouldLoadRef("111A7ctasuNUug8BoK4VJNuAFJ73rnH8bH5zqd5HrDj.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A79KGpeDUjYhRJP1n1AwYgwU9KEWmc2TNNc3KQjV.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // helloworld
		pRef := shouldLoadRef("111A85JAZugtAkQErbDe3eAaTw56DPLku8QGymJUCt2.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A5w1GcnTsht82duVrnWdVHVNyrxCUVcSPLtgQCPR.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // member
		pRef := shouldLoadRef("111A7UqbgvFXj9vkCAaNYSAkWLapu62eU5AUSv3y4JY.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A72gPKWyrF9c7yzDoccRoPQ62g1uQQDBecWJwAYr.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // migrationadmin
		pRef := shouldLoadRef("111A8DhUhw5pzyvzVg1qXomNEHXs7kDtJRQGSD1PUpc.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A6516TVnMLh8DAzTWbtEJrgZkESeCpdn2viV6D61.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // migrationshard
		pRef := shouldLoadRef("111A7FNYLZLYXYWZPbkMhCAPwV9nYrWWE7L57CtdJCj.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A66L3aoDPf2wedyRo2gyns8ghV9vdeJdJntVaGEf.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // nodedomain
		pRef := shouldLoadRef("111A6NKbCjpzFr9MttfcWV8vX8eFjiyGPPfSH1AMtwN.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A7Q5FK2ebPG9WnSiUc4iqF45w9oYkJkRjEtBohGe.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // noderecord
		pRef := shouldLoadRef("111A5fZeApbGhcsLrbfGy82kKLgapF93GhNPMLSYaPY.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A86xPKUQ1ZxSscgv5brbw93LkwiVhUWgGrYYsMar.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // pkshard
		pRef := shouldLoadRef("111A5x8N1VJTm7BKYgzSe6TWHcFi98QZgw3AnkYiKML.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A5tzn16hnKGCZCyYA8Dv9FALvPYYQu4VA41SVx6s.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // rootdomain
		pRef := shouldLoadRef("111A84uiiTD1LXAHNP4GMA6YJFjbnCdkRia2pCqwBV5.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A63R5cAgGHC5DJffqF16vUkCuSVj3GExbMLy56cS.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	{ // wallet
		pRef := shouldLoadRef("111A5gmRD1ZbHjQh7DgH9SrCK4a1qfwEUP5xAir6i8L.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A5e49cJW6GKGegWBhtgrJs7nFh1kSWhBtT2VgK4t.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	return rv
}

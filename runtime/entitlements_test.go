/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2022 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/tests/checker"
	. "github.com/onflow/cadence/runtime/tests/utils"
)

func TestRuntimeAccountEntitlementSaveAndLoadSuccess(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntime()
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) entitlement X
			access(all) entitlement Y
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				signer.save(3, to: /storage/foo)
				signer.link<auth(Test.X, Test.Y) &Int>(/public/foo, target: /storage/foo)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let cap = signer.getCapability<auth(Test.X, Test.Y) &Int>(/public/foo)
				let ref = cap.borrow()!
				let downcastRef = ref as! auth(Test.X, Test.Y) &Int
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log:     func(message string) {},
		emitEvent: func(event cadence.Event) error {
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 1}}, nil
		},
		updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
			accountCodes[location] = code
			return nil
		},
		getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
			code = accountCodes[location]
			return code, nil
		},
	}

	nextTransactionLocation := newTransactionLocationGenerator()

	err := rt.ExecuteTransaction(
		Script{
			Source: deployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction1,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction2,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

}

func TestRuntimeAccountEntitlementSaveAndLoadFail(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntime()
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) entitlement X
			access(all) entitlement Y
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				signer.save(3, to: /storage/foo)
				signer.link<auth(Test.X, Test.Y) &Int>(/public/foo, target: /storage/foo)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let cap = signer.getCapability<auth(Test.X) &Int>(/public/foo)
				let ref = cap.borrow()!
				let downcastRef = ref as! auth(Test.X, Test.Y) &Int
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log:     func(message string) {},
		emitEvent: func(event cadence.Event) error {
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 1}}, nil
		},
		updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
			accountCodes[location] = code
			return nil
		},
		getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
			code = accountCodes[location]
			return code, nil
		},
	}

	nextTransactionLocation := newTransactionLocationGenerator()

	err := rt.ExecuteTransaction(
		Script{
			Source: deployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction1,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction2,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)

	require.ErrorAs(t, err, &interpreter.ForceCastTypeMismatchError{})
}

func TestRuntimeAccountEntitlementAttachmentMap(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) entitlement X
			access(all) entitlement Y

			access(all) entitlement mapping M {
				X -> Y
			}

			access(all) resource R {}
			
			access(M) attachment A for R {
				access(Y) fun foo() {}
			}

			access(all) fun createRWithA(): @R {
				return <-attach A() to <-create R()
			}
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r <- Test.createRWithA()
				signer.save(<-r, to: /storage/foo)
				signer.link<auth(Test.X) &Test.R>(/public/foo, target: /storage/foo)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let cap = signer.getCapability<auth(Test.X) &Test.R>(/public/foo)
				let ref = cap.borrow()!
				ref[Test.A]!.foo()
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log:     func(message string) {},
		emitEvent: func(event cadence.Event) error {
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 1}}, nil
		},
		updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
			accountCodes[location] = code
			return nil
		},
		getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
			code = accountCodes[location]
			return code, nil
		},
	}

	nextTransactionLocation := newTransactionLocationGenerator()

	err := rt.ExecuteTransaction(
		Script{
			Source: deployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction1,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction2,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)

	require.NoError(t, err)
}

func TestRuntimeAccountExportEntitledRef(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntime()
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) entitlement X

			access(all) resource R {}

			access(all) fun createR(): @R {
				return <-create R()
			}
		}
	`))

	script := []byte(`
		import Test from 0x1
		access(all) fun main(): &Test.R { 
			let r <- Test.createR()
			let authAccount = getAuthAccount(0x1)
			authAccount.save(<-r, to: /storage/foo)
			let ref = authAccount.borrow<auth(Test.X) &Test.R>(from: /storage/foo)!
			return ref
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log:     func(message string) {},
		emitEvent: func(event cadence.Event) error {
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 1}}, nil
		},
		updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
			accountCodes[location] = code
			return nil
		},
		getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
			code = accountCodes[location]
			return code, nil
		},
	}

	nextTransactionLocation := newTransactionLocationGenerator()
	nextScriptLocation := newScriptLocationGenerator()

	err := rt.ExecuteTransaction(
		Script{
			Source: deployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	value, err := rt.ExecuteScript(
		Script{
			Source: script,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextScriptLocation(),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "A.0000000000000001.Test.R(uuid: 0)", value.String())
}

func TestRuntimeAccountEntitlementNamingConflict(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntime()
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) entitlement X

			access(all) resource R {
				access(X) fun foo() {}
			}

			access(all) fun createR(): @R {
				return <-create R()
			}
		}
	`))

	otherDeployTx := DeploymentTransaction("OtherTest", []byte(`
		access(all) contract OtherTest {
			access(all) entitlement X
		}
	`))

	script := []byte(`
		import Test from 0x1
		import OtherTest from 0x1

		access(all) fun main(){ 
			let r <- Test.createR()
			let ref = &r as auth(OtherTest.X) &Test.R
			ref.foo()
			destroy r
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log:     func(message string) {},
		emitEvent: func(event cadence.Event) error {
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 1}}, nil
		},
		updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
			accountCodes[location] = code
			return nil
		},
		getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
			code = accountCodes[location]
			return code, nil
		},
	}

	nextTransactionLocation := newTransactionLocationGenerator()
	nextScriptLocation := newScriptLocationGenerator()

	err := rt.ExecuteTransaction(
		Script{
			Source: deployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: otherDeployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	_, err = rt.ExecuteScript(
		Script{
			Source: script,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextScriptLocation(),
		},
	)

	var checkerErr *sema.CheckerError
	require.ErrorAs(t, err, &checkerErr)

	errs := checker.RequireCheckerErrors(t, checkerErr, 1)

	var accessError *sema.InvalidAccessError
	require.ErrorAs(t, errs[0], &accessError)
}

func TestRuntimeAccountEntitlementCapabilityCasting(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) entitlement X
			access(all) entitlement Y

			access(all) resource R {}

			access(all) fun createR(): @R {
				return <-create R()
			}
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r <- Test.createR()
				signer.save(<-r, to: /storage/foo)
				signer.link<auth(Test.X) &Test.R>(/public/foo, target: /storage/foo)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let capX = signer.getCapability<auth(Test.X) &Test.R>(/public/foo)
				let upCap = capX as Capability<&Test.R>
				let downCap = upCap as! Capability<auth(Test.X) &Test.R>
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log:     func(message string) {},
		emitEvent: func(event cadence.Event) error {
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 1}}, nil
		},
		updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
			accountCodes[location] = code
			return nil
		},
		getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
			code = accountCodes[location]
			return code, nil
		},
	}

	nextTransactionLocation := newTransactionLocationGenerator()

	err := rt.ExecuteTransaction(
		Script{
			Source: deployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction1,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction2,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)

	require.ErrorAs(t, err, &interpreter.ForceCastTypeMismatchError{})
}

func TestRuntimeAccountEntitlementCapabilityDictionary(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) entitlement X
			access(all) entitlement Y

			access(all) resource R {}

			access(all) fun createR(): @R {
				return <-create R()
			}
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r <- Test.createR()
				signer.save(<-r, to: /storage/foo)
				signer.link<auth(Test.X) &Test.R>(/public/foo, target: /storage/foo)

				let r2 <- Test.createR()
				signer.save(<-r2, to: /storage/bar)
				signer.link<auth(Test.Y) &Test.R>(/public/bar, target: /storage/bar)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let capX = signer.getCapability<auth(Test.X) &Test.R>(/public/foo)
				let capY = signer.getCapability<auth(Test.Y) &Test.R>(/public/bar)

				let dict: {Type: Capability<&Test.R>} = {}
				dict[capX.getType()] = capX
				dict[capY.getType()] = capY

				let newCapX = dict[capX.getType()]!
				let ref = newCapX.borrow()!
				let downCast = ref as! auth(Test.X) &Test.R
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log:     func(message string) {},
		emitEvent: func(event cadence.Event) error {
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 1}}, nil
		},
		updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
			accountCodes[location] = code
			return nil
		},
		getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
			code = accountCodes[location]
			return code, nil
		},
	}

	nextTransactionLocation := newTransactionLocationGenerator()

	err := rt.ExecuteTransaction(
		Script{
			Source: deployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction1,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction2,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)

	require.ErrorAs(t, err, &interpreter.ForceCastTypeMismatchError{})
}

func TestRuntimeAccountEntitlementGenericCapabilityDictionary(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) entitlement X
			access(all) entitlement Y

			access(all) resource R {}

			access(all) fun createR(): @R {
				return <-create R()
			}
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r <- Test.createR()
				signer.save(<-r, to: /storage/foo)
				signer.link<auth(Test.X) &Test.R>(/public/foo, target: /storage/foo)

				let r2 <- Test.createR()
				signer.save(<-r2, to: /storage/bar)
				signer.link<auth(Test.Y) &Test.R>(/public/bar, target: /storage/bar)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let capX = signer.getCapability<auth(Test.X) &Test.R>(/public/foo)
				let capY = signer.getCapability<auth(Test.Y) &Test.R>(/public/bar)

				let dict: {Type: Capability} = {}
				dict[capX.getType()] = capX
				dict[capY.getType()] = capY

				let newCapX = dict[capX.getType()]!
				let ref = newCapX.borrow<auth(Test.X) &Test.R>()!
				let downCast = ref as! auth(Test.X) &Test.R
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log:     func(message string) {},
		emitEvent: func(event cadence.Event) error {
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 1}}, nil
		},
		updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
			accountCodes[location] = code
			return nil
		},
		getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
			code = accountCodes[location]
			return code, nil
		},
	}

	nextTransactionLocation := newTransactionLocationGenerator()

	err := rt.ExecuteTransaction(
		Script{
			Source: deployTx,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction1,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	err = rt.ExecuteTransaction(
		Script{
			Source: transaction2,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextTransactionLocation(),
		},
	)

	require.NoError(t, err)
}

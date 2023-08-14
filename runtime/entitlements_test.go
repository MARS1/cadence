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
            prepare(signer: auth(Storage, Capabilities) &Account) {
                signer.storage.save(3, to: /storage/foo)
                let cap = signer.capabilities.storage.issue<auth(Test.X, Test.Y) &Int>(/storage/foo)
                signer.capabilities.publish(cap, at: /public/foo)
            }
        }
     `)

	transaction2 := []byte(`
        import Test from 0x1
        transaction {
            prepare(signer: &Account) {
                let ref = signer.capabilities.borrow<auth(Test.X, Test.Y) &Int>(/public/foo)!
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
            prepare(signer: auth(Storage, Capabilities) &Account) {
                signer.storage.save(3, to: /storage/foo)
                let cap = signer.capabilities.storage.issue<auth(Test.X, Test.Y) &Int>(/storage/foo)
                signer.capabilities.publish(cap, at: /public/foo)
            }
        }
     `)

	transaction2 := []byte(`
        import Test from 0x1
        transaction {
            prepare(signer: &Account) {
                let ref = signer.capabilities.borrow<auth(Test.X) &Int>(/public/foo)!
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
            prepare(signer: auth(Storage, Capabilities) &Account) {
                let r <- Test.createRWithA()
                signer.storage.save(<-r, to: /storage/foo)
                let cap = signer.capabilities.storage.issue<auth(Test.X) &Test.R>(/storage/foo)
                signer.capabilities.publish(cap, at: /public/foo)
            }
        }
     `)

	transaction2 := []byte(`
        import Test from 0x1

        transaction {
            prepare(signer: &Account) {
                let ref = signer.capabilities.borrow<auth(Test.X) &Test.R>(/public/foo)!
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
            let authAccount = getAuthAccount<auth(Storage) &Account>(0x1)
            authAccount.storage.save(<-r, to: /storage/foo)
            let ref = authAccount.storage.borrow<auth(Test.X) &Test.R>(from: /storage/foo)!
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
	require.Equal(t, "A.0000000000000001.Test.R(uuid: 1)", value.String())
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

        access(all) fun main() {
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
            prepare(signer: auth(Storage, Capabilities) &Account) {
                let r <- Test.createR()
                signer.storage.save(<-r, to: /storage/foo)
                let cap = signer.capabilities.storage.issue<auth(Test.X) &Test.R>(/storage/foo)
                signer.capabilities.publish(cap, at: /public/foo)
            }
        }
     `)

	transaction2 := []byte(`
        import Test from 0x1
        transaction {
            prepare(signer: &Account) {
                let capX = signer.capabilities.get<auth(Test.X) &Test.R>(/public/foo)!
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
            prepare(signer: auth(Storage, Capabilities) &Account) {
                let r <- Test.createR()
                signer.storage.save(<-r, to: /storage/foo)

                let capFoo = signer.capabilities.storage.issue<auth(Test.X) &Test.R>(/storage/foo)
                signer.capabilities.publish(capFoo, at: /public/foo)

                let r2 <- Test.createR()
                signer.storage.save(<-r2, to: /storage/bar)

                let capBar = signer.capabilities.storage.issue<auth(Test.Y) &Test.R>(/storage/bar)
                signer.capabilities.publish(capBar, at: /public/bar)
            }
        }
     `)

	transaction2 := []byte(`
        import Test from 0x1
        transaction {
            prepare(signer: &Account) {
                let capX = signer.capabilities.get<auth(Test.X) &Test.R>(/public/foo)!
                let capY = signer.capabilities.get<auth(Test.Y) &Test.R>(/public/bar)!

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
            prepare(signer: auth(Storage, Capabilities) &Account) {
                let r <- Test.createR()
                signer.storage.save(<-r, to: /storage/foo)

                let capFoo = signer.capabilities.storage.issue<auth(Test.X) &Test.R>(/storage/foo)
                signer.capabilities.publish(capFoo, at: /public/foo)

                let r2 <- Test.createR()
                signer.storage.save(<-r2, to: /storage/bar)

                let capBar = signer.capabilities.storage.issue<auth(Test.Y) &Test.R>(/storage/bar)
                signer.capabilities.publish(capBar, at: /public/bar)
            }
        }
     `)

	transaction2 := []byte(`
        import Test from 0x1
        transaction {
            prepare(signer: &Account) {
                let capX = signer.capabilities.get<auth(Test.X) &Test.R>(/public/foo)!
                let capY = signer.capabilities.get<auth(Test.Y) &Test.R>(/public/bar)!

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

func TestRuntimeCapabilityEntitlements(t *testing.T) {

	t.Parallel()

	address := common.MustBytesToAddress([]byte{0x1})

	test := func(t *testing.T, script string) {
		runtime := newTestInterpreterRuntime()

		accountCodes := map[common.Location][]byte{}

		runtimeInterface := &testRuntimeInterface{
			storage: newTestLedger(nil, nil),
			getSigningAccounts: func() ([]Address, error) {
				return []Address{address}, nil
			},
			resolveLocation: singleIdentifierLocationResolver(t),
			updateAccountContractCode: func(location common.AddressLocation, code []byte) error {
				accountCodes[location] = code
				return nil
			},
			getAccountContractCode: func(location common.AddressLocation) (code []byte, err error) {
				code = accountCodes[location]
				return code, nil
			},
			emitEvent: func(event cadence.Event) error {
				return nil
			},
		}

		_, err := runtime.ExecuteScript(
			Script{
				Source: []byte(script),
			},
			Context{
				Interface: runtimeInterface,
				Location:  common.ScriptLocation{},
			},
		)
		require.NoError(t, err)
	}

	t.Run("can borrow with supertype", func(t *testing.T) {
		t.Parallel()

		test(t, `
          access(all)
          entitlement X

          access(all)
          entitlement Y

          access(all)
          resource R {}

          access(all)
          fun main() {
              let account = getAuthAccount<auth(Storage, Capabilities) &Account>(0x1)

              let r <- create R()
              account.storage.save(<-r, to: /storage/foo)

              let issuedCap = account.capabilities.storage.issue<auth(X, Y) &R>(/storage/foo)
              account.capabilities.publish(issuedCap, at: /public/foo)

              let ref = account.capabilities.borrow<auth(X | Y) &R>(/public/foo)
              assert(ref != nil, message: "failed borrow")
          }
        `)
	})

	t.Run("cannot borrow with supertype then downcast", func(t *testing.T) {
		t.Parallel()

		test(t, `
          access(all)
          entitlement X

          access(all)
          entitlement Y

          access(all)
          resource R {}

          access(all)
          fun main() {
              let account = getAuthAccount<auth(Storage, Capabilities) &Account>(0x1)

              let r <- create R()
              account.storage.save(<-r, to: /storage/foo)

              let issuedCap = account.capabilities.storage.issue<auth(X, Y) &R>(/storage/foo)
              account.capabilities.publish(issuedCap, at: /public/foo)

              let ref = account.capabilities.borrow<auth(X | Y) &R>(/public/foo)
              assert(ref != nil, message: "failed borrow")

              let ref2 = ref! as? auth(X, Y) &R
              assert(ref2 == nil, message: "invalid cast")
          }
        `)
	})

	t.Run("can borrow with two types", func(t *testing.T) {
		t.Parallel()

		test(t, `
          access(all)
	      entitlement X

          access(all)
	      entitlement Y

          access(all)
	      resource R {}

          access(all)
	      fun main() {
               let account = getAuthAccount<auth(Storage, Capabilities) &Account>(0x1)

               let r <- create R()
               account.storage.save(<-r, to: /storage/foo)

               let issuedCap = account.capabilities.storage.issue<auth(X, Y) &R>(/storage/foo)
               account.capabilities.publish(issuedCap, at: /public/foo)

	           let ref = account.capabilities.borrow<auth(X, Y) &R>(/public/foo)
               assert(ref != nil, message: "failed borrow")

               let ref2 = ref! as? auth(X, Y) &R
               assert(ref2 != nil, message: "failed cast")
	      }
	    `)
	})

	t.Run("upcast runtime entitlements", func(t *testing.T) {
		t.Parallel()

		test(t, `
          access(all)
	      entitlement X

          access(all)
	      struct S {}

          access(all)
	      fun main() {
              let account = getAuthAccount<auth(Storage, Capabilities) &Account>(0x1)

	          let s = S()
	          account.storage.save(s, to: /storage/foo)

	          let issuedCap = account.capabilities.storage.issue<auth(X) &S>(/storage/foo)
              account.capabilities.publish(issuedCap, at: /public/foo)

	          let cap: Capability<auth(X) &S> = account.capabilities.get<auth(X) &S>(/public/foo)!

	          let runtimeType = cap.getType()

	          let upcastCap = cap as Capability<&S>
	          let upcastRuntimeType = upcastCap.getType()

	          assert(runtimeType != upcastRuntimeType)
	      }
	    `)
	})

	t.Run("upcast runtime type", func(t *testing.T) {
		t.Parallel()

		test(t, `
          access(all)
	      struct S {}

          access(all)
	      fun main() {
              let account = getAuthAccount<auth(Storage, Capabilities) &Account>(0x1)

	          let s = S()
	          account.storage.save(s, to: /storage/foo)

	          let issuedCap = account.capabilities.storage.issue<&S>(/storage/foo)
              account.capabilities.publish(issuedCap, at: /public/foo)

	          let cap: Capability<&S> = account.capabilities.get<&S>(/public/foo)!

	          let runtimeType = cap.getType()
	          let upcastCap = cap as Capability<&AnyStruct>
	          let upcastRuntimeType = upcastCap.getType()
	          assert(runtimeType == upcastRuntimeType)
	       }
	    `)
	})

	t.Run("can check with supertype", func(t *testing.T) {
		t.Parallel()

		test(t, `
          access(all)
	      entitlement X

          access(all)
	      entitlement Y

          access(all)
	      resource R {}

          access(all)
	      fun main() {
              let account = getAuthAccount<auth(Storage, Capabilities) &Account>(0x1)

	          let r <- create R()
	          account.storage.save(<-r, to: /storage/foo)

	          let issuedCap = account.capabilities.storage.issue<auth(X, Y) &R>(/storage/foo)
              account.capabilities.publish(issuedCap, at: /public/foo)

	          let cap = account.capabilities.get<auth(X | Y) &R>(/public/foo)!
	          assert(cap.check())
	      }
	    `)
	})

	t.Run("cannot borrow with subtype", func(t *testing.T) {
		t.Parallel()

		test(t, `
          access(all)
	      entitlement X

          access(all)
	      entitlement Y

          access(all)
	      resource R {}

          access(all)
	      fun main() {
              let account = getAuthAccount<auth(Storage, Capabilities) &Account>(0x1)

	          let r <- create R()
	          account.storage.save(<-r, to: /storage/foo)

	          let issuedCap = account.capabilities.storage.issue<auth(X) &R>(/storage/foo)
              account.capabilities.publish(issuedCap, at: /public/foo)

	          let ref = account.capabilities.borrow<auth(X, Y) &R>(/public/foo)
	          assert(ref == nil)
	      }
	    `)
	})

	t.Run("cannot get with subtype", func(t *testing.T) {
		t.Parallel()

		test(t, `
          access(all)
	      entitlement X

          access(all)
	      entitlement Y

          access(all)
	      resource R {}

          access(all)
	      fun main() {
              let account = getAuthAccount<auth(Storage, Capabilities) &Account>(0x1)

	          let r <- create R()
	          account.storage.save(<-r, to: /storage/foo)

	          let issuedCap = account.capabilities.storage.issue<auth(X) &R>(/storage/foo)
              account.capabilities.publish(issuedCap, at: /public/foo)

	          let cap = account.capabilities.get<auth(X, Y) &R>(/public/foo)
	          assert(cap == nil)
	      }
	    `)
	})
}

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"
	. "github.com/onflow/cadence/runtime/tests/utils"
)

func newTestInterpreterRuntimeWithAttachments() testInterpreterRuntime {
	rt := newTestInterpreterRuntime()
	rt.interpreterRuntime.defaultConfig.AttachmentsEnabled = true
	return rt
}

func TestAccountAttachmentSaveAndLoad(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()

	var logs []string
	var events []string
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) resource R {
				access(all) fun foo(): Int {
					return 3
				}
			}
			access(all) attachment A for R {
				access(all) fun foo(): Int {
					return base.foo()
				}
			}
			access(all) fun makeRWithA(): @R {
				return <- attach A() to <-create R()
			}
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r <- Test.makeRWithA()
				signer.save(<-r, to: /storage/foo)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r <- signer.load<@Test.R>(from: /storage/foo)!
				let i = r[Test.A]!.foo()
				destroy r
				log(i)
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log: func(message string) {
			logs = append(logs, message)
		},
		emitEvent: func(event cadence.Event) error {
			events = append(events, event.String())
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

	require.Equal(t, []string{"3"}, logs)
}

func TestAccountAttachmentExportFailure(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()

	logs := make([]string, 0)
	events := make([]string, 0)
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) resource R {}
			access(all) attachment A for R {}
			access(all) fun makeRWithA(): @R {
				return <- attach A() to <-create R()
			}
		}
	`))

	script := []byte(`
		import Test from 0x1
		access(all) fun main(): &Test.A? { 
			let r <- Test.makeRWithA()
			let a = r[Test.A]
			destroy r
			return a
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log: func(message string) {
			logs = append(logs, message)
		},
		emitEvent: func(event cadence.Event) error {
			events = append(events, event.String())
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

	_, err = rt.ExecuteScript(
		Script{
			Source: script,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextScriptLocation(),
		},
	)
	require.Error(t, err)
	require.ErrorAs(t, err, &interpreter.DestroyedResourceError{})
}

func TestAccountAttachmentExport(t *testing.T) {

	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()

	var logs []string
	var events []string
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) resource R {}
			access(all) attachment A for R {}
			access(all) fun makeRWithA(): @R {
				return <- attach A() to <-create R()
			}
		}
	`))

	script := []byte(`
		import Test from 0x1
		access(all) fun main(): &Test.A? { 
			let r <- Test.makeRWithA()
			let authAccount = getAuthAccount(0x1)
			authAccount.save(<-r, to: /storage/foo)
			let ref = authAccount.borrow<&Test.R>(from: /storage/foo)!
			let a = ref[Test.A]
			return a
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log: func(message string) {
			logs = append(logs, message)
		},
		emitEvent: func(event cadence.Event) error {
			events = append(events, event.String())
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

	v, err := rt.ExecuteScript(
		Script{
			Source: script,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextScriptLocation(),
		},
	)
	require.NoError(t, err)
	require.IsType(t, cadence.Optional{}, v)
	require.IsType(t, cadence.Attachment{}, v.(cadence.Optional).Value)
	require.Equal(t, "A.0000000000000001.Test.A()", v.(cadence.Optional).Value.String())
}

func TestAccountAttachedExport(t *testing.T) {

	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()

	var logs []string
	var events []string
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) resource R {}
			access(all) attachment A for R {}
			access(all) fun makeRWithA(): @R {
				return <- attach A() to <-create R()
			}
		}
	`))

	script := []byte(`
		import Test from 0x1
		access(all) fun main(): @Test.R { 
			return <-Test.makeRWithA()
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log: func(message string) {
			logs = append(logs, message)
		},
		emitEvent: func(event cadence.Event) error {
			events = append(events, event.String())
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

	v, err := rt.ExecuteScript(
		Script{
			Source: script,
		},
		Context{
			Interface: runtimeInterface1,
			Location:  nextScriptLocation(),
		},
	)
	require.NoError(t, err)

	require.IsType(t, cadence.Resource{}, v)
	require.Len(t, v.(cadence.Resource).Fields, 2)
	require.IsType(t, cadence.Attachment{}, v.(cadence.Resource).Fields[1])
	require.Equal(t, "A.0000000000000001.Test.A()", v.(cadence.Resource).Fields[1].String())
}

func TestAccountAttachmentSaveAndBorrow(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()

	var logs []string
	var events []string
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) resource interface I {
				access(all) fun foo(): Int
			}
			access(all) resource R: I {
				access(all) fun foo(): Int {
					return 3
				}
			}
			access(all) attachment A for I {
				access(all) fun foo(): Int {
					return base.foo()
				}
			}
			access(all) fun makeRWithA(): @R {
				return <- attach A() to <-create R()
			}
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r <- Test.makeRWithA()
				signer.save(<-r, to: /storage/foo)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r = signer.borrow<&{Test.I}>(from: /storage/foo)!
				let a: &Test.A = r[Test.A]!
				let i = a.foo()
				log(i)
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log: func(message string) {
			logs = append(logs, message)
		},
		emitEvent: func(event cadence.Event) error {
			events = append(events, event.String())
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

	require.Equal(t, []string{"3"}, logs)
}

func TestAccountAttachmentCapability(t *testing.T) {
	t.Parallel()

	storage := newTestLedger(nil, nil)
	rt := newTestInterpreterRuntimeWithAttachments()

	var logs []string
	var events []string
	accountCodes := map[Location][]byte{}

	deployTx := DeploymentTransaction("Test", []byte(`
		access(all) contract Test {
			access(all) resource interface I {
				access(all) fun foo(): Int
			}
			access(all) resource R: I {
				access(all) fun foo(): Int {
					return 3
				}
			}
			access(all) attachment A for I {
				access(all) fun foo(): Int {
					return base.foo()
				}
			}
			access(all) fun makeRWithA(): @R {
				return <- attach A() to <-create R()
			}
		}
	`))

	transaction1 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let r <- Test.makeRWithA()
				signer.save(<-r, to: /storage/foo)
				let cap = signer.capabilities.storage.issue<&{Test.I}>(/storage/foo)!
				signer.inbox.publish(cap, name: "foo", recipient: 0x2)
			}
		}
	 `)

	transaction2 := []byte(`
		import Test from 0x1
		transaction {
			prepare(signer: AuthAccount) {
				let cap = signer.inbox.claim<&{Test.I}>("foo", provider: 0x1)!
				let ref = cap.borrow()!
				let i = ref[Test.A]!.foo()
				log(i)
			}
		}
	 `)

	runtimeInterface1 := &testRuntimeInterface{
		storage: storage,
		log: func(message string) {
			logs = append(logs, message)
		},
		emitEvent: func(event cadence.Event) error {
			events = append(events, event.String())
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

	runtimeInterface2 := &testRuntimeInterface{
		storage: storage,
		log: func(message string) {
			logs = append(logs, message)
		},
		emitEvent: func(event cadence.Event) error {
			events = append(events, event.String())
			return nil
		},
		resolveLocation: singleIdentifierLocationResolver(t),
		getSigningAccounts: func() ([]Address, error) {
			return []Address{[8]byte{0, 0, 0, 0, 0, 0, 0, 2}}, nil
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
			Interface: runtimeInterface2,
			Location:  nextTransactionLocation(),
		},
	)
	require.NoError(t, err)

	require.Equal(t, []string{"3"}, logs)
}

func TestRuntimeAttachmentStorage(t *testing.T) {
	t.Parallel()

	address := common.MustBytesToAddress([]byte{0x1})

	newRuntime := func() (testInterpreterRuntime, *testRuntimeInterface) {
		runtime := newTestInterpreterRuntime()
		runtime.defaultConfig.AttachmentsEnabled = true

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
		return runtime, runtimeInterface
	}

	t.Run("save and load", func(t *testing.T) {

		t.Parallel()

		runtime, runtimeInterface := newRuntime()

		const script = `
          access(all)
          resource R {}

          access(all)
          attachment A for R {

              access(all)
              fun foo(): Int { return 3 }
          }

          access(all)
          fun main(): Int {
              let authAccount = getAuthAccount(0x1)

              let r <- create R()
              let r2 <- attach A() to <-r
              authAccount.save(<-r2, to: /storage/foo)
              let r3 <- authAccount.load<@R>(from: /storage/foo)!
              let i = r3[A]?.foo()!
              destroy r3
              return i
          }
        `
		result, err := runtime.ExecuteScript(
			Script{
				Source: []byte(script),
			},
			Context{
				Interface: runtimeInterface,
				Location:  common.ScriptLocation{},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, cadence.NewInt(3), result)
	})

	t.Run("save and borrow", func(t *testing.T) {
		t.Parallel()

		runtime, runtimeInterface := newRuntime()

		const script = `
		  access(all)
	      resource R {}

		  access(all)
	      attachment A for R {

	          access(all)
	          fun foo(): Int { return 3 }
	      }

	      access(all)
          fun main(): Int {
              let authAccount = getAuthAccount(0x1)

	          let r <- create R()
	          let r2 <- attach A() to <-r
	          authAccount.save(<-r2, to: /storage/foo)
	          let r3 = authAccount.borrow<&R>(from: /storage/foo)!
	          return r3[A]?.foo()!
	      }
	    `

		result, err := runtime.ExecuteScript(
			Script{
				Source: []byte(script),
			},
			Context{
				Interface: runtimeInterface,
				Location:  common.ScriptLocation{},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, cadence.NewInt(3), result)
	})

	t.Run("capability", func(t *testing.T) {
		t.Parallel()

		runtime, runtimeInterface := newRuntime()

		const script = `
		  access(all)
	      resource R {}

		  access(all)
	      attachment A for R {

	          access(all)
	          fun foo(): Int { return 3 }
	      }

	      access(all)
          fun main(): Int {
              let authAccount = getAuthAccount(0x1)
              let pubAccount = getAccount(0x1)

	          let r <- create R()
	          let r2 <- attach A() to <-r
	          authAccount.save(<-r2, to: /storage/foo)
	          let cap = authAccount.capabilities.storage
                  .issue<&R>(/storage/foo)
              authAccount.capabilities.publish(cap, at: /public/foo)

	          let ref = pubAccount.capabilities.borrow<&R>(/public/foo)!
	          return ref[A]?.foo()!
	      }
	    `

		result, err := runtime.ExecuteScript(
			Script{
				Source: []byte(script),
			},
			Context{
				Interface: runtimeInterface,
				Location:  common.ScriptLocation{},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, cadence.NewInt(3), result)
	})

	t.Run("capability interface", func(t *testing.T) {

		t.Parallel()

		runtime, runtimeInterface := newRuntime()

		const script = `
	      access(all)
	      resource R: I {}

	      access(all)
	      resource interface I {}

	      access(all)
	      attachment A for I {

	          access(all)
	          fun foo(): Int { return 3 }
	      }

	      access(all)
          fun main(): Int {
              let authAccount = getAuthAccount(0x1)
              let pubAccount = getAccount(0x1)

	          let r <- create R()
	          let r2 <- attach A() to <-r
	          authAccount.save(<-r2, to: /storage/foo)
	          let cap = authAccount.capabilities.storage
                    .issue<&{I}>(/storage/foo)
              authAccount.capabilities.publish(cap, at: /public/foo)

	          let ref = pubAccount.capabilities.borrow<&{I}>(/public/foo)!
	          return ref[A]?.foo()!
	      }
	    `

		result, err := runtime.ExecuteScript(
			Script{
				Source: []byte(script),
			},
			Context{
				Interface: runtimeInterface,
				Location:  common.ScriptLocation{},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, cadence.NewInt(3), result)
	})
}

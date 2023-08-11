/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright Dapper Labs, Inc.
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

package interpreter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterpretArrayFunctionEntitlements(t *testing.T) {

	t.Parallel()

	t.Run("mutable reference", func(t *testing.T) {
		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            let array: [String] = ["foo", "bar"]

            fun test() {
                var arrayRef = &array as auth(Mutate) &[String]

                // Public functions
                arrayRef.contains("hello")
                arrayRef.firstIndex(of: "hello")
                arrayRef.slice(from: 1, upTo: 1)
                arrayRef.concat(["hello"])

                // Insertable functions
                arrayRef.append("baz")
                arrayRef.appendAll(["baz"])
                arrayRef.insert(at:0, "baz")

                // Removable functions
                arrayRef.remove(at: 1)
                arrayRef.removeFirst()
                arrayRef.removeLast()
            }
        `)

		_, err := inter.Invoke("test")
		require.NoError(t, err)
	})

	t.Run("non auth reference", func(t *testing.T) {
		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            let array: [String] = ["foo", "bar"]

            fun test() {
                var arrayRef = &array as &[String]

                // Public functions
                arrayRef.contains("hello")
                arrayRef.firstIndex(of: "hello")
                arrayRef.slice(from: 1, upTo: 1)
                arrayRef.concat(["hello"])
            }
	        `)

		_, err := inter.Invoke("test")
		require.NoError(t, err)
	})

	t.Run("insertable reference", func(t *testing.T) {
		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            let array: [String] = ["foo", "bar"]

            fun test() {
                var arrayRef = &array as auth(Insert) &[String]

                // Public functions
                arrayRef.contains("hello")
                arrayRef.firstIndex(of: "hello")
                arrayRef.slice(from: 1, upTo: 1)
                arrayRef.concat(["hello"])

                // Insertable functions
                arrayRef.append("baz")
                arrayRef.appendAll(["baz"])
                arrayRef.insert(at:0, "baz")
            }
	        `)

		_, err := inter.Invoke("test")
		require.NoError(t, err)
	})

	t.Run("removable reference", func(t *testing.T) {
		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            let array: [String] = ["foo", "bar", "baz"]

            fun test() {
                var arrayRef = &array as auth(Remove) &[String]

                // Public functions
                arrayRef.contains("hello")
                arrayRef.firstIndex(of: "hello")
                arrayRef.slice(from: 1, upTo: 1)
                arrayRef.concat(["hello"])

                // Removable functions
                arrayRef.remove(at: 1)
                arrayRef.removeFirst()
                arrayRef.removeLast()
            }
	        `)

		_, err := inter.Invoke("test")
		require.NoError(t, err)
	})
}

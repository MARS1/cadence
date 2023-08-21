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

package checker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/sema"
)

func TestCheckBasicEntitlementDeclaration(t *testing.T) {

	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheck(t, `
			entitlement E
		`)

		assert.NoError(t, err)
		entitlement := checker.Elaboration.EntitlementType("S.test.E")
		assert.Equal(t, "E", entitlement.String())
	})

	t.Run("access(self) access", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			access(self) entitlement E 
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAccessModifierError{}, errs[0])
	})
}

func TestCheckBasicEntitlementMappingDeclaration(t *testing.T) {

	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheck(t, `
			entitlement mapping M {}
		`)

		assert.NoError(t, err)
		entitlement := checker.Elaboration.EntitlementMapType("S.test.M")
		assert.Equal(t, "M", entitlement.String())
	})

	t.Run("with mappings", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheck(t, `
			entitlement A 
			entitlement B
			entitlement C
			entitlement mapping M {
				A -> B
				B -> C
			}
		`)

		assert.NoError(t, err)
		entitlement := checker.Elaboration.EntitlementMapType("S.test.M")
		assert.Equal(t, "M", entitlement.String())
		assert.Equal(t, 2, len(entitlement.Relations))
	})

	t.Run("access(self) access", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			access(self) entitlement mapping M {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAccessModifierError{}, errs[0])
	})
}

func TestCheckBasicEntitlementMappingNonEntitlements(t *testing.T) {

	t.Parallel()

	t.Run("resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement A 
			resource B {}
			entitlement mapping M {
				A -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[1])
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement A 
			struct B {}
			entitlement mapping M {
				A -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[1])
	})

	t.Run("attachment", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement A 
			attachment B for AnyStruct {}
			entitlement mapping M {
				A -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[1])
	})

	t.Run("interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement A 
			resource interface B {}
			entitlement mapping M {
				A -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[1])
	})

	t.Run("contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement B
			contract A {}
			entitlement mapping M {
				A -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[1])
	})

	t.Run("event", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement B
			event A()
			entitlement mapping M {
				A -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[1])
	})

	t.Run("enum", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement B
			enum A: UInt8 {}
			entitlement mapping M {
				A -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[1])
	})

	t.Run("simple type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement B
			entitlement mapping M {
				Int -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[0])
	})

	t.Run("other mapping", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement B
			entitlement mapping A {}
			entitlement mapping M {
				A -> B
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementTypeInMapError{}, errs[0])
	})
}

func TestCheckEntitlementDeclarationNesting(t *testing.T) {
	t.Parallel()
	t.Run("in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement E
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("in contract interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract interface C {
				entitlement E
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("in resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource R {
				entitlement E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in resource interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource interface R {
				entitlement E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in attachment", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			attachment A for AnyStruct {
				entitlement E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct S {
				entitlement E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct interface S {
				entitlement E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in enum", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			enum X: UInt8 {
				entitlement E
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEnumCaseError{}, errs[1])
	})
}

func TestCheckEntitlementMappingDeclarationNesting(t *testing.T) {
	t.Parallel()
	t.Run("in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement mapping M {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("in contract interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract interface C {
				entitlement mapping M {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("in resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource R {
				entitlement mapping M {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in resource interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource interface R {
				entitlement mapping M {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in attachment", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			attachment A for AnyStruct {
				entitlement mapping M {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct S {
				entitlement mapping M {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct interface S {
				entitlement mapping M {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in enum", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			enum X: UInt8 {
				entitlement mapping M {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEnumCaseError{}, errs[1])
	})
}

func TestCheckBasicEntitlementAccess(t *testing.T) {

	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface S {
				access(E) let foo: String
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("multiple entitlements conjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement A
			entitlement B
			entitlement C
			resource interface R {
				access(A, B) let foo: String
				access(B, C) fun bar()
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("multiple entitlements disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement A
			entitlement B
			entitlement C
			resource interface R {
				access(A | B) let foo: String
				access(B | C) fun bar()
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("valid in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement E
				struct interface S {
					access(E) let foo: String
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("valid in contract interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract interface C {
				entitlement E
				struct interface S {
					access(E) let foo: String
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("qualified", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement E
				struct interface S {
					access(E) let foo: String
				}
			}
			resource R {
				access(C.E) fun bar() {}
			}
		`)

		assert.NoError(t, err)
	})
}

func TestCheckBasicEntitlementMappingAccess(t *testing.T) {

	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) let foo: auth(M) &String
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("optional valid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) let foo: auth(M) &String?
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("non-reference field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("non-auth reference field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) let foo: &String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("non-reference container field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) let foo: [String]
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("mismatched entitlement mapping", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			entitlement mapping N {}
			struct interface S {
				access(M) let foo: auth(N) &String
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[1])
	})

	t.Run("mismatched entitlement mapping to set", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			entitlement N
			struct interface S {
				access(M) let foo: auth(N) &String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("function", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) fun foo() 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("accessor function in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			contract interface S {
				access(M) fun foo(): auth(M) &Int 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("accessor function no container", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			access(M) fun foo(): auth(M) &Int {
				return &1 as auth(M) &Int
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[1])
	})

	t.Run("accessor function", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) fun foo(): auth(M) &Int 
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function non mapped return", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement mapping M {}
			struct interface S {
				access(M) fun foo(): auth(X) &Int 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("accessor function non mapped access", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement mapping M {}
			struct interface S {
				access(X) fun foo(): auth(M) &Int 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("accessor function optional", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) fun foo(): auth(M) &Int? 
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with impl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct S {
				access(M) fun foo(): auth(M) &Int {
					return &1 as auth(M) &Int
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with impl wrong mapping", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			entitlement mapping N {}
			struct S {
				access(M) fun foo(): auth(M) &Int {
					return &1 as auth(N) &Int
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})

	t.Run("accessor function with impl subtype", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement mapping M {
				X -> Y
				X -> Z
			}
			struct S {
				access(M) fun foo(): auth(M) &Int {
					return &1 as auth(Y, Z) &Int
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with impl supertype", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement mapping M {
				X -> Y
				X -> Z
			}
			var x: [auth(Y) &Int] = []
			struct S {
				access(M) fun foo(): auth(M) &Int {
					let r =  &1 as auth(M) &Int
					x[0] = r
					return r
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
	})

	t.Run("accessor function with impl invalid cast", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			entitlement mapping M {
				E -> F
			}
			struct S {
				access(M) fun foo(): auth(M) &Int {
					let x = &1 as auth(M) &Int
					// cannot cast, because M may be access(all)
					let y: auth(F) &Int = x
					return y
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(F) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(M) &Int")
	})

	t.Run("accessor function with complex impl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			var x: [AnyStruct] = []
			struct S {
				access(M) fun foo(cond: Bool): auth(M) &Int {
					if(cond) {
						let r = x[0]
						if let ref = x as? auth(M) &Int {
							return ref
						} else {
							return &2 as auth(M) &Int
						}
					} else {
						let r = &3 as auth(M) &Int
						x.append(r)
						return r
					}
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with downcast impl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement mapping M {
				X -> Y
				X -> Z
			}
			struct T {
				access(Y) fun foo() {}
			}
			struct S {
				access(M) fun foo(cond: Bool): auth(M) &T {
					let x = &T() as auth(M) &T
					if let y = x as? auth(Y) &T {
						y.foo()
					} 
					return x
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with no downcast impl", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheckWithOptions(t, `
			entitlement X
			entitlement Y
			entitlement mapping M {
				X -> Y
			}
			struct T {
				access(Y) fun foo() {}
			}
			struct S {
				access(M) fun foo(cond: Bool): auth(M) &T {
					let x = &T() as auth(M) &T
					x.foo()
					return x
				}
			}
		`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.Y"),
				},
				sema.Conjunction,
			),
		)
		// in this case `M` functions like a generic name for an entitlement,
		// so we use `M` as the access for `x` here
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.NewEntitlementMapAccess(checker.Elaboration.EntitlementMapType("S.test.M")),
		)
	})

	t.Run("accessor function with object access impl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement mapping M {
				X -> Y
			}
			struct T {
				access(Y) fun getRef(): auth(Y) &Int {
					return &1 as auth(Y) &Int
				}
			}
			struct S {
				access(M) let t: auth(M) &T
				access(M) fun foo(cond: Bool): auth(M) &Int {
					// success because we have self is fully entitled to the domain of M
					return self.t.getRef() 
				}
				init() {
					self.t = &T() as auth(Y) &T
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with invalid object access impl", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheckWithOptions(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement mapping M {
				X -> Y
			}
			struct T {
				access(Z) fun getRef(): auth(Y) &Int {
					return &1 as auth(Y) &Int
				}
			}
			struct S {
				access(M) let t: auth(M) &T
				access(M) fun foo(cond: Bool): auth(M) &Int {
					// invalid bc we have no Z entitlement
					return self.t.getRef() 
				}
				init() {
					self.t = &T() as auth(Y) &T
				}
			}
		`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.Z"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.Y"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"reference needs entitlement `Z`",
		)
	})

	t.Run("accessor function with mapped object access impl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement mapping M {
				X -> Y
			}
			entitlement mapping N {
				Y -> Z
			}
			struct T {
				access(N) fun getRef(): auth(N) &Int {
					return &1 as auth(N) &Int
				}
			}
			struct S {
				access(M) let t: auth(M) &T
				access(X) fun foo(cond: Bool): auth(Z) &Int {
					return self.t.getRef() 
				}
				init() {
					self.t = &T() as auth(Y) &T
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with composed mapping object access impl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement mapping M {
				X -> Y
			}
			entitlement mapping N {
				Y -> Z
			}
			entitlement mapping NM {
				X -> Z
			}
			struct T {
				access(N) fun getRef(): auth(N) &Int {
					return &1 as auth(N) &Int
				}
			}
			struct S {
				access(M) let t: auth(M) &T
				access(NM) fun foo(cond: Bool): auth(NM) &Int {
					return self.t.getRef() 
				}
				init() {
					self.t = &T() as auth(Y) &T
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with invalid composed mapping object access impl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement Q
			entitlement mapping M {
				X -> Y
			}
			entitlement mapping N {
				Y -> Z
			}
			entitlement mapping NM {
				X -> Q
			}
			struct T {
				access(N) fun getRef(): auth(N) &Int {
					return &1 as auth(N) &Int
				}
			}
			struct S {
				access(M) let t: auth(M) &T
				access(NM) fun foo(cond: Bool): auth(NM) &Int {
					return self.t.getRef() 
				}
				init() {
					self.t = &T() as auth(Y) &T
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(NM) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(Z) &Int")
	})

	t.Run("accessor function with superset composed mapping object access input", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement A 
			entitlement B
			entitlement mapping M {
				X -> Y
				A -> B
			}
			entitlement mapping N {
				Y -> Z
			}
			entitlement mapping NM {
				X -> Z
			}
			struct T {
				access(N) fun getRef(): auth(N) &Int {
					return &1 as auth(N) &Int
				}
			}
			struct S {
				access(M) let t: auth(M) &T
				access(NM) fun foo(cond: Bool): auth(NM) &Int {
					return self.t.getRef() 
				}
				init() {
					self.t = &T() as auth(Y, B) &T
				}
			}`)

		assert.NoError(t, err)
	})

	t.Run("accessor function with composed mapping object access skipped step", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement A 
			entitlement B
			entitlement mapping M {
				X -> Y
				A -> B
			}
			entitlement mapping N {
				Y -> Z
			}
			entitlement mapping NM {
				X -> Z
				A -> B
			}
			struct T {
				access(N) fun getRef(): auth(N) &Int {
					return &1 as auth(N) &Int
				}
			}
			struct S {
				access(M) let t: auth(M) &T
				access(NM) fun foo(cond: Bool): auth(NM) &Int {
					// the B entitlement doesn't pass through the mapping N
					return self.t.getRef() 
				}
				init() {
					self.t = &T() as auth(Y, B) &T
				}
			}`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(NM) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(Z) &Int")
	})

	t.Run("accessor function with composed mapping object access included intermediate step", func(t *testing.T) {

		t.Parallel()

		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement Z
			entitlement A 
			entitlement B
			entitlement mapping M {
				X -> Y
				A -> B
			}
			entitlement mapping N {
				Y -> Z
				B -> B
			}
			entitlement mapping NM {
				X -> Z
				A -> B
			}
			struct T {
				access(N) fun getRef(): auth(N) &Int {
					return &1 as auth(N) &Int
				}
			}
			struct S {
				access(M) let t: auth(M) &T
				access(NM) fun foo(cond: Bool): auth(NM) &Int {
					return self.t.getRef() 
				}
				init() {
					self.t = &T() as auth(Y, B) &T
				}
			}`)

		assert.NoError(t, err)
	})

	t.Run("accessor function array", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) fun foo(): [auth(M) &Int]
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("accessor function with invalid mapped ref arg", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			struct interface S {
				access(M) fun foo(arg: auth(M) &Int): auth(M) &Int 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("multiple mappings conjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {} 
			entitlement mapping N {}
			resource interface R {
				access(M, N) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMultipleMappedEntitlementError{}, errs[0])
	})

	t.Run("multiple mappings conjunction with regular", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {} 
			entitlement N
			resource interface R {
				access(M, N) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMultipleMappedEntitlementError{}, errs[0])
	})

	t.Run("multiple mappings disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {} 
			entitlement mapping N {}
			resource interface R {
				access(M | N) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMultipleMappedEntitlementError{}, errs[0])
	})

	t.Run("multiple mappings disjunction with regular", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement M 
			entitlement mapping N {}
			resource interface R {
				access(M | N) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("valid in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement mapping M {} 
				struct interface S {
					access(M) let foo: auth(M) &String
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("valid in contract interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract interface C {
				entitlement mapping M {} 
				struct interface S {
					access(M) let foo: auth(M) &String
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("qualified", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement mapping M {} 
				struct interface S {
					access(M) let foo: auth(M) &String
				}
			}
			resource interface R {
				access(C.M) let bar: auth(C.M) &String
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("ref array field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			resource interface R {
				access(M) let foo: [auth(M) &Int]
			}
		`)

		assert.NoError(t, err)
	})
}

func TestCheckInvalidEntitlementAccess(t *testing.T) {

	t.Parallel()

	t.Run("invalid variable decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			access(E) var x: String = ""
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid fun decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			access(E) fun foo() {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid contract field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			contract C {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid contract interface field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			contract interface C {
				access(E) fun foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid event", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource I {
				access(E) event Foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[1])
	})

	t.Run("invalid enum case", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			enum X: UInt8 {
				access(E) case red
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAccessModifierError{}, errs[0])
	})

	t.Run("missing entitlement declaration fun", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
	})

	t.Run("missing entitlement declaration field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct interface S {
				access(E) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
	})
}

func TestCheckInvalidEntitlementMappingAuth(t *testing.T) {
	t.Parallel()

	t.Run("invalid variable annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			let x: auth(M) &Int = 3
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})

	t.Run("invalid param annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			fun foo(x: auth(M) &Int) {

			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("invalid return annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			fun foo(): auth(M) &Int {}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
		require.IsType(t, &sema.MissingReturnStatementError{}, errs[1])
	})

	t.Run("invalid ref expr annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			let x = &1 as auth(M) &Int
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("invalid failable annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			let x = &1 as &Int
			let y = x as? auth(M) &Int
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("invalid type param annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			fun foo(x: Capability<auth(M) &Int>) {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("invalid type argument annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement mapping M {}
			let x = authAccount.borrow<auth(M) &Int>(from: /storage/foo)
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("invalid cast annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			let x = &1 as &Int
			let y = x as auth(M) &Int
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("capability field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			resource interface R {
				access(M) let foo: Capability<auth(M) &Int>
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("optional ref field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			resource interface R {
				access(M) let foo: (auth(M) &Int)?
			}
		`)

		// exception made for optional reference fields
		assert.NoError(t, err)
	})

	t.Run("fun ref field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			resource interface R {
				access(M) let foo: fun(auth(M) &Int): auth(M) &Int
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("optional fun ref field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			resource interface R {
				access(M) let foo: fun((auth(M) &Int?))
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("mapped ref unmapped field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			struct interface S {
				access(E) var x: auth(M) &String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("mapped nonref unmapped field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			struct interface S {
				access(E) var x: fun(auth(M) &String): Int
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
	})

	t.Run("mapped field unmapped ref", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			struct interface S {
				access(M) var x: auth(E) &String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("different map", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			entitlement mapping N {
				E -> F
			}
			struct interface S {
				access(M) var x: auth(N) &String
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidMappedAuthorizationOutsideOfFieldError{}, errs[0])
		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[1])
	})
}

func TestCheckInvalidEntitlementMappingAccess(t *testing.T) {

	t.Parallel()

	t.Run("invalid variable decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			access(M) var x: String = ""
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("nonreference field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			resource interface R {
				access(M) let foo: Int
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("optional nonreference field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			resource interface R {
				access(M) let foo: Int?
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("invalid fun decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			access(M) fun foo() {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("invalid contract field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			contract C {
				access(M) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("invalid contract interface field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			contract interface C {
				access(M) fun foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[0])
	})

	t.Run("invalid event", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			resource I {
				access(M) event Foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
		require.IsType(t, &sema.InvalidMappedEntitlementMemberError{}, errs[1])
	})

	t.Run("invalid enum case", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping M {}
			enum X: UInt8 {
				access(M) case red
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAccessModifierError{}, errs[0])
	})

	t.Run("missing entitlement mapping declaration fun", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource R {
				access(M) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
	})

	t.Run("missing entitlement mapping declaration field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct interface S {
				access(M) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
	})
}

func TestCheckNonEntitlementAccess(t *testing.T) {

	t.Parallel()

	t.Run("resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("resource interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource interface E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("attachment", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			attachment E for AnyStruct {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("struct interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("event", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			event E()
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("contract interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract interface E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("enum", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			enum E: UInt8 {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})
}

func TestCheckEntitlementInheritance(t *testing.T) {

	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				access(E) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("valid interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(E) fun foo() 
			}
			struct interface S: I {
				access(E) fun foo() 
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("valid mapped", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		 	entitlement X
			entitlement Y
			entitlement mapping M {
				X -> Y
			}
			struct interface I {
				access(M) let x: auth(M) &String
			}
			struct S: I {
				access(M) let x: auth(M) &String
				init() {
					self.x = &"foo" as auth(Y) &String
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("valid mapped interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		 	entitlement X
			entitlement Y
			entitlement mapping M {
				X -> Y
			}
			struct interface I {
				access(M) let x: auth(M) &String
			}
			struct interface S: I {
				access(M) let x: auth(M) &String
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("mismatched mapped", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement mapping M {}
			entitlement mapping N {
				X -> Y
			}
			struct interface I {
				access(M) let x: auth(M) &String
			}
			struct S: I {
				access(N) let x: auth(N) &String
				init() {
					self.x = &"foo" as auth(Y) &String
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("mismatched mapped interfaces", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement mapping M {}
			entitlement mapping N {
				X -> Y
			}
			struct interface I {
				access(M) let x: auth(M) &String
			}
			struct interface S: I {
				access(N) let x: auth(N) &String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InterfaceMemberConflictError{}, errs[0])
	})

	t.Run("access(all) subtyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(all) fun foo() 
			}
			struct S: I {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access(all) subtyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(all) var x: String
			}
			struct S: I {
				access(E) var x: String
				init() {
					self.x = ""
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access(all) supertying invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				access(all) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access(all) supertyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(E) var x: String
			}
			struct S: I {
				access(all) var x: String
				init() {
					self.x = ""
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access contract subtyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(contract) fun foo() 
			}
			struct S: I {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access account subtyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(account) fun foo() 
			}
			struct S: I {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access account supertying invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				access(account) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access contract supertying invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				access(contract) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access(self) supertying invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				access(self) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("invalid map subtype with regular conjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			struct interface I {
				access(E, F) var x: auth(E, F) &String
			}
			struct S: I {
				access(M) var x: auth(M) &String 

				init() {
					self.x = &"foo" as auth(F) &String
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("invalid map supertype with regular conjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			struct interface I {
				access(M) var x: auth(M) &String
			}
			struct S: I {
				access(E, F) var x: auth(E, F) &String 

				init() {
					self.x = &"foo" as auth(E, F) &String
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("invalid map subtype with regular disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			struct interface I {
				access(E | F) var x: auth(E | F) &String
			}
			struct S: I {
				access(M) var x: auth(M) &String 

				init() {
					self.x = &"foo" as auth(F) &String
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("invalid map supertype with regular disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			struct interface I {
				access(M) var x: auth(M) &String
			}
			struct S: I {
				access(E | F) var x: auth(E | F) &String 

				init() {
					self.x = &"foo" as auth(F) &String
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("expanded entitlements valid in disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(F) fun foo() 
			}
			struct S: I, J {
				access(E | F) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("more expanded entitlements valid in disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			entitlement G
			struct interface I {
				access(E | G) fun foo() 
			}
			struct S: I {
				access(E | F | G) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("reduced entitlements valid with conjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			struct interface I {
				access(E, G) fun foo() 
			}
			struct interface J {
				access(E, F) fun foo() 
			}
			struct S: I, J {
				access(E) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("more reduced entitlements valid with conjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			struct interface I {
				access(E, F, G) fun foo() 
			}
			struct S: I {
				access(E, F) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("expanded entitlements invalid in conjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(F) fun foo() 
			}
			struct S: I, J {
				access(E, F) fun foo() {}
			}
		`)

		// this conforms to neither I nor J
		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
		require.IsType(t, &sema.ConformanceError{}, errs[1])
	})

	t.Run("more expanded entitlements invalid in conjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			entitlement G
			struct interface I {
				access(E, F) fun foo() 
			}
			struct S: I {
				access(E, F, G) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("expanded entitlements invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(F) fun foo() 
			}
			struct S: I, J {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("reduced entitlements invalid with disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(E | F) fun foo() 
			}
			struct S: I, J {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("more reduced entitlements invalid with disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			struct interface I {
				access(E | F | G) fun foo() 
			}
			struct S: I {
				access(E | G) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("overlapped entitlements invalid with disjunction", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			struct interface J {
				access(E | F) fun foo() 
			}
			struct S: J {
				access(E | G) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("overlapped entitlements invalid with disjunction/conjunction subtype", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			struct interface J {
				access(E | F) fun foo() 
			}
			struct S: J {
				access(E, G) fun foo() {}
			}
		`)

		// implementation is more specific because it requires both, but interface only guarantees one
		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("disjunction/conjunction subtype valid when sets are the same", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface J {
				access(E | E) fun foo() 
			}
			struct S: J {
				access(E, E) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("overlapped entitlements valid with conjunction/disjunction subtype", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			struct interface J {
				access(E, F) fun foo() 
			}
			struct S: J {
				access(E | G) fun foo() {}
			}
		`)

		// implementation is less specific because it only requires one, but interface guarantees both
		assert.NoError(t, err)
	})

	t.Run("different entitlements invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			entitlement G 
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(F) fun foo() 
			}
			struct S: I, J {
				access(E | G) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("default function entitlements", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement mapping M {
				E -> F
			}
			entitlement G
			struct interface I {
				access(M) fun foo(): auth(M) &Int {
					return &1 as auth(M) &Int
				}
			}
			struct S: I {}
			fun test() {
				let s = S()
				let ref = &s as auth(E) &S
				let i: auth(F) &Int = s.foo()
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("attachment default function entitlements", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			entitlement mapping M {
				E -> F
			}
			entitlement mapping N {
				G -> E
			}
			struct interface I {
				access(M) fun foo(): auth(M) &Int {
					return &1 as auth(M) &Int
				}
			}
			struct S {}
			access(N) attachment A for S: I {}
			fun test() {
				let s = attach A() to S()
				let ref = &s as auth(G) &S
				let i: auth(F) &Int = s[A]!.foo()
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("attachment inherited default function entitlements", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			entitlement mapping M {
				E -> F
			}
			entitlement mapping N {
				G -> E
			}
			struct interface I {
				access(M) fun foo(): auth(M) &Int {
					return &1 as auth(M) &Int
				}
			}
			struct interface I2: I {}
			struct S {}
			access(N) attachment A for S: I2 {}
			fun test() {
				let s = attach A() to S()
				let ref = &s as auth(G) &S
				let i: auth(F) &Int = s[A]!.foo()
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("attachment default function entitlements no attachment mapping", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F 
			entitlement G
			entitlement mapping M {
				E -> F
			}
			entitlement mapping N {
				G -> E
			}
			struct interface I {
				access(M) fun foo(): auth(M) &Int {
					return &1 as auth(M) &Int
				}
			}
			struct S {}
			attachment A for S: I {}
			fun test() {
				let s = attach A() to S()
				let ref = &s as auth(G) &S
				let i: auth(F) &Int = s[A]!.foo() // mismatch
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// because A is declared with no mapping, all its references are unentitled
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
	})
}

func TestCheckEntitlementTypeAnnotation(t *testing.T) {

	t.Parallel()

	t.Run("invalid local annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			let x: E = ""
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})

	t.Run("invalid param annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			access(all) fun foo(e: E) {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid return annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource interface I {
				access(all) fun foo(): E 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid field annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource interface I {
				let e: E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid conformance annotation", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource R: E {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidConformanceError{}, errs[0])
	})

	t.Run("invalid array annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource interface I {
				let e: [E]
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid fun annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource interface I {
				let e: (fun (E): Void)
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid enum conformance", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			enum X: E {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEnumRawTypeError{}, errs[0])
	})

	t.Run("invalid dict annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource interface I {
				let e: {E: E}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		// key
		require.IsType(t, &sema.InvalidDictionaryKeyTypeError{}, errs[0])
		// value
		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[1])
	})

	t.Run("invalid fun annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource interface I {
				let e: (fun (E): Void)
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("runtype type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			let e = Type<E>()
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("type arg", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E
			let e = authAccount.load<E>(from: /storage/foo)
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		// entitlements are not storable either
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})

	t.Run("intersection", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E
			resource interface I {
				let e: {E}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidIntersectedTypeError{}, errs[0])
		require.IsType(t, &sema.AmbiguousIntersectionTypeError{}, errs[1])
	})

	t.Run("reference", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E
			resource interface I {
				let e: &E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("capability", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E
			resource interface I {
				let e: Capability<&E>
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[1])
	})

	t.Run("optional", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E
			resource interface I {
				let e: E?
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})
}

func TestCheckEntitlementMappingTypeAnnotation(t *testing.T) {

	t.Parallel()

	t.Run("invalid local annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			let x: E = ""
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})

	t.Run("invalid param annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			access(all) fun foo(e: E) {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid return annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			resource interface I {
				access(all) fun foo(): E 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid field annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			resource interface I {
				let e: E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid conformance annotation", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			resource R: E {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidConformanceError{}, errs[0])
	})

	t.Run("invalid array annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			resource interface I {
				let e: [E]
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid fun annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			resource interface I {
				let e: (fun (E): Void)
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid enum conformance", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			enum X: E {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEnumRawTypeError{}, errs[0])
	})

	t.Run("invalid dict annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			resource interface I {
				let e: {E: E}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		// key
		require.IsType(t, &sema.InvalidDictionaryKeyTypeError{}, errs[0])
		// value
		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[1])
	})

	t.Run("runtype type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			let e = Type<E>()
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("type arg", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement mapping E {}
			let e = authAccount.load<E>(from: /storage/foo)
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		// entitlements are not storable either
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})

	t.Run("intersection", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement mapping E {}
			resource interface I {
				let e: {E}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidIntersectedTypeError{}, errs[0])
		require.IsType(t, &sema.AmbiguousIntersectionTypeError{}, errs[1])
	})

	t.Run("reference", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement mapping E {}
			resource interface I {
				let e: &E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("capability", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement mapping E {}
			resource interface I {
				let e: Capability<&E>
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[1])
	})

	t.Run("optional", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement mapping E {}
			resource interface I {
				let e: E?
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})
}

func TestCheckAttachmentEntitlementAccessAnnotation(t *testing.T) {

	t.Parallel()
	t.Run("mapping allowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement mapping E {}
			access(E) attachment A for AnyStruct {}
		`)

		assert.NoError(t, err)
	})

	t.Run("entitlement set not allowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			access(E, F) attachment A for AnyStruct {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("mapping allowed in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		contract C {
			entitlement X 
			entitlement Y
			entitlement mapping E {
				X -> Y
			}	
			access(E) attachment A for AnyStruct {
				access(Y) fun foo() {}
			}
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("entitlement set not allowed in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		contract C {
			entitlement E
			access(E) attachment A for AnyStruct {}
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

}

func TestCheckEntitlementSetAccess(t *testing.T) {

	t.Parallel()

	runTest := func(refType string, memberName string, valid bool) {
		t.Run(fmt.Sprintf("%s on %s", memberName, refType), func(t *testing.T) {
			t.Parallel()
			_, err := ParseAndCheckAccount(t, fmt.Sprintf(`
				entitlement X
				entitlement Y
				entitlement Z

				struct R {
					access(all) fun p() {}

					access(X) fun x() {}
					access(Y) fun y() {}
					access(Z) fun z() {}

					access(X, Y) fun xy() {}
					access(Y, Z) fun yz() {}
					access(X, Z) fun xz() {}
					
					access(X, Y, Z) fun xyz() {}

					access(X | Y) fun xyOr() {}
					access(Y | Z) fun yzOr() {}
					access(X | Z) fun xzOr() {}

					access(X | Y | Z) fun xyzOr() {}

					access(self) fun private() {}
					access(contract) fun c() {}
					access(account) fun a() {}
				}

				fun test(ref: %s) {
					ref.%s()
				}
			`, refType, memberName))

			if valid {
				assert.NoError(t, err)
			} else {
				errs := RequireCheckerErrors(t, err, 1)

				require.IsType(t, &sema.InvalidAccessError{}, errs[0])
			}
		})
	}

	tests := []struct {
		refType    string
		memberName string
		valid      bool
	}{
		{"&R", "p", true},
		{"&R", "x", false},
		{"&R", "xy", false},
		{"&R", "xyz", false},
		{"&R", "xyOr", false},
		{"&R", "xyzOr", false},
		{"&R", "private", false},
		{"&R", "a", true},
		{"&R", "c", false},

		{"auth(X) &R", "p", true},
		{"auth(X) &R", "x", true},
		{"auth(X) &R", "y", false},
		{"auth(X) &R", "xy", false},
		{"auth(X) &R", "xyz", false},
		{"auth(X) &R", "xyOr", true},
		{"auth(X) &R", "xyzOr", true},
		{"auth(X) &R", "private", false},
		{"auth(X) &R", "a", true},
		{"auth(X) &R", "c", false},

		{"auth(X, Y) &R", "p", true},
		{"auth(X, Y) &R", "x", true},
		{"auth(X, Y) &R", "y", true},
		{"auth(X, Y) &R", "xy", true},
		{"auth(X, Y) &R", "xyz", false},
		{"auth(X, Y) &R", "xyOr", true},
		{"auth(X, Y) &R", "xyzOr", true},
		{"auth(X, Y) &R", "private", false},
		{"auth(X, Y) &R", "a", true},
		{"auth(X, Y) &R", "c", false},

		{"auth(X, Y, Z) &R", "p", true},
		{"auth(X, Y, Z) &R", "x", true},
		{"auth(X, Y, Z) &R", "y", true},
		{"auth(X, Y, Z) &R", "z", true},
		{"auth(X, Y, Z) &R", "xy", true},
		{"auth(X, Y, Z) &R", "xyz", true},
		{"auth(X, Y, Z) &R", "xyOr", true},
		{"auth(X, Y, Z) &R", "xyzOr", true},
		{"auth(X, Y, Z) &R", "private", false},
		{"auth(X, Y, Z) &R", "a", true},
		{"auth(X, Y, Z) &R", "c", false},

		{"auth(X | Y) &R", "p", true},
		{"auth(X | Y) &R", "x", false},
		{"auth(X | Y) &R", "y", false},
		{"auth(X | Y) &R", "xy", false},
		{"auth(X | Y) &R", "xyz", false},
		{"auth(X | Y) &R", "xyOr", true},
		{"auth(X | Y) &R", "xzOr", false},
		{"auth(X | Y) &R", "yzOr", false},
		{"auth(X | Y) &R", "xyzOr", true},
		{"auth(X | Y) &R", "private", false},
		{"auth(X | Y) &R", "a", true},
		{"auth(X | Y) &R", "c", false},

		{"auth(X | Y | Z) &R", "p", true},
		{"auth(X | Y | Z) &R", "x", false},
		{"auth(X | Y | Z) &R", "y", false},
		{"auth(X | Y | Z) &R", "xy", false},
		{"auth(X | Y | Z) &R", "xyz", false},
		{"auth(X | Y | Z) &R", "xyOr", false},
		{"auth(X | Y | Z) &R", "xzOr", false},
		{"auth(X | Y | Z) &R", "yzOr", false},
		{"auth(X | Y | Z) &R", "xyzOr", true},
		{"auth(X | Y | Z) &R", "private", false},
		{"auth(X | Y | Z) &R", "a", true},
		{"auth(X | Y | Z) &R", "c", false},

		{"R", "p", true},
		{"R", "x", true},
		{"R", "y", true},
		{"R", "xy", true},
		{"R", "xyz", true},
		{"R", "xyOr", true},
		{"R", "xzOr", true},
		{"R", "yzOr", true},
		{"R", "xyzOr", true},
		{"R", "private", false},
		{"R", "a", true},
		{"R", "c", false},
	}

	for _, test := range tests {
		runTest(test.refType, test.memberName, test.valid)
	}

}

func TestCheckEntitlementMapAccess(t *testing.T) {

	t.Parallel()
	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct Q {
			access(Y) fun foo() {}
		}
		struct interface S {
			access(M) let x: auth(M) &Q
		}
		fun foo(s: auth(X) &{S}) {
			s.x.foo()
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("basic with optional access", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct Q {
			access(Y) fun foo() {}
		}
		struct interface S {
			access(M) let x: auth(M) &Q
		}
		fun foo(s: auth(X) &{S}?) {
			s?.x?.foo()
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("basic with optional access return", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct Q {}
		struct interface S {
			access(M) let x: auth(M) &Q
		}
		fun foo(s: auth(X) &{S}?): auth(Y) &Q? {
			return s?.x
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("basic with optional full entitled map", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y
		entitlement E
		entitlement F
		entitlement mapping M {
			X -> Y
			E -> F
		}
		struct S {
			access(M) let foo: auth(M) &Int
			init() {
				self.foo = &3 as auth(F, Y) &Int
			}
		}
		fun test(): &Int {
			let s: S? = S()
			let i: auth(F, Y) &Int? = s?.foo
			return i!
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("basic with optional partial map", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y
		entitlement E
		entitlement F
		entitlement mapping M {
			X -> Y
			E -> F
		}
		struct S {
			access(M) let foo: auth(M) &Int
			init() {
				self.foo = &3 as auth(F, Y) &Int
			}
		}
		fun test(): &Int {
			let s = S()
			let ref = &s as auth(X) &S?
			let i: auth(F, Y) &Int? = ref?.foo
			return i!
		}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		typeMismatchError := &sema.TypeMismatchError{}
		require.ErrorAs(t, errs[0], &typeMismatchError)
		require.Equal(t, typeMismatchError.ExpectedType.QualifiedString(), "S?")
		require.Equal(t, typeMismatchError.ActualType.QualifiedString(), "S")

		require.ErrorAs(t, errs[1], &typeMismatchError)
		require.Equal(t, typeMismatchError.ExpectedType.QualifiedString(), "auth(F, Y) &Int?")
		require.Equal(t, typeMismatchError.ActualType.QualifiedString(), "auth(Y) &Int?")
	})

	t.Run("basic with optional function call return invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement E
		entitlement F 
		entitlement mapping M {
			X -> Y
			E -> F
		}
		struct S {
			access(M) fun foo(): auth(M) &Int {
				return &1 as auth(M) &Int
			}
		}
		fun foo(s: auth(X) &S?): auth(X, Y) &Int? {
			return s?.foo()
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(X, Y) &Int?")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(Y) &Int?")
	})

	t.Run("multiple outputs", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement E
		entitlement F
		entitlement mapping M {
			X -> Y
			E -> F
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(X | E) &{S}) {
			let x: auth(Y | F) &Int = ref.x
			let x2: auth(Y, F) &Int = ref.x
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(Y, F) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(Y | F) &Int")
	})

	t.Run("optional", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct interface S {
			access(M) let x: auth(M) &Int?
		}
		fun foo(ref: auth(X) &{S}) {
			let x: auth(Y) &Int? = ref.x
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("do not retain entitlements", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(X) &{S}) {
			let x: auth(X) &Int = ref.x
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// X is not retained in the entitlements for ref
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(X) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(Y) &Int")
	})

	t.Run("different views", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement A
		entitlement B
		entitlement mapping M {
			X -> Y
			A -> B
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(A) &{S}) {
			let x: auth(Y) &Int = ref.x
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// access gives B, not Y
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(Y) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(B) &Int")
	})

	t.Run("safe disjoint", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement A
		entitlement B
		entitlement mapping M {
			X -> Y
			A -> B
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(A | X) &{S}) {
			let x: auth(B | Y) &Int = ref.x
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("unrepresentable disjoint", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement A
		entitlement B
		entitlement C
		entitlement mapping M {
			X -> Y
			X -> C
			A -> B
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(A | X) &{S}) {
			let x = ref.x
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.UnrepresentableEntitlementMapOutputError{}, errs[0])
	})

	t.Run("unrepresentable disjoint with dedup", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement A
		entitlement B
		entitlement mapping M {
			X -> Y
			X -> B
			A -> B
			A -> Y
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(A | X) &{S}) {
			let x = ref.x
		}
		`)

		// theoretically this should be allowed, because ((Y & B) | (Y & B)) simplifies to
		// just (Y & B), but this would require us to build in a simplifier for boolean expressions,
		// which is a lot of work for an edge case that is very unlikely to come up
		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.UnrepresentableEntitlementMapOutputError{}, errs[0])
	})

	t.Run("multiple output", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(X) &{S}) {
			let x: auth(Y, Z) &Int = ref.x
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("unmapped entitlements do not pass through map", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement D
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(D) &{S}) {
			let x1: auth(D) &Int = ref.x
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// access results in access(all) access because D is not mapped
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(D) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&Int")
	})

	t.Run("multiple output with upcasting", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(X) &{S}) {
			let x: auth(Z) &Int = ref.x
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("multiple inputs", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement A
		entitlement B 
		entitlement C
		entitlement X
		entitlement Y
		entitlement Z
		entitlement mapping M {
			A -> C
			B -> C
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref1: auth(A) &{S}, ref2: auth(B) &{S}) {
			let x1: auth(C) &Int = ref1.x
			let x2: auth(C) &Int = ref2.x
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("multiple inputs and outputs", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement A
		entitlement B 
		entitlement C
		entitlement X
		entitlement Y
		entitlement Z
		entitlement mapping M {
			A -> B
			A -> C
			X -> Y
			X -> Z
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(A, X) &{S}) {
			let x: auth(B, C, Y, Z) &Int = ref.x
			let upRef = ref as auth(A) &{S}
			let upX: auth(B, C) &Int = upRef.x
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("multiple inputs and outputs mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement A
		entitlement B 
		entitlement C
		entitlement X
		entitlement Y
		entitlement Z
		entitlement mapping M {
			A -> B
			A -> C
			X -> Y
			X -> Z
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: auth(A, X) &{S}) {
			let upRef = ref as auth(A) &{S}
			let upX: auth(X, Y) &Int = upRef.x
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// access gives B & C, not X & Y
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(X, Y) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(B, C) &Int")
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: &{S}) {
			let x: &Int = ref.x
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("unauthorized downcast", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct interface S {
			access(M) let x: auth(M) &Int
		}
		fun foo(ref: &{S}) {
			let x: auth(Y) &Int = ref.x
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// result is not authorized
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(Y) &Int")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&Int")
	})

	t.Run("basic with init", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {
			access(M) let x: auth(M) &Int
			init() {
				self.x = &1 as auth(Y, Z) &Int
			}
		}
		let ref = &S() as auth(X) &S
		let x = ref.x
		`)

		assert.NoError(t, err)
	})

	t.Run("basic with update", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {
			access(M) var x: auth(M) &Int
			init() {
				self.x = &1 as auth(Y, Z) &Int
			}
			fun updateX(x: auth(Y, Z) &Int) {
				self.x = x
			}
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("basic with update error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {
			access(M) var x: auth(M) &Int
			init() {
				self.x = &1 as auth(Y, Z) &Int
			}
			fun updateX(x: auth(Z) &Int) {
				self.x = x
			}
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// init of map needs full authorization of codomain
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
	})

	t.Run("basic with unauthorized init", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {
			access(M) let x: auth(M) &Int
			init() {
				self.x = &1 as &Int
			}
		}
		let ref = &S() as auth(X) &S
		let x = ref.x
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// init of map needs full authorization of codomain
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
	})

	t.Run("basic with underauthorized init", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {
			access(M) let x: auth(M) &Int
			init() {
				self.x = &1 as auth(Y) &Int
			}
		}
		let ref = &S() as auth(X) &S
		let x = ref.x
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// init of map needs full authorization of codomain
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
	})

	t.Run("basic with underauthorized disjunction init", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {
			access(M) let x: auth(M) &Int
			init() {
				self.x = (&1 as auth(Y) &Int) as auth(Y | Z) &Int
			}
		}
		let ref = &S() as auth(X) &S
		let x = ref.x
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// init of map needs full authorization of codomain
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
	})

	t.Run("basic with non-reference init", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {
			access(M) let x: auth(M) &Int
			init() {
				self.x = 1
			}
		}
		let ref = &S() as auth(X) &S
		let x = ref.x
		`)

		errs := RequireCheckerErrors(t, err, 1)

		// init of map needs full authorization of codomain
		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
	})
}

func TestCheckAttachmentEntitlements(t *testing.T) {

	t.Parallel()
	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {}
		access(M) attachment A for S {
			access(Y) fun entitled() {
				let a: auth(Y) &A = self 
				let b: &S = base 
			} 
			access(all) fun unentitled() {
				let a: auth(X, Y) &A = self // err
				let b: auth(X) &S = base // err
			}
		}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(X, Y) &A")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(Y) &A")
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
		require.Equal(t, errs[1].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(X) &S")
		require.Equal(t, errs[1].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&S")
	})

	t.Run("base type with too few requirements", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {}
		access(M) attachment A for S {
			require entitlement X
			access(all) fun unentitled() {
				let b: &S = base 
			} 
			access(all) fun entitled() {
				let b: auth(X, Y) &S = base
			}
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(X, Y) &S")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(X) &S")
	})

	t.Run("base type with no requirements", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {}
		access(M) attachment A for S {
			access(all) fun unentitled() {
				let b: &S = base 
			} 
			access(all) fun entitled() {
				let b: auth(X) &S = base
			}
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(X) &S")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&S")
	})

	t.Run("base type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {}
		access(M) attachment A for S {
			require entitlement X
			require entitlement Y
			access(all) fun unentitled() {
				let b: &S = base 
			} 
			access(all) fun entitled() {
				let b: auth(X, Y) &S = base
			}
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("multiple mappings", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement E
		entitlement F
		entitlement mapping M {
			X -> Y
			E -> F
		}
		struct S {
			access(E, X) fun foo() {}
		}
		access(M) attachment A for S {
			access(F, Y) fun entitled() {
				let a: auth(F, Y) &A = self 
			} 
			access(all) fun unentitled() {
				let a: auth(F, Y, E) &A = self // err
			}
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(F, Y, E) &A")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(Y, F) &A")
	})

	t.Run("missing in codomain", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement E
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {}
		access(M) attachment A for S {
			access(E) fun entitled() {} 
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAttachmentEntitlementError{}, errs[0])
		require.Equal(t, errs[0].(*sema.InvalidAttachmentEntitlementError).InvalidEntitlement.QualifiedString(), "E")
	})

	t.Run("missing in codomain in set", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement E
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {}
		access(M) attachment A for S {
			access(Y | E | Z) fun entitled() {} 
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAttachmentEntitlementError{}, errs[0])
		require.Equal(t, errs[0].(*sema.InvalidAttachmentEntitlementError).InvalidEntitlement.QualifiedString(), "E")
	})

	t.Run("multiple missing in codomain", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement E
		entitlement F
		entitlement mapping M {
			E -> F
		}
		struct S {}
		access(M) attachment A for S {
			access(F, X, E) fun entitled() {} 
		}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidAttachmentEntitlementError{}, errs[0])
		require.Equal(t, errs[0].(*sema.InvalidAttachmentEntitlementError).InvalidEntitlement.QualifiedString(), "X")
		require.IsType(t, &sema.InvalidAttachmentEntitlementError{}, errs[1])
		require.Equal(t, errs[1].(*sema.InvalidAttachmentEntitlementError).InvalidEntitlement.QualifiedString(), "E")
	})

	t.Run("mapped field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {
			access(Y) fun foo() {}
		}
		access(M) attachment A for S {
			access(M) let x: auth(M) &S
			init() {
				self.x = &S() as auth(Y) &S
			}
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("access(all) decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		struct S {}
		attachment A for S {
			access(Y) fun entitled() {} 
			access(Y) let entitledField: Int
			access(all) fun unentitled() {
				let a: auth(Y) &A = self // err
				let b: auth(X) &S = base // err
			}
			init() {
				self.entitledField = 3
			}
		}
		`)

		errs := RequireCheckerErrors(t, err, 4)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(Y) &A")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&A")
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
		require.Equal(t, errs[1].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(X) &S")
		require.Equal(t, errs[1].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&S")
		require.IsType(t, &sema.InvalidAttachmentEntitlementError{}, errs[2])
		require.IsType(t, &sema.InvalidAttachmentEntitlementError{}, errs[3])
	})

	t.Run("non mapped entitlement decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		struct S {}
		access(X) attachment A for S {
			
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

}

func TestCheckAttachmentAccessEntitlements(t *testing.T) {

	t.Parallel()
	t.Run("basic owned fully entitled", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z 
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct S {}
		access(M) attachment A for S {
			access(Y, Z) fun foo() {}
		}
		let s = attach A() to S()
		let a: auth(Y, Z) &A = s[A]!
		`)

		assert.NoError(t, err)
	})

	t.Run("basic owned intersection fully entitled", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z 
		entitlement mapping M {
			X -> Y
			X -> Z
		}
		struct interface I {}
		struct S: I {}
		access(M) attachment A for I {
			access(Y, Z) fun foo() {}
		}
		let s: {I} = attach A() to S()
		let a: auth(Y, Z) &A = s[A]!
		`)

		assert.NoError(t, err)
	})

	t.Run("basic reference mapping", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement E
		entitlement F
		entitlement mapping M {
			X -> Y
			E -> F
		}
		struct S {
			access(X, E) fun foo() {}
		}
		access(M) attachment A for S {
			access(Y, F) fun foo() {}
		}
		let s = attach A() to S()
		let yRef = &s as auth(X) &S
		let fRef = &s as auth(E) &S
		let bothRef = &s as auth(X, E) &S
		let a1: auth(Y) &A = yRef[A]!
		let a2: auth(F) &A = fRef[A]!
		let a3: auth(F) &A = yRef[A]! // err
		let a4: auth(Y) &A = fRef[A]! // err
		let a5: auth(Y, F) &A = bothRef[A]!
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(F) &A?")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(Y) &A?")

		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
		require.Equal(t, errs[1].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(Y) &A?")
		require.Equal(t, errs[1].(*sema.TypeMismatchError).ActualType.QualifiedString(), "auth(F) &A?")
	})

	t.Run("access(all) access entitled attachment", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {}
		access(M) attachment A for S {
			access(Y) fun foo() {}
		}
		let s = attach A() to S()
		let ref = &s as &S
		let a1: auth(Y) &A = ref[A]!
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(Y) &A?")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&A?")
	})

	t.Run("entitled access access(all) attachment", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {
			access(X) fun foo() {}
		}
		access(all) attachment A for S {
			access(all) fun foo() {}
		}
		let s = attach A() to S()
		let ref = &s as auth(X) &S
		let a1: auth(Y) &A = ref[A]!
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(Y) &A?")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&A?")
	})

	t.Run("access(all) access access(all) attachment", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement mapping M {
			X -> Y
		}
		struct S {}
		access(all) attachment A for S {}
		let s = attach A() to S()
		let ref = &s as &S
		let a1: auth(Y) &A = ref[A]!
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.TypeMismatchError{}, errs[0])
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ExpectedType.QualifiedString(), "auth(Y) &A?")
		require.Equal(t, errs[0].(*sema.TypeMismatchError).ActualType.QualifiedString(), "&A?")
	})

	t.Run("unrepresentable access mapping", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y 
		entitlement Z
		entitlement E
		entitlement F
		entitlement G
		entitlement mapping M {
			X -> Y
			X -> Z
			E -> F
			E -> G
		}
		struct S {
			access(X, E) fun foo() {}
		}
		access(M) attachment A for S {
			access(Y, Z, F, G) fun foo() {}
		}
		let s = attach A() to S()
		let ref = (&s as auth(X) &S) as auth(X | E) &S
		let a1 = ref[A]!
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.UnrepresentableEntitlementMapOutputError{}, errs[0])
		require.IsType(t, &sema.InvalidTypeIndexingError{}, errs[1])
	})
}

func TestCheckEntitlementConditions(t *testing.T) {
	t.Parallel()

	t.Run("use of function on owned value", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		struct S {
			view access(X) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: S) {
			pre {
				r.foo(): ""
			} 
			post {
				r.foo(): ""
			}
			r.foo()
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("use of function on entitled referenced value", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		struct S {
			view access(X) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: auth(X) &S) {
			pre {
				r.foo(): ""
			} 
			post {
				r.foo(): ""
			}
			r.foo()
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("use of function on unentitled referenced value", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheckWithOptions(t, `
		entitlement X
		struct S {
			view access(X) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: &S) {
			pre {
				r.foo(): ""
			} 
			post {
				r.foo(): ""
			}
			r.foo()
		}
		`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 3)
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.X"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.UnauthorizedAccess,
		)
		require.IsType(t, &sema.InvalidAccessError{}, errs[1])
		require.IsType(t, &sema.InvalidAccessError{}, errs[2])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"reference needs entitlement `X`",
		)
	})

	t.Run("result value usage struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		struct S {
			view access(X) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: S): S {
			post {
				result.foo(): ""
			}
			return r
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("result value usage reference", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheckWithOptions(t, `
		entitlement X
		struct S {
			view access(X) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: S): &S {
			post {
				result.foo(): ""
			}
			return &r as auth(X) &S
		}
		`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.X"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.UnauthorizedAccess,
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"reference needs entitlement `X`",
		)
	})

	t.Run("result value usage reference authorized", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		struct S {
			view access(X) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: S): auth(X) &S {
			post {
				result.foo(): ""
			}
			return &r as auth(X) &S
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("result value usage resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y
		resource R {
			view access(X) fun foo(): Bool {
				return true
			}
			view access(X, Y) fun bar(): Bool {
				return true
			}
		}
		fun bar(r: @R): @R {
			post {
				result.foo(): ""
				result.bar(): ""
			}
			return <-r
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("optional result value usage resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y
		resource R {
			view access(X) fun foo(): Bool {
				return true
			}
			view access(X, Y) fun bar(): Bool {
				return true
			}
		}
		fun bar(r: @R): @R? {
			post {
				result?.foo()!: ""
				result?.bar()!: ""
			}
			return <-r
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("result value inherited entitlement resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y
		resource interface I {
			access(X, Y) view fun foo(): Bool {
				return true
			}
		}
		resource R: I {}
		fun bar(r: @R): @R {
			post {
				result.foo(): ""
			}
			return <-r
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("result value usage unentitled resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		resource R {
			view fun foo(): Bool {
				return true
			}
		}
		fun bar(r: @R): @R {
			post {
				result.foo(): ""
			}
			return <-r
		}
		`)

		assert.NoError(t, err)
	})

	t.Run("result value usage, variable-sized resource array", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
			resource R {}

			fun foo(r: @[R]): @[R] {
				post {
					bar(result): ""
				}
				return <-r
			}

			// 'result' variable should have all the entitlements available for arrays.
			view fun bar(_ r: auth(Mutate, Insert, Remove) &[R]): Bool {
				return true
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("result value usage, constant-sized resource array", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
			resource R {}

			fun foo(r: @[R; 5]): @[R; 5] {
				post {
					bar(result): ""
				}
				return <-r
			}

			// 'result' variable should have all the entitlements available for arrays.
			view fun bar(_ r: auth(Mutate, Insert, Remove) &[R; 5]): Bool {
				return true
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("result value usage, resource dictionary", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
			resource R {}

			fun foo(r: @{String:R}): @{String:R} {
				post {
					bar(result): ""
				}
				return <-r
			}

			// 'result' variable should have all the entitlements available for dictionaries.
			view fun bar(_ r: auth(Mutate, Insert, Remove) &{String:R}): Bool {
				return true
			}
		`)

		assert.NoError(t, err)
	})
}

func TestCheckEntitledWriteAndMutateNotAllowed(t *testing.T) {

	t.Parallel()

	t.Run("basic owned", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct S {
				access(E) var x: Int
				init() {
					self.x = 1
				}
			}
			fun foo() {
				let s = S()
				s.x = 3
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAssignmentAccessError{}, errs[0])
	})

	t.Run("basic authorized", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct S {
				access(E) var x: Int
				init() {
					self.x = 1
				}
			}
			fun foo() {
				let s = S()
				let ref = &s as auth(E) &S
				ref.x = 3
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAssignmentAccessError{}, errs[0])
	})

	t.Run("mapped owned", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement X
			entitlement Y
			entitlement mapping M {
				X -> Y
			}
			struct S {
				access(M) var x: auth(M) &Int
				init() {
					self.x = &1 as auth(Y) &Int
				}
			}
			fun foo() {
				let s = S()
				s.x = &1 as auth(Y) &Int
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAssignmentAccessError{}, errs[0])
	})

	t.Run("mapped authorized", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
		entitlement X
		entitlement Y
		entitlement mapping M {
			X -> Y
		}
		struct S {
			access(M) var x: auth(M) &Int
			init() {
				self.x = &1 as auth(Y) &Int
			}
		}
		fun foo() {
			let s = S()
			let ref = &s as auth(X) &S
			ref.x = &1 as auth(Y) &Int
		}
		`)

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAssignmentAccessError{}, errs[0])
	})

	t.Run("basic mutate", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct S {
				access(E) var x: [Int]
				init() {
					self.x = [1]
				}
			}
			fun foo() {
				let s = S()
				s.x.append(3)
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("basic authorized", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckWithOptions(t, `
			entitlement E
			struct S {
				access(E) var x: [Int]
				init() {
					self.x = [1]
				}
			}
			fun foo() {
				let s = S()
				let ref = &s as auth(E) &S
				ref.x.append(3)
			}
		`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					sema.InsertEntitlement,
					sema.MutateEntitlement,
				},
				sema.Disjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.UnauthorizedAccess,
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"reference needs one of entitlements `Insert` or `Mutate`",
		)
	})
}

func TestCheckAttachmentRequireEntitlements(t *testing.T) {
	t.Parallel()
	t.Run("entitlements allowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			attachment A for AnyStruct {
				require entitlement E
				require entitlement F
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("entitlement mapping disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement mapping M {}
			attachment A for AnyStruct {
				require entitlement E
				require entitlement M
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("event disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			event M()
			attachment A for AnyStruct {
				require entitlement E
				require entitlement M
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("struct disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct M {}
			attachment A for AnyStruct {
				require entitlement E
				require entitlement M
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("struct interface disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct interface M {}
			attachment A for AnyStruct {
				require entitlement E
				require entitlement M
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("resource disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource M {}
			attachment A for AnyStruct {
				require entitlement E
				require entitlement M
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("resource interface disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			resource interface M {}
			attachment A for AnyStruct {
				require entitlement E
				require entitlement M
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("attachment disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			attachment M for AnyResource {}
			attachment A for AnyStruct {
				require entitlement E
				require entitlement M
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("enum disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			enum M: UInt8 {}
			attachment A for AnyStruct {
				require entitlement E
				require entitlement M
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("int disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			attachment A for AnyStruct {
				require entitlement E
				require entitlement Int
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementRequirement{}, errs[0])
	})

	t.Run("duplicates disallowed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			attachment A for AnyStruct {
				require entitlement E
				require entitlement E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DuplicateEntitlementRequirementError{}, errs[0])
	})
}

func TestCheckAttachProvidedEntitlements(t *testing.T) {
	t.Parallel()
	t.Run("all provided", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			struct S {}
			attachment A for S {
				require entitlement E
				require entitlement F
			}
			fun foo() {
				let s = attach A() to S() with (E, F)
			}

		`)
		assert.NoError(t, err)
	})

	t.Run("extra provided", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			entitlement G
			struct S {}
			attachment A for S {
				require entitlement E
				require entitlement F
			}
			fun foo() {
				let s = attach A() to S() with (E, F, G)
			}

		`)
		assert.NoError(t, err)
	})

	t.Run("one missing", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			struct S {}
			attachment A for S {
				require entitlement E
				require entitlement F
			}
			fun foo() {
				let s = attach A() to S() with (E)
			}

		`)
		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.RequiredEntitlementNotProvidedError{}, errs[0])
		require.Equal(t, errs[0].(*sema.RequiredEntitlementNotProvidedError).RequiredEntitlement.Identifier, "F")
	})

	t.Run("one missing with extra provided", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			entitlement G
			struct S {}
			attachment A for S {
				require entitlement E
				require entitlement F
			}
			fun foo() {
				let s = attach A() to S() with (E, G)
			}

		`)
		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.RequiredEntitlementNotProvidedError{}, errs[0])
		require.Equal(t, errs[0].(*sema.RequiredEntitlementNotProvidedError).RequiredEntitlement.Identifier, "F")
	})

	t.Run("two missing", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement F
			struct S {}
			attachment A for S {
				require entitlement E
				require entitlement F
			}
			fun foo() {
				let s = attach A() to S() 
			}

		`)
		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.RequiredEntitlementNotProvidedError{}, errs[0])
		require.Equal(t, errs[0].(*sema.RequiredEntitlementNotProvidedError).RequiredEntitlement.Identifier, "E")

		require.IsType(t, &sema.RequiredEntitlementNotProvidedError{}, errs[1])
		require.Equal(t, errs[1].(*sema.RequiredEntitlementNotProvidedError).RequiredEntitlement.Identifier, "F")
	})

	t.Run("mapping provided", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			entitlement mapping M {}
			struct S {}
			attachment A for S {
				require entitlement E
			}
			fun foo() {
				let s = attach A() to S() with (M)
			}

		`)
		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNonEntitlementProvidedError{}, errs[0])

		require.IsType(t, &sema.RequiredEntitlementNotProvidedError{}, errs[1])
		require.Equal(t, errs[1].(*sema.RequiredEntitlementNotProvidedError).RequiredEntitlement.Identifier, "E")
	})

	t.Run("int provided", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct S {}
			attachment A for S {
				require entitlement E
			}
			fun foo() {
				let s = attach A() to S() with (UInt8)
			}

		`)
		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNonEntitlementProvidedError{}, errs[0])

		require.IsType(t, &sema.RequiredEntitlementNotProvidedError{}, errs[1])
		require.Equal(t, errs[1].(*sema.RequiredEntitlementNotProvidedError).RequiredEntitlement.Identifier, "E")
	})

	t.Run("struct provided", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E
			struct S {}
			attachment A for S {
				require entitlement E
			}
			fun foo() {
				let s = attach A() to S() with (S)
			}

		`)
		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNonEntitlementProvidedError{}, errs[0])

		require.IsType(t, &sema.RequiredEntitlementNotProvidedError{}, errs[1])
		require.Equal(t, errs[1].(*sema.RequiredEntitlementNotProvidedError).RequiredEntitlement.Identifier, "E")
	})
}

func TestCheckBuiltinEntitlements(t *testing.T) {

	t.Parallel()

	t.Run("builtin", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            struct S {
                access(Mutate) fun foo() {}
                access(Insert) fun bar() {}
                access(Remove) fun baz() {}
            }

            fun main() {
                let s = S()
                let mutableRef = &s as auth(Mutate) &S
                let insertableRef = &s as auth(Insert) &S
                let removableRef = &s as auth(Remove) &S
            }
        `)

		assert.NoError(t, err)
	})

	t.Run("redefine", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            entitlement Mutate
            entitlement Insert
            entitlement Remove
        `)

		errs := RequireCheckerErrors(t, err, 3)

		require.IsType(t, &sema.RedeclarationError{}, errs[0])
		require.IsType(t, &sema.RedeclarationError{}, errs[1])
		require.IsType(t, &sema.RedeclarationError{}, errs[2])
	})

}

func TestCheckIdentityMapping(t *testing.T) {

	t.Parallel()

	t.Run("owned value", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            struct S {
                access(Identity) fun foo(): auth(Identity) &AnyStruct {
                    let a: AnyStruct = "hello"
                    return &a as auth(Identity) &AnyStruct
                }
            }

            fun main() {
                let s = S()

                // OK
                let resultRef1: &AnyStruct = s.foo()

                // Error: Must return an unauthorized ref
                let resultRef2: auth(Mutate) &AnyStruct = s.foo()
            }
        `)

		errors := RequireCheckerErrors(t, err, 1)
		typeMismatchError := &sema.TypeMismatchError{}
		require.ErrorAs(t, errors[0], &typeMismatchError)

		require.IsType(t, &sema.ReferenceType{}, typeMismatchError.ActualType)
		actualReference := typeMismatchError.ActualType.(*sema.ReferenceType)

		require.IsType(t, sema.EntitlementSetAccess{}, actualReference.Authorization)
		actualAuth := actualReference.Authorization.(sema.EntitlementSetAccess)

		assert.Equal(t, 0, actualAuth.Entitlements.Len())
	})

	t.Run("unauthorized ref", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            struct S {
                access(Identity) fun foo(): auth(Identity) &AnyStruct {
                    let a: AnyStruct = "hello"
                    return &a as auth(Identity) &AnyStruct
                }
            }

            fun main() {
                let s = S()

                let ref = &s as &S

                // OK
                let resultRef1: &AnyStruct = ref.foo()

                // Error: Must return an unauthorized ref
                let resultRef2: auth(Mutate) &AnyStruct = ref.foo()
            }
        `)

		errors := RequireCheckerErrors(t, err, 1)
		typeMismatchError := &sema.TypeMismatchError{}
		require.ErrorAs(t, errors[0], &typeMismatchError)
	})

	t.Run("basic entitled ref", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            struct S {
                access(Identity) fun foo(): auth(Identity) &AnyStruct {
                    let a: AnyStruct = "hello"
                    return &a as auth(Identity) &AnyStruct
                }
            }

            fun main() {
                let s = S()

                let mutableRef = &s as auth(Mutate) &S
                let ref1: auth(Mutate) &AnyStruct = mutableRef.foo()

                let insertableRef = &s as auth(Insert) &S
                let ref2: auth(Insert) &AnyStruct = insertableRef.foo()

                let removableRef = &s as auth(Remove) &S
                let ref3: auth(Remove) &AnyStruct = removableRef.foo()
            }
        `)

		assert.NoError(t, err)
	})

	t.Run("entitlement set ref", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            struct S {
                access(Identity) fun foo(): auth(Identity) &AnyStruct {
                    let a: AnyStruct = "hello"
                    return &a as auth(Identity) &AnyStruct
                }
            }

            fun main() {
                let s = S()

                let ref1 = &s as auth(Insert | Remove) &S
                let resultRef1: auth(Insert | Remove) &AnyStruct = ref1.foo()

                let ref2 = &s as auth(Insert, Remove) &S
                let resultRef2: auth(Insert, Remove) &AnyStruct = ref2.foo()
            }
        `)

		assert.NoError(t, err)
	})

	t.Run("owned value, with entitlements", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            entitlement A
            entitlement B
            entitlement C

            struct X {
               access(A | B) var s: String

               init() {
                   self.s = "hello"
               }

               access(C) fun foo() {}
            }

            struct Y {

                // Reference
                access(Identity) var x1: auth(Identity) &X

                // Optional reference
                access(Identity) var x2: auth(Identity) &X?

                // Function returning a reference
                access(Identity) fun getX(): auth(Identity) &X {
                    let x = X()
                    return &x as auth(Identity) &X
                }

                // Function returning an optional reference
                access(Identity) fun getOptionalX(): auth(Identity) &X? {
                    let x: X? = X()
                    return &x as auth(Identity) &X?
                }

                init() {
                    let x = X()
                    self.x1 = &x as auth(A, B, C) &X
                    self.x2 = nil
                }
            }

            fun main() {
                let y = Y()

                let ref1: auth(A, B, C) &X = y.x1

                let ref2: auth(A, B, C) &X? = y.x2

                let ref3: auth(A, B, C) &X = y.getX()

                let ref4: auth(A, B, C) &X? = y.getOptionalX()
            }
        `)

		assert.NoError(t, err)
	})

	t.Run("owned value, with entitlements, function typed field", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            entitlement A
            entitlement B
            entitlement C

            struct X {
               access(A | B) var s: String

               init() {
                   self.s = "hello"
               }

               access(C) fun foo() {}
            }

            struct Y {

                access(Identity) let fn: (fun (): X)

                init() {
                    self.fn = fun(): X {
                        return X()
                    }
                }
            }

            fun main() {
                let y = Y()
                let v = y.fn()
            }
        `)

		errors := RequireCheckerErrors(t, err, 1)
		invalidMapping := &sema.InvalidMappedEntitlementMemberError{}
		require.ErrorAs(t, errors[0], &invalidMapping)
	})

	t.Run("owned value, with entitlements, function ref typed field", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
            entitlement A
            entitlement B
            entitlement C

            struct X {
               access(A | B) var s: String

               init() {
                   self.s = "hello"
               }

               access(C) fun foo() {}
            }

            struct Y {

                access(Identity) let fn: auth(Identity) &(fun (): X)?

                init() {
                    self.fn = nil
                }
            }

            fun main() {
                let y = Y()
                let v: auth(A, B, C) &(fun (): X) = y.fn
            }
        `)

		errors := RequireCheckerErrors(t, err, 1)
		typeMismatchError := &sema.TypeMismatchError{}
		require.ErrorAs(t, errors[0], &typeMismatchError)

		actualType := typeMismatchError.ActualType
		require.IsType(t, &sema.OptionalType{}, actualType)
		optionalType := actualType.(*sema.OptionalType)

		require.IsType(t, &sema.ReferenceType{}, optionalType.Type)
		referenceType := optionalType.Type.(*sema.ReferenceType)

		require.IsType(t, sema.EntitlementSetAccess{}, referenceType.Authorization)
		auth := referenceType.Authorization.(sema.EntitlementSetAccess)

		// Entitlements of function return type `X` must NOT be
		// available for the reference typed field.
		require.Equal(t, 0, auth.Entitlements.Len())
	})
}

func TestCheckEntitlementErrorReporting(t *testing.T) {
	t.Run("three or more conjunction", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheckWithOptions(t, `
		entitlement X
		entitlement Y
		entitlement Z
		entitlement A
		entitlement B
		struct S {
			view access(X, Y, Z) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: auth(A, B) &S) {
			r.foo()
		}
	`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.X"),
					checker.Elaboration.EntitlementType("S.test.Y"),
					checker.Elaboration.EntitlementType("S.test.Z"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.A"),
					checker.Elaboration.EntitlementType("S.test.B"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"reference needs all of entitlements `X`, `Y`, and `Z`",
		)
	})

	t.Run("has one entitlement of three", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheckWithOptions(t, `
		entitlement X
		entitlement Y
		entitlement Z
		entitlement A
		entitlement B
		struct S {
			view access(X, Y, Z) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: auth(A, B, Y) &S) {
			r.foo()
		}
	`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.X"),
					checker.Elaboration.EntitlementType("S.test.Y"),
					checker.Elaboration.EntitlementType("S.test.Z"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.A"),
					checker.Elaboration.EntitlementType("S.test.B"),
					checker.Elaboration.EntitlementType("S.test.Y"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"reference needs all of entitlements `X` and `Z`",
		)
	})

	t.Run("has one entitlement of three", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheckWithOptions(t, `
		entitlement X
		entitlement Y
		entitlement Z
		entitlement A
		entitlement B
		struct S {
			view access(X | Y | Z) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: auth(A, B) &S) {
			r.foo()
		}
	`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.X"),
					checker.Elaboration.EntitlementType("S.test.Y"),
					checker.Elaboration.EntitlementType("S.test.Z"),
				},
				sema.Disjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.A"),
					checker.Elaboration.EntitlementType("S.test.B"),
				},
				sema.Conjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"reference needs one of entitlements `X`, `Y`, or `Z`",
		)
	})

	t.Run("no suggestion for disjoint possession set", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheckWithOptions(t, `
		entitlement X
		entitlement Y
		entitlement Z
		entitlement A
		entitlement B
		struct S {
			view access(X | Y | Z) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: auth(A | B) &S) {
			r.foo()
		}
	`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.X"),
					checker.Elaboration.EntitlementType("S.test.Y"),
					checker.Elaboration.EntitlementType("S.test.Z"),
				},
				sema.Disjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			sema.NewEntitlementSetAccess(
				[]*sema.EntitlementType{
					checker.Elaboration.EntitlementType("S.test.A"),
					checker.Elaboration.EntitlementType("S.test.B"),
				},
				sema.Disjunction,
			),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"",
		)
	})

	t.Run("no suggestion for self access requirement", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckWithOptions(t, `
		entitlement A
		entitlement B
		struct S {
			view access(self) fun foo(): Bool {
				return true
			}
		}
		fun bar(r: auth(A, B) &S) {
			r.foo()
		}
	`, ParseAndCheckOptions{Config: &sema.Config{SuggestionsEnabled: true}})

		errs := RequireCheckerErrors(t, err, 1)
		require.IsType(t, &sema.InvalidAccessError{}, errs[0])
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).RestrictingAccess,
			sema.PrimitiveAccess(ast.AccessSelf),
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).PossessedAccess,
			nil,
		)
		assert.Equal(
			t,
			errs[0].(*sema.InvalidAccessError).SecondaryError(),
			"",
		)
	})
}

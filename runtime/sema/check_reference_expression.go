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

package sema

import (
	"github.com/onflow/cadence/runtime/ast"
)

// VisitReferenceExpression checks a reference expression
func (checker *Checker) VisitReferenceExpression(referenceExpression *ast.ReferenceExpression) Type {

	resultType := checker.expectedType
	if resultType == nil {
		checker.report(
			&TypeAnnotationRequiredError{
				Cause: "cannot infer type from reference expression:",
				Pos:   referenceExpression.Expression.StartPosition(),
			},
		)
		return InvalidType
	}

	// Check the result type and ensure it is a reference type
	var isOpt bool
	var referenceType *ReferenceType
	var expectedLeftType, returnType Type

	if !resultType.IsInvalidType() {
		var ok bool
		// Reference expressions may reference a value which has an optional type.
		// For example, the result of indexing into a dictionary is an optional:
		//
		// let ints: {Int: String} = {0: "zero"}
		// let ref: &T? = &ints[0] as &T?   // read as (&T)?
		//
		// In this case the reference expression's type is an optional type.
		// Unwrap it one level to get the actual reference type
		var optType *OptionalType
		optType, isOpt = resultType.(*OptionalType)
		if isOpt {
			resultType = optType.Type
		}

		referenceType, ok = resultType.(*ReferenceType)
		if !ok {
			checker.report(
				&NonReferenceTypeReferenceError{
					ActualType: resultType,
					Range:      ast.NewRangeFromPositioned(checker.memoryGauge, referenceExpression),
				},
			)
		} else {
			expectedLeftType = referenceType.Type
			returnType = referenceType
			if isOpt {
				expectedLeftType = &OptionalType{Type: expectedLeftType}
				returnType = &OptionalType{Type: returnType}
			}
		}
	}

	// Type-check the referenced expression

	referencedExpression := referenceExpression.Expression

	beforeErrors := len(checker.errors)

	referencedType, actualType := checker.visitExpression(referencedExpression, expectedLeftType)

	// check that the type of the referenced value is not itself a reference
	var requireNoReferenceNesting func(actualType Type)
	requireNoReferenceNesting = func(actualType Type) {
		switch nestedReference := actualType.(type) {
		case *ReferenceType:
			checker.report(&NestedReferenceError{
				Type:  nestedReference,
				Range: checker.expressionRange(referenceExpression),
			})
		case *OptionalType:
			requireNoReferenceNesting(nestedReference.Type)
		}
	}
	requireNoReferenceNesting(actualType)

	hasErrors := len(checker.errors) > beforeErrors
	if !hasErrors {
		// If the reference type was an optional type,
		// we proposed an optional type to the referenced expression.
		//
		// Check that it actually has an optional type

		// If the reference type was a non-optional type,
		// check that the referenced expression does not have an optional type

		// Do not report an error if the `expectedLeftType` is unknown

		if _, ok := actualType.(*OptionalType); ok != isOpt && expectedLeftType != nil {
			checker.report(&TypeMismatchError{
				ExpectedType: expectedLeftType,
				ActualType:   actualType,
				Expression:   referencedExpression,
				Range:        checker.expressionRange(referenceExpression),
			})
		}
	}

	if referenceType == nil {
		return InvalidType
	}

	checker.checkUnusedExpressionResourceLoss(referencedType, referencedExpression)

	checker.Elaboration.SetReferenceExpressionBorrowType(referenceExpression, returnType)

	return returnType
}

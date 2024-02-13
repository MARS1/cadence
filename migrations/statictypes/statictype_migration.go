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

package statictypes

import (
	"fmt"

	"github.com/onflow/cadence/migrations"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/sema"
)

type StaticTypeMigration struct {
	compositeTypeConverter CompositeTypeConverterFunc
	interfaceTypeConverter InterfaceTypeConverterFunc
}

type CompositeTypeConverterFunc func(staticType *interpreter.CompositeStaticType) interpreter.StaticType
type InterfaceTypeConverterFunc func(staticType *interpreter.InterfaceStaticType) interpreter.StaticType

var _ migrations.ValueMigration = &StaticTypeMigration{}

func NewStaticTypeMigration() *StaticTypeMigration {
	return &StaticTypeMigration{}
}

func (m *StaticTypeMigration) WithCompositeTypeConverter(converterFunc CompositeTypeConverterFunc) *StaticTypeMigration {
	m.compositeTypeConverter = converterFunc
	return m
}

func (m *StaticTypeMigration) WithInterfaceTypeConverter(converterFunc InterfaceTypeConverterFunc) *StaticTypeMigration {
	m.interfaceTypeConverter = converterFunc
	return m
}

func (*StaticTypeMigration) Name() string {
	return "StaticTypeMigration"
}

// Migrate migrates `AuthAccount` and `PublicAccount` types inside `TypeValue`s,
// to the account reference type (&Account).
func (m *StaticTypeMigration) Migrate(
	_ interpreter.StorageKey,
	_ interpreter.StorageMapKey,
	value interpreter.Value,
	_ *interpreter.Interpreter,
) (newValue interpreter.Value, err error) {
	switch value := value.(type) {
	case interpreter.TypeValue:
		convertedType := m.maybeConvertStaticType(value.Type)
		if convertedType == nil {
			return
		}
		return interpreter.NewTypeValue(nil, convertedType), nil

	case *interpreter.CapabilityValue:
		convertedBorrowType := m.maybeConvertStaticType(value.BorrowType)
		if convertedBorrowType == nil {
			return
		}
		return interpreter.NewUnmeteredCapabilityValue(value.ID, value.Address, convertedBorrowType), nil

	case *interpreter.PathCapabilityValue: //nolint:staticcheck
		convertedBorrowType := m.maybeConvertStaticType(value.BorrowType)
		if convertedBorrowType == nil {
			return
		}
		return &interpreter.PathCapabilityValue{ //nolint:staticcheck
			BorrowType: convertedBorrowType,
			Path:       value.Path,
			Address:    value.Address,
		}, nil

	case interpreter.PathLinkValue: //nolint:staticcheck
		convertedBorrowType := m.maybeConvertStaticType(value.Type)
		if convertedBorrowType == nil {
			return
		}
		return interpreter.PathLinkValue{ //nolint:staticcheck
			Type:       convertedBorrowType,
			TargetPath: value.TargetPath,
		}, nil

	case *interpreter.AccountCapabilityControllerValue:
		convertedBorrowType := m.maybeConvertStaticType(value.BorrowType)
		if convertedBorrowType == nil {
			return
		}
		borrowType := convertedBorrowType.(*interpreter.ReferenceStaticType)
		return interpreter.NewUnmeteredAccountCapabilityControllerValue(borrowType, value.CapabilityID), nil

	case *interpreter.StorageCapabilityControllerValue:
		convertedBorrowType := m.maybeConvertStaticType(value.BorrowType)
		if convertedBorrowType == nil {
			return
		}
		borrowType := convertedBorrowType.(*interpreter.ReferenceStaticType)
		return interpreter.NewUnmeteredStorageCapabilityControllerValue(
			borrowType,
			value.CapabilityID,
			value.TargetPath,
		), nil
	}

	return
}

func (m *StaticTypeMigration) maybeConvertStaticType(staticType interpreter.StaticType) interpreter.StaticType {
	switch staticType := staticType.(type) {
	case *interpreter.ConstantSizedStaticType:
		convertedType := m.maybeConvertStaticType(staticType.Type)
		if convertedType != nil {
			return interpreter.NewConstantSizedStaticType(nil, convertedType, staticType.Size)
		}

	case *interpreter.VariableSizedStaticType:
		convertedType := m.maybeConvertStaticType(staticType.Type)
		if convertedType != nil {
			return interpreter.NewVariableSizedStaticType(nil, convertedType)
		}

	case *interpreter.DictionaryStaticType:
		convertedKeyType := m.maybeConvertStaticType(staticType.KeyType)
		convertedValueType := m.maybeConvertStaticType(staticType.ValueType)
		if convertedKeyType != nil && convertedValueType != nil {
			return interpreter.NewDictionaryStaticType(nil, convertedKeyType, convertedValueType)
		}
		if convertedKeyType != nil {
			return interpreter.NewDictionaryStaticType(nil, convertedKeyType, staticType.ValueType)
		}
		if convertedValueType != nil {
			return interpreter.NewDictionaryStaticType(nil, staticType.KeyType, convertedValueType)
		}

	case *interpreter.CapabilityStaticType:
		convertedBorrowType := m.maybeConvertStaticType(staticType.BorrowType)
		if convertedBorrowType != nil {
			return interpreter.NewCapabilityStaticType(nil, convertedBorrowType)
		}

	case *interpreter.IntersectionStaticType:

		var convertedInterfaceTypes []*interpreter.InterfaceStaticType

		var convertedInterfaceType bool

		for _, interfaceStaticType := range staticType.Types {
			convertedType := m.maybeConvertStaticType(interfaceStaticType)

			// lazily allocate the slice
			if convertedInterfaceTypes == nil {
				convertedInterfaceTypes = make([]*interpreter.InterfaceStaticType, 0, len(staticType.Types))
			}

			var replacement *interpreter.InterfaceStaticType
			if convertedType != nil {
				var ok bool
				replacement, ok = convertedType.(*interpreter.InterfaceStaticType)
				if !ok {
					panic(fmt.Errorf(
						"invalid non-interface replacement in intersection type %s: %s replaced by %s",
						staticType,
						interfaceStaticType,
						convertedType,
					))
				}

				convertedInterfaceType = true
			} else {
				replacement = interfaceStaticType
			}
			convertedInterfaceTypes = append(convertedInterfaceTypes, replacement)
		}

		legacyType := staticType.LegacyType
		var convertedLegacyType interpreter.StaticType
		if legacyType != nil {
			convertedLegacyType = m.maybeConvertStaticType(legacyType)
		}

		// If the interface set has at least two items,
		// then force it to be re-stored/re-encoded,
		// even if the interface types in the set have not changed.
		if len(staticType.Types) >= 2 || convertedInterfaceType || convertedLegacyType != nil {

			intersectionType := interpreter.NewIntersectionStaticType(nil, convertedInterfaceTypes)

			if convertedLegacyType != nil {
				intersectionType.LegacyType = convertedLegacyType
			} else {
				intersectionType.LegacyType = staticType.LegacyType
			}

			return intersectionType
		}

	case *interpreter.OptionalStaticType:
		convertedInnerType := m.maybeConvertStaticType(staticType.Type)
		if convertedInnerType != nil {
			return interpreter.NewOptionalStaticType(nil, convertedInnerType)
		}

	case *interpreter.ReferenceStaticType:
		// TODO: Reference of references must not be allowed?
		convertedReferencedType := m.maybeConvertStaticType(staticType.ReferencedType)
		if convertedReferencedType != nil {
			switch convertedReferencedType {

			// If the converted type is already an account reference, then return as-is.
			// i.e: Do not create reference to a reference.
			case authAccountReferenceType,
				unauthorizedAccountReferenceType:
				return convertedReferencedType

			default:
				return interpreter.NewReferenceStaticType(nil, staticType.Authorization, convertedReferencedType)
			}
		}

	case interpreter.FunctionStaticType:
		// Non-storable

	case *interpreter.CompositeStaticType:
		compositeTypeConverter := m.compositeTypeConverter
		if compositeTypeConverter != nil {
			return compositeTypeConverter(staticType)
		}

	case *interpreter.InterfaceStaticType:
		interfaceTypeConverter := m.interfaceTypeConverter
		if interfaceTypeConverter != nil {
			return interfaceTypeConverter(staticType)
		}

	case dummyStaticType:
		// This is for testing the migration.
		// i.e: wrapper was only to make it possible to use as a dictionary-key.
		// Ignore the wrapper, and continue with the inner type.
		return m.maybeConvertStaticType(staticType.PrimitiveStaticType)

	case interpreter.PrimitiveStaticType:
		// Is it safe to do so?
		switch staticType {
		case interpreter.PrimitiveStaticTypePublicAccount: //nolint:staticcheck
			return unauthorizedAccountReferenceType

		case interpreter.PrimitiveStaticTypeAuthAccount: //nolint:staticcheck
			return authAccountReferenceType

		case interpreter.PrimitiveStaticTypeAuthAccountCapabilities, //nolint:staticcheck
			interpreter.PrimitiveStaticTypePublicAccountCapabilities: //nolint:staticcheck
			return interpreter.PrimitiveStaticTypeAccount_Capabilities

		case interpreter.PrimitiveStaticTypeAuthAccountAccountCapabilities: //nolint:staticcheck
			return interpreter.PrimitiveStaticTypeAccount_AccountCapabilities

		case interpreter.PrimitiveStaticTypeAuthAccountStorageCapabilities: //nolint:staticcheck
			return interpreter.PrimitiveStaticTypeAccount_StorageCapabilities

		case interpreter.PrimitiveStaticTypeAuthAccountContracts, //nolint:staticcheck
			interpreter.PrimitiveStaticTypePublicAccountContracts: //nolint:staticcheck
			return interpreter.PrimitiveStaticTypeAccount_Contracts

		case interpreter.PrimitiveStaticTypeAuthAccountKeys, //nolint:staticcheck
			interpreter.PrimitiveStaticTypePublicAccountKeys: //nolint:staticcheck
			return interpreter.PrimitiveStaticTypeAccount_Keys

		case interpreter.PrimitiveStaticTypeAuthAccountInbox: //nolint:staticcheck
			return interpreter.PrimitiveStaticTypeAccount_Inbox

		case interpreter.PrimitiveStaticTypeAccountKey: //nolint:staticcheck
			return interpreter.AccountKeyStaticType
		}

	default:
		panic(errors.NewUnreachableError())
	}

	return nil
}

var authAccountEntitlements = []common.TypeID{
	sema.StorageType.ID(),
	sema.ContractsType.ID(),
	sema.KeysType.ID(),
	sema.InboxType.ID(),
	sema.CapabilitiesType.ID(),
}

var authAccountReferenceType = func() *interpreter.ReferenceStaticType {
	auth := interpreter.NewEntitlementSetAuthorization(
		nil,
		func() []common.TypeID {
			return authAccountEntitlements
		},
		len(authAccountEntitlements),
		sema.Conjunction,
	)
	return interpreter.NewReferenceStaticType(
		nil,
		auth,
		interpreter.PrimitiveStaticTypeAccount,
	)
}()

var unauthorizedAccountReferenceType = interpreter.NewReferenceStaticType(
	nil,
	interpreter.UnauthorizedAccess,
	interpreter.PrimitiveStaticTypeAccount,
)
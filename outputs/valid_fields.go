package outputs

import "reflect"

func CheckValidField(obj interface{}, fieldType reflect.Type) error {
	fKind := fieldType.Kind()
	isPtr := fKind == reflect.Ptr
	if isPtr {
		fKind = fieldType.Elem().Kind()
	}

	switch fKind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String,
		reflect.Array, reflect.Slice,
		reflect.Map, reflect.Struct:
		return nil
	default:
		// we don't support Interface, Func, UnsafePointer, Chan as decode targets
		// we also don't support double pointers: **<type>
		return ErrInvalidContractField{
			ObjectType: reflect.TypeOf(obj),
			FieldType:  fieldType,
			Message:    "invalid output field type",
		}
	}
}

func CheckValidConnectionField(obj interface{}, fieldType reflect.Type) error {
	fKind := fieldType.Kind()
	isPtr := fKind == reflect.Ptr
	if isPtr {
		fKind = fieldType.Elem().Kind()
	}
	switch fKind {
	case reflect.Struct:
		return nil
	}
	return ErrInvalidContractField{
		ObjectType: reflect.TypeOf(obj),
		FieldType:  fieldType,
		Message:    "connections outputs must be decoded into a struct",
	}
}

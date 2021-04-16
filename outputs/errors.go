package outputs

import (
	"fmt"
	"reflect"
)

type ErrInvalidContractField struct {
	ObjectType reflect.Type
	FieldType  reflect.Type
	Message    string
}

func (e ErrInvalidContractField) Error() string {
	return fmt.Sprintf("invalid contract field for (%s, %s), connection outputs must be decoded into a struct", e.ObjectType.Name(), e.FieldType.Name())
}

type ErrMissingRequiredConnection struct {
	ConnectionName string
	ConnectionType string
}

func (e ErrMissingRequiredConnection) Error() string {
	return fmt.Sprintf("required connection missing (name=%s, type=%s)", e.ConnectionName, e.ConnectionType)
}

type ErrMissingRequiredOutput struct {
	Name string
}

func (e ErrMissingRequiredOutput) Error() string {
	return fmt.Sprintf("required output missing (name=%s)", e.Name)
}

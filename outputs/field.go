package outputs

import (
	"encoding/json"
	"fmt"
	"github.com/vmihailenco/tagparser"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"reflect"
)

var (
	StructTag               = "ns"
	StructTagConnectionName = "connectionName"
	StructTagConnectionType = "connectionType"
	StructTagOptional       = "optional"
)

/*
Field is a representation of a struct tag that is specific to nullstone outputs
You can define an output mapping directly to a workspace or to outputs through a connection in that workspace

Examples:
type Outputs struct {
  Output1        string            `ns:"output1"`
  OptionalOutput string            `ns:"optional_output,optional"`
  MapOutput      map[string]string `ns:"map_output"`

  Dependency DependencyOutputs `ns:",connectionType:some-dependency"`
}

type DependencyOutputs struct {
  Output2 string `ns:"output2"`
}

Notes:
  All fields that that map connections must be a well-defined struct
  If you want to ignore a member in the struct, use `ns:"-"`
  If you want to make a field/connection optional, add `ns:"output,optional"`
*/
type Field struct {
	Field          reflect.StructField
	Tag            string
	Name           string
	ConnectionType string
	ConnectionName string
	Optional       bool
}

func (f Field) SafeSet(sourceObj interface{}, outputs types.Outputs) error {
	sourceType := reflect.TypeOf(sourceObj)
	if sourceType.Kind() != reflect.Ptr {
		return fmt.Errorf("source object must be a pointer")
	}
	if sourceType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("source object must be a pointer to a struct")
	}

	var item types.OutputItem
	var ok bool
	if outputs != nil {
		item, ok = outputs[f.Name]
	}
	if !ok {
		if f.Optional {
			return nil
		}
		return ErrMissingRequiredOutput{
			Name: f.Name,
		}
	}

	objVal := reflect.ValueOf(sourceObj).Elem()
	fieldVal := objVal.FieldByName(f.Field.Name)
	rawJsonEncoded, _ := json.Marshal(item.Value)
	if err := json.Unmarshal(rawJsonEncoded, fieldVal.Addr().Interface()); err != nil {
		return fmt.Errorf("could not deserialize output value from output %q: %w", f.Name, err)
	}
	return nil
}

func (f Field) InitializeConnectionValue(obj interface{}) interface{} {
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Ptr {
		return fmt.Errorf("source object must be a pointer")
	}
	if objType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("source object must be a pointer to a struct")
	}

	objVal := reflect.ValueOf(obj).Elem()
	fieldVal := objVal.FieldByName(f.Field.Name)
	if fieldVal.Kind() == reflect.Ptr {
		newPtr := reflect.New(f.Field.Type.Elem())
		fieldVal.Set(newPtr)
		return newPtr.Interface()
	}
	return fieldVal.Addr().Interface()
}

func GetFields(typ reflect.Type) []Field {
	fields := make([]Field, 0)
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		field := Field{
			Field: f,
		}
		field.Tag = f.Tag.Get(StructTag)
		if field.Tag == "-" {
			continue
		}

		structured := tagparser.Parse(field.Tag)
		field.Name = structured.Name
		field.ConnectionName = structured.Options[StructTagConnectionName]
		field.ConnectionType = structured.Options[StructTagConnectionType]
		field.Optional = structured.HasOption(StructTagOptional)

		fields = append(fields, field)
	}

	return fields
}

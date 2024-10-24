// Code generated by github.com/actgardner/gogen-avro/v10. DO NOT EDIT.
package dto

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/actgardner/gogen-avro/v10/compiler"
	"github.com/actgardner/gogen-avro/v10/vm"
	"github.com/actgardner/gogen-avro/v10/vm/types"
)

var _ = fmt.Printf

type User struct {
	Name string `json:"name"`

	Favorite_number int64 `json:"favorite_number"`

	Favorite_color string `json:"favorite_color"`
}

const UserAvroCRC64Fingerprint = "\x8d\t\x1dg\r\xa8\xf9r"

func NewUser() User {
	r := User{}
	return r
}

func DeserializeUser(r io.Reader) (User, error) {
	t := NewUser()
	deser, err := compiler.CompileSchemaBytes([]byte(t.Schema()), []byte(t.Schema()))
	if err != nil {
		return t, err
	}

	err = vm.Eval(r, deser, &t)
	return t, err
}

func DeserializeUserFromSchema(r io.Reader, schema string) (User, error) {
	t := NewUser()

	deser, err := compiler.CompileSchemaBytes([]byte(schema), []byte(t.Schema()))
	if err != nil {
		return t, err
	}

	err = vm.Eval(r, deser, &t)
	return t, err
}

func writeUser(r User, w io.Writer) error {
	var err error
	err = vm.WriteString(r.Name, w)
	if err != nil {
		return err
	}
	err = vm.WriteLong(r.Favorite_number, w)
	if err != nil {
		return err
	}
	err = vm.WriteString(r.Favorite_color, w)
	if err != nil {
		return err
	}
	return err
}

func (r User) Serialize(w io.Writer) error {
	return writeUser(r, w)
}

func (r User) Schema() string {
	return "{\"fields\":[{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"favorite_number\",\"type\":\"long\"},{\"name\":\"favorite_color\",\"type\":\"string\"}],\"name\":\"kafkapracticum.User\",\"type\":\"record\"}"
}

func (r User) SchemaName() string {
	return "kafkapracticum.User"
}

func (_ User) SetBoolean(v bool)    { panic("Unsupported operation") }
func (_ User) SetInt(v int32)       { panic("Unsupported operation") }
func (_ User) SetLong(v int64)      { panic("Unsupported operation") }
func (_ User) SetFloat(v float32)   { panic("Unsupported operation") }
func (_ User) SetDouble(v float64)  { panic("Unsupported operation") }
func (_ User) SetBytes(v []byte)    { panic("Unsupported operation") }
func (_ User) SetString(v string)   { panic("Unsupported operation") }
func (_ User) SetUnionElem(v int64) { panic("Unsupported operation") }

func (r *User) Get(i int) types.Field {
	switch i {
	case 0:
		w := types.String{Target: &r.Name}

		return w

	case 1:
		w := types.Long{Target: &r.Favorite_number}

		return w

	case 2:
		w := types.String{Target: &r.Favorite_color}

		return w

	}
	panic("Unknown field index")
}

func (r *User) SetDefault(i int) {
	switch i {
	}
	panic("Unknown field index")
}

func (r *User) NullField(i int) {
	switch i {
	}
	panic("Not a nullable field index")
}

func (_ User) AppendMap(key string) types.Field { panic("Unsupported operation") }
func (_ User) AppendArray() types.Field         { panic("Unsupported operation") }
func (_ User) HintSize(int)                     { panic("Unsupported operation") }
func (_ User) Finalize()                        {}

func (_ User) AvroCRC64Fingerprint() []byte {
	return []byte(UserAvroCRC64Fingerprint)
}

func (r User) MarshalJSON() ([]byte, error) {
	var err error
	output := make(map[string]json.RawMessage)
	output["name"], err = json.Marshal(r.Name)
	if err != nil {
		return nil, err
	}
	output["favorite_number"], err = json.Marshal(r.Favorite_number)
	if err != nil {
		return nil, err
	}
	output["favorite_color"], err = json.Marshal(r.Favorite_color)
	if err != nil {
		return nil, err
	}
	return json.Marshal(output)
}

func (r *User) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	var val json.RawMessage
	val = func() json.RawMessage {
		if v, ok := fields["name"]; ok {
			return v
		}
		return nil
	}()

	if val != nil {
		if err := json.Unmarshal([]byte(val), &r.Name); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no value specified for name")
	}
	val = func() json.RawMessage {
		if v, ok := fields["favorite_number"]; ok {
			return v
		}
		return nil
	}()

	if val != nil {
		if err := json.Unmarshal([]byte(val), &r.Favorite_number); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no value specified for favorite_number")
	}
	val = func() json.RawMessage {
		if v, ok := fields["favorite_color"]; ok {
			return v
		}
		return nil
	}()

	if val != nil {
		if err := json.Unmarshal([]byte(val), &r.Favorite_color); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no value specified for favorite_color")
	}
	return nil
}

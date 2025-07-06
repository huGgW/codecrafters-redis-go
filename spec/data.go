package spec

import (
	"fmt"
	"reflect"
)

// TODO: add RESP3 specs when needed

type Data interface {
	data()

	Value() any
	Type() reflect.Type
	Incomplete() bool
}

func Value[T any](data Data) (T, error) {
	var zero T

	assignType := reflect.TypeOf(zero)
	if !data.Type().AssignableTo(assignType) {
		return zero, fmt.Errorf("type %v cannot be assigned to %v", data, assignType)
	}

	return data.Value().(T), nil
}

type SimpleStringData struct {
	S string
}

func SimpleStringOf(s string) *SimpleStringData {
	return &SimpleStringData{S: s}
}

func (s *SimpleStringData) data()              {}
func (s *SimpleStringData) Value() any         { return s.S }
func (s *SimpleStringData) Type() reflect.Type { return reflect.TypeFor[string]() }
func (s *SimpleStringData) Incomplete() bool   { return false }

type SimpleErrorData struct {
	Err error
}

func (e *SimpleErrorData) data()              {}
func (e *SimpleErrorData) Value() any         { return e.Err }
func (e *SimpleErrorData) Type() reflect.Type { return reflect.TypeFor[error]() }
func (e *SimpleErrorData) Incomplete() bool   { return false }

type IntegerData struct {
	I int64
}

func (i *IntegerData) data()              {}
func (i *IntegerData) Value() any         { return i.I }
func (i *IntegerData) Type() reflect.Type { return reflect.TypeFor[int64]() }
func (i *IntegerData) Incomplete() bool   { return false }

type BulkStringData struct {
	Len int
	S   string
}

func NullBulkString() *BulkStringData {
	return &BulkStringData{Len: -1, S: ""}
}

func BulkStringOf(s string) *BulkStringData {
	return &BulkStringData{Len: len(s), S: s}
}

func (b *BulkStringData) data()              {}
func (b *BulkStringData) Value() any         { return b.S }
func (b *BulkStringData) Type() reflect.Type { return reflect.TypeFor[string]() }
func (b *BulkStringData) Incomplete() bool {
	return len(b.S) < b.Len
}

func (b *BulkStringData) IsNull() bool {
	return b.Len == -1
}

type ArrayData struct {
	Len int
	A   []Data
}

func (a *ArrayData) data()              {}
func (a *ArrayData) Value() any         { return a.A }
func (a *ArrayData) Type() reflect.Type { return reflect.TypeFor[[]Data]() }
func (a *ArrayData) Incomplete() bool {
	return len(a.A) < a.Len
}

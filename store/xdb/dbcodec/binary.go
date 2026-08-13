package dbcodec

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

var _ dbtype.Codec = (*Binary)(nil)

type Binary struct{}

func (b Binary) Name() string {
	return "binary"
}

func (b Binary) Encode(a any) (any, error) {
	if a == nil {
		return nil, nil
	}

	switch v := a.(type) {
	case []byte:
		return v, nil
	}
	rv := reflect.ValueOf(a)
	if rv.Kind() == reflect.Array && rv.Type().Elem().Kind() == reflect.Uint8 {
		nb := make([]byte, rv.Len())
		reflect.Copy(reflect.ValueOf(nb), rv)
		return nb, nil
	}

	return nil, b.invalidType(a)
}

func (b Binary) invalidType(a any) error {
	return fmt.Errorf("unsupported type %T, want []byte or [N]byte", a)
}

func (b Binary) Decode(str string, a any) error {
	if a == nil {
		return errors.New("destination is nil")
	}

	rv := reflect.ValueOf(a)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%T is a non-nil pointer ", a)
	}

	rv = rv.Elem()

	if (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) || rv.Type().Elem().Kind() != reflect.Uint8 {
		return b.invalidType(a)
	}

	switch rv.Kind() {
	case reflect.Slice:
		nb := []byte(str)
		rv.SetBytes(nb)
		return nil
	case reflect.Array:
		nb := []byte(str)
		if len(nb) != rv.Len() {
			return fmt.Errorf("length mismatch, got %d bytes, want %d", len(nb), rv.Len())
		}

		reflect.Copy(rv, reflect.ValueOf(nb))
		return nil

	default:
		return b.invalidType(a)
	}
}

func (b Binary) Kind() dbtype.Kind {
	return dbtype.KindBinary
}

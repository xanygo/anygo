package dbcodec

import (
	"errors"
	"fmt"
	"uuid"

	"github.com/xanygo/anygo/xdb/dbtype"
	"github.com/xanygo/anygo/xerror"
)

type UUID struct{}

const UUIDName = "uuid"

var _ dbtype.Codec = (*Text)(nil)
var _ dbtype.HasKind = (*Text)(nil)

func (t UUID) Kind() dbtype.Kind {
	return dbtype.KindUUID
}

func (t UUID) Name() string {
	return UUIDName
}

func (t UUID) Encode(obj any) (any, error) {
	switch tv := obj.(type) {
	case uuid.UUID:
		return tv[:], nil
	default:
		return nil, fmt.Errorf("uuid encode %w %T", xerror.ErrInvalidType, obj)
	}
}

var uuidZsroStr1 = "00000000-0000-0000-0000-000000000000"

func (t UUID) Decode(str string, obj any) error {
	switch v := obj.(type) {
	case *uuid.UUID:
		if v == nil {
			return errors.New("nil *uuid.UUID")
		}
		var id uuid.UUID
		switch len(str) {
		case 16:
			copy(id[:], str)
			*v = id
			return nil
		default:
			if str == uuidZsroStr1 {
				*v = id
				return nil
			}
			id, err := uuid.Parse(str)
			if err != nil {
				return err
			}
			*v = id
			return nil
		}

	default:
		return fmt.Errorf("unsupported UUID decode type %T with %q", obj, str)
	}
}

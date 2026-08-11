package xdb

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/dialect"
)

func NewSchema(client Queryer) (*Schema, error) {
	d, err := dialect.Find(client.Driver())
	if err != nil {
		return nil, err
	}
	return &Schema{
		client:  client,
		dialect: d,
	}, nil
}

func MustNewSchema(client Queryer) *Schema {
	s, err := NewSchema(client)
	if err != nil {
		panic(err)
	}
	return s
}

type Schema struct {
	dialect dbtype.Dialect
	client  Queryer
}

func (s *Schema) getDescDialect() (dbtype.DescDialect, error) {
	dd, ok := s.dialect.(dbtype.DescDialect)
	if ok {
		return dd, nil
	}
	return nil, fmt.Errorf("dialect %T is not DescDialect", s.dialect)
}

func (s *Schema) Databases(ctx context.Context) ([]string, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return nil, err
	}
	return dd.Databases(ctx, s.client)
}

func (s *Schema) CurrentDatabase(ctx context.Context) (string, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return "", err
	}
	return dd.CurrentDatabase(ctx, s.client)
}

func (s *Schema) Tables(ctx context.Context) ([]string, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return nil, err
	}
	return dd.Tables(ctx, s.client)
}

func (s *Schema) TableExists(ctx context.Context, table string) (bool, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return false, err
	}
	return dd.TableExists(ctx, s.client, table)
}

func (s *Schema) TableColumns(ctx context.Context, table string) ([]string, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return nil, err
	}
	return dd.TableColumns(ctx, s.client, table)
}

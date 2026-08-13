package xdb

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/dialect"
)

func NewSchemaAPI(client HasDriver) (*SchemaAPI, error) {
	d, err := dialect.Find(client.Driver())
	if err != nil {
		return nil, err
	}
	return &SchemaAPI{
		client:  client,
		dialect: d,
	}, nil
}

func MustNewSchemaAPI(client Queryer) *SchemaAPI {
	s, err := NewSchemaAPI(client)
	if err != nil {
		panic(err)
	}
	return s
}

type SchemaAPI struct {
	dialect dbtype.Dialect
	client  HasDriver
}

func (s *SchemaAPI) getDescDialect() (dbtype.DescDialect, error) {
	dd, ok := s.dialect.(dbtype.DescDialect)
	if ok {
		return dd, nil
	}
	return nil, fmt.Errorf("dialect %T is not DescDialect", s.dialect)
}

func (s *SchemaAPI) getQueryer() (Queryer, error) {
	q, ok := s.client.(Queryer)
	if ok {
		return q, nil
	}
	return nil, fmt.Errorf("client %T is not Queryer", s.client)
}

func (s *SchemaAPI) getExecer() (Execer, error) {
	q, ok := s.client.(Execer)
	if ok {
		return q, nil
	}
	return nil, fmt.Errorf("client %T is not Execer", s.client)
}

// Databases 返回数据库名称列表
func (s *SchemaAPI) Databases(ctx context.Context) ([]string, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return nil, err
	}
	q, err := s.getQueryer()
	if err != nil {
		return nil, err
	}
	return dd.Databases(ctx, q)
}

// CurrentDatabase 返回当前连接的数据库名称
func (s *SchemaAPI) CurrentDatabase(ctx context.Context) (string, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return "", err
	}
	q, err := s.getQueryer()
	if err != nil {
		return "", err
	}
	return dd.CurrentDatabase(ctx, q)
}

// Tables 返回当前数据库中所有表的名称列表
func (s *SchemaAPI) Tables(ctx context.Context) ([]string, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return nil, err
	}
	q, err := s.getQueryer()
	if err != nil {
		return nil, err
	}
	return dd.Tables(ctx, q)
}

// TableExists 判断指定的表名是否存在
func (s *SchemaAPI) TableExists(ctx context.Context, table string) (bool, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return false, err
	}
	q, err := s.getQueryer()
	if err != nil {
		return false, err
	}
	return dd.TableExists(ctx, q, table)
}

// TableColumns 返回指定表的字段列表
func (s *SchemaAPI) TableColumns(ctx context.Context, table string) ([]string, error) {
	dd, err := s.getDescDialect()
	if err != nil {
		return nil, err
	}
	q, err := s.getQueryer()
	if err != nil {
		return nil, err
	}
	return dd.TableColumns(ctx, q, table)
}

func (s *SchemaAPI) DropTable(ctx context.Context, table string) error {
	qe, err := s.getExecer()
	if err != nil {
		return err
	}
	query := fmt.Sprintf("DROP TABLE %s", s.dialect.QuoteIdentifier(table))
	_, err = Exec(ctx, qe, query)
	return err
}

func (s *SchemaAPI) DropTableIfExists(ctx context.Context, table string) error {
	qe, err := s.getExecer()
	if err != nil {
		return err
	}
	has, err := s.TableExists(ctx, table)
	if err != nil || !has {
		return err
	}
	query := fmt.Sprintf("DROP TABLE %s", s.dialect.QuoteIdentifier(table))
	_, err = Exec(ctx, qe, query)
	return err
}

func (s *SchemaAPI) CreateDatabase(ctx context.Context, db string) error {
	qe, err := s.getExecer()
	if err != nil {
		return err
	}
	query := fmt.Sprintf("CREATE DATABASE %s", s.dialect.QuoteIdentifier(db))
	_, err = Exec(ctx, qe, query)
	return err
}

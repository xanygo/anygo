//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xservice

import (
	"context"
	"time"

	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xpp"
)

var DefaultLoader = &Loader{}

type Loader struct {
	Registry   Registry
	IDC        string
	FlushCycle time.Duration
}

func (l *Loader) getIDC() string {
	if l.IDC != "" {
		return l.IDC
	}
	return xattr.IDC()
}

func (l *Loader) getRegistry() Registry {
	if l.Registry == nil {
		return DefaultRegistry()
	}
	return l.Registry
}

func (l *Loader) getFlushCycle() time.Duration {
	if l.FlushCycle >= time.Second {
		return l.FlushCycle
	}
	return 5 * time.Second
}

func (l *Loader) Load(ctx context.Context, filenames ...string) error {
	var done bool
	var successList []Service
	reg := l.getRegistry()
	defer func() {
		if !done {
			for _, ser := range successList {
				xpp.TryStopWorker(ctx, ser)
			}
		}
	}()
	for _, name := range filenames {
		ser, err := l.loadOne(ctx, name)
		if err == nil {
			successList = append(successList, ser)
		} else {
			return err
		}
	}

	for _, ser := range successList {
		old := reg.Upsert(ser)
		if old != nil {
			xpp.TryStopWorker(ctx, old)
		}
	}
	done = true
	return nil
}

func (l *Loader) loadOne(ctx context.Context, name string) (Service, error) {
	cfg, err := ParserConfigFile(name)
	if err != nil {
		return nil, err
	}
	ser, err := cfg.Parser(l.getIDC())
	if err != nil {
		return nil, err
	}
	err = xpp.TryStartWorker(ctx, l.getFlushCycle(), ser)
	return ser, err
}

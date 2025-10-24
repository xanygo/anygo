//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xservice

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xpp"
)

type Loader struct {
	Registry Registry    // 可选，存储管理 Service 的组件，为 nil 时，使用 efaultRegistry()
	IDC      string      // 可选，当前机房名称，为空时从 xattr.IDC() 取值
	Logger   xlog.Logger // 可选，打印加载日志，为空时日志会输出到黑洞

	watchFiles *xmap.Sync[string, fs.FileInfo]
	once       sync.Once

	reloadWorker *xpp.CycleWorker
	WatchReload  time.Duration // 可选，控制定期检查 service 配置文件是否更新的周期，值 >=1s 时生效
}

func (l *Loader) getIDC() string {
	if l.IDC != "" {
		return l.IDC
	}
	return xattr.IDC()
}

func (l *Loader) getLogger() xlog.Logger {
	if l.Logger != nil {
		return l.Logger
	}
	return &xlog.NopLogger{}
}

func (l *Loader) getRegistry() Registry {
	if l.Registry == nil {
		return DefaultRegistry()
	}
	return l.Registry
}

func (l *Loader) LoadDir(ctx context.Context, patterns ...string) error {
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		files = append(files, matches...)
	}
	return l.Load(ctx, files...)
}

func (l *Loader) initOnce() {
	if l.WatchReload < time.Second {
		return
	}
	l.watchFiles = &xmap.Sync[string, fs.FileInfo]{}
	l.reloadWorker = &xpp.CycleWorker{
		WorkerName: "CheckReloadServiceFile",
		Do:         l.checkReload,
		Cycle:      l.WatchReload,
	}
	l.reloadWorker.Start(context.Background())
}

func (l *Loader) checkReload(ctx context.Context) error {
	ctx = xlog.NewContext(ctx)
	xlog.AddAttr(ctx, xlog.String("action", "checkReload"))

	lg := l.getLogger()
	l.watchFiles.Range(func(name string, value fs.FileInfo) bool {
		info, err := os.Stat(name)
		if err != nil {
			lg.Warn(ctx, err.Error(), xlog.String("fileName", name), xlog.ErrorAttr("error", err))
			return true
		}
		if info.ModTime().After(value.ModTime()) {
			err = l.Load(ctx, name)
			if err == nil {
				l.watchFiles.Store(name, info)
			}
		}
		return true
	})
	return nil
}

func (l *Loader) Load(ctx context.Context, filenames ...string) error {
	l.once.Do(l.initOnce)

	ctx = xlog.NewContext(ctx)
	xlog.AddAttr(ctx, xlog.String("Action", "xservice.Loader.Load"))
	lg := l.getLogger()
	reg := l.getRegistry()

	parserOne := func(ctx context.Context, name string) error {
		ctx = xlog.ForkContext(ctx)
		name = filepath.Clean(name)
		xlog.AddAttr(ctx, xlog.String("fileName", name))
		baseName := filepath.Base(name)
		if strings.HasPrefix(baseName, "_") || strings.HasPrefix(baseName, ".") {
			lg.Info(ctx, "ignored")
			return nil
		}
		info, err := os.Stat(name)
		if err != nil {
			lg.Error(ctx, err.Error())
			return err
		}
		if l.WatchReload >= time.Second && !l.watchFiles.Exists(name) {
			l.watchFiles.Store(name, info)
		}
		ser, err := l.loadOneStart(ctx, name)
		if err != nil {
			lg.Error(ctx, err.Error())
			return err
		}
		old := reg.Upsert(ser)
		if old != nil {
			err = old.Stop(ctx)
			xlog.AddAttr(ctx, xlog.Bool("Upsert", true), xlog.ErrorAttr("stop", err))
		}
		lg.Info(ctx, "loaded", xlog.String("serviceName", ser.Name()))
		return nil
	}

	var wg xsync.WaitGo
	for _, name := range filenames {
		wg.GoErr(func() error {
			return parserOne(ctx, name)
		})
	}
	return wg.Wait()
}

func (l *Loader) loadOneStart(ctx context.Context, name string) (Service, error) {
	cfg, err := ParserConfigFile(name)
	if err != nil {
		return nil, err
	}
	ser, err := cfg.Parser(l.getIDC())
	if err != nil {
		return nil, err
	}

	err = ser.Start(ctx)
	return ser, err
}

func (l *Loader) Stop(ctx context.Context) {
	l.once.Do(l.initOnce)
	l.reloadWorker.Stop(ctx)
}

var defaultLoader = &xsync.OnceInit[*Loader]{
	New: func() *Loader {
		return &Loader{
			Logger:      xlog.Default(),
			WatchReload: time.Second,
		}
	},
}

func DefaultLoader() *Loader {
	return defaultLoader.Load()
}

func LoadDir(ctx context.Context, patterns ...string) error {
	return DefaultLoader().LoadDir(ctx, patterns...)
}

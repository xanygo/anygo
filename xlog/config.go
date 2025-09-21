//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xlog

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xio/xfs"
)

var _ DispatchWriter = (*FileConfig)(nil)

// FileConfig 文件配置信息
type FileConfig struct {
	// FileName 日志文件名，必填，可用环境变量配置，如 {xattr.LogDir}/access/access.log
	FileName string `yaml:"FileName"`

	// ExtRule 切分规则，可选，如 no,1hour,1day 等，默认为 1hour
	// 全部支持的规则详见 xio/xfc.ExtRules
	ExtRule string `yaml:"ExtRule"`

	// MaxFiles 切分后保留的文件个数，可选，默认为 48
	MaxFiles int `yaml:"MaxFiles"`

	// MaxDelay 日志内容落盘最大延迟时间，可选，如 1s
	// 需要支持 time.ParseDuration 方法解析
	MaxDelay string `yaml:"MaxDelay"`

	// Dispatch 日志按等级分发写入不同后缀的日志文件的规则，可选
	// 若不配置，所有等级日志写入一份日志文件
	Dispatch []FileConfigDispatch `yaml:"Dispatch"`

	once xsync.OnceDoValueErr[map[Level]io.WriteCloser]
}

func (fc *FileConfig) getMaxFiles() int {
	if fc.MaxFiles <= 0 {
		return 48
	}
	return fc.MaxFiles
}

func (fc *FileConfig) getExtRule() string {
	if fc.ExtRule == "" {
		return "1hour"
	}
	return fc.ExtRule
}

func (fc *FileConfig) getMaxDelay() time.Duration {
	d, _ := time.ParseDuration(fc.MaxDelay)
	if d > time.Millisecond {
		return d
	}
	return 50 * time.Millisecond
}

func (fc *FileConfig) init() (map[Level]io.WriteCloser, error) {
	return fc.once.Do(fc.initWriters)
}

func (fc *FileConfig) initWriters() (map[Level]io.WriteCloser, error) {
	if fc.FileName == "" {
		return nil, errors.New("empty log FileName")
	}
	result := make(map[Level]io.WriteCloser, len(allLevels))
	if len(fc.Dispatch) == 0 {
		rw := &xfs.Rotator{
			Path:     fc.FileName,
			ExtRule:  fc.getExtRule(),
			MaxFiles: fc.getMaxFiles(),
			MaxDelay: fc.getMaxDelay(),
		}
		for _, l := range allLevels {
			result[l] = rw
		}
		return result, nil
	}
	files := make(map[string]bool, len(fc.Dispatch))
	for _, item := range fc.Dispatch {
		fp := item.getFilePath(fc.FileName)
		if files[fp] {
			return nil, fmt.Errorf("duplicated log file path %q with Dispatch.Ext=%q", fp, item.Ext)
		}
		files[fp] = true
		rw := &xfs.Rotator{
			Path:     fp,
			ExtRule:  fc.getExtRule(),
			MaxFiles: fc.getMaxFiles(),
			MaxDelay: fc.getMaxDelay(),
		}
		for _, l := range item.getLevels() {
			if _, has := result[l]; has {
				continue
			}
			result[l] = rw
		}
	}
	return result, nil
}

func (fc *FileConfig) Writers() map[Level]io.WriteCloser {
	ws, err := fc.initWriters()
	if err != nil {
		panic(err)
	}
	return ws
}

type FileConfigDispatch struct {
	Ext   string  `yaml:"Ext"`
	Level []Level `yaml:"Level"`
}

func (fc FileConfigDispatch) getFilePath(fp string) string {
	if fc.Ext == "" {
		return fp
	}
	return fp + fc.Ext
}

func (fc FileConfigDispatch) getLevels() []Level {
	if len(fc.Level) == 0 {
		return allLevels
	}
	return fc.Level
}

var defaultDispatch = []FileConfigDispatch{
	{
		Ext:   "",
		Level: []Level{LevelInfo},
	},
	{
		Ext:   ".debug",
		Level: []Level{LevelDebug},
	},
	{
		Ext:   ".wf",
		Level: []Level{LevelWarn, LevelError},
	},
}

func ParserFileConfig(fp string) (*FileConfig, error) {
	obj := &FileConfig{}
	if err := xcfg.Parse(fp, obj); err != nil {
		return nil, err
	}
	_, err := obj.init()
	if err != nil {
		return nil, err
	}
	return nil, err
}

type FileLoggerOpt struct {
	CfgPath    string
	Cfg        *FileConfig
	NewHandler NewHandlerFunc
}

func (fo FileLoggerOpt) NewLogger() (Logger, error) {
	fn := fo.NewHandler
	if fn == nil {
		fn = defaultJSONHandler
	}
	if fo.CfgPath != "" && xcfg.Exists(fo.CfgPath) {
		return NewFileLogger(fo.CfgPath, fn)
	}
	if fo.Cfg == nil {
		return nil, errors.New("both CfgPath and Cfg all empty")
	}
	ws, err := fo.Cfg.init()
	if err != nil {
		return nil, err
	}
	return &Simple{
		Handler: NewDispatchHandler(ws, fn),
	}, nil
}

func (fo FileLoggerOpt) MustNewLogger() Logger {
	lg, err := fo.NewLogger()
	if err == nil {
		return lg
	}
	panic(err)
}

func NewFileLogger(fp string, fn NewHandlerFunc) (Logger, error) {
	fc, err := ParserFileConfig(fp)
	if err != nil {
		return nil, err
	}
	if fn == nil {
		fn = defaultJSONHandler
	}
	return &Simple{
		Handler: NewDispatchHandler(fc.Writers(), fn),
	}, nil
}

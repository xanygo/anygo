//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xservice

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xbus"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xnaming"
	"github.com/xanygo/anygo/xnet/xproxy"
)

type Config struct {
	Name string `json:"Name" yaml:"Name" validator:"required"`

	ConnectTimeout   int64  `json:"ConnectTimeout"   yaml:"ConnectTimeout"`   // 连接超时,可选，单位毫秒
	ConnectRetry     int    `json:"ConnectRetry"     yaml:"ConnectRetry"`     // 连接重试次数，默认为 0
	WriteTimeout     int64  `json:"WriteTimeout"     yaml:"WriteTimeout"`     // 写超时时间，单位毫秒
	ReadTimeout      int64  `json:"ReadTimeout"      yaml:"ReadTimeout"`      // 读超时时间，单位毫秒
	HandshakeTimeout int64  `json:"HandshakeTimeout" yaml:"HandshakeTimeout"` // 握手超时时间，单位毫秒
	Protocol         string `json:"Protocol"         yaml:"Protocol"`         // 交互协议
	WorkerCycle      string `json:"WorkerCycle"         yaml:"WorkerCycle"`   // 后台任务运行周期

	UseProxy string         `json:"UseProxy" yaml:"UseProxy"`       // 将另外一个service 当做代理
	Proxy    *xproxy.Config `json:"Proxy"             yaml:"Proxy"` // 当子服务是代理时使用

	Retry           int                `json:"Retry"             yaml:"Retry"`
	MaxResponseSize int64              `json:"MaxResponseSize"   yaml:"MaxResponseSize"`
	HTTP            *HTTPPart          `json:"HTTP"              yaml:"HTTP"`
	ConnPool        *ConnPoolPart      `json:"ConnPool"          yaml:"ConnPool"`
	TLS             *xoption.TLSConfig `json:"TLS"               yaml:"TLS"`
	DownStream      DownStreamPart     `json:"DownStream"        yaml:"DownStream" validator:"required,dive,required"`

	Extra map[string]any // 其他字段
}

var _ xcodec.DecodeExtra = (*Config)(nil)

func (c *Config) NeedDecodeExtra() string {
	return "Extra"
}

// ConnPoolPart 连接池配置参数
type ConnPoolPart struct {
	Name            string `json:"Name" yaml:"Name"`                       // 连接池名称，可选，默认为 Short,可选 Long
	MaxOpen         int    `json:"MaxOpen" yaml:"MaxOpen"`                 // 最大打开数量,<= 0 为不限制
	MaxIdle         int    `json:"MaxIdle" yaml:"MaxIdle"`                 // 最大空闲数，应 <= MaxOpen,<=0 为不允许存在 Idle 元素
	MaxLifeTime     int    `json:"MaxLifeTime" yaml:"MaxLifeTime"`         // 最大使用时长,单位毫秒，超过后将被销毁, <=0 为不限制
	MaxIdleTime     int    `json:"MaxIdleTime" yaml:"MaxIdleTime"`         // 最大空闲等待时间,单位毫秒，超过后将被销毁, <=0 为不限制
	MaxPoolIdleTime int    `json:"MaxPoolIdleTime" yaml:"MaxPoolIdleTime"` // 单位毫秒，当超过此时长未被使用后,关闭并清理对应的 Pool,<=0 时使用默认值 10 minute
}

func (cp *ConnPoolPart) GetName() string {
	if cp == nil || cp.Name == "" {
		return xdial.Short
	}
	return cp.Name
}

func (cp *ConnPoolPart) GetOption() xpool.Option {
	if cp == nil {
		return xpool.Option{}
	}
	return xpool.Option{
		MaxOpen:         cp.MaxOpen,
		MaxIdle:         cp.MaxIdle,
		MaxLifeTime:     time.Duration(cp.MaxLifeTime) * time.Millisecond,
		MaxIdleTime:     time.Duration(cp.MaxIdleTime) * time.Millisecond,
		MaxPoolIdleTime: time.Duration(cp.MaxPoolIdleTime) * time.Millisecond,
	}
}

type DownStreamPart struct {
	LoadBalancer string                       `json:"LoadBalancer" yaml:"LoadBalancer"`
	Address      []string                     `json:"Address" yaml:"Address"`
	IDC          map[string]DownStreamIDCPart `json:"IDC" yaml:"IDC"`
}

func (c *DownStreamPart) getIDCAddress(idc string) []string {
	if len(c.IDC) == 0 {
		return nil
	}
	return c.IDC[idc].Address
}

type DownStreamIDCPart struct {
	Address []string `json:"Address" yaml:"Address" validator:"required,dive,required"`
}

type HTTPPart struct {
	Host   string      `json:"Host" yaml:"Host"` // 主机名，可选
	Header http.Header `json:"Header" yaml:"Header"`
}

func (ho *HTTPPart) Clone() *HTTPPart {
	return &HTTPPart{
		Host:   ho.Host,
		Header: ho.Header.Clone(),
	}
}

// Parser 解析为 Service 类型（需要Start 后才能使用）
func (c *Config) Parser(idc string) (Service, error) {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return nil, errors.New("name is empty")
	}
	opt := xoption.NewDynamic()
	xoption.SetConnectTimeout(opt, time.Duration(c.ConnectTimeout)*time.Millisecond)
	xoption.SetConnectRetry(opt, c.ConnectRetry)
	xoption.SetWriteTimeout(opt, time.Duration(c.WriteTimeout)*time.Millisecond)
	xoption.SetReadTimeout(opt, time.Duration(c.ReadTimeout)*time.Millisecond)
	xoption.SetHandshakeTimeout(opt, time.Duration(c.HandshakeTimeout)*time.Millisecond)
	xoption.SetRetry(opt, c.Retry)
	xoption.SetMaxResponseSize(opt, c.MaxResponseSize)
	SetOptConnPool(opt, c.ConnPool)
	xoption.SetProtocol(opt, c.Protocol)
	if c.TLS != nil {
		tc, err := c.TLS.Parser()
		if err != nil {
			return nil, err
		}
		xoption.SetTLSConfig(opt, tc)
	}

	if c.UseProxy != "" {
		if c.UseProxy == c.Name {
			return nil, fmt.Errorf("invalid UseProxy=%q for service %q", c.UseProxy, c.Name)
		}
		xoption.SetUseProxy(opt, c.UseProxy)
	}
	if c.WorkerCycle != "" {
		cycle, err := time.ParseDuration(c.WorkerCycle)
		if err != nil {
			return nil, fmt.Errorf("invalid WorkerCycle=%q for service %q", c.WorkerCycle, c.Name)
		}
		if cycle < 100*time.Millisecond {
			return nil, fmt.Errorf("invalid WorkerCycle=%q for service %q", c.WorkerCycle, c.Name)
		}
		xoption.SetWorkerCycle(opt, cycle)
	}

	if c.Proxy != nil {
		xproxy.SetOptConfig(opt, c.Proxy)
	}
	if c.HTTP != nil {
		SetOptHTTP(opt, *c.HTTP)
	}

	for k, v := range c.Extra {
		xoption.SetExtra(opt, k, v)
	}

	impl := &serviceImpl{
		broker: xbus.NewBroker(),
		name:   c.Name,
		opt:    opt,
	}

	primaryAddress := c.DownStream.getIDCAddress(idc)
	fallbackAddress := c.DownStream.Address

	if len(primaryAddress) == 0 && len(fallbackAddress) == 0 {
		return nil, errors.New("empty downstream address list")
	}

	ap, err := xbalance.New(c.DownStream.LoadBalancer)
	if err != nil {
		return nil, err
	}
	impl.broker.MustRegisterConsumer(xnaming.Topic, ap)
	impl.balancer = ap

	impl.connector = &connector{}

	poolOpt := c.ConnPool.GetOption()
	pool, err := xdial.NewGroupPool(c.ConnPool.GetName(), &poolOpt, impl.connector)
	if err != nil {
		return nil, err
	}
	impl.pool = pool

	nw, err := xnaming.NewWorker(idc, xoption.WorkerCycle(opt), primaryAddress, fallbackAddress)
	if err != nil {
		return nil, err
	}
	impl.nw = nw
	impl.broker.RegisterProducer(nw)
	return impl, nil
}

func ParserConfigFile(path string) (*Config, error) {
	cfg := &Config{}
	if err := xcfg.Parse(path, cfg); err != nil {
		return nil, err
	}
	data := map[string]any{}
	if err := xcfg.Parse(path, &data); err != nil {
		return nil, err
	}
	baseName := filepath.Base(path)
	pureName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	if pureName != cfg.Name {
		return nil, fmt.Errorf("service Name expected %q, got %q", pureName, cfg.Name)
	}
	return cfg, nil
}

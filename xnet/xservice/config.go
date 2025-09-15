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

	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xnaming"
	"github.com/xanygo/anygo/xnet/xproxy"
	"github.com/xanygo/anygo/xoption"
)

type Config struct {
	Name            string         `json:"Name" yaml:"Name" validator:"required"`
	ConnectTimeout  int64          `json:"ConnectTimeout" yaml:"ConnectTimeout"` // 连接超时,可选
	ConnectRetry    int            `json:"ConnectRetry" yaml:"ConnectRetry"`
	WriteTimeout    int64          `json:"WriteTimeout" yaml:"WriteTimeout"`
	ReadTimeout     int64          `json:"ReadTimeout" yaml:"ReadTimeout"`
	Retry           int            `json:"Retry" yaml:"Retry"`
	MaxResponseSize int64          `json:"MaxResponseSize" yaml:"MaxResponseSize"`
	Proxy           *xproxy.Config `json:"Proxy" yaml:"Proxy"`
	HTTP            *HTTPPart      `json:"HTTP" yaml:"HTTP"`
	TLS             *xoption.TLSConfig
	DownStream      DownStreamPart `json:"DownStream" yaml:"DownStream" validator:"required,dive,required"`
}

type DownStreamPart struct {
	LoadBalancer string   `json:"LoadBalancer" yaml:"LoadBalancer"`
	Address      []string `json:"Address" yaml:"Address"`
	IDC          map[string]DownStreamIDCPart
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
	Host   string      `json:"Host" yaml:"Host"`   // 主机名，可选
	HTTPS  bool        `json:"HTTPS" yaml:"HTTPS"` // 是否发起 HTTPS 请求，可选，默认 false
	Header http.Header `json:"Header" yaml:"Header"`
}

func (ho *HTTPPart) Clone() *HTTPPart {
	return &HTTPPart{
		Host:   ho.Host,
		HTTPS:  ho.HTTPS,
		Header: ho.Header.Clone(),
	}
}

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
	xoption.SetRetry(opt, c.Retry)
	xoption.SetMaxResponseSize(opt, c.MaxResponseSize)
	if c.TLS != nil {
		tc, err := c.TLS.Parser()
		if err != nil {
			return nil, err
		}
		xoption.SetTLSConfig(opt, tc)
	}
	if c.Proxy != nil {
		xproxy.SetOptConfig(opt, c.Proxy)
	}
	if c.HTTP != nil {
		SetOptHTTP(opt, *c.HTTP)
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
	nw, err := xnaming.NewWorker(idc, primaryAddress, fallbackAddress)
	if err != nil {
		return nil, err
	}
	impl.nw = nw
	impl.broker.RegisterProducer(nw)
	return impl, nil
}

func ParserConfigFile(path string) (*Config, error) {
	var cfg Config
	if err := xcfg.Parse(path, &cfg); err != nil {
		return nil, err
	}
	baseName := filepath.Base(path)
	pureName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	if pureName != cfg.Name {
		return nil, fmt.Errorf("service Name expected %q, got %q", pureName, cfg.Name)
	}
	return &cfg, nil
}

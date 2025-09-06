//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xservice

import (
	"errors"
	"strings"
	"time"

	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xnaming"
)

type Config struct {
	Name           string           `json:"Name" yaml:"Name" validator:"required"`
	ConnectTimeout int64            `json:"ConnectTimeout" yaml:"ConnectTimeout"`
	ConnectRetry   int              `json:"ConnectRetry" yaml:"ConnectRetry"`
	WriteTimeout   int64            `json:"WriteTimeout" yaml:"WriteTimeout"`
	ReadTimeout    int64            `json:"ReadTimeout" yaml:"ReadTimeout"`
	Retry          int              `json:"Retry" yaml:"Retry"`
	HTTP           HTTPOption       `json:"HTTP" yaml:"HTTP"`
	DownStream     ConfigDownStream `json:"DownStream" yaml:"DownStream" validator:"required,dive,required"`
}

type ConfigDownStream struct {
	LoadBalancer string   `json:"LoadBalancer" yaml:"LoadBalancer"`
	Address      []string `json:"Address" yaml:"Address"`
	IDC          map[string]ConfigDownStreamIDC
}

func (c *ConfigDownStream) getIDCAddress(idc string) []string {
	if len(c.IDC) == 0 {
		return nil
	}
	return c.IDC[idc].Address
}

type ConfigDownStreamIDC struct {
	Address []string `json:"Address" yaml:"Address" validator:"required,dive,required"`
}

func (c *Config) Parser(idc string) (Service, error) {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return nil, errors.New("name is empty")
	}
	impl := &serviceImpl{
		broker: xbus.NewBroker(),
		name:   c.Name,
		opt: Option{
			ConnectTimeout: time.Duration(c.ConnectTimeout) * time.Millisecond,
			ConnectRetry:   c.ConnectRetry,
			WriteTimeout:   time.Duration(c.WriteTimeout) * time.Millisecond,
			ReadTimeout:    time.Duration(c.ReadTimeout) * time.Millisecond,
			Retry:          c.Retry,
			HTTP:           c.HTTP.Clone(),
		},
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
	return &cfg, nil
}

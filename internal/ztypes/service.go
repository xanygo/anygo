//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package ztypes

import (
	"encoding/json"
	"errors"
	"strings"
)

// 这里定义 xnet/xservice 相关 package 需要使用的类型

type ServiceCommand struct {
	Path string
	Args []string
	Dir  string
}

func (sc *ServiceCommand) LoadFromStr(str string) error {
	err := json.Unmarshal([]byte(str), &sc)
	if err != nil {
		return err
	}
	sc.Path = strings.TrimSpace(sc.Path)
	if sc.Path == "" {
		return errors.New("empty command Path")
	}
	sc.Dir = strings.TrimSpace(sc.Dir)
	return nil
}

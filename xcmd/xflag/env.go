//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-05

package xflag

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// EnvStringVar 优先使用环境变量作为默认值
func EnvStringVar(p *string, name string, envKey string, value string, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		value = ev
	}
	usage += envUsage(envKey, ev)
	flag.StringVar(p, name, value, usage)
}

func envUsage(key string, value string) string {
	return fmt.Sprintf(" [env %q = %q]", key, value)
}

func EnvString(name string, envKey string, value string, usage string) *string {
	p := new(string)
	EnvStringVar(p, name, envKey, value, usage)
	return p
}

func EnvIntVar(p *int, name string, envKey string, value int, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.Atoi(ev)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += envUsage(envKey, ev)
	flag.IntVar(p, name, value, usage)
}

func EnvInt(name string, envKey string, value int, usage string) *int {
	p := new(int)
	EnvIntVar(p, name, envKey, value, usage)
	return p
}

func EnvInt64Var(p *int64, name string, envKey string, value int64, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseInt(ev, 10, 64)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += envUsage(envKey, ev)
	flag.Int64Var(p, name, value, usage)
}

func EnvInt64(name string, envKey string, value int64, usage string) *int64 {
	p := new(int64)
	EnvInt64Var(p, name, envKey, value, usage)
	return p
}

func EnvUint64Var(p *uint64, name string, envKey string, value uint64, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseUint(ev, 10, 64)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += envUsage(envKey, ev)
	flag.Uint64Var(p, name, value, usage)
}

func EnvUint64(name string, envKey string, value uint64, usage string) *uint64 {
	p := new(uint64)
	EnvUint64Var(p, name, envKey, value, usage)
	return p
}

func EnvUintVar(p *uint, name string, envKey string, value uint, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseUint(ev, 10, strconv.IntSize)
		if err == nil {
			value = uint(nv)
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += envUsage(envKey, ev)
	flag.UintVar(p, name, value, usage)
}

func EnvUint(name string, envKey string, value uint, usage string) *uint {
	p := new(uint)
	EnvUintVar(p, name, envKey, value, usage)
	return p
}

func EnvDurationVar(p *time.Duration, name string, envKey string, value time.Duration, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := time.ParseDuration(ev)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += envUsage(envKey, ev)
	flag.DurationVar(p, name, value, usage)
}

func EnvDuration(name string, envKey string, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	EnvDurationVar(p, name, envKey, value, usage)
	return p
}

func EnvBoolVar(p *bool, name string, envKey string, value bool, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseBool(ev)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += envUsage(envKey, ev)
	flag.BoolVar(p, name, value, usage)
}

func EnvBool(name string, envKey string, value bool, usage string) *bool {
	p := new(bool)
	EnvBoolVar(p, name, envKey, value, usage)
	return p
}

func EnvFloat64Var(p *float64, name string, envKey string, value float64, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseFloat(ev, 64)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += envUsage(envKey, ev)
	flag.Float64Var(p, name, value, usage)
}

func EnvFloat64(name string, envKey string, value float64, usage string) *float64 {
	p := new(float64)
	EnvFloat64Var(p, name, envKey, value, usage)
	return p
}

func EnvFloat32Var(p *float32, name string, envKey string, value float32, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseFloat(ev, 32)
		if err == nil {
			value = float32(nv)
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += envUsage(envKey, ev)
	Float32Var(p, name, value, usage)
}

func EnvFloat32(name string, envKey string, value float32, usage string) *float32 {
	p := new(float32)
	EnvFloat32Var(p, name, envKey, value, usage)
	return p
}

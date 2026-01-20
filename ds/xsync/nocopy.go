//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-20

package xsync

type noCopy struct{}

func (*noCopy) Lock() {}

func (*noCopy) Unlock() {}

//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-05

package xmetric

import (
	"context"

	"github.com/xanygo/anygo/ds/xsync"
)

var defaultProvider = xsync.OnceInit[Provider]{
	New: func() Provider {
		return &InnerProvider{}
	},
}

func DefaultProvider() Provider {
	return defaultProvider.Load()
}

func SetDefaultProvider(p Provider) {
	defaultProvider.Store(p)
}

type Provider interface {
	Tracer(name string) Tracer
}

type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

func Start(ctx context.Context, name string) (context.Context, Span) {
	return DefaultProvider().Tracer("default").Start(ctx, name)
}

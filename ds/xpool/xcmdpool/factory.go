//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-03

package xcmdpool

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xlog"
)

var _ xpool.Factory[*command] = (*commandFactory)(nil)

type commandFactory struct {
	P *Command
}

func (c *commandFactory) New(ctx context.Context) (*command, error) {
	cmd := exec.CommandContext(c.P.rootCtx, c.P.Path, c.P.Args...)

	logCtx := xlog.NewContext(ctx)
	xlog.AddMetaAttr(logCtx, xlog.String("Command", c.P.Path))
	cmd.Stderr = xlog.AsWriter(logCtx, xlog.Default(), xlog.LevelInfo)

	if c.P.Setup != nil {
		c.P.Setup(cmd)
	}
	pw, err1 := cmd.StdinPipe()
	if err1 != nil {
		return nil, fmt.Errorf("get StdinPipe failed: %w", err1)
	}
	pr, err2 := cmd.StdoutPipe()
	if err2 != nil {
		return nil, fmt.Errorf("get StdoutPipe failed: %w", err2)
	}
	err := cmd.Start()
	if err != nil {
		if xattr.IsDebugMode() {
			xlog.Warn(ctx, "command start failed", xlog.String("cmd", cmd.String()), xlog.ErrorAttr("error", err))
		}
		return nil, err
	}

	if xattr.IsDebugMode() {
		xlog.Debug(ctx, "command started", xlog.String("cmd", cmd.String()), xlog.Int("pid", cmd.Process.Pid))
	}

	rw := &readWriter{
		r: bufio.NewReader(pr),
		w: pw,
	}
	// time.Sleep(10 * time.Second)
	go func() {
		_ = cmd.Wait()
		_ = pw.Close()
		_ = pr.Close()
	}()

	nc := &command{
		cmd: cmd,
		rw:  rw,
	}
	return nc, nil
}

var _ io.ReadWriter = (*readWriter)(nil)

type readWriter struct {
	w io.Writer
	r *bufio.Reader
}

func (rw *readWriter) Read(p []byte) (n int, err error) {
	return rw.r.Read(p)
}

func (rw *readWriter) Write(p []byte) (n int, err error) {
	return rw.w.Write(p)
}

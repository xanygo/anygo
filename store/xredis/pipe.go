//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-15

package xredis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xoption"
)

// TxPipelined 使用 MULTI + EXEC 批量执行
func (c *Client) TxPipelined(ctx context.Context, fn func(ctx context.Context, pipe *Pipeline) error) ([]Cmder, error) {
	pip := c.NewTxPipeline()
	err := fn(ctx, pip)
	if err != nil {
		return nil, err
	}
	err = pip.Exec(ctx)
	return pip.cmds, err
}

func (c *Client) NewTxPipeline() *Pipeline {
	pipe := &Pipeline{
		client: c,
		tx:     true,
	}
	pipe.init()
	return pipe
}

// Pipelined 批量执行，非事务，不需要主动调用 Exec 方法提交。任意一条命令失败，不会影响其他命令。
func (c *Client) Pipelined(ctx context.Context, fn func(ctx context.Context, pipe *Pipeline) error) ([]Cmder, error) {
	pip := c.NewPipeline()
	err := fn(ctx, pip)
	if err != nil {
		return nil, err
	}
	err = pip.Exec(ctx)
	return pip.cmds, err
}

// NewPipeline  批量执行，非事务，需要需要主动调用 Exec 方法提交。任意一条命令失败，不会影响其他命令。
func (c *Client) NewPipeline() *Pipeline {
	pipe := &Pipeline{
		client: c,
	}
	pipe.init()
	return pipe
}

var errPipeNotExecuted = errors.New("pipeline has not been executed")

// Pipeline 批量命令
type Pipeline struct {
	tx     bool
	cmds   []Cmder
	client *Client
}

func (pipe *Pipeline) init() {}

func (pipe *Pipeline) NewAnyCmd(args ...any) *AnyCmd {
	cmd := NewAnyCmd(args...)
	pipe.addCmd(cmd)
	return cmd
}

func (pipe *Pipeline) addCmd(cmds ...Cmder) {
	for _, c := range cmds {
		c.SetErr(errPipeNotExecuted)
	}
	pipe.cmds = append(pipe.cmds, cmds...)
}

// Process 将命令加入队列
func (pipe *Pipeline) Process(ctx context.Context, cmds ...Cmder) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
	}
	pipe.addCmd(cmds...)
	return nil
}

func (pipe *Pipeline) Exec(ctx context.Context) error {
	if len(pipe.cmds) == 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
	}
	req := &pipeRequest{
		tx:   pipe.tx,
		cmds: pipe.cmds,
	}
	resp := &pipeResponse{
		tx:   pipe.tx,
		cmds: pipe.cmds,
	}
	err := pipe.client.invoke(ctx, req, resp)
	return err
}

func (pipe *Pipeline) Cmds() []Cmder {
	return pipe.cmds
}

var _ xrpc.Request = (*pipeRequest)(nil)

type pipeRequest struct {
	tx   bool
	cmds []Cmder
}

func (req *pipeRequest) String() string {
	return "pipeRequest"
}

func (req *pipeRequest) Protocol() string {
	return Protocol
}

func (req *pipeRequest) APIName() string {
	if req.tx {
		return "TxPipeline"
	}
	return "Pipeline"
}

func (req *pipeRequest) WriteTo(ctx context.Context, w *xnet.ConnNode, opt xoption.Reader) error {
	if req.tx {
		return req.writeTx(w)
	}
	bf := bp.Get()
	defer bp.Put(bf)

	for i := 0; i < len(req.cmds); i++ {
		mr := resp3.NewRequest(resp3.DataTypeAny, req.cmds[i].Args()...)
		bf.Reset()
		_, err := w.Write(mr.Bytes(bf))
		if err != nil {
			return err
		}
	}
	return nil
}

func (req *pipeRequest) writeTx(w *xnet.ConnNode) error {
	bf := bp.Get()
	defer bp.Put(bf)
	mr := resp3.NewRequest(resp3.DataTypeSimpleString, "MULTI")
	bf.Reset()
	_, err := w.Write(mr.Bytes(bf))
	if err != nil {
		return err
	}
	rd := bufio.NewReader(w)
	reply, err := resp3.ReadByType(rd, resp3.DataTypeSimpleString)
	if err = resp3.ToOkStatus(reply, err); err != nil {
		return err
	}
	for i := 0; i < len(req.cmds); i++ {
		mr = resp3.NewRequest(resp3.DataTypeAny, req.cmds[i].Args()...)
		bf.Reset()
		_, err = w.Write(mr.Bytes(bf))
		if err != nil {
			return err
		}
		reply, err = resp3.ReadByType(rd, resp3.DataTypeSimpleString)
		str, err := resp3.ToString(reply, err)
		if err != nil {
			return err
		}
		if str != "QUEUED" {
			return fmt.Errorf("invalid reply %q, expect is 'QUEUED'", str)
		}
	}
	mr = resp3.NewRequest(resp3.DataTypeAny, "EXEC")
	bf.Reset()
	_, err = w.Write(mr.Bytes(bf))
	return err
}

var _ xrpc.Response = (*pipeResponse)(nil)

type pipeResponse struct {
	tx   bool
	cmds []Cmder
	err  error
}

func (resp *pipeResponse) String() string {
	return "pipeResponse"
}

func (resp *pipeResponse) LoadFrom(ctx context.Context, req xrpc.Request, rd io.Reader, opt xoption.Reader) error {
	br := bufio.NewReader(rd)
	if resp.tx {
		return resp.readTx(ctx, br)
	}

	for _, cmd := range resp.cmds {
		reply, err := resp3.ReadOneElement(br)
		cmd.SetReply(reply, err)
	}
	return nil
}

func (resp *pipeResponse) readTx(ctx context.Context, rd resp3.Reader) error {
	reply, err := resp3.ReadByType(rd, resp3.DataTypeArray)
	if err != nil {
		return err
	}
	arr, ok := reply.(resp3.Array)
	if !ok {
		return fmt.Errorf("invalid type: %T, expect array", reply)
	}
	if len(arr) != len(resp.cmds) {
		return fmt.Errorf("invalid cmds reply length: %d != %d", len(resp.cmds), len(arr))
	}
	for i := 0; i < len(resp.cmds); i++ {
		resp.cmds[i].SetReply(arr[i], nil)
	}
	return nil
}

func (resp *pipeResponse) ErrCode() int64 {
	return xerror.ErrCode(resp.err, 0)
}

func (resp *pipeResponse) ErrMsg() string {
	if resp.err == nil {
		return ""
	}
	return resp.err.Error()
}

func (resp *pipeResponse) Unwrap() any {
	return nil
}

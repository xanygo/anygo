//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-12-30

package xredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) ACLCat(ctx context.Context, category string) ([]string, error) {
	args := []any{"ACL", "CAT"}
	if category != "" {
		args = append(args, category)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

func (c *Client) ACLDelUser(ctx context.Context, usernames ...string) (int64, error) {
	if len(usernames) == 0 {
		return 0, errors.New("usernames is empty")
	}
	args := []any{"ACL", "DELUSER"}
	args = xslice.Append(args, usernames...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) ACLDryRun(ctx context.Context, username string, command string, arg ...any) error {
	args := []any{"ACL", "DRYRUN", username, command}
	args = xslice.Append(args, arg...)
	cmd := resp3.NewRequest(resp3.DataTypeAny, args...)
	resp := c.do(ctx, cmd)
	if resp.err != nil {
		return resp.err
	}
	switch rv := resp.result.(type) {
	case resp3.SimpleString:
		return resp3.ToOkStatus(rv, nil)
	case resp3.BulkString:
		return errors.New(rv.String())
	default:
		return fmt.Errorf("invalid reply type: %T", rv)
	}
}

func (c *Client) ACLGenPass(ctx context.Context, bits int) (string, error) {
	args := []any{"ACL", "GENPASS"}
	if bits > 0 {
		args = append(args, bits)
	}
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) ACLGetUser(ctx context.Context, username string) (map[string]any, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "ACL", "GETUSER", username)
	resp := c.do(ctx, cmd)
	return resp3.ToStringAnyMap(resp.result, resp.err)
}

func (c *Client) ACLList(ctx context.Context) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ACL", "LIST")
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

func (c *Client) ACLLoad(ctx context.Context) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "ACL", "LOAD")
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) ACLLog(ctx context.Context, count int) ([]map[string]any, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ACL", "LOG", count)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(0)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(arr))
	for _, v := range arr {
		ma, err := resp3.ToStringAnyMap(v, nil)
		if err != nil {
			return nil, err
		}
		result = append(result, ma)
	}
	return result, nil
}

func (c *Client) ACLLogReset(ctx context.Context) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "ACL", "LOG", "RESET")
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) ACLSave(ctx context.Context) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "ACL", "SAVE")
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) ACLSetUser(ctx context.Context, username string, rules ...string) error {
	if len(rules) == 0 {
		return errors.New("rules is empty")
	}
	args := []any{"ACL", "SETUSER", username}
	args = xslice.Append(args, rules...)
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) ACLUsers(ctx context.Context) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ACL", "USERS")
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

func (c *Client) ACLWhoAmI(ctx context.Context) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "ACL", "WHOAMI")
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

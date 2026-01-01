//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-01

package xredis

// https://redis.io/docs/latest/commands/cms.incrby/

import (
	"context"
	"errors"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) CMSIncrBy(ctx context.Context, key string, item string, increment int64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "CMS.INCRBY", key, item, increment)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(1)
	if err != nil {
		return 0, err
	}
	return resp3.ToInt64(arr[0], nil)
}

func (c *Client) CMSIncrByN(ctx context.Context, key string, items ...ItemIncrement) ([]int64, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"CMS.INCRBY", key}
	for _, item := range items {
		args = append(args, item.Item, item.Increment)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(len(items))
	if err != nil {
		return nil, err
	}
	result := make([]int64, len(arr))
	for i, ret := range arr {
		result[i], err = resp3.ToInt64(ret, err)
	}
	return result, err
}

type ItemIncrement struct {
	Item      string
	Increment int64
}

func (c *Client) CMSInfo(ctx context.Context, key string) (CMSInfo, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "CMS.INFO", key)
	resp := c.do(ctx, cmd)
	mp, err := resp3.ToMap(resp.result, resp.err)
	if err != nil {
		return CMSInfo{}, err
	}
	info := CMSInfo{}
	for k, v := range mp {
		ks, err := resp3.ToString(k, nil)
		if err != nil {
			return info, err
		}
		switch ks {
		case "width":
			info.Width, err = resp3.ToInt64(v, nil)
		case "depth":
			info.Depth, err = resp3.ToInt64(v, nil)
		case "count":
			info.Count, err = resp3.ToInt64(v, nil)
		}
		if err != nil {
			return info, err
		}
	}
	return info, nil
}

type CMSInfo struct {
	Width int64
	Depth int64
	Count int64
}

func (c *Client) CMSInitByDim(ctx context.Context, key string, width int64, depth int64) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "CMS.INITBYDIM", key, width, depth)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) CMSInitByProb(ctx context.Context, key string, error float64, probability float64) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "CMS.INITBYPROB", key, error, probability)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) CMSMerge(ctx context.Context, destination string, sources ...string) error {
	if len(sources) == 0 {
		return errNoItems
	}
	args := []any{"CMS.MERGE", destination, len(sources)}
	args = xslice.Append(args, sources...)
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) CMSMergeWithWeight(ctx context.Context, destination string, sources []string, weights []int64) error {
	if len(sources) == 0 {
		return errNoItems
	}
	if len(weights) == 0 {
		return errors.New("no source")
	}
	if len(sources) != len(weights) {
		return errors.New("sources.len!=weights.len")
	}
	args := []any{"CMS.MERGE", destination, len(sources)}
	args = xslice.Append(args, sources...)
	args = append(args, "WEIGHTS")
	args = xslice.Append(args, weights...)
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) CMSQuery(ctx context.Context, key string, items ...string) ([]int64, error) {
	if len(items) == 0 {
		return nil, errNoItems
	}
	args := []any{"CMS.QUERY", key}
	args = xslice.Append(args, items...)
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64Slice(resp.result, resp.err, len(items))
}

//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-02

package xredis

import (
	"context"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

// SAdd 将指定成员添加到键所存储的集合中。
//
// 对于已经存在于集合中的成员，会被忽略。
// 如果键不存在，会先创建一个新集合，再添加指定成员。
// 如果键对应的值不是集合类型，则返回错误。
func (c *Client) SAdd(ctx context.Context, key string, members ...string) (int64, error) {
	if len(members) == 0 {
		return 0, errNoMembers
	}
	return c.doKeyValuesIntResult(ctx, "SADD", key, members...)
}

// SCard 返回存储在给定键上的集合的基数（元素数量）
func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "SCARD", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// SDiff 返回第一个集合与所有后续集合的差集所得到的成员
func (c *Client) SDiff(ctx context.Context, key string, keys ...string) ([]string, error) {
	args := make([]any, 2, 2+len(keys))
	args[0] = "SDIFF"
	args[1] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// SDiffStore 该命令等同于 SDIFF，但不是返回结果集，而是将结果存储到指定的目标键中。
// 如果目标键已存在，则会被覆盖
func (c *Client) SDiffStore(ctx context.Context, destination string, key string, keys ...string) (int64, error) {
	args := make([]any, 3, 3+len(keys))
	args[0] = "SDIFFSTORE"
	args[1] = destination
	args[2] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// SInter 返回所有给定集合的交集所得到的成员
func (c *Client) SInter(ctx context.Context, key string, keys ...string) ([]string, error) {
	args := make([]any, 2, 2+len(keys))
	args[0] = "SINTER"
	args[1] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// SIsMember 返回给定的成员是否存在于键所存储的集合中
//
//	若 key 或者 member 不存在，返回 false,nil
//	若 key 类型错误，或者其他错误，会返回 false,error
//	若 member 存在于 Set 中，返回 true,nil
func (c *Client) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "SISMEMBER", key, member)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBool(resp.result, resp.err, 1)
}

// SMembers 返回键所存储的集合中的所有成员
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "SMEMBERS", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// SMIsMember 返回给定的成员是否存在于键所存储的集合中
//
//	若 key 或者 member 不存在，返回 false,nil
//	若 key 类型错误，或者其他错误，会返回 false,error
//	若 member 存在于 Set 中，返回 true,nil
func (c *Client) SMIsMember(ctx context.Context, key string, members ...string) ([]bool, error) {
	if len(members) == 0 {
		return nil, errNoMembers
	}
	args := make([]any, 2, 2+len(members))
	args[0] = "SMISMEMBER"
	args[1] = key
	for _, k := range members {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBools(resp.result, resp.err, len(members), 1)
}

// SMove 将指定成员从源集合移动到目标集合。该操作是原子的，在任意时刻，对于其他客户端而言，该元素要么属于源集合，要么属于目标集合。
//
// 如果源集合不存在或不包含指定元素，则不执行任何操作，并返回 0。
// 否则，该元素会从源集合中移除，并添加到目标集合中。
// 如果目标集合中已存在该元素，则只会从源集合中移除该元素。
func (c *Client) SMove(ctx context.Context, source string, destination string, member string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "SMOVE", source, destination, member)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBool(resp.result, resp.err, 1)
}

// SPop 移除并返回键所存储集合中的一个随机成员
//
//	若 key 不存在，会返回 "",false,nil
//	若 key 存在( Set 不为空 )，会返回 "value",true,nil
//	其他错误，返回 "",false,error
func (c *Client) SPop(ctx context.Context, key string) (string, bool, error) {
	ret, err := c.SPopN(ctx, key, 1)
	if err != nil {
		return "", false, err
	}
	if len(ret) == 1 {
		return ret[0], true, nil
	}
	return "", false, nil
}

// SPopN 移除并返回键所存储集合中的最多 count 个随机成员
//
//	若 key 不存在，会返回 nil,nil
//	若 key 存在( Set 不为空 )，会返回 []{非空},nil
//	其他错误，返回 nil,error
func (c *Client) SPopN(ctx context.Context, key string, count int) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "SPOP", key, count)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// SRandMember 随机返回该键所存储集合中的最多 count 个元素
//
// 若 key 不存在，返回 nil,nil
func (c *Client) SRandMember(ctx context.Context, key string, count int) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "SRANDMEMBER", key, count)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// SRem 从键所存储的集合中移除指定成员。
//
//	若 key 或者 member 不存在，会返回 0
//	若对应的值不是集合类型，则返回错误
func (c *Client) SRem(ctx context.Context, key string, members ...string) (int64, error) {
	if len(members) == 0 {
		return 0, errNoMembers
	}
	args := make([]any, 2, 2+len(members))
	args[0] = "SREM"
	args[1] = key
	for _, mem := range members {
		args = append(args, mem)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// SUnion 返回所有给定集合的并集所得到的成员
func (c *Client) SUnion(ctx context.Context, key string, keys ...string) ([]string, error) {
	args := make([]any, 2, 2+len(keys))
	args[0] = "SREM"
	args[1] = key
	for _, mem := range keys {
		args = append(args, mem)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSet, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// SUnionStore 该命令等同于 SUNION，但不是返回结果集，而是将结果存储到指定的目标键中。
// 如果目标键已存在，则会被覆盖。
func (c *Client) SUnionStore(ctx context.Context, destination string, key string, keys ...string) ([]string, error) {
	args := make([]any, 3, 3+len(keys))
	args[0] = "SUNIONSTORE"
	args[1] = destination
	args[2] = key
	for _, mem := range keys {
		args = append(args, mem)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSet, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

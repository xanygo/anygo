//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-02

package xredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

type Z = resp3.Z

func (c *Client) ZAdd(ctx context.Context, key string, score float64, member string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZADD", key, score, member)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) ZAddOpt(ctx context.Context, key string, opt []string, score float64, member string) (int, error) {
	args := make([]any, 2, len(opt)+3)
	args[0] = "ZADD"
	args[1] = key
	for _, v := range opt {
		args = append(args, v)
	}
	args = append(args, score, member)
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) ZAddMap(ctx context.Context, key string, members map[string]float64) (int, error) {
	args := make([]any, 2, len(members)*2+2)
	args[0] = "ZADD"
	args[1] = key
	for member, score := range members {
		args = append(args, score, member)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) ZAddMapOpt(ctx context.Context, key string, opt []string, members map[string]float64) (int, error) {
	args := make([]any, 2, len(members)*2+2)
	args[0] = "ZADD"
	args[1] = key
	for _, v := range opt {
		args = append(args, v)
	}
	for member, score := range members {
		args = append(args, score, member)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// ZCard 返回键所存储有序集合的元素数量
func (c *Client) ZCard(ctx context.Context, key string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZCARD", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// ZCount 返回键所存储的有序集合中，分数在 min 和 max 范围内的元素数量
func (c *Client) ZCount(ctx context.Context, key string, min float64, max float64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZCOUNT", key, min, max)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) ZDiff(ctx context.Context, key string, keys ...string) ([]string, error) {
	args := make([]any, 3, len(key)+3)
	args[0] = "ZDIFF"
	args[1] = 1 + len(keys)
	args[2] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZDiffWithScores(ctx context.Context, key string, keys ...string) (map[string]float64, error) {
	args := make([]any, 3, len(key)+4)
	args[0] = "ZDIFF"
	args[1] = 1 + len(keys)
	args[2] = key
	for _, k := range keys {
		args = append(args, k)
	}
	args = append(args, "WITHSCORES")
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToMapFloat64(resp.result, resp.err)
}

// ZDiffStore 计算第一个有序集合与所有后续有序集合的差集，并将结果存储到目标键中。输入键的总数由 numkeys 指定。
//
// 不存在的键会被视为空集合。 如果目标键已存在，则会被覆盖。
func (c *Client) ZDiffStore(ctx context.Context, destination string, key string, keys ...string) (int, error) {
	args := make([]any, 4, 4+len(keys))
	args[0] = "ZDIFFSTORE"
	args[1] = destination
	args[2] = len(keys) + 1
	args[3] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// ZIncrBy 将键所存储有序集合中指定成员的分数增加 increment。
//
// 如果成员不存在，则将其添加到集合中，分数为 increment（相当于之前的分数为 0.0）。
// 如果键不存在，则创建一个新的有序集合，并将该成员作为唯一成员。
// 如果键存在但不是有序集合类型，则返回错误。
// 分数应为数值的字符串表示，可为双精度浮点数，也可以提供负值来减少分数。
func (c *Client) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZINCRBY", key, increment, member)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

// ZInter 计算由指定键给出的多个有序集合的交集
func (c *Client) ZInter(ctx context.Context, key string, keys ...string) ([]string, error) {
	args := make([]any, 3, 3+len(keys))
	args[0] = "ZINTER"
	args[1] = len(keys) + 1
	args[1] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZInterWithScores(ctx context.Context, key string, keys ...string) ([]Z, error) {
	args := make([]any, 3, 4+len(keys))
	args[0] = "ZINTER"
	args[1] = len(keys) + 1
	args[1] = key
	for _, k := range keys {
		args = append(args, k)
	}
	args = append(args, "WITHSCORES")
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZInterStore 计算由指定键给出的多个有序集合的交集,并将结果存储到目标键中
func (c *Client) ZInterStore(ctx context.Context, destination string, key string, keys ...string) (int, error) {
	args := make([]any, 4, 4+len(keys))
	args[0] = "ZINTERSTORE"
	args[1] = destination
	args[2] = len(keys) + 1
	args[3] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// ZLexCount 当有序集合中的所有元素具有相同分数时，为了强制按字典序排序，该命令返回键所存储的有序集合中，值在 min 和 max 之间的元素数量。
//
// min 和 max 参数的含义与 ZRANGEBYLEX 中描述的相同
func (c *Client) ZLexCount(ctx context.Context, key string, min float64, max float64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZLEXCOUNT", key, min, max)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZMPop 提供的键列表中，第一个非空的有序集合中弹出一个或多个元素，这些元素以 成员-分数对（member-score pairs）的形式返回。
//
// 返回值：
// fromKey: 结果来自那个 key
func (c *Client) ZMPop(ctx context.Context, key string, keys []string, min bool, count int) (fromKey string, members []Z, err error) {
	args := make([]any, 3, 6+len(keys))
	args[0] = "ZMPOP"
	args[1] = len(keys) + 1
	args[2] = key
	for _, k := range keys {
		args = append(args, k)
	}
	if min {
		args[2] = "MIN"
	} else {
		args[2] = "MAX"
	}
	if count > 0 {
		args = append(args, "COUNT", count)
	}

	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	if resp.err != nil {
		return "", nil, resp.err
	}
	arr, ok := resp.result.(resp3.Array)
	if !ok || len(arr) != 2 {
		return "", nil, errors.New("not ZMPOP reply")
	}
	fromKey, err = resp3.ToString(arr[0], nil)
	members, err = resp3.ToZSlice(arr[1], err)
	return fromKey, members, err
}

// ZMScore 返回键所存储有序集合中指定成员的分数
//
// 对于不存在于有序集合中的成员，返回 nil。
func (c *Client) ZMScore(ctx context.Context, key string, members ...string) (map[string]float64, error) {
	args := make([]any, 2, 2+len(members))
	args[0] = "ZMSCORE"
	args[1] = key
	for _, member := range members {
		args = append(args, member)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToMapFloat64WithKeys(resp.result, resp.err, members)
}

func (c *Client) ZPopMax(ctx context.Context, key string, count int) ([]Z, error) {
	args := make([]any, 2, 4)
	args[0] = "ZPOPMAX"
	args[1] = key
	if count > 0 {
		args = append(args, "COUNT", count)
	}

	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

func (c *Client) ZPopMin(ctx context.Context, key string, count int) ([]Z, error) {
	args := make([]any, 2, 4)
	args[0] = "ZPOPMIN"
	args[1] = key
	if count > 0 {
		args = append(args, "COUNT", count)
	}

	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

func (c *Client) ZRandMember(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANDMEMBER", key)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) ZRandMemberN(ctx context.Context, key string, count int) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANDMEMBER", key, count)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZRandMemberWithScores(ctx context.Context, key string, count int) ([]Z, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANDMEMBER", key, count, "WITHSCORES")
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZRange 返回存储在指定键中的有序集合中指定范围的元素。
func (c *Client) ZRange(ctx context.Context, key string, start int64, stop int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZRangeRev(ctx context.Context, key string, start int64, stop int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "REV")
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZRangeByScore(ctx context.Context, key string, start int64, stop int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "BYSCORE")
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZRangeRevByScore(ctx context.Context, key string, start int64, stop int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "BYSCORE", "REV")
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZRangeByLex(ctx context.Context, key string, start int64, stop int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "BYLEX")
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZRangeRevByLex(ctx context.Context, key string, start int64, stop int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "BYLEX", "REV")
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) ZRangeWithScore(ctx context.Context, key string, start int64, stop int64) ([]Z, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "WITHSCORES")
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

func (c *Client) ZRangeRevWithScore(ctx context.Context, key string, start int64, stop int64) ([]Z, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "REV", "WITHSCORES")
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

func (c *Client) ZRangeByScoreWithScore(ctx context.Context, key string, start int64, stop int64) ([]Z, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "BYSCORE", "WITHSCORES")
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

func (c *Client) ZRangeRevByScoreWithScore(ctx context.Context, key string, start int64, stop int64) ([]Z, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "BYSCORE", "REV", "WITHSCORES")
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

func (c *Client) ZRangeByLexWithScore(ctx context.Context, key string, start int64, stop int64) ([]Z, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "BYLEX", "WITHSCORES")
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

func (c *Client) ZRangeRevByLexWithScore(ctx context.Context, key string, start int64, stop int64) ([]Z, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANGE", key, start, stop, "BYLEX", "REV", "WITHSCORES")
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZRank 返回键所存储有序集合中指定成员的排名，按分数从低到高排序
//
// 排名（索引）从 0 开始，即分数最低的成员排名为 0。
func (c *Client) ZRank(ctx context.Context, key string, member string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZRANK", key, member)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) ZRevRank(ctx context.Context, key string, member string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZREVRANK", key, member)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) ZRankWithScore(ctx context.Context, key string, member string) (int64, float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANK", key, member, "WITHSCORE")
	resp := c.do(ctx, cmd)
	if resp.err != nil {
		return 0, 0, resp.err
	}
	arr, _ := resp.result.(resp3.Array)
	if len(arr) != 2 {
		return 0, 0, fmt.Errorf("expected 2 arrays, got %d", len(arr))
	}
	rank, err1 := resp3.ToInt64(arr[0], nil)
	score, err2 := resp3.ToFloat64(arr[1], err1)
	return rank, score, err2
}

// ZRevRankWithScore 返回键所存储有序集合中指定成员的排名，按分数从高到低排序。
//
// 排名（索引）从 0 开始，即分数最高的成员排名为 0。
func (c *Client) ZRevRankWithScore(ctx context.Context, key string, member string) (int64, float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZREVRANK", key, member, "WITHSCORE")
	resp := c.do(ctx, cmd)
	if resp.err != nil {
		return 0, 0, resp.err
	}
	arr, _ := resp.result.(resp3.Array)
	if len(arr) != 2 {
		return 0, 0, fmt.Errorf("expected 2 arrays, got %d", len(arr))
	}
	rank, err1 := resp3.ToInt64(arr[0], nil)
	score, err2 := resp3.ToFloat64(arr[1], err1)
	return rank, score, err2
}

// ZRem 从键所存储的有序集合中移除指定成员。
//
// 不存在的成员会被忽略。
// 如果键存在但不是有序集合类型，则返回错误。
func (c *Client) ZRem(ctx context.Context, key string, members ...string) (int64, error) {
	args := make([]any, 2, 2+len(members))
	args[0] = "ZREM"
	args[1] = key
	for _, member := range members {
		args = append(args, member)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZRemRangeByLex 删除值在 min 与 max 之间的所有元素
func (c *Client) ZRemRangeByLex(ctx context.Context, key string, min string, max string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZREMRANGEBYLEX", key, min, max)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZRemRangeByRank 移除键所存储有序集合中，排名在 start 与 stop 之间的所有元素
//
// start 和 stop 是 0 为起始索引 的排名，其中 0 表示分数最低的元素。
// 这两个索引可以为负数，表示从分数最高的元素开始的偏移。例如：
// -1 表示分数最高的元素
// -2 表示分数第二高的元素
// 依此类推。
func (c *Client) ZRemRangeByRank(ctx context.Context, key string, start int64, stop int64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZREMRANGEBYRANK", key, start, stop)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) ZRemRangeByScore(ctx context.Context, key string, min string, max string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZREMRANGEBYSCORE", key, min, max)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZScore 返回键所存储有序集合中指定成员的分数。
//
// 如果成员不存在于有序集合中，或者键不存在，则返回 nil
func (c *Client) ZScore(ctx context.Context, key string, member string) (float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeDouble, "ZSCORE", key, member)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

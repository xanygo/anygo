//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-02

package xredis

import (
	"context"
	"errors"
	"io"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

type Z = resp3.Z

// ZAdd 将所有指定的成员及其分数添加到存储在给定键的有序集合中。
// 如果某个指定的成员已经存在于有序集合中，则它的分数会被更新，并且该元素会被重新插入到正确的位置以保证排序顺序正确。
//
// 如果给定的键不存在，则会创建一个新的有序集合，并将指定的成员作为唯一的元素加入其中，就像原本集合为空一样。
// 如果该键已存在但其类型不是有序集合，则返回一个错误。
//
// 分数的值应当是双精度浮点数的字符串表示形式，+inf 和 -inf ( math.Inf(1)、 math.Inf(-1) ) 也是有效的取值。
//
// 返回值：bool - member是否新写入的 ( 只是 score 更新，会返回 false )
func (c *Client) ZAdd(ctx context.Context, key string, score float64, member string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZADD", key, score, member)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBool(resp.result, resp.err, 1)
}

// ZAddIncr 对有序集合 key 中的 member 执行分数递增操作（ZADD INCR）。
//
//	如果 member 不存在，则以 score 作为初始分数添加。
//	如果 key 不存在，则先创建一个空的有序集合。
//
//	返回值：递增后的最新分数。
func (c *Client) ZAddIncr(ctx context.Context, key string, score float64, member string) (float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeDouble, "ZADD", key, "INCR", score, member)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

// ZAddOpt ZAdd 的增强版本，支持如下 Options：
//
//	XX：仅更新已存在的元素，不添加新元素。
//	NX：仅添加新元素，不更新已存在的元素。
//	LT：仅在新分数小于当前分数时更新已存在的元素。此标志不会阻止添加新元素。
//	GT：仅在新分数大于当前分数时更新已存在的元素。此标志不会阻止添加新元素。
//	CH：修改返回值的含义，使其从“新增元素的数量”变为“发生变化的元素总数”（CH 是 changed 的缩写）。
//	   发生变化的元素包括新增的元素，以及分数被更新的已有元素。
//	   对于那些分数没有变化的元素，则不会被计入。
//	INCR：不支持
//	注意：通常情况下，ZADD 的返回值只计算新增元素的数量。
//
// 返回值：
//
//	1.空回复（Null reply）：如果由于与 XX / NX / LT / GT 选项中的某一个产生冲突而导致操作被中止时返回。
//	2.当未使用 CH 选项时，返回新增成员的数量。
//	3.当使用 CH 选项时，返回新增或被更新成员的数量。
func (c *Client) ZAddOpt(ctx context.Context, key string, opt []string, score float64, member string) (int64, error) {
	args := make([]any, 2, len(opt)+4)
	args[0] = "ZADD"
	args[1] = key
	for _, v := range opt {
		args = append(args, v)
	}
	args = append(args, score, member)
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZAddMap 批量新增、更新有序集合
//
// 返回值： 新写入member的个数 ( 只是 score 更新，计数会忽略 )
func (c *Client) ZAddMap(ctx context.Context, key string, members map[string]float64) (int64, error) {
	if len(members) == 0 {
		return 0, errNoMembers
	}
	args := make([]any, 2, len(members)*2+2)
	args[0] = "ZADD"
	args[1] = key
	for member, score := range members {
		args = append(args, score, member)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZAddMapOpt 批量新增、更新有序集合, ZAddMap 的增强版本，支持 Option：
// [NX | XX] [GT | LT] [CH], 具体含义详见 ZAddOpt
func (c *Client) ZAddMapOpt(ctx context.Context, key string, opt []string, members map[string]float64) (int64, error) {
	if len(members) == 0 {
		return 0, errNoMembers
	}
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
	return resp3.ToInt64(resp.result, resp.err)
}

// ZCard 返回键所存储有序集合的元素数量
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZCARD", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZCount 返回键所存储的有序集合中，分数在 min 和 max 范围内的元素数量
func (c *Client) ZCount(ctx context.Context, key string, min float64, max float64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZCOUNT", key, min, max)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZDiff 计算有序集合的差集，返回只存在于 key 中、但不存在于 keys 中的成员。
//
// 当未提供 keys 参数时返回错误
func (c *Client) ZDiff(ctx context.Context, key string, keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return nil, errNoKeys
	}
	args := make([]any, 3, len(key)+3)
	args[0] = "ZDIFF"
	args[1] = 1 + len(keys)
	args[2] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

// ZDiffWithScores 计算有序集合的差集，并返回成员及其对应的 score。
//
// 该方法对应 Redis 的 ZDIFF 命令（WITHSCORES）。
// 返回结果仅包含存在于 key 中、但不存在于 keys 中的成员。
// 当未提供 keys 参数时返回错误。
func (c *Client) ZDiffWithScores(ctx context.Context, key string, keys ...string) ([]Z, error) {
	if len(keys) == 0 {
		return nil, errNoKeys
	}
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
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZDiffStore 计算第一个有序集合与所有后续有序集合的差集，并将结果存储到目标键中。
//
// 不存在的键会被视为空集合。 如果目标键已存在，则会被覆盖。
// 若 keys 为空，则相当于将 key 拷贝到 destination 中。
//
// 返回值：存储到 destination 的元素个数
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
	cmd := resp3.NewRequest(resp3.DataTypeDouble, "ZINCRBY", key, increment, member)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

// ZInter 计算由指定键给出的多个有序集合的交集。
//
//	若 keys 为空，则返回 key 里所有的 member。
//	不存在的键会被视为空集合。
func (c *Client) ZInter(ctx context.Context, key string, keys ...string) ([]string, error) {
	args := make([]any, 3, 3+len(keys))
	args[0] = "ZINTER"
	args[1] = len(keys) + 1
	args[2] = key
	for _, k := range keys {
		args = append(args, k)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

// ZInterWithScores 计算多个有序集合（Sorted Set）的交集，并返回结果成员及其分值（score）。
//
// 该方法等价于执行 Redis 命令：
//
//	ZINTER key key1 key2 ... WITHSCORES
//
// 参数说明：
//   - key：第一个参与交集运算的有序集合键
//   - keys：其余参与交集运算的有序集合键,可以为空（keys为空时，返回 key 全集）
//
// 只有同时存在于所有指定有序集合中的成员才会出现在结果中。
// 返回结果按 score 升序排列。
//
// 如果指定的 key 不存在，则返回空结果。
func (c *Client) ZInterWithScores(ctx context.Context, key string, keys ...string) ([]Z, error) {
	args := make([]any, 3, 4+len(keys))
	args[0] = "ZINTER"
	args[1] = len(keys) + 1
	args[2] = key
	for _, k := range keys {
		args = append(args, k)
	}
	args = append(args, "WITHSCORES")
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZInterStore 计算由指定键给出的多个有序集合的交集,并将结果存储到目标键中
//
// 参数说明：
//   - destination: 存储目标，若已存在则覆盖
//   - key：第一个参与交集运算的有序集合键
//   - keys：其余参与交集运算的有序集合键,可以为空（keys为空时，存储 key 全集）
func (c *Client) ZInterStore(ctx context.Context, destination string, key string, keys ...string) (int64, error) {
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
	return resp3.ToInt64(resp.result, resp.err)
}

// ZLexCount 当有序集合中的所有元素具有相同分数时，为了强制按字典序排序，该命令返回键所存储的有序集合中，值在 min 和 max 之间的元素数量。
//
//   - 合法的 start 和 stop 必须以 ( 或 [ 开头，用于指定范围边界是开区间还是闭区间。
//   - start 和 stop 可以使用特殊值 + 或 -，分别表示正无穷和负无穷的字符串。
func (c *Client) ZLexCount(ctx context.Context, key string, min string, max string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZLEXCOUNT", key, min, max)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZMPop 提供的键列表中，第一个非空的有序集合中弹出一个或多个元素，这些元素以 成员-分数对（member-score pairs）的形式返回。
//
// 参数说明：
//
//   - key: 首个 POP 的有序集合键
//   - keys: 其余参与 POP 的有序集合键,可以为空
//   - min: 弹出方向，若为 true，则使用 MIN 规则（从 score 最小的元素 开始） 否则 MAX（从score 最大的元素开始）
//   - count: POP 的个数限制，>0 时值有效，否则使用为 redis Server 默认值 (1)
//
// 返回值：
//   - fromKey: 结果来自那个 key
//   - members: pop 的元素
//   - err: 错误
//
// 若 key 不存在，会返回 “”, nil, nil
func (c *Client) ZMPop(ctx context.Context, key string, keys []string, min bool, count int) (fromKey string, members []Z, err error) {
	args := make([]any, 3, 6+len(keys))
	args[0] = "ZMPOP"
	args[1] = len(keys) + 1
	args[2] = key
	for _, k := range keys {
		args = append(args, k)
	}
	if min {
		args = append(args, "MIN")
	} else {
		args = append(args, "MAX")
	}
	if count > 0 {
		args = append(args, "COUNT", count)
	}

	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	if resp.err != nil {
		if errors.Is(resp.err, ErrNil) {
			return "", nil, nil
		}
		return "", nil, resp.err
	}
	arr, err := resp.asResp3Array(2)
	if err != nil {
		return "", nil, err
	}
	fromKey, err = resp3.ToString(arr[0], nil)
	members, err = resp3.ToZSlice(arr[1], err)
	return fromKey, members, err
}

// ZMScore 返回键所存储有序集合中指定成员的分数
//
// 对于不存在于有序集合中的成员，返回 nil。
func (c *Client) ZMScore(ctx context.Context, key string, members ...string) (map[string]float64, error) {
	if len(members) == 0 {
		return nil, errNoMembers
	}
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

// ZPopMax 从有序集合中弹出分值最大的成员。
// 返回值包含成员及其对应的 score，按分值从大到小排序。
//
// 参数 count 可指定弹出成员的数量，若 count <= 0，则只弹出一个成员。
// 对应 Redis 的 ZPOPMAX 命令。
func (c *Client) ZPopMax(ctx context.Context, key string, count int) ([]Z, error) {
	args := make([]any, 2, 3)
	args[0] = "ZPOPMAX"
	args[1] = key
	if count > 0 {
		args = append(args, count)
	}

	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZPopMin 从有序集合中弹出分值最小的成员。
// 返回值包含成员及其对应的 score，按分值从小到大排序。
//
// 参数 count 可指定弹出成员的数量，若 count <= 0，则只弹出一个成员。
// 对应 Redis 的 ZPOPMIN 命令。
func (c *Client) ZPopMin(ctx context.Context, key string, count int) ([]Z, error) {
	args := make([]any, 2, 4)
	args[0] = "ZPOPMIN"
	args[1] = key
	if count > 0 {
		args = append(args, count)
	}

	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZRandMember 从有序集合中随机返回一个成员（不包含 score）。
//
// 对应 Redis 的 ZRANDMEMBER 命令。
//
// 如果 key 不存在，返回: "", false, nil
func (c *Client) ZRandMember(ctx context.Context, key string) (string, bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "ZRANDMEMBER", key)
	resp := c.do(ctx, cmd)
	mem, err := resp3.ToString(resp.result, resp.err)
	if err != nil && errors.Is(err, ErrNil) {
		return "", false, nil
	}
	return mem, err == nil, err
}

// ZRandMemberN 从有序集合中随机返回指定数量的成员（不包含 score）。
//
// 参数 count 指定返回的成员数量。
// 对应 Redis 的 ZRANDMEMBER 命令，
//
// 如果集合为空或 key 不存在，返回 nil,nil 。
func (c *Client) ZRandMemberN(ctx context.Context, key string, count int) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANDMEMBER", key, count)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

// ZRandMemberWithScores 从有序集合中随机返回指定数量的成员及其 score。
//
// 参数 count 指定返回的成员数量。
// 对应 Redis 的 ZRANDMEMBER 命令（WITHSCORES）。
// 返回结果包含成员及其分值
//
// 如果 key 不存在，返回 nil,nil。
func (c *Client) ZRandMemberWithScores(ctx context.Context, key string, count int) ([]Z, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANDMEMBER", key, count, "WITHSCORES")
	resp := c.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZRange 返回存储在指定键中的有序集合中指定范围的元素。
func (c *Client) ZRange(ctx context.Context, key string, opt ZRangeBy) ([]string, error) {
	return opt.zRange(ctx, c, key, "")
}

// ZRangeBy ZRANGEXXX 命令的配置参数
//
// 使用不同的算法，Start 和 Stop 赋值时有区别的：
//
// 1.使用 BYLEX 作为排序算法时：
//   - 合法的 Start 和 Stop 必须以 ( 或 [ 开头，用于指定范围边界是开区间还是闭区间。
//   - Start 和 Stop 可以使用特殊值 + 或 -，分别表示正无穷和负无穷的字符串。
//
// 2. 使用 非 BYLEX 排序算法时：
//   - Start 和 Stop 可以取 -inf 和 +inf，分别表示负无穷和正无穷。
//   - 默认情况下，Start 和 Stop 指定的分数区间是闭区间（包含边界），即 score >= Start and score <= Stop。
//   - 如果希望使用开区间（不包含边界），可以在分数前加上字符 "("，如 Start="(5", stop=10 ->  score > 5 and score <= 10
type ZRangeBy struct {
	Start   string // 必填, 起始分数
	Stop    string // 必填，截止分数
	Reverse bool   // 可选，按 score 降序排列
	Offset  int64  // 可选
	Count   int64  // 可选
}

// ZRANGE key start stop [BYSCORE | BYLEX] [REV] [LIMIT offset count] [WITHSCORES]
func (zb ZRangeBy) zRange(ctx context.Context, client *Client, key string, by string) ([]string, error) {
	args := []any{"ZRANGE", key, zb.Start, zb.Stop}
	if by != "" {
		args = append(args, by)
	}
	if zb.Reverse {
		args = append(args, "REV")
	}
	if zb.Count > 0 {
		args = append(args, "LIMIT", zb.Offset, zb.Count)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := client.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

func (zb ZRangeBy) zRangeWithScore(ctx context.Context, client *Client, key string, by string) ([]Z, error) {
	args := []any{"ZRANGE", key, zb.Start, zb.Stop}
	if by != "" {
		args = append(args, by)
	}
	if zb.Reverse {
		args = append(args, "REV")
	}
	if zb.Count > 0 {
		args = append(args, "LIMIT", zb.Offset, zb.Count)
	}
	args = append(args, "WITHSCORES")
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := client.do(ctx, cmd)
	return resp3.ToZSlice(resp.result, resp.err)
}

// ZRangeByScore 返回有序集合中指定分值区间的成员（按分值升序排列）。
//
// 对应 Redis 的 ZRANGE 命令（BYSCORE）。
func (c *Client) ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) ([]string, error) {
	return opt.zRange(ctx, c, key, "BYSCORE")
}

// ZRangeByLex 返回有序集合中按字典序（Lexicographical order）指定区间的成员。
//
// 对应 Redis 的 ZRANGE 命令（BYLEX）。
func (c *Client) ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) ([]string, error) {
	return opt.zRange(ctx, c, key, "BYLEX")
}

// ZRangeWithScore 返回有序集合中指定区间的成员及其分值（score）。
//
// 对应 Redis 的 ZRANGE 命令（WITHSCORES）。
func (c *Client) ZRangeWithScore(ctx context.Context, key string, opt ZRangeBy) ([]Z, error) {
	return opt.zRangeWithScore(ctx, c, key, "")
}

// ZRangeByScoreWithScore 返回有序集合中指定分值区间的成员及其分值（score），按分值升序排列。
//
// 对应 Redis 的 ZRANGE 命令（BYSCORE WITHSCORES）。
func (c *Client) ZRangeByScoreWithScore(ctx context.Context, key string, opt ZRangeBy) ([]Z, error) {
	return opt.zRangeWithScore(ctx, c, key, "BYSCORE")
}

// ZRangeByLexWithScore 返回有序集合中按字典序（Lexicographical order）指定区间的成员及其分值（score）。
//
// 对应 Redis 的 ZRANGE 命令（BYLEX WITHSCORES）。
func (c *Client) ZRangeByLexWithScore(ctx context.Context, key string, opt ZRangeBy) ([]Z, error) {
	return opt.zRangeWithScore(ctx, c, key, "BYLEX")
}

// ZRank 返回键所存储有序集合中指定成员的排名，按分数从低到高排序
//
// 排名（索引）从 0 开始，即分数最低的成员排名为 0。
// 若 key 或者 member 不存在，会返回 ErrNil
func (c *Client) ZRank(ctx context.Context, key string, member string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZRANK", key, member)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZRevRank 返回有序集合中指定成员的排名（按分值降序排列）。
//
// 对应 Redis 的 ZREVRANK 命令。
// 若 key 或者 member 不存在，会返回 ErrNil
func (c *Client) ZRevRank(ctx context.Context, key string, member string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZREVRANK", key, member)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZRankWithScore 返回有序集合中指定成员的排名和分值（score）。
//
// 排名按分值升序计算，0 表示分值最高的成员。
// 参数 key 为有序集合键，member 为要查询的成员。
// 对应 Redis 的 ZRANK 命令（WITHSCORES）。
func (c *Client) ZRankWithScore(ctx context.Context, key string, member string) (int64, float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "ZRANK", key, member, "WITHSCORE")
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(2)
	if err != nil {
		return 0, 0, err
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
	arr, err := resp.asResp3Array(2)
	if err != nil {
		return 0, 0, err
	}
	rank, err1 := resp3.ToInt64(arr[0], nil)
	score, err2 := resp3.ToFloat64(arr[1], err1)
	return rank, score, err2
}

// ZRem 从键所存储的有序集合中移除指定成员。
//
// 若 key 不存在 或者 member 不存在，会忽略掉，即返回 0，nil。
// 如果键存在但不是有序集合类型，则返回错误。
func (c *Client) ZRem(ctx context.Context, key string, members ...string) (int64, error) {
	if len(members) == 0 {
		return 0, errNoMembers
	}
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
//
//   - 合法的 min 和 max 必须以 ( 或 [ 开头，用于指定范围边界是开区间还是闭区间。
//   - min 和 max 可以使用特殊值 + 或 -，分别表示正无穷和负无穷的字符串。
//   - 若 key 不存在，会返回 0, nil
func (c *Client) ZRemRangeByLex(ctx context.Context, key string, min string, max string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZREMRANGEBYLEX", key, min, max)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZRemRangeByRank 移除键所存储有序集合中，排名在 [start 与 stop] 之间的所有元素
//
// start 和 stop 是 0 为起始索引的排名，是闭区间，包含 start 和 stop 本身，其中 0 表示分数最低的元素。
//
// 使用负数，表示从分数最高的元素开始的偏移, 例如：
//   - -1 表示分数最高的元素
//   - -2 表示分数第二高的元素
//   - 依此类推。
//
// 若 key 不存在，会返回 0, nil
func (c *Client) ZRemRangeByRank(ctx context.Context, key string, start int64, stop int64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZREMRANGEBYRANK", key, start, stop)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZRemRangeByScore 删除有序集合中指定分值区间的成员。
//
// 参数 key 为有序集合键，min 和 max 指定分值区间（包含 min 和 max）。
// 对应 Redis 的 ZREMRANGEBYSCORE 命令。
func (c *Client) ZRemRangeByScore(ctx context.Context, key string, min string, max string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "ZREMRANGEBYSCORE", key, min, max)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// ZScore 返回键所存储有序集合中指定成员的分数。
//
// 如果成员不存在于有序集合中，或者键不存在，则返回 ErrNil
func (c *Client) ZScore(ctx context.Context, key string, member string) (float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeDouble, "ZSCORE", key, member)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

// ZScan 遍历集合
//
// 若 key 不存在，会返回 0，nil,nil
func (c *Client) ZScan(ctx context.Context, key string, cursor uint64, pattern string, count int64) (uint64, []Z, error) {
	args := []any{"ZSCAN", key, cursor}
	if pattern != "" {
		args = append(args, "MATCH", pattern)
	}
	if count > 0 {
		args = append(args, "COUNT", count)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(2)
	if err != nil {
		return 0, nil, err
	}
	nextCursor, err := resp3.ToUint64(arr[0], nil)
	result, err2 := resp3.ToZSliceFlat(arr[1], err)
	return nextCursor, result, err2
}

// ZScanWalk 使用 ZScan 遍历该 key 满足条件的所有的值
//
// 回调方法 walk 返回 err = io.EOF 表示正常提前终止, 返回其他 error 表示提前异常终止
func (c *Client) ZScanWalk(ctx context.Context, key string, cursor uint64, pattern string, count int64, walk func(cursor uint64, m []Z) error) error {
	for {
		next, items, err := c.ZScan(ctx, key, cursor, pattern, count)
		if err != nil {
			return err
		}
		if err = walk(next, items); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if next == 0 {
			return nil
		}
		cursor = next
	}
}

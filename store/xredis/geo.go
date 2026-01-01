//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-12-31

package xredis

// https://redis.io/docs/latest/commands/geoadd/

import (
	"context"
	"errors"
	"fmt"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) GEOAdd(ctx context.Context, key string, items ...GeoMember) (int, error) {
	if len(items) == 0 {
		return 0, errNoValues
	}
	args := []any{"GEOADD", key}
	for _, item := range items {
		args = item.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) GEOAddWithOption(ctx context.Context, key string, opt *GEOAddOption, items ...GeoMember) (int, error) {
	if len(items) == 0 {
		return 0, errNoValues
	}
	args := []any{"GEOADD", key}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	for _, item := range items {
		args = item.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

type GeoMember struct {
	Longitude float64
	Latitude  float64
	Member    string
}

func (m GeoMember) appendArgs(args []any) []any {
	return append(args, m.Longitude, m.Latitude, m.Member)
}

type GEOAddOption struct {
	NX bool
	XX bool
	CH bool
}

func (opt *GEOAddOption) appendArgs(args []any) []any {
	if opt.NX {
		args = append(args, "NX")
	} else if opt.XX {
		args = append(args, "XX")
	}
	if opt.CH {
		args = append(args, "CH")
	}
	return args
}

func (c *Client) GEODist(ctx context.Context, key string, member1, member2 string, unit string) (float64, error) {
	args := []any{"GEODIST", key, member1, member2}
	if unit != "" {
		args = append(args, unit)
	}
	cmd := resp3.NewRequest(resp3.DataTypeDouble, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

func (c *Client) GEOHash(ctx context.Context, key string, members ...string) ([]*string, error) {
	if len(members) == 0 {
		return nil, errNoMembers
	}
	args := []any{"GEOHASH", key}
	args = xslice.Append(args, members...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToPtrStringSlice(resp.result, resp.err, len(members))
}

func (c *Client) GEOPos(ctx context.Context, key string, members ...string) ([]*GeoPos, error) {
	if len(members) == 0 {
		return nil, errNoMembers
	}
	args := []any{"GEOPOS", key}
	args = xslice.Append(args, members...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(len(members))
	if err != nil {
		return nil, err
	}
	result := make([]*GeoPos, 0, len(members))
	for _, item := range arr {
		if item.DataType() == resp3.DataTypeNull {
			result = append(result, nil)
			continue
		}
		point, err := resp3.ToSlice(item, nil)
		err = xslice.CheckLenIn(point, err, 2)
		if err != nil {
			return nil, err
		}
		pos := &GeoPos{}
		pos.Longitude, err = resp3.ToFloat64(point[0], nil)
		pos.Latitude, err = resp3.ToFloat64(point[1], err)
		if err != nil {
			return nil, err
		}
		result = append(result, pos)
	}
	return result, nil
}

type GeoPos struct {
	Longitude float64
	Latitude  float64
}

// GEOSearch 搜索，会忽略 option 里的 WithXXX 配置
func (c *Client) GEOSearch(ctx context.Context, key string, opt *GEOSearchOption) ([]string, error) {
	if opt == nil {
		return nil, errors.New("option is required")
	}
	args := []any{"GEOSEARCH", key}
	args = opt.appendArgs1(args)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

func (c *Client) GEOSearchLocation(ctx context.Context, key string, opt *GEOSearchOption) ([]GeoLocation, error) {
	if opt == nil {
		return nil, errors.New("option is required")
	}
	args := []any{"GEOSEARCH", key}
	args = opt.appendArgs1(args)
	args = opt.appendArgs2(args)
	withCount := opt.withLen()
	if withCount == 0 {
		return nil, errors.New("withXXX option is required")
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(0)
	if err != nil {
		return nil, err
	}
	result := make([]GeoLocation, 0, len(arr))
	for _, item := range arr {
		itemArr, err := resp3.ToSlice(item, nil)
		err = xslice.CheckLenIn(itemArr, err, 1+withCount)
		if err != nil {
			return nil, err
		}
		gl := GeoLocation{}
		gl.Member, err = resp3.ToString(itemArr[0], err)
		for i := 1; i < len(itemArr); i++ {
			switch rv := itemArr[i].(type) {
			case resp3.BulkString:
				gl.Dist, err = resp3.ToFloat64(rv, err)
			case resp3.Integer:
				gl.GeoHash, err = resp3.ToInt64(rv, err)
			case resp3.Array:
				point, err1 := resp3.ToSlice(rv, err)
				if err2 := xslice.CheckLenIn(point, err1, 2); err2 != nil {
					return nil, err
				}
				gl.Longitude, err = resp3.ToFloat64(point[0], err)
				gl.Latitude, err = resp3.ToFloat64(point[1], err)
			default:
				return nil, fmt.Errorf("unknown type: %T", rv)
			}
			if err != nil {
				return nil, err
			}
		}
		result = append(result, gl)
	}
	return result, nil
}

type GeoLocation struct {
	Member              string
	Longitude, Latitude float64
	Dist                float64
	GeoHash             int64
}

type GEOSearchOption struct {
	Member string // 有值时，使用 FromMember，和 FromLonLat 二选一

	Longitude float64 // 有值时使用 FromLonLat
	Latitude  float64 // 有值时使用 FromLonLat

	Radius     float64 // 搜索半径，可选
	RadiusUnit string  // 半径单位，可选，：M | KM（默认） | FT | MI

	BoxWidth  float64 // 搜索宽度，ByBox
	BoxHeight float64 // 搜索高度,ByBox
	BoxUnit   string  // 长度单位，可选，：M | KM（默认） | FT | MI

	Sort string // 排序，ASC or DESC，可选，默认空，不排序

	Count    int
	CountAny bool

	// 以下属性在使用 GEOSearch 时，不需要填写，但是在使用 GEOSearchLocation 时，至少需要有一个
	WithCoord bool

	WithDist bool
	WithHash bool
}

func (opt *GEOSearchOption) appendArgs1(args []any) []any {
	if opt.Member != "" {
		args = append(args, "FROMMEMBER", opt.Member)
	} else if opt.Latitude != 0 || opt.Longitude != 0 {
		args = append(args, "FROMLONLAT", opt.Longitude, opt.Latitude)
	}
	if opt.Radius > 0 {
		args = append(args, "BYRADIUS", opt.Radius)
		if opt.RadiusUnit != "" {
			args = append(args, opt.RadiusUnit)
		}
	} else if opt.BoxWidth > 0 {
		args = append(args, "BYBOX", opt.BoxWidth, opt.BoxHeight)
		if opt.BoxUnit != "" {
			args = append(args, opt.BoxUnit)
		}
	}

	if opt.Sort != "" {
		args = append(args, opt.Sort)
	}

	if opt.Count > 0 {
		args = append(args, "COUNT", opt.Count)
		if opt.CountAny {
			args = append(args, "ANY")
		}
	}
	return args
}

func (opt *GEOSearchOption) withLen() int {
	var count int
	if opt.WithCoord {
		count++
	}

	if opt.WithDist {
		count++
	}

	if opt.WithHash {
		count++
	}
	return count
}

func (opt *GEOSearchOption) appendArgs2(args []any) []any {
	if opt.WithCoord {
		args = append(args, "WITHCOORD")
	}
	if opt.WithDist {
		args = append(args, "WITHDIST")
	}
	if opt.WithHash {
		args = append(args, "WITHHASH")
	}
	return args
}

// GEOSearchStore 搜索并将结果存储
func (c *Client) GEOSearchStore(ctx context.Context, destination, source string, opt *GEOSearchStoreOption) (int64, error) {
	if opt == nil {
		return 0, errors.New("option is required")
	}
	args := []any{"GEOSEARCHSTORE", destination, source}
	args = opt.appendArgs1(args)
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

type GEOSearchStoreOption struct {
	Member string // 有值时，使用 FromMember，和 FromLonLat 二选一

	Longitude float64 // 有值时使用 FromLonLat
	Latitude  float64 // 有值时使用 FromLonLat

	Radius     float64 // 搜索半径，可选
	RadiusUnit string  // 半径单位，可选，：M | KM（默认） | FT | MI

	BoxWidth  float64 // 搜索宽度，ByBox
	BoxHeight float64 // 搜索高度,ByBox
	BoxUnit   string  // 长度单位，可选，：M | KM（默认） | FT | MI

	Sort string // 排序，ASC or DESC，可选，默认空，不排序

	Count    int
	CountAny bool

	StoreDist bool
}

func (opt *GEOSearchStoreOption) appendArgs1(args []any) []any {
	if opt.Member != "" {
		args = append(args, "FROMMEMBER", opt.Member)
	} else if opt.Latitude != 0 || opt.Longitude != 0 {
		args = append(args, "FROMLONLAT", opt.Longitude, opt.Latitude)
	}
	if opt.Radius > 0 {
		args = append(args, "BYRADIUS", opt.Radius)
		if opt.RadiusUnit != "" {
			args = append(args, opt.RadiusUnit)
		}
	} else if opt.BoxWidth > 0 {
		args = append(args, "BYBOX", opt.BoxWidth, opt.BoxHeight)
		if opt.BoxUnit != "" {
			args = append(args, opt.BoxUnit)
		}
	}

	if opt.Sort != "" {
		args = append(args, opt.Sort)
	}

	if opt.Count > 0 {
		args = append(args, "COUNT", opt.Count)
		if opt.CountAny {
			args = append(args, "ANY")
		}
	}

	if opt.StoreDist {
		args = append(args, "STOREDIST")
	}
	return args
}

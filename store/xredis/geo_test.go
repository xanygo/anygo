//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-12-31

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestGEO(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Skipf("create redis-server skipped: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())
	_, client, errClient := NewClientByURI("demo", ts.URI())
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	members1 := []GeoMember{
		{Longitude: 13.361389, Latitude: 38.115556, Member: "Palermo"},
		{Longitude: 15.087269, Latitude: 37.502669, Member: "Catania"},
		{Longitude: 12.758489, Latitude: 38.788135, Member: "edge1"},
		{Longitude: 17.241510, Latitude: 38.788135, Member: "edge2"},
	}

	t.Run("GEOHash", func(t *testing.T) {
		got, err := client.GEOHash(ctx, "GEOHash-1", "m1")
		xt.NoError(t, err)
		xt.Equal(t, []*string{nil}, got)

		added, err := client.GEOAdd(ctx, "GEOHash-1", members1...)
		xt.NoError(t, err)
		xt.Equal(t, 4, added)

		got, err = client.GEOHash(ctx, "GEOHash-1", "Palermo", "Catania")
		xt.NoError(t, err)
		a := "sqc8b49rny0"
		b := "sqdtr74hyu0"
		xt.Equal(t, []*string{&a, &b}, got)
	})

	t.Run("GEOSearch", func(t *testing.T) {
		added, err := client.GEOAdd(ctx, "GEOSearch-1", members1...)
		xt.NoError(t, err)
		xt.Equal(t, 4, added)

		opt1 := &GEOSearchOption{
			Longitude:  15,
			Latitude:   37,
			Radius:     200,
			RadiusUnit: "km",
			Sort:       "ASC",
		}
		sr, err := client.GEOSearch(ctx, "GEOSearch-1", opt1)
		xt.NoError(t, err)
		xt.Equal(t, []string{"Catania", "Palermo"}, sr)

		opt2 := &GEOSearchOption{
			Longitude: 15,
			Latitude:  37,

			BoxWidth:  400,
			BoxHeight: 400,
			BoxUnit:   "KM",

			Sort:      "ASC",
			WithCoord: true,
			WithDist:  true,
			WithHash:  true,
		}
		lr, err := client.GEOSearchLocation(ctx, "GEOSearch-1", opt2)
		xt.NoError(t, err)
		xt.NotEmpty(t, lr)
		xt.Len(t, lr, 4)
		for _, item := range lr {
			xt.NotEmpty(t, item.Member)
			xt.NotEmpty(t, item.Dist)
			xt.NotEmpty(t, item.GeoHash)
			xt.NotEmpty(t, item.Latitude)
			xt.NotEmpty(t, item.Longitude)
		}
	})
}

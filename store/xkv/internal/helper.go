package internal

import "github.com/xanygo/anygo/ds/xcmp"

func ParserMinMax(min, max string) (func(num float64) bool, error) {
	minBound := &xcmp.Bound[float64]{}
	if err := minBound.ParserMin(min); err != nil {
		return nil, err
	}

	maxBound := &xcmp.Bound[float64]{}
	if err := maxBound.ParserMax(max); err != nil {
		return nil, err
	}
	return func(num float64) bool {
		return minBound.MatchMin(num) && maxBound.MatchMax(num)
	}, nil
}

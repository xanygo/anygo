//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-06-01

package xws

import (
	"fmt"

	"github.com/xanygo/anygo/xcodec"
)

func decode(mt MessageType, data []byte) (*Message, error) {
	switch mt {
	case TextMessage, BinaryMessage:
		var tr Message
		err := xcodec.Decode(xcodec.JSON, data, &tr)
		if err != nil {
			return nil, err
		}
		tr.Type = mt
		return &tr, nil
	default:
		return nil, fmt.Errorf("invalid message type: %v", mt)
	}
}

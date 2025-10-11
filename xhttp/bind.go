//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package xhttp

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xvalidator"
)

func Bind(r *http.Request, obj any) error {
	defer r.Body.Close()
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		bf, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		err = xcodec.JSON.Decode(bf, obj)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("not support Content-Type: %s", ct)
	}
	return xvalidator.Validate(obj)
}

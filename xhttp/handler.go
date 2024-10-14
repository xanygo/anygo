//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-04

package xhttp

import "net/http"

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-08

package xhttp

import (
	"net/http"
)

func WriteHeader(w http.ResponseWriter, header http.Header) {
	for key, vs := range header {
		for _, value := range vs {
			w.Header().Add(key, value)
		}
	}
}

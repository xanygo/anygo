//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package caption

import (
	"image"
	"net/http"
)

type Caption interface {
	http.Handler
	Image() image.Image
	Code() string
}

package images

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func Base64(filename string) (v string, err error) {
	p, pErr := os.ReadFile(filename)
	if pErr != nil {
		err = pErr
		return
	}
	var mimeType string
	if strings.HasSuffix(strings.ToLower(filename), ".png") {
		mimeType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(filename), ".jpg") ||
		strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
		mimeType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(filename), ".gif") {
		mimeType = "image/gif"
	} else if strings.HasSuffix(strings.ToLower(filename), ".webp") {
		mimeType = "image/webp"
	} else {
		mimeType = "image/jpeg" // 默认
	}
	base64String := base64.StdEncoding.EncodeToString(p)
	v = fmt.Sprintf("data:%s;base64,%s", mimeType, base64String)
	return
}

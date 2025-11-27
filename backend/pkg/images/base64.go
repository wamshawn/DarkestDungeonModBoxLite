package images

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Encode(filename string) (v string, err error) {
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
		err = errors.New("unsupported file type")
		return
	}
	base64String := base64.StdEncoding.EncodeToString(p)
	v = fmt.Sprintf("data:%s;base64,%s", mimeType, base64String)
	return
}

func EncodeBytes(filename string, src []byte) (v string, err error) {
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
		err = errors.New("unsupported file type")
		return
	}
	base64String := base64.StdEncoding.EncodeToString(src)
	v = fmt.Sprintf("data:%s;base64,%s", mimeType, base64String)
	return
}

func Decode(filename string, encoded string) (v string, err error) {
	idx := strings.Index(encoded, "data:")
	if idx == -1 {
		err = errors.New("malformed image")
		return
	}
	encoded = strings.TrimLeft(encoded, "data:")
	idx = strings.Index(encoded, ";base64,")
	if idx == -1 {
		err = errors.New("malformed image")
		return
	}

	sub := ""
	mimeType := encoded[:idx]
	switch mimeType {
	case "image/png":
		sub = ".png"
	case "image/jpeg", "image/jpg":
		sub = ".jpg"
	case "image/gif":
		sub = ".gif"
	case "image/webp":
		sub = ".webp"
	default:
		err = errors.New("malformed image")
		return
	}
	encoded = strings.TrimLeft(encoded, mimeType)
	idx = strings.Index(encoded, ";base64,")
	if idx == -1 {
		err = errors.New("malformed image")
		return
	}
	encoded = strings.TrimLeft(encoded, ";base64,")
	b, decodeErr := base64.StdEncoding.DecodeString(encoded)
	if decodeErr != nil {
		err = errors.New("malformed image")
		return
	}

	if ext := filepath.Ext(filename); ext != "" && ext != sub {
		idx = strings.LastIndexByte(filename, '.')
		filename = filename[:idx]
	}
	filename = filename + sub
	wErr := os.WriteFile(filename, b, 0644)
	if wErr != nil {
		err = wErr
		return
	}
	v = filename
	return
}

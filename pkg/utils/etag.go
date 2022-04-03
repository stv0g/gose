package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
)

const (
	MaxPartCount = 10000
)

func IsValidETag(et string) bool {
	p := strings.SplitN(et, "-", 2)

	if etag, err := hex.DecodeString(p[0]); err != nil || len(etag) != md5.Size {
		return false
	}

	if num, err := strconv.Atoi(p[1]); err != nil || num > MaxPartCount || num <= 0 {
		return false
	}

	return true
}

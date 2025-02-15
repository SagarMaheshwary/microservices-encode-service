package helper

import (
	"crypto/rand"
	"encoding/base32"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func GetRootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))

	return filepath.Dir(d)
}

func UniqueString(length int) string {
	b := make([]byte, 32)

	rand.Read(b)

	return strings.ToLower(base32.StdEncoding.EncodeToString(b)[:length])
}

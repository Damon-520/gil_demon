package cryptox

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
)

type Md5x struct {
}

func NewMd5x() *Md5x {
	return &Md5x{}
}

// MD5Hash MD5哈希值
func (m *Md5x) MD5Hash(b []byte) string {
	h := md5.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MD5HashString MD5哈希值
func (m *Md5x) MD5HashString(s string) string {
	return m.MD5Hash([]byte(s))
}

// SHA1Hash SHA1哈希值
func (m *Md5x) SHA1Hash(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// SHA1HashString SHA1哈希值
func (m *Md5x) SHA1HashString(s string) string {
	return m.SHA1Hash([]byte(s))
}

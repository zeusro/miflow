package miioservice

import (
	"bytes"
	"compress/gzip"
	"crypto/rc4"
	"io"
)

// newRC4Cipher 创建 RC4 cipher，用于 MIoT 解密。
func newRC4Cipher(key []byte) (*rc4.Cipher, error) {
	return rc4.NewCipher(key)
}

// gzipDecode 解压 gzip 数据。
func gzipDecode(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

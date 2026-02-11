package miioservice

import (
	"bytes"
	"compress/gzip"
	"crypto/rc4"
	"io"
)

func newRC4Cipher(key []byte) (*rc4.Cipher, error) {
	return rc4.NewCipher(key)
}

func gzipDecode(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

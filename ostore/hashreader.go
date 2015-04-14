package ostore

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
)

type HashReader struct {
	h hash.Hash
	r io.Reader
}

func NewHashReader(r io.Reader) *HashReader {
	return &HashReader{sha512.New(), r}
}

func (hr *HashReader) Read(p []byte) (int, error) {
	n, err := hr.r.Read(p)
	if err != nil {
		return n, err
	}
	wn := 0
	tn := 0
	for wn, err = hr.h.Write(p[:n]); err != nil && wn < n; {
		tn, err = hr.h.Write(p[wn:n])
		wn += tn
	}
	return n, err
}

func (hr *HashReader) Digest() []byte {
	return hr.h.Sum(nil)
}

func (hr *HashReader) HexDigest() string {
	return hex.EncodeToString(hr.Digest())
}

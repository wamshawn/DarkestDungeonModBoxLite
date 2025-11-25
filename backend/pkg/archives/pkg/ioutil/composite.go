package ioutil

import (
	"errors"
	"io"
)

func NewCompositeByteReader(b []byte, r io.Reader) io.Reader {
	return &CompositeByteReader{
		n: 0,
		b: b,
		r: r,
	}
}

type CompositeByteReader struct {
	n int
	b []byte
	r io.Reader
	e error
}

func (cbr *CompositeByteReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		err = io.ErrShortBuffer
		return
	}
	if cbr.n < len(cbr.b) {
		cp := copy(p, cbr.b[cbr.n:])
		cbr.n += cp
		n = cp
		if n == len(p) {
			return
		}
	}
	if cbr.e != nil {
		err = cbr.e
		return
	}
	nn, rErr := cbr.r.Read(p[n:])
	n += nn
	if errors.Is(rErr, io.EOF) {
		cbr.e = io.EOF
		if n == 0 {
			err = io.EOF
		}
		return
	}
	err = rErr
	return
}

package ws

import (
	"bytes"
	"compress/zlib"
	"io"
)

func Compress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := zlib.NewWriter(buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	w.Close()
	return buf.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

package ws

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
)

func CompressGZIP(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := gzip.NewWriter(buf)
	w.Write(data)
	w.Close()
	return buf.Bytes(), nil
}

func DecompressGZIP(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func CompressFlate(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w, err := flate.NewWriter(buf, 1)
	if err != nil {
		return nil, err
	}

	w.Write(data)
	w.Close()
	return buf.Bytes(), nil
}

func DecompressFlate(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	r := flate.NewReader(bytes.NewReader(data))
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

package ws

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"io"
)

const defaultLevel = flate.BestCompression

func Compress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w, err := flate.NewWriter(buf, defaultLevel)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	w.Close()

	encodedMsg := make([]byte, base64.StdEncoding.EncodedLen(len(buf.Bytes())))
	base64.StdEncoding.Encode(encodedMsg, buf.Bytes())

	return encodedMsg, nil
}

func Decompress(data []byte) ([]byte, error) {
	decodedMsg := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	if _, err := base64.StdEncoding.Decode(decodedMsg, data); err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	r := flate.NewReader(bytes.NewReader(decodedMsg))
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

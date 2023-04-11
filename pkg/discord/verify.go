package discord

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

func Verify(r *http.Request, key ed25519.PublicKey) error {

	signature := r.Header.Get("X-Signature-Ed25519")
	if signature == "" {
		return fmt.Errorf("signature can not empty")
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}

	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		return fmt.Errorf("signature is not valid")
	}

	timestamp := r.Header.Get("X-Signature-Timestamp")
	if timestamp == "" {
		return fmt.Errorf("timestamp can not empty")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	body := append([]byte(timestamp), bodyBytes...)

	result := ed25519.Verify(key, body, sig)
	if !result {
		return fmt.Errorf("signature is not valid")
	}
	return nil
}

package autoroute

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// A keySigner is used to verify the integrity of SessionKeys at the system borders
type keySigner struct {
	key []byte
}

// newKeySigner creates a new keySigner with the given key
func newKeySigner(key string) *keySigner {
	return &keySigner{
		key: []byte(key),
	}
}

// Sign appends an HMAC to a value
func (ks *keySigner) Sign(val string) (string, error) {
	h := hmac.New(sha256.New, ks.key)
	_, err := h.Write([]byte(val))
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("%s.%s", val, hex.EncodeToString(h.Sum(nil))), nil
}

// Verify checks a value signed with Sign
func (ks *keySigner) Verify(pubVal string) (string, error) {
	h := hmac.New(sha256.New, ks.key)

	spl := strings.Split(pubVal, ".")
	if len(spl) != 2 {
		return "", errors.New("invalid token")
	}

	_, err := h.Write([]byte(spl[0]))
	if err != nil {
		return "", nil
	}

	hmacBytes, err := hex.DecodeString(spl[1])
	if err != nil {
		return "", err
	}

	if !hmac.Equal(hmacBytes, h.Sum(nil)) {
		return "", errors.New("invalid signature")
	}

	return spl[0], nil
}

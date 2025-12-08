package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func VerifyMPSignature(secret []byte, signatureHeader string, body []byte) bool {
	if signatureHeader == "" {
		return false
	}

	// Esperamos: "ts=1702000000, v1=7cb2..."
	parts := strings.Split(signatureHeader, ",")
	if len(parts) != 2 {
		return false
	}

	var ts string
	var v1 string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "ts=") {
			ts = strings.TrimPrefix(p, "ts=")
		}
		if strings.HasPrefix(p, "v1=") {
			v1 = strings.TrimPrefix(p, "v1=")
		}
	}

	if ts == "" || v1 == "" {
		return false
	}

	// 1) Calcular SHA-256 del body
	bodyHash := sha256.Sum256(body)
	digest := hex.EncodeToString(bodyHash[:])

	// 2) Armar el string base SEGÚN MP:
	// ts=<ts>:digest=<body_sha256>
	signBase := "ts=" + ts + ":digest=" + digest

	// 3) HMAC-SHA256(secret, signBase)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signBase))
	expectedMAC := mac.Sum(nil)
	expectedHex := hex.EncodeToString(expectedMAC)

	// 4) Comparar en constante time
	return hmac.Equal([]byte(expectedHex), []byte(v1))
}

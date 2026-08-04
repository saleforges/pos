package ipaymu

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

// buildCallbackSignature mirrors the documented iPaymu verification flow:
// normalize values, sort keys (json.Marshal on map does this), serialize
// without HTML escaping, escape "/" as "\/", HMAC-SHA256 with VA secret.
func buildCallbackSignature(va string, raw map[string]interface{}) string {
	normalized := normalizeValues(raw)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(normalized)
	canonical := strings.ReplaceAll(strings.TrimRight(buf.String(), "\n"), "/", "\\/")

	h := hmac.New(sha256.New, []byte(va))
	h.Write([]byte(canonical))
	return hex.EncodeToString(h.Sum(nil))
}

func TestVerifyCallback(t *testing.T) {
	t.Parallel()

	va := "1179001234567890"
	payload := map[string]interface{}{
		"reference_id":    "5",
		"status":          "berhasil",
		"status_code":     "1",
		"amount":          "30000",
		"via":             "qris",
		"trx_id":          "TRX123",
		"is_escrow":       "0",
		"buyer_name":      "Pak Budi",
		"additional_info": "[]",
	}

	t.Run("valid signature passes", func(t *testing.T) {
		sig := buildCallbackSignature(va, payload)
		if !VerifyCallback(va, sig, payload) {
			t.Error("expected valid signature to pass")
		}
	})

	t.Run("tampered payload fails", func(t *testing.T) {
		sig := buildCallbackSignature(va, payload)
		tampered := map[string]interface{}{}
		for k, v := range payload {
			tampered[k] = v
		}
		tampered["amount"] = "99999"
		if VerifyCallback(va, sig, tampered) {
			t.Error("expected tampered payload to fail")
		}
	})

	t.Run("wrong secret fails", func(t *testing.T) {
		sig := buildCallbackSignature(va, payload)
		if VerifyCallback("different-va", sig, payload) {
			t.Error("expected wrong va to fail")
		}
	})

	t.Run("empty signature fails", func(t *testing.T) {
		if VerifyCallback(va, "", payload) {
			t.Error("expected empty signature to fail")
		}
	})

	t.Run("missing additional_info normalizes to empty array", func(t *testing.T) {
		noAdditional := map[string]interface{}{
			"reference_id": "5", "status": "berhasil", "status_code": "1",
			"amount": "30000", "via": "qris", "trx_id": "TRX123", "is_escrow": "0",
		}
		sig := buildCallbackSignature(va, noAdditional)
		if !VerifyCallback(va, sig, noAdditional) {
			t.Error("expected signature with missing additional_info to pass")
		}
	})

	// Regression: real payload captured from the iPaymu sandbox "Tes Notify"
	// simulator. Its X-Signature must verify with our implementation.
	t.Run("real sandbox simulator payload", func(t *testing.T) {
		realVA := "0000001326904469"
		realSig := "7cc161bded6ca01562e783cc38fe48e45bbc14077cf59120f4dd25c72323c06b"
		realPayload := map[string]interface{}{
			"trx_id":                  "222704",
			"sid":                     "a1b14861-5782-4623-9ee1-f5b77191b7ff",
			"reference_id":            "29",
			"status":                  "berhasil",
			"status_code":             "1",
			"sub_total":               "15000",
			"total":                   "15270",
			"amount":                  "15270",
			"fee":                     "375",
			"paid_off":                "14895",
			"created_at":              "2026-08-05 02:08:47",
			"expired_at":              "2026-08-06 02:08:47",
			"paid_at":                 "2026-08-05 02:12:12",
			"settlement_status":       "settled",
			"transaction_status_code": "7",
			"is_escrow":               "true",
			"system_notes":            "Sandbox notify",
			"via":                     "qris",
			"channel":                 "qris",
			"payment_no":              "",
			"buyer_name":              "ilhamp",
			"buyer_email":             "sss@gmail.com",
			"buyer_phone":             "2222",
			"additional_info":         "[]",
			"url":                     "https://api-dev.saleforges.com/v1/payments/ipaymu/callback",
		}
		if !VerifyCallback(realVA, realSig, realPayload) {
			t.Error("expected real sandbox simulator signature to verify")
		}
	})
}

func TestSignRequest(t *testing.T) {
	t.Parallel()

	body := []byte(`{"product":["Es Teh"],"qty":[2],"price":[30000]}`)
	sig := SignRequest("POST", "1179001234567890", "api-key-123", body)

	// Recompute manually per the documented formula.
	bodyHash := sha256.Sum256(body)
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))
	stringToSign := "POST:1179001234567890:" + bodyHashHex + ":api-key-123"
	h := hmac.New(sha256.New, []byte("api-key-123"))
	h.Write([]byte(stringToSign))
	expected := hex.EncodeToString(h.Sum(nil))

	if sig != expected {
		t.Errorf("signature mismatch: got %s want %s", sig, expected)
	}
}

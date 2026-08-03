package ipaymu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// buildCallbackSignature mirrors the documented iPaymu verification flow to
// produce a valid signature for testing.
func buildCallbackSignature(va string, raw map[string]interface{}) string {
	// Normalize exactly like VerifyCallback does.
	normalized := normalizeValues(raw)
	keys := make([]string, 0, len(normalized))
	for k := range normalized {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.Quote(k))
		sb.WriteString(":")
		b, _ := json.Marshal(normalized[k])
		sb.Write(b)
	}
	sb.WriteString("}")

	h := hmac.New(sha256.New, []byte(va))
	h.Write([]byte(sb.String()))
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

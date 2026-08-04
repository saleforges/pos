package ipaymu

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SignRequest builds the iPaymu v2 signature for outgoing API requests:
// stringToSign = METHOD:VA:sha256(body).lower():APIKey
// signature = HMAC-SHA256(stringToSign, APIKey)
func SignRequest(method, va, apiKey string, body []byte) string {
	bodyHash := sha256.Sum256(body)
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))
	stringToSign := method + ":" + va + ":" + bodyHashHex + ":" + apiKey

	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

// Timestamp returns the iPaymu timestamp header format (YYYYMMDDHHmmss).
func Timestamp() string {
	return time.Now().UTC().Format("20060102150405")
}

// VerifyCallback validates an iPaymu callback X-Signature. The secret key
// is the merchant VA. Per iPaymu docs the payload is: values normalized
// (strings to int/bool, additional_info as []), keys sorted ascending,
// JSON-serialized WITHOUT HTML escaping, then every "/" escaped as "\/"
// (their example uses JSON.stringify + replace(/\\\//g, '\\\/')) before
// HMAC-SHA256.
func VerifyCallback(va, signature string, raw map[string]interface{}) bool {
	if signature == "" {
		return false
	}
	normalized := normalizeValues(raw)

	// json.Marshal on a map sorts keys ascending automatically.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(normalized); err != nil {
		return false
	}
	canonical := strings.ReplaceAll(strings.TrimRight(buf.String(), "\n"), "/", "\\/")

	h := hmac.New(sha256.New, []byte(va))
	h.Write([]byte(canonical))
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.ToLower(signature)))
}

// normalizeValues converts callback form/string values to the types iPaymu
// expects before signature validation (per their docs).
func normalizeValues(raw map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		s := fmt.Sprintf("%v", v)
		switch k {
		case "is_escrow":
			result[k] = s == "true" || s == "1"
		case "trx_id", "status_code", "transaction_status_code", "paid_off":
			if n, err := strconv.Atoi(s); err == nil {
				result[k] = n
			} else {
				result[k] = s
			}
		case "additional_info":
			if s == "[]" {
				result[k] = []interface{}{}
			} else {
				result[k] = s
			}
		default:
			result[k] = s
		}
	}
	if _, ok := result["additional_info"]; !ok {
		result["additional_info"] = []interface{}{}
	}
	return result
}

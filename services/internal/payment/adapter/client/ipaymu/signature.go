package ipaymu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
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
// is the merchant VA. Payload values must be normalized (strings to
// int/bool, missing additional_info as []), keys sorted ascending, then
// serialized to JSON before HMAC-SHA256.
func VerifyCallback(va, signature string, raw map[string]interface{}) bool {
	if signature == "" {
		return false
	}
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
		val := normalized[k]
		b, err := json.Marshal(val)
		if err != nil {
			return false
		}
		sb.Write(b)
	}
	sb.WriteString("}")

	h := hmac.New(sha256.New, []byte(va))
	h.Write([]byte(sb.String()))
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

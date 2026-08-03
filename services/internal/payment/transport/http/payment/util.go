package payment

import (
	"strconv"
	"strings"
)

func str(raw map[string]interface{}, key string) string {
	v, ok := raw[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func statusCode(raw map[string]interface{}) int {
	switch v := raw["status_code"].(type) {
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return 0
	}
}

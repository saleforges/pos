package payment

import (
	"fmt"
	"strconv"
	"strings"
)

func str(raw map[string]interface{}, key string) string {
	v, ok := raw[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		// JSON numbers (reference_id as number) arrive as float64.
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", t)
	}
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

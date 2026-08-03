package payment

import "testing"

func TestStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  map[string]interface{}
		key  string
		want string
	}{
		{"string value", map[string]interface{}{"status": "berhasil"}, "status", "berhasil"},
		{"json number", map[string]interface{}{"reference_id": float64(23)}, "reference_id", "23"},
		{"missing key", map[string]interface{}{"status": "x"}, "amount", ""},
		{"nil value", map[string]interface{}{"channel": nil}, "channel", ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := str(tt.raw, tt.key); got != tt.want {
				t.Errorf("str(%v, %q) = %q, want %q", tt.raw, tt.key, got, tt.want)
			}
		})
	}
}

func TestStatusCode(t *testing.T) {
	t.Parallel()

	if got := statusCode(map[string]interface{}{"status_code": float64(1)}); got != 1 {
		t.Errorf("float64 status_code = %d, want 1", got)
	}
	if got := statusCode(map[string]interface{}{"status_code": "1"}); got != 1 {
		t.Errorf("string status_code = %d, want 1", got)
	}
	if got := statusCode(map[string]interface{}{}); got != 0 {
		t.Errorf("missing status_code = %d, want 0", got)
	}
}

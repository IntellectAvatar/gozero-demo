package i18n

import (
	"net/http"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   Lang
	}{
		{"zh-CN", "zh-CN,en;q=0.9", ZH},
		{"zh-TW", "zh-TW", ZH},
		{"en-US", "en-US,en;q=0.9", EN},
		{"empty fallback", "", EN},
		{"unknown fallback", "fr-FR", EN},
		{"ja-JP fallback", "ja-JP", EN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/", nil)
			r.Header.Set("Accept-Language", tt.header)
			if got := Detect(r); got != tt.want {
				t.Errorf("Detect() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestMsg(t *testing.T) {
	tests := []struct {
		name string
		lang Lang
		key  string
		want string
	}{
		{"zh success", ZH, "success", "成功"},
		{"zh not_found", ZH, "not_found", "资源不存在"},
		{"zh wrong_password", ZH, "wrong_password", "用户名或密码错误"},
		{"en success", EN, "success", "Success"},
		{"en unauthorized", EN, "unauthorized", "Unauthorized, please login first"},
		{"missing key fallback to en", ZH, "nonexistent", "nonexistent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Msg(tt.lang, tt.key); got != tt.want {
				t.Errorf("Msg(%s, %s) = %s, want %s", tt.lang, tt.key, got, tt.want)
			}
		})
	}
}

func TestMessagesCoverAllKeys(t *testing.T) {
	for k := range enMessages {
		if _, ok := zhMessages[k]; !ok {
			t.Errorf("zhMessages missing key: %s", k)
		}
	}
}

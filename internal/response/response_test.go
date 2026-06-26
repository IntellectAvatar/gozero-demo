package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOkWrapsSingleObjectInArray(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	Ok(w, r, map[string]string{"key": "value"})

	var body Body
	json.NewDecoder(w.Body).Decode(&body)

	if body.Code != CodeSuccess {
		t.Errorf("Code = %d, want %d", body.Code, CodeSuccess)
	}
	if body.Message != "Success" {
		t.Errorf("Message = %s, want Success", body.Message)
	}
	// Data 必须是数组
	if len(body.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(body.Data))
	}
}

func TestOkListWithTotal(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	OkList(w, r, []any{"a", "b"}, 99)

	var body Body
	json.NewDecoder(w.Body).Decode(&body)

	if len(body.Data) != 2 {
		t.Errorf("len(Data) = %d, want 2", len(body.Data))
	}
	if body.Total == nil || *body.Total != 99 {
		t.Errorf("Total = %v, want 99", body.Total)
	}
}

func TestOkMsgNilDataReturnsEmptyArray(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	OkMsg(w, r, nil, "password_changed")

	var body Body
	json.NewDecoder(w.Body).Decode(&body)

	if len(body.Data) != 0 {
		t.Errorf("len(Data) = %d, want 0 (nil → empty array)", len(body.Data))
	}
}

func TestFailEmptyData(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	Fail(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")

	var body Body
	json.NewDecoder(w.Body).Decode(&body)

	if body.Code != CodeUnauthorized {
		t.Errorf("Code = %d, want %d", body.Code, CodeUnauthorized)
	}
	if len(body.Data) != 0 {
		t.Errorf("len(Data) = %d, want 0 (errors produce empty Data)", len(body.Data))
	}
}

func TestI18nLanguageSwitch(t *testing.T) {
	t.Run("zh-CN", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept-Language", "zh-CN")
		Ok(w, r, nil)
		var body Body
		json.NewDecoder(w.Body).Decode(&body)
		if body.Message != "成功" {
			t.Errorf("Message = %s, want 成功", body.Message)
		}
	})
	t.Run("default en", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		Ok(w, r, nil)
		var body Body
		json.NewDecoder(w.Body).Decode(&body)
		if body.Message != "Success" {
			t.Errorf("Message = %s, want Success", body.Message)
		}
	})
}

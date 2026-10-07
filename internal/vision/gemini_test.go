package vision

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fake(t *testing.T, status int, body string, check func(r *http.Request, req map[string]any)) Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(raw, &req)
		if check != nil {
			check(r, req)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return Client{Key: "secret-key", Model: "test-model", BaseURL: srv.URL, HTTP: srv.Client()}
}

func answer(items string) string {
	b, _ := json.Marshal(items)
	return `{"candidates":[{"content":{"parts":[{"text":` + string(b) + `}]}}]}`
}

func TestLocateConvertsBoxesToPixels(t *testing.T) {
	c := fake(t, 200, answer(`[{"kind":"tavern","name":"Пьяный гусь","note":"Таверна у моста","box_2d":[400,100,600,300]},
		{"kind":"dragon-nest","name":"?","box_2d":[0,0,1000,1000]},{"kind":"city","name":"x","box_2d":[1,2]}]`),
		func(r *http.Request, req map[string]any) {
			if r.Header.Get("x-goog-api-key") != "secret-key" || !strings.HasSuffix(r.URL.Path, "/models/test-model:generateContent") {
				t.Errorf("запрос: %s, ключ %q", r.URL.Path, r.Header.Get("x-goog-api-key"))
			}
			if !strings.Contains(r.URL.String(), "generateContent") || strings.Contains(r.URL.RawQuery, "secret") {
				t.Error("ключ не должен попадать в адрес запроса")
			}
		})
	got, err := c.Locate(context.Background(), "image/jpeg", []byte{1, 2, 3}, 2000, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("ожидалось 2 места (рамка из двух чисел отброшена), получено %d", len(got))
	}
	if got[0].X != 400 || got[0].Y != 500 || got[0].Kind != "tavern" || got[0].Name != "Пьяный гусь" {
		t.Errorf("таверна: %+v", got[0])
	}
	if got[1].Kind != "other" {
		t.Errorf("неизвестный вид должен стать other: %+v", got[1])
	}
}

func TestLocateErrorsAreReadable(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
		want   string
	}{
		"ключ":   {400, `{"error":{"message":"API key not valid. Please pass a valid API key.","status":"INVALID_ARGUMENT"}}`, "неверный ключ"},
		"регион": {400, `{"error":{"message":"User location is not supported for the API use.","status":"FAILED_PRECONDITION"}}`, "регионе"},
		"лимит":  {429, `{"error":{"message":"Resource exhausted","status":"RESOURCE_EXHAUSTED"}}`, "лимит"},
		"пусто":  {200, answer(`[]`), "ни одного места"},
		"мусор":  {200, `{"candidates":[{"content":{"parts":[{"text":"не json"}]}}]}`, "не список"},
		"отказ":  {200, `{"promptFeedback":{"blockReason":"SAFETY"}}`, "отказался"},
	}
	for name, tc := range cases {
		c := fake(t, tc.status, tc.body, nil)
		_, err := c.Locate(context.Background(), "image/png", []byte{1}, 100, 100)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := (Client{}).Locate(context.Background(), "image/png", nil, 1, 1); err == nil {
		t.Error("без ключа – ошибка")
	}
}

// Package vision – распознавание мест на картинке карты через Google Gemini API (бесплатный тариф подходит).
package vision

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultModel  = "gemini-3.8-flash"
	FallbackModel = "gemini-flash-latest" // если основная модель перегружена
	busyRetries   = 2                     // повторов при «модель перегружена» (503) перед переходом на резервную
	busyWait      = 3 * time.Second
	defaultBase   = "https://generativelanguage.googleapis.com/v1beta"
	maxFound      = 60
	boxScale      = 1000 // Gemini возвращает рамки в долях от 0 до 1000
	timeout       = 2 * time.Minute
	maxBody       = 8 << 20
)

// Found – место, найденное на картинке. X, Y – центр рамки в пикселях картинки.
type Found struct {
	Kind, Name, Note string
	X, Y             float64
}

// Client – клиент Gemini. BaseURL и HTTP подменяются в тестах.
type Client struct {
	Key, Model, BaseURL string
	HTTP                *http.Client
	retryWait           time.Duration // пауза между повторами (в тестах – короткая)
}

// Kinds – виды мест, которые модель может вернуть (совпадают с model.PlaceKinds без «room»).
var Kinds = []string{"capital", "city", "town", "village", "tavern", "temple", "shop", "smithy", "castle", "port",
	"dungeon", "cave", "ruins", "lair", "camp", "tower", "shrine", "mine", "forest", "mountain", "swamp", "lake", "other"}

const prompt = `You are helping a tabletop RPG game master. Analyze this fantasy map image and find the notable locations a GM would use:
settlements (capital, city, town, village), buildings and sites (tavern, temple, shop, smithy, castle, port, tower, shrine),
and wild places (dungeon, cave, ruins, lair, camp, mine, forest, mountain, swamp, lake). Use "other" only if nothing fits.
If the map has text labels, use them as names exactly as written. Otherwise invent a short evocative name in Russian.
For each location write a one-sentence note in Russian describing what is visible there.
Return at most 40 locations, the most important ones first. box_2d is [ymin, xmin, ymax, xmax] normalized to 0-1000.`

type item struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Note string `json:"note"`
	Box  []int  `json:"box_2d"`
}

// Locate отправляет картинку модели и возвращает найденные места в пикселях картинки w×h.
func (c Client) Locate(ctx context.Context, mime string, data []byte, w, h int) ([]Found, error) {
	if strings.TrimSpace(c.Key) == "" {
		return nil, errors.New("не задан ключ Gemini API – укажите его в настройках карты")
	}
	body, err := json.Marshal(c.request(mime, data))
	if err != nil {
		return nil, err
	}
	model := c.Model
	if model == "" {
		model = DefaultModel
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	models := []string{model}
	if model != FallbackModel {
		models = append(models, FallbackModel)
	}
	wait := c.retryWait
	if wait == 0 {
		wait = busyWait
	}
	var lastErr error
	for _, m := range models {
		for try := 0; try <= busyRetries; try++ {
			raw, status, err := c.post(ctx, m, body)
			switch {
			case err != nil:
				return nil, err
			case status == http.StatusOK:
				return parse(raw, w, h)
			case status != http.StatusServiceUnavailable:
				return nil, apiError(status, raw)
			}
			lastErr = apiError(status, raw)
			if m == FallbackModel || try == busyRetries {
				break // резервную модель не мучаем повторами; основную – после busyRetries
			}
			select {
			case <-ctx.Done():
				return nil, lastErr
			case <-time.After(wait):
			}
		}
	}
	return nil, lastErr
}

// post отправляет запрос к модели и возвращает тело ответа и код.
func (c Client) post(ctx context.Context, model string, body []byte) ([]byte, int, error) {
	base := c.BaseURL
	if base == "" {
		base = defaultBase
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/models/"+model+":generateContent", bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.Key)
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: timeout}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("нет связи с Gemini API: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	return raw, resp.StatusCode, err
}

func (c Client) request(mime string, data []byte) map[string]any {
	str := map[string]any{"type": "STRING"}
	return map[string]any{
		"contents": []any{map[string]any{"parts": []any{
			map[string]any{"text": prompt},
			map[string]any{"inline_data": map[string]any{"mime_type": mime, "data": data}}, // []byte → base64 при маршалинге
		}}},
		"generationConfig": map[string]any{
			"temperature":        0.4,
			"response_mime_type": "application/json",
			"response_schema": map[string]any{"type": "ARRAY", "items": map[string]any{"type": "OBJECT",
				"properties": map[string]any{"kind": map[string]any{"type": "STRING", "enum": Kinds}, "name": str, "note": str,
					"box_2d": map[string]any{"type": "ARRAY", "items": map[string]any{"type": "INTEGER"}}},
				"required": []string{"kind", "name", "box_2d"}}},
		},
	}
}

// apiError превращает ответ об ошибке в понятное сообщение (неверный ключ, регион, лимиты).
func apiError(status int, raw []byte) error {
	var e struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &e)
	msg := e.Error.Message
	if msg == "" {
		msg = strings.TrimSpace(string(raw))
	}
	switch {
	case status == http.StatusServiceUnavailable:
		return fmt.Errorf("Gemini сейчас перегружен, попробуйте через минуту (%s)", msg)
	case status == http.StatusTooManyRequests:
		return fmt.Errorf("Gemini: превышен лимит бесплатного тарифа, попробуйте позже (%s)", msg)
	case strings.Contains(msg, "location is not supported"):
		return fmt.Errorf("Gemini недоступен в вашем регионе: %s", msg)
	case status == http.StatusBadRequest && strings.Contains(strings.ToLower(msg), "api key"):
		return fmt.Errorf("Gemini: неверный ключ API (%s)", msg)
	}
	return fmt.Errorf("Gemini API, код %d: %s", status, msg)
}

func parse(raw []byte, w, h int) ([]Found, error) {
	var r struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason"`
		} `json:"promptFeedback"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("Gemini вернул непонятный ответ: %w", err)
	}
	if r.PromptFeedback.BlockReason != "" {
		return nil, fmt.Errorf("Gemini отказался обработать картинку: %s", r.PromptFeedback.BlockReason)
	}
	if len(r.Candidates) == 0 || len(r.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("Gemini не вернул результата")
	}
	var text strings.Builder
	for _, p := range r.Candidates[0].Content.Parts {
		text.WriteString(p.Text)
	}
	var items []item
	if err := json.Unmarshal([]byte(text.String()), &items); err != nil {
		return nil, fmt.Errorf("Gemini вернул не список мест: %w", err)
	}
	out := make([]Found, 0, len(items))
	for _, it := range items {
		if len(it.Box) != 4 || len(out) >= maxFound {
			continue
		}
		ymin, xmin, ymax, xmax := clamp(it.Box[0]), clamp(it.Box[1]), clamp(it.Box[2]), clamp(it.Box[3])
		kind := it.Kind
		if !contains(Kinds, kind) {
			kind = "other"
		}
		out = append(out, Found{Kind: kind, Name: strings.TrimSpace(it.Name), Note: strings.TrimSpace(it.Note),
			X: float64(xmin+xmax) / 2 / boxScale * float64(w), Y: float64(ymin+ymax) / 2 / boxScale * float64(h)})
	}
	if len(out) == 0 {
		return nil, errors.New("Gemini не нашёл на карте ни одного места")
	}
	return out, nil
}

func clamp(v int) int { return max(0, min(boxScale, v)) }

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

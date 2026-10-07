package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"heroesbook/internal/vision"
)

func TestGenerateWorldAndDungeonMaps(t *testing.T) {
	a := newApp()
	m, err := a.GenerateWorld("Эльдория", 7, []string{"human", "dwarf"}, 3, "small")
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "Эльдория" || len(m.Places) < 4 {
		t.Fatalf("мир: %q, мест %d", m.Name, len(m.Places))
	}
	if img, _ := a.GetMapImage(m.ID); !strings.HasPrefix(img, "data:image/jpeg") {
		t.Error("у мира нет картинки")
	}
	if _, err := a.GenerateWorld("", 1, []string{"дракон-эльф"}, 3, "small"); err == nil {
		t.Error("неизвестная раса должна отклоняться")
	}
	d, err := a.GenerateDungeon("", 3, 6, 2)
	if err != nil || d.Name != "Подземелье" || len(d.Places) < 3 {
		t.Fatalf("подземелье: %+v %v", d.Name, err)
	}
}

func TestPlacesEditAndPopulate(t *testing.T) {
	a := newApp()
	m, err := a.GenerateDungeon("x", 3, 5, 2)
	if err != nil {
		t.Fatal(err)
	}
	m, err = a.AddPlace(m.ID, 10, 10, "tavern", "")
	if err != nil {
		t.Fatal(err)
	}
	p := m.Places[len(m.Places)-1]
	if p.Name != "Таверна" || len(p.NPCs) != 0 {
		t.Fatalf("новое место: %+v", p)
	}
	if _, err := a.AddPlace(m.ID, 1, 1, "космодром", ""); err == nil {
		t.Error("неизвестный вид места – ошибка")
	}
	m, err = a.PopulateMap(m.ID, 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Places[len(m.Places)-1]; len(got.NPCs) == 0 {
		t.Error("в таверне после населения должны появиться люди")
	}
	p = m.Places[len(m.Places)-1]
	p.Name, p.Race = "Пьяный гусь", "dwarf"
	if m, err = a.UpdatePlace(m.ID, p); err != nil || m.Places[len(m.Places)-1].Name != "Пьяный гусь" {
		t.Fatalf("переименование: %v", err)
	}
	n := len(m.Places)
	if m, err = a.RemovePlace(m.ID, p.ID); err != nil || len(m.Places) != n-1 {
		t.Fatalf("удаление: %v", err)
	}
	if err := a.Undo(); err != nil || len(a.Maps()[0].Places) != n {
		t.Errorf("отмена должна вернуть место: %v", err)
	}
}

type fakeVision struct{ found []vision.Found }

func (f fakeVision) Locate(context.Context, string, []byte, int, int) ([]vision.Found, error) {
	return f.found, nil
}

func TestRecognizeMapAddsAndPopulates(t *testing.T) {
	a := newApp()
	m, err := a.GenerateDungeon("x", 3, 4, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.RecognizeMap(m.ID, 3); err == nil || !strings.Contains(err.Error(), "ключ") {
		t.Errorf("без ключа – понятная ошибка, получено %v", err)
	}
	if _, err := a.SetAISettings("AIza-test-key-1234", "", false); err != nil {
		t.Fatal(err)
	}
	old := visionClient
	defer func() { visionClient = old }()
	visionClient = func(s aiSettings) interface {
		Locate(ctx context.Context, mime string, data []byte, w, h int) ([]vision.Found, error)
	} {
		if s.GeminiKey != "AIza-test-key-1234" {
			t.Errorf("ключ не дошёл до клиента: %q", s.GeminiKey)
		}
		return fakeVision{[]vision.Found{{Kind: "village", Name: "Ольховка", X: 50, Y: 60}, {Kind: "village", Name: "дубль", X: 52, Y: 61}, {Kind: "lair", Name: "Нора", X: 300, Y: 200}}}
	}
	before := len(m.Places)
	m, err = a.RecognizeMap(m.ID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Places) != before+2 {
		t.Fatalf("ожидалось +2 места (дубль отброшен), стало %d из %d", len(m.Places), before)
	}
	for _, p := range m.Places[before:] {
		if p.Kind == "village" && len(p.NPCs) == 0 {
			t.Error("найденная деревня должна быть населена")
		}
	}
}

func TestAISettingsNeverExposeKey(t *testing.T) {
	a := newApp()
	v, err := a.SetAISettings("AIzaSECRETkey9876", "gemini-test", false)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(v)
	if strings.Contains(string(raw), "SECRET") || !v.HasKey || v.KeyHint != "…9876" || v.Model != "gemini-test" {
		t.Errorf("настройки для интерфейса: %s", raw)
	}
	if v, _ = a.SetAISettings("", "", true); v.HasKey {
		t.Error("ключ должен удаляться")
	}
	if _, err := a.SetAISettings("bad key with spaces", "", false); err == nil {
		t.Error("ключ с пробелами – ошибка")
	}
	if strings.Contains(strings.Join(a.Log(), " "), "SECRET") {
		t.Error("ключ попал в журнал")
	}
}

func TestAIModelNameIsSafe(t *testing.T) {
	a := newApp()
	for _, bad := range []string{"x%2F..", "../models", "a b", `x\y`} {
		if _, err := a.SetAISettings("", bad, false); err == nil {
			t.Errorf("модель %q должна отклоняться", bad)
		}
	}
	if _, err := a.SetAISettings("", "gemini-3.8-flash", false); err != nil {
		t.Errorf("обычное имя модели: %v", err)
	}
}

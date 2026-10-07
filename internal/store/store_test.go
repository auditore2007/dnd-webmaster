package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"heroesbook/internal/model"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	f := &FileStore{Path: filepath.Join(t.TempDir(), "state.json")}
	in := State{Characters: []model.Character{{ID: "a", Name: "Лира"}}, Log: []string{"x"}, SeedVer: 2}
	if err := f.Save(in); err != nil {
		t.Fatal(err)
	}
	out, err := f.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Characters) != 1 || out.Characters[0].Name != "Лира" || out.SeedVer != 2 {
		t.Errorf("прочитано не то: %+v", out)
	}
	b, _ := os.ReadFile(f.Path)
	if strings.Contains(string(b), "\n  ") {
		t.Error("состояние пишется с отступами – файл раздувается")
	}
}

func TestMissingFileIsEmptyState(t *testing.T) {
	f := &FileStore{Path: filepath.Join(t.TempDir(), "none.json")}
	s, err := f.Load()
	if err != nil || len(s.Characters) != 0 {
		t.Fatalf("ожидалось пустое состояние, получено %+v, %v", s, err)
	}
}

func TestCorruptFileMovesToBak(t *testing.T) {
	f := &FileStore{Path: filepath.Join(t.TempDir(), "state.json")}
	if err := os.WriteFile(f.Path, []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Load(); err == nil {
		t.Fatal("ожидалась ошибка для повреждённого файла")
	}
	if _, err := os.Stat(f.Path + ".bak"); err != nil {
		t.Errorf("копия .bak не создана: %v", err)
	}
}

func TestBlobs(t *testing.T) {
	f := &FileStore{Path: filepath.Join(t.TempDir(), "state.json")}
	if err := f.SaveBlob("map-1", "data:image/png;base64,AA"); err != nil {
		t.Fatal(err)
	}
	if got, _ := f.LoadBlob("map-1"); got != "data:image/png;base64,AA" {
		t.Errorf("blob = %q", got)
	}
	if err := f.SaveBlob("map-1", ""); err != nil {
		t.Fatal(err)
	}
	if got, _ := f.LoadBlob("map-1"); got != "" {
		t.Error("пустая строка должна удалять blob")
	}
	if err := f.SaveBlob("missing", ""); err != nil {
		t.Errorf("удаление несуществующего blob не должно быть ошибкой: %v", err)
	}
}

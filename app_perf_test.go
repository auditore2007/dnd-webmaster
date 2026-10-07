package main

import (
	"strings"
	"testing"

	"heroesbook/internal/store"
)

// Снимки игры хранятся отдельными файлами: state.json пишется на каждый клик и не должен тащить их за собой.
func TestSnapshotDataLivesInBlobs(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	if err := a.SaveSnapshot("до боя"); err != nil {
		t.Fatal(err)
	}
	ms := a.store.(*memStore)
	sn := ms.s.Snapshots
	if len(sn) != 1 || len(sn[0].Data) != 0 {
		t.Fatalf("в state.json не должно быть данных снимка: %+v", sn)
	}
	if ms.blobs["snap-"+sn[0].ID] == "" {
		t.Fatal("данные снимка не записаны в отдельный файл")
	}
	if _, err := a.ApplyDamage(h.ID, 3); err != nil {
		t.Fatal(err)
	}
	if err := a.LoadSnapshot(sn[0].ID); err != nil {
		t.Fatal(err)
	}
	if got := a.Characters()[0].HP; got != h.HP {
		t.Errorf("после загрузки HP=%d, ожидалось %d", got, h.HP)
	}
	if err := a.DeleteSnapshot(sn[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, left := ms.blobs["snap-"+sn[0].ID]; left {
		t.Error("файл удалённого снимка остался")
	}
}

func TestOldInlineSnapshotsMigrateToBlobs(t *testing.T) {
	ms := &memStore{s: store.State{Snapshots: []store.Snapshot{{ID: "s1", Name: "старое", Data: []byte(`{"characters":[]}`)}}}}
	a := &App{store: ms, rng: fixed{9}}
	a.startup(nil)
	if len(ms.s.Snapshots[0].Data) != 0 || ms.blobs["snap-s1"] == "" {
		t.Fatalf("снимок не перенесён: %+v", ms.s.Snapshots)
	}
	if err := a.LoadSnapshot("s1"); err != nil {
		t.Fatal(err)
	}
}

func TestQuickSaveKeepsFiveBlobs(t *testing.T) {
	a := newApp()
	hero(t, a)
	for i := 0; i < 7; i++ {
		if _, err := a.QuickSave(); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(a.Snapshots()); n != 5 {
		t.Errorf("быстрых сохранений %d, ожидалось 5", n)
	}
	if n := len(a.store.(*memStore).blobs); n != 5 {
		t.Errorf("файлов снимков %d, ожидалось 5", n)
	}
}

// Undo должен работать и после бросков, которые пишут в журнал без сохранения на диск.
func TestUndoAfterRollKeepsLog(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	if _, err := a.Roll("1d6", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.ApplyDamage(h.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := a.Undo(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(a.Log(), "\n"), "Бросок 1d6") {
		t.Error("отмена потеряла строку броска из журнала")
	}
}

func TestCheckLabelUsesAbilityName(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	v, err := a.Check(h.ID, "str", "check", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(v.Label, "Сила") || strings.Contains(v.Label, "str") {
		t.Errorf("подпись проверки %q, ожидалось название характеристики", v.Label)
	}
	s, err := a.Check(h.ID, "dex", "save", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s.Label, "спасбросок Ловкости") && !strings.Contains(s.Label, "Ловкость") {
		t.Errorf("подпись спасброска %q", s.Label)
	}
}

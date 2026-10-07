package main

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"heroesbook/internal/store"
)

// Снимки игры. Сами данные лежат в отдельных файлах snap-<id>: state.json пишется после каждого действия,
// и двадцать полных копий игры внутри него делали каждое сохранение в разы медленнее.

const maxSnapshots = 20

func snapBlob(id string) string { return "snap-" + id }

// saveSnapshot добавляет снимок текущей игры (без самих снимков). Вызывающий держит мьютекс и сохраняет состояние.
func (a *App) saveSnapshot(name string) (store.Snapshot, error) {
	s := a.state
	s.Snapshots = nil
	b, err := json.Marshal(s)
	if err != nil {
		return store.Snapshot{}, err
	}
	now := time.Now().Format("02.01.2006 15:04")
	if name = strings.TrimSpace(name); name == "" {
		name = "Сохранение " + now
	}
	sn := store.Snapshot{ID: newID(), Name: name, Time: now}
	if bs, ok := a.store.(blobStore); ok {
		if err := bs.SaveBlob(snapBlob(sn.ID), string(b)); err != nil {
			return store.Snapshot{}, err
		}
	} else {
		sn.Data = b
	}
	a.state.Snapshots = append(a.state.Snapshots, sn)
	if n := len(a.state.Snapshots); n > maxSnapshots {
		a.dropSnapshots(a.state.Snapshots[:n-maxSnapshots])
		a.state.Snapshots = append([]store.Snapshot(nil), a.state.Snapshots[n-maxSnapshots:]...)
	}
	return sn, nil
}

// snapshotData возвращает сохранённую игру: из файла снимка или (старый формат) из самого state.json.
func (a *App) snapshotData(sn store.Snapshot) ([]byte, error) {
	if len(sn.Data) > 0 {
		return sn.Data, nil
	}
	bs, ok := a.store.(blobStore)
	if !ok {
		return nil, errors.New("сохранение пустое")
	}
	s, err := bs.LoadBlob(snapBlob(sn.ID))
	if err != nil {
		return nil, err
	}
	if s == "" {
		return nil, errors.New("файл сохранения не найден")
	}
	return []byte(s), nil
}

// dropSnapshots откладывает удаление файлов снимков до успешного persist: если запись state.json не удастся,
// он не будет ссылаться на уже удалённые файлы.
func (a *App) dropSnapshots(list []store.Snapshot) {
	for _, sn := range list {
		a.drops = append(a.drops, snapBlob(sn.ID))
	}
}

// flushDrops удаляет отложенные файлы; ошибка удаления не мешает игре (останется лишний файл).
func (a *App) flushDrops() {
	bs, ok := a.store.(blobStore)
	if ok {
		for _, name := range a.drops {
			_ = bs.SaveBlob(name, "")
		}
	}
	a.drops = nil
}

// migrateSnapshots переносит снимки старого формата (данные внутри state.json) в отдельные файлы.
func (a *App) migrateSnapshots() {
	bs, ok := a.store.(blobStore)
	if !ok {
		return
	}
	for i := range a.state.Snapshots {
		sn := &a.state.Snapshots[i]
		if len(sn.Data) == 0 {
			continue
		}
		if bs.SaveBlob(snapBlob(sn.ID), string(sn.Data)) == nil {
			sn.Data = nil
		}
	}
}

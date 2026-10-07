// Package store — хранение состояния. Repository-интерфейс позволяет заменить файл на БД.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"heroesbook/internal/bestiary"
	"heroesbook/internal/combat"
	"heroesbook/internal/model"
)

type Pin struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Text string  `json:"text"`
}

// MapMeta — карта из галереи; сама картинка лежит в отдельном файле map-<id>.
type MapMeta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Pins []Pin  `json:"pins"`
	Grid int    `json:"grid"` // размер клетки в пикселях картинки, 0 — сетки нет
	Feet int    `json:"feet"` // сколько футов в клетке
}

// Snapshot — именованная копия игры; Data хранит состояние без самих снимков.
type Snapshot struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Time string          `json:"time"`
	Data json.RawMessage `json:"data"`
}

type State struct {
	Characters []model.Character     `json:"characters"`
	Encounter  *combat.Encounter     `json:"encounter"`
	Library    []model.Item          `json:"library"`
	Log        []string              `json:"log"`
	Pins       []Pin                 `json:"pins,omitempty"` // устарело: метки переехали в Maps
	Maps       []MapMeta             `json:"maps"`
	Monsters   []bestiary.Monster    `json:"monsters"` // собственные существа
	Spells     []model.SpellTemplate `json:"spells"`   // собственные заклинания
	Snapshots  []Snapshot            `json:"snapshots"`
	SeedVer    int                   `json:"seedVer"` // версия каталога SRD, уже добавленная в библиотеку
}

type Store interface {
	Load() (State, error)
	Save(State) error
}

type FileStore struct{ Path string }

func Default() (*FileStore, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &FileStore{Path: filepath.Join(dir, "heroes-book", "state.json")}, nil
}

// Load: если файл повреждён, он переименовывается в .bak, а игра стартует с пустого состояния.
func (f *FileStore) Load() (State, error) {
	b, err := os.ReadFile(f.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return State{}, nil
	}
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		bak := f.Path + ".bak"
		_ = os.Rename(f.Path, bak)
		return State{}, fmt.Errorf("файл сохранения повреждён, копия: %s", bak)
	}
	return s, nil
}

// Save пишет во временный файл и переименовывает его: при сбое старое сохранение остаётся целым.
func (f *FileStore) Save(s State) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := f.Path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, f.Path)
}

// Blob — большие данные (своя карта) хранятся рядом с состоянием отдельным файлом, чтобы не раздувать state.json.
func (f *FileStore) blobPath(name string) string {
	return filepath.Join(filepath.Dir(f.Path), name+".blob")
}

func (f *FileStore) SaveBlob(name, data string) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return err
	}
	if data == "" {
		err := os.Remove(f.blobPath(name))
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	tmp := f.blobPath(name) + ".tmp"
	if err := os.WriteFile(tmp, []byte(data), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, f.blobPath(name))
}

func (f *FileStore) LoadBlob(name string) (string, error) {
	b, err := os.ReadFile(f.blobPath(name))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return string(b), err
}

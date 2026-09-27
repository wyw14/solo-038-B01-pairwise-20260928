package service

import (
	"path/filepath"
	"time"

	"github.com/wyw14/solo-038/internal/journal"
)

func runtimeSaveSnapshot(dataDir string, values map[string]any) error {
	return journal.SaveSnapshot(filepath.Join(dataDir, "runtime-snapshot.json"), journal.Snapshot{Version: uint64(time.Now().UnixNano()), Values: values})
}

func runtimeLoadSnapshot(dataDir string) (journal.Snapshot, error) {
	return journal.LoadSnapshot(filepath.Join(dataDir, "runtime-snapshot.json"))
}

package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Snapshot struct {
	Version uint64         `json:"version"`
	Values  map[string]any `json:"values"`
}

func SaveSnapshot(path string, snapshot Snapshot) error {
	if snapshot.Version == 0 {
		return fmt.Errorf("snapshot version is required")
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := path + ".next"
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return err
	}
	file, err := os.Open(temporary)
	if err != nil {
		return err
	}
	if err = file.Sync(); err == nil {
		err = file.Close()
	} else {
		_ = file.Close()
	}
	if err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func LoadSnapshot(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, err
	}
	if snapshot.Version == 0 {
		return Snapshot{}, fmt.Errorf("snapshot version is invalid")
	}
	return snapshot, nil
}

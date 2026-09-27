package journal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wyw14/solo-038/internal/model"
)

type Store struct {
	mu       sync.Mutex
	path     string
	state    *State
	failNext error
}

func NewStore(path string) *Store {
	return &Store{path: path, state: &State{}}
}

func (store *Store) FailNext(err error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.failNext = err
}

func (store *Store) Append(event model.Event) (uint64, error) {
	if err := event.Validate(); err != nil {
		return 0, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	sequence := store.state.Issued()
	if store.failNext != nil {
		err := store.failNext
		store.failNext = nil
		store.state.MarkFailure(err.Error())
		return sequence, err
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o755); err != nil {
		return sequence, err
	}
	file, err := os.OpenFile(store.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return sequence, err
	}
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(event); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		store.state.MarkFailure(err.Error())
		return sequence, fmt.Errorf("append journal event: %w", err)
	}
	if err := store.state.MarkDurable(sequence); err != nil {
		return sequence, err
	}
	return sequence, nil
}

func (store *Store) Events() ([]model.Event, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	file, err := os.Open(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	events := []model.Event{}
	for scanner.Scan() {
		var event model.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		if err := event.Validate(); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}

func (store *Store) State() *State {
	return store.state
}

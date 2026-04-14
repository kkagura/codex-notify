package dedupe

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"codex-notify/internal/event"
)

const retention = 7 * 24 * time.Hour

type Store struct {
	path string
	mu   sync.Mutex
}

type state struct {
	Entries map[string]int64 `json:"entries"`
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func Key(env event.Envelope) string {
	if env.TurnID != "" {
		return "turn:" + env.TurnID
	}

	if env.ThreadID != "" && env.Type != "" {
		h := sha1.Sum(env.RawJSON)
		return "thread:" + env.ThreadID + ":type:" + env.Type + ":sha1:" + hex.EncodeToString(h[:8])
	}

	h := sha1.Sum(env.RawJSON)
	return "sha1:" + hex.EncodeToString(h[:])
}

func (s *Store) Seen(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.load()
	if err != nil {
		return false, err
	}

	now := time.Now()
	prune(st, now)
	_, ok := st.Entries[key]
	if ok {
		if err := s.save(st); err != nil {
			return false, err
		}
	}
	return ok, nil
}

func (s *Store) Mark(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.load()
	if err != nil {
		return err
	}

	now := time.Now()
	prune(st, now)
	st.Entries[key] = now.Unix()
	return s.save(st)
}

func (s *Store) load() (*state, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &state{Entries: map[string]int64{}}, nil
		}
		return nil, err
	}

	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		return &state{Entries: map[string]int64{}}, nil
	}
	if st.Entries == nil {
		st.Entries = map[string]int64{}
	}
	return &st, nil
}

func (s *Store) save(st *state) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func prune(st *state, now time.Time) {
	cutoff := now.Add(-retention).Unix()
	for key, ts := range st.Entries {
		if ts < cutoff {
			delete(st.Entries, key)
		}
	}
}

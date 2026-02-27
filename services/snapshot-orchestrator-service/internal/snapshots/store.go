package snapshots

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("snapshot not found")

type Store struct {
	mu        sync.RWMutex
	snapshots map[string]Snapshot
	now       func() time.Time
}

func NewStore() *Store {
	return &Store{
		snapshots: make(map[string]Snapshot),
		now:       time.Now,
	}
}

func (s *Store) Create(ctx context.Context, projectID string, mode Mode, sourceRef map[string]any) (Snapshot, error) {
	_ = ctx
	snapshot := Snapshot{
		SnapshotID: newSnapshotID(),
		ProjectID:  projectID,
		Mode:       mode,
		Status:     StatusCreated,
		CreatedAt:  s.now(),
		SourceRef:  sourceRef,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[snapshot.SnapshotID] = snapshot
	ready := snapshot
	ready.Status = StatusReady
	s.snapshots[snapshot.SnapshotID] = ready
	return snapshot, nil
}

func (s *Store) Get(ctx context.Context, snapshotID string) (Snapshot, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	snapshot, ok := s.snapshots[snapshotID]
	if !ok {
		return Snapshot{}, ErrNotFound
	}
	return snapshot, nil
}

func newSnapshotID() string {
	var bytes [16]byte
	_, err := rand.Read(bytes[:])
	if err != nil {
		// crypto/rand should not fail in normal operation; fall back to time entropy.
		return "snap_" + hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return "snap_" + hex.EncodeToString(bytes[:])
}

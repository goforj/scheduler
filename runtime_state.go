package scheduler

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type JobEventType string

const (
	JobStarted   JobEventType = "job_started"
	JobSucceeded JobEventType = "job_succeeded"
	JobFailed    JobEventType = "job_failed"
	JobSkipped   JobEventType = "job_skipped"
)

type JobSkipReason string

const (
	JobSkipPaused JobSkipReason = "paused"
)

type JobEvent struct {
	Type       JobEventType
	JobID      uuid.UUID
	Name       string
	TargetKind string
	Attempt    int
	Duration   time.Duration
	Error      error
	Reason     string
	OccurredAt time.Time
}

type JobObserver interface {
	OnJobEvent(ctx context.Context, event JobEvent)
}

type JobObserverFunc func(ctx context.Context, event JobEvent)

func (f JobObserverFunc) OnJobEvent(ctx context.Context, event JobEvent) { f(ctx, event) }

type runtimeState struct {
	mu         sync.RWMutex
	pausedAll  bool
	pausedJobs map[uuid.UUID]bool
	metadata   map[uuid.UUID]JobMetadata
	attempts   map[uuid.UUID]int
	observers  []JobObserver
}

func newRuntimeState() *runtimeState {
	return &runtimeState{
		pausedJobs: make(map[uuid.UUID]bool),
		metadata:   make(map[uuid.UUID]JobMetadata),
		attempts:   make(map[uuid.UUID]int),
	}
}

func (s *runtimeState) pauseAll() {
	s.mu.Lock()
	s.pausedAll = true
	s.mu.Unlock()
}

func (s *runtimeState) resumeAll() {
	s.mu.Lock()
	s.pausedAll = false
	s.mu.Unlock()
}

func (s *runtimeState) isPausedAll() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pausedAll
}

func (s *runtimeState) pauseJob(id uuid.UUID) {
	s.mu.Lock()
	s.pausedJobs[id] = true
	s.mu.Unlock()
}

func (s *runtimeState) resumeJob(id uuid.UUID) {
	s.mu.Lock()
	delete(s.pausedJobs, id)
	s.mu.Unlock()
}

func (s *runtimeState) isJobPaused(id uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pausedJobs[id]
}

func (s *runtimeState) isExecutionPaused(id uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pausedAll || s.pausedJobs[id]
}

func (s *runtimeState) setMetadata(meta JobMetadata) {
	s.mu.Lock()
	s.metadata[meta.ID] = meta
	s.mu.Unlock()
}

func (s *runtimeState) metadataCopy() map[uuid.UUID]JobMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[uuid.UUID]JobMetadata, len(s.metadata))
	for id, meta := range s.metadata {
		meta.Paused = s.pausedAll || s.pausedJobs[id]
		out[id] = meta
	}
	return out
}

func (s *runtimeState) metadataList() []JobMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]JobMetadata, 0, len(s.metadata))
	for id, meta := range s.metadata {
		meta.Paused = s.pausedAll || s.pausedJobs[id]
		out = append(out, meta)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].ID.String() < out[j].ID.String()
		}
		return out[i].Name < out[j].Name
	})

	return out
}

func (s *runtimeState) nextAttempt(id uuid.UUID) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[id]++
	return s.attempts[id]
}

func (s *runtimeState) addObserver(observer JobObserver) {
	if observer == nil {
		return
	}
	s.mu.Lock()
	s.observers = append(s.observers, observer)
	s.mu.Unlock()
}

func (s *runtimeState) emit(ctx context.Context, event JobEvent) {
	s.mu.RLock()
	observers := append([]JobObserver(nil), s.observers...)
	s.mu.RUnlock()

	for _, observer := range observers {
		observer.OnJobEvent(ctx, event)
	}
}

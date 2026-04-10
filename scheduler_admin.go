package scheduler

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidScheduleID = errors.New("invalid schedule id")

// Admin exposes schedule inspection and control operations for operator surfaces.
type Admin struct {
	s *Scheduler
}

// ScheduleInfo describes a scheduled job for UI and admin surfaces.
type ScheduleInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Schedule   string   `json:"schedule"`
	Handler    string   `json:"handler"`
	HandlerRaw string   `json:"handler_raw"`
	Next       string   `json:"next"`
	NextRun    string   `json:"next_run"`
	Tags       []string `json:"tags"`
	Paused     bool     `json:"paused"`
}

// Admin returns a facade for scheduler inspection and control operations.
// @group Metadata
func (s *Scheduler) Admin() *Admin {
	return &Admin{s: s}
}

// ListSchedules returns a stable UI-friendly snapshot of the configured schedules.
// @group Metadata
func (a *Admin) ListSchedules() []ScheduleInfo {
	jobInfos := a.s.JobsInfo()
	nextRunByID := map[uuid.UUID]time.Time{}
	for _, job := range a.s.Jobs() {
		next, err := job.NextRun()
		if err != nil {
			continue
		}
		nextRunByID[job.ID()] = next
	}

	now := time.Now()
	out := make([]ScheduleInfo, 0, len(jobInfos))
	for _, info := range jobInfos {
		nextRun := nextRunByID[info.ID]
		jobType := normalizeScheduleType(info.TargetKind)
		handler := strings.TrimSpace(info.Handler)
		if jobType == "" {
			if strings.Contains(info.Name, ":") {
				jobType = "command"
			} else {
				jobType = "function"
			}
		}
		schedule := strings.TrimSpace(info.Schedule)
		if schedule == "" {
			schedule = "-"
		}
		if handler == "" {
			if jobType == "command" {
				handler = strings.TrimSpace(info.Command)
				if handler == "" {
					handler = "-"
				}
			} else {
				handler = info.Name
			}
		}
		name := strings.TrimSpace(info.Name)
		if name == "" {
			name = strings.TrimSpace(info.Command)
		}
		if name == "" {
			name = "-"
		}
		rawHandler := handler
		out = append(out, ScheduleInfo{
			ID:         info.ID.String(),
			Name:       formatScheduleDisplayName(name),
			Type:       jobType,
			Schedule:   schedule,
			Handler:    formatScheduleDisplayName(handler),
			HandlerRaw: rawHandler,
			Next:       formatRelativeTime(now, nextRun),
			NextRun:    nextRun.Format(time.RFC3339),
			Tags:       info.Tags,
			Paused:     info.Paused,
		})
	}
	return out
}

// PauseSchedule pauses a schedule by id.
// @group Metadata
func (a *Admin) PauseSchedule(id string) error {
	jobID, err := parseScheduleID(id)
	if err != nil {
		return err
	}
	return a.s.PauseJob(jobID)
}

// ResumeSchedule resumes a paused schedule by id.
// @group Metadata
func (a *Admin) ResumeSchedule(id string) error {
	jobID, err := parseScheduleID(id)
	if err != nil {
		return err
	}
	return a.s.ResumeJob(jobID)
}

// RestartSchedule resumes a schedule if needed, runs it immediately, and restores pause state.
// @group Metadata
func (a *Admin) RestartSchedule(id string) error {
	jobID, err := parseScheduleID(id)
	if err != nil {
		return err
	}
	found := false
	wasAllPaused := a.s.IsPausedAll()
	wasJobPaused := a.s.IsJobPaused(jobID)
	if wasAllPaused {
		if err := a.s.ResumeAll(); err != nil {
			return err
		}
		defer func() {
			_ = a.s.PauseAll()
		}()
	} else if wasJobPaused {
		if err := a.s.ResumeJob(jobID); err != nil {
			return err
		}
		defer func() {
			_ = a.s.PauseJob(jobID)
		}()
	}
	for _, job := range a.s.Jobs() {
		if job.ID() != jobID {
			continue
		}
		found = true
		if err := job.RunNow(); err != nil {
			return err
		}
		break
	}
	if !found {
		return nil
	}
	return nil
}

// IsPausedAll reports whether all schedules are currently paused.
// @group Metadata
func (a *Admin) IsPausedAll() bool {
	return a.s.IsPausedAll()
}

// PauseAll halts job execution without removing schedule definitions.
// @group Metadata
func (a *Admin) PauseAll() error {
	return a.s.PauseAll()
}

// ResumeAll restarts job execution for all schedules.
// @group Metadata
func (a *Admin) ResumeAll() error {
	return a.s.ResumeAll()
}

func parseScheduleID(id string) (uuid.UUID, error) {
	jobID, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, ErrInvalidScheduleID
	}
	return jobID, nil
}

func normalizeScheduleType(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return "function"
	}
	return trimmed
}

func formatRelativeTime(now, next time.Time) string {
	if next.IsZero() {
		return "-"
	}
	if next.Before(now) {
		return "now"
	}
	return "in " + time.Until(next).Round(time.Second).String()
}

func formatScheduleDisplayName(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	if strings.Contains(value, ":") {
		return value
	}
	if idx := strings.LastIndex(value, "/"); idx >= 0 && idx < len(value)-1 {
		value = value[idx+1:]
	}
	value = strings.ReplaceAll(value, "(*", "")
	value = strings.ReplaceAll(value, ").", ".")
	return value
}

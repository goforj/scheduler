package scheduler

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

func TestGetJobsListUsesMetadata(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	jb := NewJobBuilder(s)

	jb.Cron("0 * * * *").Command("hello:world", "weekly")
	jb.Every(2).Minutes().Name("funcJob").Do(func() {})

	entries := jb.getJobsList()
	require.Len(t, entries, 2)

	// Ensure sorting by name and job metadata
	require.Equal(t, "funcJob", entries[0].Name)
	require.Equal(t, jobScheduleInterval, entries[0].ScheduleType)
	require.Equal(t, jobTargetCommand, entries[1].TargetKind)
	require.Contains(t, entries[1].Schedule, "0 * * * *")
}

func TestJobsListHelpers(t *testing.T) {
	now := time.Now()
	next := now.Add(90 * time.Minute)

	require.Equal(t, "cron 0 * * * *", formatSchedule("0 * * * *", jobScheduleCron))
	require.Equal(t, "every 1m0s", formatSchedule("1m0s", jobScheduleInterval))
	require.Equal(t, "other", formatSchedule("other", jobScheduleUnknown))

	require.Equal(t, []string{"env=dev"}, filterTags([]string{"env=dev", "cron=1", "name=foo"}))

	require.Equal(t, "(unnamed)", nameForEntry(&jobEntry{}))
	require.Equal(t, "target", nameForEntry(&jobEntry{Target: "target"}))

	sched, kind := scheduleFromTags([]string{"interval=1m"})
	require.Equal(t, "1m", sched)
	require.Equal(t, jobScheduleInterval, kind)

	require.Contains(t, renderTarget("handler", jobTargetFunction, "name"), "handler")
	require.Contains(t, renderTarget("", jobTargetCommand, "name"), "-")

	require.Contains(t, renderNextRunWithWidth(time.Time{}, nil, 5), "-")
	require.Contains(t, renderNextRunWithWidth(next, nil, 8), "in")
	require.Contains(t, renderNextRunWithWidth(next, nil, 8), "Dec")
	require.Contains(t, renderJobType(jobTargetCommand), "command")

	require.Equal(t, "in 59s", humanizeDuration(59*time.Second))
	require.Equal(t, "in 2m", humanizeDuration(2*time.Minute))
	require.Equal(t, "in 1h30m", humanizeDuration(90*time.Minute))
	require.Equal(t, "in 2d", humanizeDuration(48*time.Hour))
	require.Equal(t, "in 2w", humanizeDuration(15*24*time.Hour))

	require.Contains(t, formatNextRun(time.Now().Add(time.Hour)), "in")
	require.NotEmpty(t, shortTimestamp(now))
}

func TestPrintJobsList(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	jb := NewJobBuilder(s)

	jb.EveryMinute().Command("hello:world")
	jb.EveryTwoMinutes().Do(func() {})

	// Capture stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	defer func() { os.Stdout = stdout }()
	os.Stdout = w

	jb.PrintJobsList()

	require.NoError(t, w.Close())
	outBytes, _ := io.ReadAll(r)
	out := string(outBytes)

	require.Contains(t, out, "Scheduler Jobs")
	require.Contains(t, out, "hello:world")
}

func TestPrintJobsListEmpty(t *testing.T) {
	// Capture stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	defer func() { os.Stdout = stdout }()
	os.Stdout = w

	NewJobBuilder(nil).PrintJobsList()

	require.NoError(t, w.Close())
	outBytes, _ := io.ReadAll(r)
	out := string(outBytes)

	require.Contains(t, out, "No jobs registered.")
}

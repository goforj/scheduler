package scheduler

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type jobEntry struct {
	Name         string
	TargetKind   jobTargetKind
	Schedule     string
	ScheduleType jobScheduleKind
	Target       string
	Tags         []string
	NextRun      time.Time
	NextRunErr   error
}

var (
	jobStyleName        = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	jobStyleType        = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	jobStyleSchedule    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	jobStyleCron        = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	jobStyleTargetCmd   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	jobStyleTargetFunc  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	jobStyleTags        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	jobStyleNextRun     = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	jobStyleNextRunErr  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	jobStyleMissingInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
)

// getJobsList builds a slice of job entries enriched with metadata from the fluent wrapper.
func (j *JobBuilder) getJobsList() []*jobEntry {
	if j == nil || j.scheduler == nil {
		return nil
	}

	if ad, ok := j.scheduler.(gocronSchedulerAdapter); ok && ad.s == nil {
		return nil
	}

	jobs := j.scheduler.Jobs()
	metadata := j.JobMetadata()
	var entries []*jobEntry

	for _, job := range jobs {
		meta := metadata[job.ID()]
		schedule := meta.Schedule
		scheduleType := jobScheduleKind(meta.ScheduleType)
		targetKind := jobTargetKind(meta.TargetKind)
		target := meta.Command
		if target == "" {
			target = meta.Handler
		}

		name := job.Name()
		if name == "" {
			name = meta.Name
		}
		tags := job.Tags()
		if len(tags) == 0 && len(meta.Tags) > 0 {
			tags = meta.Tags
		}

		if targetKind == jobTargetFunction && meta.Handler != "" {
			if name == "" || strings.Contains(name, "/") || strings.Contains(name, "(*") {
				name = meta.Handler
			}
		}

		if schedule == "" {
			schedule, scheduleType = scheduleFromTags(tags)
		}
		if targetKind == "" {
			targetKind = jobTargetFunction
		}
		if name == "" {
			name = target
		}
		if name == "" {
			name = "(unnamed)"
		}

		next, err := job.NextRun()

		entries = append(entries, &jobEntry{
			Name:         name,
			TargetKind:   targetKind,
			Schedule:     schedule,
			ScheduleType: scheduleType,
			Target:       target,
			Tags:         tags,
			NextRun:      next,
			NextRunErr:   err,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}

// PrintJobsList renders and prints the scheduler job table to stdout.
// @group Diagnostics
//
// Example: print current jobs
//
//	s, _ := gocron.NewScheduler()
//	s.Start()
//	defer s.Shutdown()
//
//	scheduler.NewJobBuilder(s).
//		EverySecond().
//		Name("heartbeat").
//		Do(func() {})
//
//	scheduler.NewJobBuilder(s).PrintJobsList()
func (j *JobBuilder) PrintJobsList() {
	entries := j.getJobsList()
	if len(entries) == 0 {
		fmt.Println("No jobs registered.")
		return
	}
	rows := jobEntriesToRows(entries)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1).
		Foreground(lipgloss.Color("15"))
	cellStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	t := table.New().
		Border(lipgloss.ASCIIBorder()).
		BorderStyle(
			lipgloss.NewStyle().Foreground(
				lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"},
			),
		).
		BorderBottom(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
		}).
		Headers("Name", "Type", "Schedule", "Handler", "Next Run", "Tags").
		Rows(rows...)

	borderColor := lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	rendered := t.Render()
	tableWidth := lipgloss.Width(rendered)
	innerWidth := tableWidth - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	borderLine := borderStyle.Render("+" + strings.Repeat("-", innerWidth) + "+")

	pipe := lipgloss.NewStyle().Foreground(borderColor).Render("|")
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15"))
	label := labelStyle.Render(fmt.Sprintf(" Scheduler Jobs › (%d)", len(entries)))

	fmt.Println(borderLine)
	fmt.Println(pipe + label)
	fmt.Println(rendered)
}

func jobEntriesToRows(entries []*jobEntry) [][]string {
	var rows [][]string
	maxRelWidth := maxRelativeWidth(entries)
	for _, entry := range entries {
		nameRaw := nameForEntry(entry)
		name := jobStyleName.Render(nameRaw)
		target := renderTarget(entry.Target, entry.TargetKind, nameRaw)
		rows = append(rows, []string{
			name,
			renderJobType(entry.TargetKind),
			renderSchedule(entry.Schedule, entry.ScheduleType),
			target,
			renderNextRunWithWidth(entry.NextRun, entry.NextRunErr, maxRelWidth),
			renderTags(entry.Tags),
		})
	}
	return rows
}

func renderJobType(kind jobTargetKind) string {
	if kind == "" {
		return jobStyleMissingInfo.Render("unknown")
	}
	return jobStyleType.Render(string(kind))
}

func renderSchedule(schedule string, kind jobScheduleKind) string {
	if schedule == "" {
		return jobStyleMissingInfo.Render("-")
	}
	schedule = formatSchedule(schedule, kind)
	switch kind {
	case jobScheduleCron:
		return jobStyleCron.Render(schedule)
	case jobScheduleInterval:
		return jobStyleSchedule.Render(schedule)
	default:
		return jobStyleSchedule.Render(schedule)
	}
}

func renderTarget(target string, kind jobTargetKind, nameRaw string) string {
	if target == "" || target == nameRaw {
		return jobStyleMissingInfo.Render("-")
	}
	if kind == jobTargetCommand {
		return jobStyleTargetCmd.Render(target)
	}
	return jobStyleTargetFunc.Render(target)
}

func nameForEntry(entry *jobEntry) string {
	if entry == nil {
		return "-"
	}
	name := entry.Name
	if name == "" && entry.Target != "" {
		name = entry.Target
	}
	if name == "" {
		return "(unnamed)"
	}
	return name
}

func renderTags(tags []string) string {
	filtered := filterTags(tags)
	if len(filtered) == 0 {
		return jobStyleMissingInfo.Render("-")
	}
	return jobStyleTags.Render(strings.Join(filtered, ", "))
}

func scheduleFromTags(tags []string) (string, jobScheduleKind) {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "cron=") {
			return strings.TrimPrefix(tag, "cron="), jobScheduleCron
		}
		if strings.HasPrefix(tag, "interval=") {
			return strings.TrimPrefix(tag, "interval="), jobScheduleInterval
		}
	}
	return "", jobScheduleUnknown
}

func filterTags(tags []string) []string {
	var filtered []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, "cron=") || strings.HasPrefix(tag, "interval=") || strings.HasPrefix(tag, "name=") {
			continue
		}
		filtered = append(filtered, tag)
	}
	return filtered
}

func formatSchedule(schedule string, kind jobScheduleKind) string {
	if schedule == "" {
		return schedule
	}
	switch kind {
	case jobScheduleCron:
		return "cron " + schedule
	case jobScheduleInterval:
		return "every " + schedule
	default:
		return schedule
	}
}

func formatNextRun(next time.Time) string {
	local := next.Local()
	relative := humanizeDuration(time.Until(local))
	short := shortTimestamp(local)
	return fmt.Sprintf("%-10s · %s", relative, short)
}

func renderNextRunWithWidth(next time.Time, err error, width int) string {
	if err != nil {
		return jobStyleNextRunErr.Render("error: " + err.Error())
	}
	if next.IsZero() {
		return jobStyleMissingInfo.Render("-")
	}
	if width <= 0 {
		return jobStyleNextRun.Render(formatNextRun(next))
	}
	local := next.Local()
	relative := humanizeDuration(time.Until(local))
	short := shortTimestamp(local)
	return jobStyleNextRun.Render(fmt.Sprintf("%-*s %s", width, relative, short))
}

func maxRelativeWidth(entries []*jobEntry) int {
	width := 0
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		if entry.NextRunErr != nil || entry.NextRun.IsZero() {
			if width < 1 {
				width = 1
			}
			continue
		}
		local := entry.NextRun.Local()
		rel := humanizeDuration(time.Until(local))
		if l := len(rel); l > width {
			width = l
		}
	}
	if width == 0 {
		width = 1
	}
	return width
}

func humanizeDuration(d time.Duration) string {
	if d < 0 {
		return "past due"
	}

	seconds := int(d.Round(time.Second).Seconds())
	if seconds < 60 {
		return fmt.Sprintf("in %ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("in %dm", minutes)
	}
	hours := minutes / 60
	if hours < 48 {
		remM := minutes % 60
		if remM == 0 {
			return fmt.Sprintf("in %dh", hours)
		}
		return fmt.Sprintf("in %dh%dm", hours, remM)
	}
	days := hours / 24
	if days < 14 {
		return fmt.Sprintf("in %dd", days)
	}
	weeks := days / 7
	return fmt.Sprintf("in %dw", weeks)
}

func shortTimestamp(t time.Time) string {
	// Example: Dec 17@1AM or Dec 17@1:15PM
	str := t.Format("Jan 2 3:04PM")
	str = strings.Replace(str, ":00AM", "AM", 1)
	str = strings.Replace(str, ":00PM", "PM", 1)
	return str
}

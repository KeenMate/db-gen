package helpers

import (
	"fmt"
	"strings"
	"time"
)

type entry struct {
	time  time.Time
	label string
}

type Timer struct {
	timers []entry
}

func NewTimer() *Timer {
	timer := &Timer{timers: []entry{}}
	timer.AddEntry("start")

	return timer
}

func (t *Timer) Finish() {
	t.AddEntry("finish")
}

func (t *Timer) String() string {
	if len(t.timers) < 2 {
		return "Timer not stopped"
	}
	var str strings.Builder

	startTime := t.timers[0].time
	endTime := t.timers[len(t.timers)-1].time
	totalExecTime := endTime.Sub(startTime)
	str.WriteString(fmt.Sprintf("Total execution time %s \n", totalExecTime))

	for i, entry := range t.timers[1 : len(t.timers)-1] {
		sinceLast := entry.time.Sub(t.timers[i].time)
		sinceStart := entry.time.Sub(startTime)
		str.WriteString(fmt.Sprintf("   %s: %s (%s)\n", entry.label, sinceLast, sinceStart))
	}
	return str.String()
}

func (t *Timer) AddEntry(label string) {
	t.timers = append(t.timers, entry{time: time.Now(), label: label})
}

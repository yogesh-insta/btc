package schedule

import (
	"fmt"
	"time"
)

func LoadLocation(name string) *time.Location {
	if name == "" {
		name = "Australia/Sydney"
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

func IsWeekday(t time.Time, loc *time.Location) bool {
	wd := t.In(loc).Weekday()
	return wd >= time.Monday && wd <= time.Friday
}

func LocalDate(t time.Time, loc *time.Location) string {
	return t.In(loc).Format("2006-01-02")
}

func ISOWeek(t time.Time, loc *time.Location) string {
	y, w := t.In(loc).ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}

// DueAlertSlot returns a slot key (e.g. "2026-07-03-08") for the earliest
// unsent alert hour that has already passed today in loc.
func DueAlertSlot(now time.Time, loc *time.Location, hours []int, sent []string) (string, bool) {
	if len(hours) == 0 {
		return "", false
	}
	local := now.In(loc)
	today := local.Format("2006-01-02")
	currentHour := local.Hour()

	sentSet := make(map[string]bool, len(sent))
	for _, s := range sent {
		sentSet[s] = true
	}

	for _, h := range hours {
		if currentHour < h {
			continue
		}
		slot := fmt.Sprintf("%s-%02d", today, h)
		if !sentSet[slot] {
			return slot, true
		}
	}
	return "", false
}

// AlertSlotLabel formats a slot key for email subjects.
func AlertSlotLabel(slot string, loc *time.Location) string {
	if len(slot) < 13 {
		return slot
	}
	datePart := slot[:10]
	hourPart := slot[11:]
	return fmt.Sprintf("%s %s:00 %s", datePart, hourPart, loc.String())
}

// DueOncePerDayHour returns true only during the target local hour and new date.
func DueOncePerDayHour(now time.Time, loc *time.Location, hour int, lastDate string) bool {
	local := now.In(loc)
	if local.Hour() != hour {
		return false
	}
	today := local.Format("2006-01-02")
	return today != lastDate
}

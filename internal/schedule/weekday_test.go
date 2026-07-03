package schedule

import (
	"testing"
	"time"
)

func TestIsWeekday(t *testing.T) {
	loc := time.UTC
	mon := time.Date(2026, 7, 6, 12, 0, 0, 0, loc)
	sat := time.Date(2026, 7, 11, 12, 0, 0, 0, loc)
	if !IsWeekday(mon, loc) {
		t.Fatal("expected Monday to be weekday")
	}
	if IsWeekday(sat, loc) {
		t.Fatal("expected Saturday not to be weekday")
	}
}

func TestISOWeek(t *testing.T) {
	loc := time.UTC
	got := ISOWeek(time.Date(2026, 7, 6, 12, 0, 0, 0, loc), loc)
	if got != "2026-W28" {
		t.Fatalf("ISOWeek = %q, want 2026-W28", got)
	}
}

func TestDueAlertSlot(t *testing.T) {
	loc := time.UTC
	hours := []int{8, 18}
	now := time.Date(2026, 7, 3, 9, 0, 0, 0, loc)

	slot, ok := DueAlertSlot(now, loc, hours, nil)
	if !ok || slot != "2026-07-03-08" {
		t.Fatalf("DueAlertSlot = %q %v, want 2026-07-03-08 true", slot, ok)
	}

	slot, ok = DueAlertSlot(now, loc, hours, []string{"2026-07-03-08"})
	if ok {
		t.Fatalf("expected no slot after morning sent, got %q", slot)
	}

	evening := time.Date(2026, 7, 3, 19, 0, 0, 0, loc)
	slot, ok = DueAlertSlot(evening, loc, hours, []string{"2026-07-03-08"})
	if !ok || slot != "2026-07-03-18" {
		t.Fatalf("DueAlertSlot evening = %q %v, want 2026-07-03-18 true", slot, ok)
	}
}

package utils

import (
	"time"
)

// Period type to represent a point in modularized time.
/**
 * e.g: Every monday at 10:00.
 * duration: period duration is one week.
 * margin: if margin is 10 hours, then this period is every monday.
 *
 * e.g: Every 18.00 in the evening.
 * duration: period duration is one day.
 * margin: if margin is 18 hours, then this period is every day at 18.00.
 */
type Period struct {
	Duration time.Duration `json:"duration"`
	Margin   time.Duration `json:"margin"`
}

type RangedPeriod struct {
	Duration    time.Duration `json:"duration"`
	StartMargin time.Duration `json:"startMargin"`
	EndMargin   time.Duration `json:"endMargin"`
}

type PeriodicPlan struct {
	Periods []RangedPeriod `json:"periods,omitempty"`
}

func (plan *PeriodicPlan) RegisterPeriod(period RangedPeriod) {
	plan.Periods = append(plan.Periods, period)
}

func (plan *PeriodicPlan) IsIn(time time.Time) bool {
	for _, rangedPeriod := range plan.Periods {
		if rangedPeriod.IsIn(time) {
			return true
		}
	}
	return false
}

/* returns start period and end period */
func (rangedPeriod *RangedPeriod) getStartEndPeriods() (Period, Period) {
	return Period{rangedPeriod.Duration, rangedPeriod.StartMargin}, Period{rangedPeriod.Duration, rangedPeriod.EndMargin}
}

func (rangedPeriod *RangedPeriod) IsIn(time time.Time) bool {
	startPeriod, endPeriod := rangedPeriod.getStartEndPeriods()
	start := startPeriod.NextPointBefore(time)
	end := endPeriod.NextPointAfter(time)
	if end.UnixNano()-start.UnixNano() > rangedPeriod.Duration.Nanoseconds() {
		return false
	}
	return time.After(start) && time.Before(end)
}

func (period *Period) NextPointBefore(t time.Time) time.Time {
	return period.NextPointAfter(t.Add(-period.Duration))
}

func (period *Period) NextPointAfter(t time.Time) time.Time {
	nowNano := t.UnixNano()
	periodNano := period.Duration.Nanoseconds()

	nextNano := nowNano - (nowNano % periodNano) // align
	nextNano += period.Margin.Nanoseconds()      // add margin

	if nowNano > nextNano {
		nextNano += period.Duration.Nanoseconds() // if margin was not enough to be next point, next period will.
	}

	newTime := time.Unix(0, nextNano)
	zone, _ := t.Zone()
	location, _ := time.LoadLocation(zone)
	result := newTime.In(location)
	return result
}

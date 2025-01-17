package utils

import (
	"testing"
	"time"
)

func TestPeriodNextPoint(t *testing.T) {
	period := Period{
		Duration: 24 * time.Hour,
		Margin:   18 * time.Hour,
	}

	testTime := time.Date(2023, 10, 1, 17, 0, 0, 0, time.UTC)
	expectedTime := time.Date(2023, 10, 1, 18, 0, 0, 0, time.UTC)

	result := period.NextPointAfter(testTime)
	if !result.Equal(expectedTime) {
		t.Errorf("FAILED 1 expected %v, got %v", expectedTime, result)
	}

	testTime = time.Date(2023, 10, 1, 19, 0, 0, 0, time.UTC)
	expectedTime = time.Date(2023, 10, 2, 18, 0, 0, 0, time.UTC)

	result = period.NextPointAfter(testTime)
	if !result.Equal(expectedTime) {
		t.Errorf("FAILED 2 expected %v, got %v", expectedTime, result)
	}
}

func TestRangedPeriodIsIn(t *testing.T) {
	rangedPeriod := RangedPeriod{
		Duration:    24 * time.Hour,
		StartMargin: 8 * time.Hour,
		EndMargin:   18 * time.Hour,
	}

	testTime := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	if !rangedPeriod.IsIn(testTime) {
		t.Errorf("expected %v to be in range", testTime)
	}

	testTime = time.Date(2023, 10, 1, 20, 0, 0, 0, time.UTC)
	if rangedPeriod.IsIn(testTime) {
		t.Errorf("expected %v to be out of range", testTime)
	}
}

func TestPeriodicPlanIsIn(t *testing.T) {
	rangedPeriod := RangedPeriod{
		Duration:    24 * time.Hour,
		StartMargin: 8 * time.Hour,
		EndMargin:   18 * time.Hour,
	}

	plan := PeriodicPlan{}
	plan.RegisterPeriod(rangedPeriod)

	testTime := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	if !plan.IsIn(testTime) {
		t.Errorf("expected %v to be in plan", testTime)
	}

	testTime = time.Date(2023, 10, 1, 20, 0, 0, 0, time.UTC)
	if plan.IsIn(testTime) {
		t.Errorf("expected %v to be out of plan", testTime)
	}
}

func TestPeriodNextPointBefore(t *testing.T) {
	period := Period{
		Duration: 24 * time.Hour,
		Margin:   18 * time.Hour,
	}

	testTime := time.Date(2023, 10, 2, 19, 0, 0, 0, time.UTC)
	expectedTime := time.Date(2023, 10, 2, 18, 0, 0, 0, time.UTC)

	result := period.NextPointBefore(testTime)
	if !result.Equal(expectedTime) {
		t.Errorf("FAILED 1 expected %v, got %v", expectedTime, result)
	}

	testTime = time.Date(2023, 10, 3, 19, 0, 0, 0, time.UTC)
	expectedTime = time.Date(2023, 10, 3, 18, 0, 0, 0, time.UTC)

	result = period.NextPointBefore(testTime)
	if !result.Equal(expectedTime) {
		t.Errorf("FAILED 2 expected %v, got %v", expectedTime, result)
	}
}

func TestRangedPeriodGetStartEndPeriods(t *testing.T) {
	rangedPeriod := RangedPeriod{
		Duration:    24 * time.Hour,
		StartMargin: 8 * time.Hour,
		EndMargin:   18 * time.Hour,
	}

	startPeriod, endPeriod := rangedPeriod.getStartEndPeriods()

	if startPeriod.Margin != 8*time.Hour {
		t.Errorf("expected start Margin to be 8 hours, got %v", startPeriod.Margin)
	}

	if endPeriod.Margin != 18*time.Hour {
		t.Errorf("expected end Margin to be 18 hours, got %v", endPeriod.Margin)
	}
}

func TestPeriodicPlanRegisterPeriod(t *testing.T) {
	rangedPeriod := RangedPeriod{
		Duration:    24 * time.Hour,
		StartMargin: 8 * time.Hour,
		EndMargin:   18 * time.Hour,
	}

	plan := PeriodicPlan{}
	plan.RegisterPeriod(rangedPeriod)

	if len(plan.Periods) != 1 {
		t.Errorf("expected 1 period, got %d", len(plan.Periods))
	}

	if plan.Periods[0] != rangedPeriod {
		t.Errorf("expected %v, got %v", rangedPeriod, plan.Periods[0])
	}
}

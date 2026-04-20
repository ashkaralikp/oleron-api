package timesheet

import (
	"math"
	"time"
)

type EstimateInput struct {
	Year          int
	Month         int
	SupportHours  float64
	OvertimeHours float64
	FixedMonthlySalary float64
	OTRate        float64
}

type EstimateResult struct {
	WholeMonthHours    float64
	HourlyRate         float64
	Scenario           string  // Full | Over | Partial
	EstimatedPay       float64
}

// Compute calculates the pay estimate using the actual number of weekdays
// in the given month rather than the fixed 173.33 average.
func Compute(in EstimateInput) EstimateResult {
	wholeMonthHours := float64(weekdaysInMonth(in.Year, in.Month)) * 8.0
	hourlyRate := 0.0
	if wholeMonthHours > 0 {
		hourlyRate = in.FixedMonthlySalary / wholeMonthHours
	}

	var scenario string
	var estimatedPay float64

	fullMonthCovered := in.SupportHours >= wholeMonthHours
	switch {
	case fullMonthCovered && in.OvertimeHours > 0:
		// Regular hours met + extra OT hours logged → pay fixed + OT
		scenario = "Over"
		estimatedPay = in.FixedMonthlySalary + (in.OvertimeHours * in.OTRate)
	case fullMonthCovered:
		// Regular hours met, no OT
		scenario = "Full"
		estimatedPay = in.FixedMonthlySalary
	default:
		// Worked less than the full month; OT not applicable
		scenario = "Partial"
		estimatedPay = in.SupportHours * hourlyRate
	}

	return EstimateResult{
		WholeMonthHours: round2(wholeMonthHours),
		HourlyRate:      round2(hourlyRate),
		Scenario:        scenario,
		EstimatedPay:    round2(estimatedPay),
	}
}

// weekdaysInMonth counts Mon–Fri days in the given month/year.
func weekdaysInMonth(year, month int) int {
	count := 0
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	for t.Month() == time.Month(month) {
		wd := t.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
		t = t.AddDate(0, 0, 1)
	}
	return count
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

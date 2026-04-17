package attendance

import (
	"context"
	"errors"
	"math"
	"time"

	"rmp-api/internal/models"
)

const gracePeriodMinutes = 15 // minutes after start_time before considered late

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Punch handles both punch-in and punch-out for the calling user.
func (s *Service) Punch(ctx context.Context, userID string) (*PunchResult, error) {
	now := time.Now()

	// Block if today is a branch holiday
	isHoliday, _, err := s.repo.IsHoliday(ctx, userID)
	if err != nil {
		return nil, err
	}
	if isHoliday {
		return nil, errors.New("today is a public holiday")
	}

	// Resolve today's day timing (nil = no timing configured, allow punch without logic)
	dayOfWeek := int(now.Weekday()) // 0=Sun ... 6=Sat
	timing, err := s.repo.FindDayTiming(ctx, userID, dayOfWeek)
	if err != nil {
		timing = nil
	}

	if timing != nil && !timing.IsWorkingDay {
		return nil, errors.New("today is not a working day")
	}

	// Check existing record for today
	existing, err := s.repo.FindTodayRecord(ctx, userID)
	if err != nil {
		// No record — do punch-in
		return s.doPunchIn(ctx, userID, now, timing)
	}

	// Record exists — do punch-out if not already done
	if existing.PunchOut != nil {
		return nil, errors.New("already punched out for today")
	}
	return s.doPunchOut(ctx, existing, now, timing)
}

func (s *Service) doPunchIn(ctx context.Context, userID string, now time.Time, timing *DayTiming) (*PunchResult, error) {
	status := "present"

	if timing != nil && timing.StartTime != nil {
		expectedStart, err := parseDayTime(now, *timing.StartTime)
		if err == nil && now.After(expectedStart.Add(time.Duration(gracePeriodMinutes)*time.Minute)) {
			status = "late_in"
		}
	}

	a, err := s.repo.CreatePunchIn(ctx, userID, now, status)
	if err != nil {
		return nil, err
	}
	return recordToResult("punch_in", a), nil
}

func (s *Service) doPunchOut(ctx context.Context, rec *models.Attendance, now time.Time, timing *DayTiming) (*PunchResult, error) {
	rawHours := now.Sub(*rec.PunchIn).Hours()
	workHours := math.Round(rawHours*100) / 100

	status := rec.Status
	wasLateIn := status == "late_in"

	if timing != nil {
		expected := expectedWorkHours(timing)

		switch {
		case expected > 0 && workHours < (expected/2):
			status = "half_day"
		case timing.EndTime != nil:
			expectedEnd, err := parseDayTime(now, *timing.EndTime)
			if err == nil && now.Before(expectedEnd) {
				if wasLateIn {
					status = "late_in_early_out"
				} else {
					status = "early_out"
				}
			}
		}
	}

	a, err := s.repo.UpdatePunchOut(ctx, rec.ID, now, workHours, status)
	if err != nil {
		return nil, err
	}
	return recordToResult("punch_out", a), nil
}

// GetToday returns the calling user's attendance record for today.
func (s *Service) GetToday(ctx context.Context, userID string) (*TodayResult, error) {
	a, err := s.repo.FindTodayRecord(ctx, userID)
	if err != nil {
		return &TodayResult{PunchedIn: false, PunchedOut: false}, nil
	}

	return &TodayResult{
		PunchedIn:  a.PunchIn != nil,
		PunchedOut: a.PunchOut != nil,
		ID:         &a.ID,
		UserID:     &a.UserID,
		WorkDate:   &a.WorkDate,
		PunchIn:    a.PunchIn,
		PunchOut:   a.PunchOut,
		WorkHours:  a.WorkHours,
		Status:     &a.Status,
		Notes:      a.Notes,
	}, nil
}

// recordToResult maps a models.Attendance to PunchResult.
func recordToResult(action string, a *models.Attendance) *PunchResult {
	return &PunchResult{
		Action:    action,
		ID:        a.ID,
		UserID:    a.UserID,
		WorkDate:  a.WorkDate,
		PunchIn:   a.PunchIn,
		PunchOut:  a.PunchOut,
		WorkHours: a.WorkHours,
		Status:    a.Status,
		Notes:     a.Notes,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// parseDayTime builds a time.Time for today from a "HH:MM:SS" string.
func parseDayTime(today time.Time, t string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05",
		today.Format("2006-01-02")+" "+t, today.Location())
}

// expectedWorkHours returns net expected hours (total − break) from day timing.
func expectedWorkHours(dt *DayTiming) float64 {
	if dt.StartTime == nil || dt.EndTime == nil {
		return 0
	}
	ref := time.Now()
	start, err1 := parseDayTime(ref, *dt.StartTime)
	end, err2 := parseDayTime(ref, *dt.EndTime)
	if err1 != nil || err2 != nil {
		return 0
	}
	return end.Sub(start).Hours() - float64(dt.BreakMinutes)/60
}

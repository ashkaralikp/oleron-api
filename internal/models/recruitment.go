package models

import "time"

type Vacancy struct {
	ID               string     `json:"id"`
	BranchID         string     `json:"branch_id"`
	CreatedBy        string     `json:"created_by"`
	Title            string     `json:"title"`
	Department       *string    `json:"department"`
	Description      *string    `json:"description"`
	Requirements     *string    `json:"requirements"`
	Positions        int        `json:"positions"`
	Status           string     `json:"status"`
	Deadline         *time.Time `json:"deadline"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ApplicationCount int        `json:"application_count,omitempty"`
}

type Application struct {
	ID          string      `json:"id"`
	VacancyID   string      `json:"vacancy_id"`
	FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
	Email       string      `json:"email"`
	Phone       *string     `json:"phone"`
	CVUrl       *string     `json:"cv_url"`
	CoverLetter *string     `json:"cover_letter"`
	Status      string      `json:"status"`
	Notes       *string     `json:"notes"`
	AppliedAt   time.Time   `json:"applied_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Interviews  []Interview `json:"interviews,omitempty"`
}

type Interview struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"application_id"`
	InterviewerID string    `json:"interviewer_id"`
	ScheduledAt   time.Time `json:"scheduled_at"`
	Type          string    `json:"type"`
	Location      *string   `json:"location"`
	Outcome       string    `json:"outcome"`
	Feedback      *string   `json:"feedback"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

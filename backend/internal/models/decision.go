package models

import "time"

type Decision struct {
	ID          int       `json:"id"`
	CandidateID int       `json:"candidate_id"`
	Decision    string    `json:"decision"`
	Notes       *string   `json:"notes"`
	DecidedBy   int       `json:"decided_by"`
	DecidedAt   time.Time `json:"decided_at"`
}

type CreateDecisionRequest struct {
	Decision string  `json:"decision" binding:"required,oneof=shortlist reject waitlist review"`
	Notes    *string `json:"notes"`
}

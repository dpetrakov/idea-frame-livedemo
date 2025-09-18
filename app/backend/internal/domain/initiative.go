package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ErrInitiativeNotFound returned when initiative is not found
var ErrInitiativeNotFound = errors.New("initiative not found")

// Initiative represents a domain initiative entity
type Initiative struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	AuthorID     uuid.UUID  `json:"authorId"`
	Author       User       `json:"author"`
	AssigneeID   *uuid.UUID `json:"assigneeId"`
	Assignee     *User      `json:"assignee"`
	Value        *int       `json:"value"`
	Speed        *int       `json:"speed"`
	Cost         *int       `json:"cost"`
	Weight       float64    `json:"weight"`
	CommentsCount int       `json:"commentsCount"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// InitiativeCreate represents data for creating a new initiative
type InitiativeCreate struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

// InitiativeUpdate represents data for updating an initiative
type InitiativeUpdate struct {
	Value      *int       `json:"value"`
	Speed      *int       `json:"speed"`
	Cost       *int       `json:"cost"`
	AssigneeID *uuid.UUID `json:"assigneeId"`
}

// Validate validates initiative creation data
func (ic *InitiativeCreate) Validate() error {
	// Title validation
	if strings.TrimSpace(ic.Title) == "" {
		return errors.New("title is required")
	}
	
	if utf8.RuneCountInString(ic.Title) > 140 {
		return errors.New("title must not exceed 140 characters")
	}

	// Description validation (optional)
	if ic.Description != nil && utf8.RuneCountInString(*ic.Description) > 10000 {
		return errors.New("description must not exceed 10000 characters")
	}

	return nil
}

// Validate validates initiative update data
func (iu *InitiativeUpdate) Validate() error {
	// Value validation (1-5 or null)
	if iu.Value != nil && (*iu.Value < 1 || *iu.Value > 5) {
		return errors.New("value must be between 1 and 5")
	}

	// Speed validation (1-5 or null)
	if iu.Speed != nil && (*iu.Speed < 1 || *iu.Speed > 5) {
		return errors.New("speed must be between 1 and 5")
	}

	// Cost validation (1-5 or null)
	if iu.Cost != nil && (*iu.Cost < 1 || *iu.Cost > 5) {
		return errors.New("cost must be between 1 and 5")
	}

	return nil
}

// CalculateWeight calculates initiative weight based on value, speed, cost
// Formula: 0.5*value + 0.3*speed - 0.2*cost
// If all attributes are null, weight = 0.0
func (i *Initiative) CalculateWeight() float64 {
	if i.Value == nil && i.Speed == nil && i.Cost == nil {
		return 0.0
	}

	value := float64(0)
	speed := float64(0)
	cost := float64(0)

	if i.Value != nil {
		value = float64(*i.Value)
	}
	if i.Speed != nil {
		speed = float64(*i.Speed)
	}
	if i.Cost != nil {
		cost = float64(*i.Cost)
	}

	// Round to 2 decimal places
	weight := 0.5*value + 0.3*speed - 0.2*cost
	return float64(int(weight*100+0.5)) / 100
}

// IsOwner checks if user is the author of the initiative
func (i *Initiative) IsOwner(userID uuid.UUID) bool {
	return i.AuthorID == userID
}

// IsAssignee checks if user is assigned to the initiative
func (i *Initiative) IsAssignee(userID uuid.UUID) bool {
	return i.AssigneeID != nil && *i.AssigneeID == userID
}

// HasAccess checks if user has access to view/edit the initiative
// For this demo, all authorized users can view all initiatives
func (i *Initiative) HasAccess(userID uuid.UUID) bool {
	return true // Public access for demo
}
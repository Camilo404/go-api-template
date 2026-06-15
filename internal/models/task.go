// Package models defines the domain entities, DTOs, and shared types
// used across the service. Validation lives with the entity it validates
// so the rules are colocated with the shape of the data.
package models

import (
	"strings"
	"time"
)

// Task is the example domain entity. Copy this file to model your own
// resource and rename the type, then update the handler/service/store
// that references it.
type Task struct {
	ID          int       `json:"id" example:"1"`
	Title       string    `json:"title" example:"Buy milk"`
	Description string    `json:"description" example:"Whole milk, 2L"`
	Completed   bool      `json:"completed" example:"false"`
	CreatedAt   time.Time `json:"created_at" format:"date-time"`
	UpdatedAt   time.Time `json:"updated_at" format:"date-time"`
}

// CreateTaskInput is the payload accepted by POST endpoints.
type CreateTaskInput struct {
	Title       string `json:"title" example:"Buy milk"`
	Description string `json:"description" example:"Whole milk, 2L"`
}

// Validate normalises and validates the create payload.
func (in *CreateTaskInput) Validate() error {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	if in.Title == "" {
		return ErrTitleRequired
	}
	if len(in.Title) > 200 {
		return ErrTitleTooLong
	}
	if len(in.Description) > 2000 {
		return ErrDescriptionTooLong
	}
	return nil
}

// UpdateTaskInput uses pointers so callers can express "leave field
// unchanged" by omitting it from the JSON body.
type UpdateTaskInput struct {
	Title       *string `json:"title,omitempty" example:"Buy milk"`
	Description *string `json:"description,omitempty" example:"Whole milk, 2L"`
	Completed   *bool   `json:"completed,omitempty" example:"true"`
}

// Validate normalises and validates the update payload.
func (in *UpdateTaskInput) Validate() error {
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return ErrTitleRequired
		}
		if len(t) > 200 {
			return ErrTitleTooLong
		}
		in.Title = &t
	}
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		if len(d) > 2000 {
			return ErrDescriptionTooLong
		}
		in.Description = &d
	}
	return nil
}

package service

import (
	"context"
	"errors"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
)

type ActivityService struct {
	repo ActivityRepository
}

func NewActivityService(repo ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

type CreateActivityInput struct {
	Type      string    `json:"type"`
	Distance  *float64  `json:"distance,omitempty"`
	Duration  *int      `json:"duration,omitempty"`
	Elevation *float64  `json:"elevation,omitempty"`
	Calories  *int      `json:"calories,omitempty"`
	StartedAt time.Time `json:"started_at"`
	Source    string    `json:"source,omitempty"`
	SourceID  string    `json:"source_id"`
}

type ActivityResponse struct {
	ID        uint     `json:"id"`
	UserID    uint     `json:"user_id"`
	Type      string   `json:"type"`
	Distance  *float64 `json:"distance"`
	Duration  *int     `json:"duration"`
	Elevation *float64 `json:"elevation"`
	Calories  *int     `json:"calories"`
	StartedAt string   `json:"started_at"`
	Source    string   `json:"source"`
	SourceID  string   `json:"source_id"`
	CreatedAt string   `json:"created_at"`
}

func (s *ActivityService) Create(ctx context.Context, userID uint, input CreateActivityInput) (*ActivityResponse, bool, error) {
	if input.SourceID == "" {
		return nil, false, errors.New("source_id is required")
	}
	if input.StartedAt.IsZero() {
		return nil, false, errors.New("started_at is required")
	}

	activityType, err := model.ParseActivityType(input.Type)
	if err != nil {
		return nil, false, err
	}

	existing, err := s.repo.FindBySourceID(input.SourceID)
	if err == nil && existing != nil {
		return toActivityResponse(existing), false, nil
	}

	source := model.SourceManual
	if input.Source != "" {
		var err error
		source, err = model.ParseSource(input.Source)
		if err != nil {
			return nil, false, err
		}
	}

	activity := &model.Activity{
		UserID:    userID,
		Type:      activityType,
		Distance:  input.Distance,
		Duration:  input.Duration,
		Elevation: input.Elevation,
		Calories:  input.Calories,
		StartedAt: input.StartedAt,
		Source:    source,
		SourceID:  input.SourceID,
	}

	if err := s.repo.Create(activity); err != nil {
		return nil, false, err
	}

	return toActivityResponse(activity), true, nil
}

func (s *ActivityService) List(ctx context.Context, userID uint, from, to *time.Time, limit, offset int) ([]ActivityResponse, error) {
	activities, err := s.repo.FindByUserID(userID, from, to, limit, offset)
	if err != nil {
		return nil, err
	}

	resp := make([]ActivityResponse, len(activities))
	for i, a := range activities {
		resp[i] = *toActivityResponse(&a)
	}
	return resp, nil
}

func toActivityResponse(a *model.Activity) *ActivityResponse {
	return &ActivityResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Type:      string(a.Type),
		Distance:  a.Distance,
		Duration:  a.Duration,
		Elevation: a.Elevation,
		Calories:  a.Calories,
		StartedAt: a.StartedAt.Format(time.RFC3339),
		Source:    string(a.Source),
		SourceID:  a.SourceID,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}

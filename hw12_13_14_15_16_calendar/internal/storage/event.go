package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type DurationSeconds time.Duration

type Event struct {
	ID          uuid.UUID
	Title       string
	Start       time.Time
	End         time.Time
	Description string
	UserID      uuid.UUID
	TimeBefore  time.Duration
}

type EventDTO struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Start       time.Time       `json:"start"`
	End         time.Time       `json:"end"`
	Description string          `json:"description"`
	UserID      uuid.UUID       `json:"userId"`
	TimeBefore  DurationSeconds `json:"timeBefore"`
}

func ToDTO(e Event) EventDTO {
	return EventDTO{
		ID:          e.ID,
		Title:       e.Title,
		Start:       e.Start,
		End:         e.End,
		Description: e.Description,
		UserID:      e.UserID,
		TimeBefore:  DurationSeconds(e.TimeBefore),
	}
}

func FromDTO(dto EventDTO) Event {
	return Event{
		ID:          dto.ID,
		Title:       dto.Title,
		Start:       dto.Start,
		End:         dto.End,
		Description: dto.Description,
		UserID:      dto.UserID,
		TimeBefore:  time.Duration(dto.TimeBefore),
	}
}

type Interval struct {
	ID    uuid.UUID
	Start time.Time
	End   time.Time
}

func (e Event) CheckValid() error {
	if e.ID == uuid.Nil {
		return &ErrInvalidEvent{
			Field:   "id",
			Message: "ID события обязателен",
		}
	}
	if e.UserID == uuid.Nil {
		return &ErrInvalidEvent{
			Field:   "userId",
			Message: "ID пользователя обязателен",
		}
	}
	if e.Start.Before(time.Now()) {
		return &ErrInvalidEvent{
			Field:   "start",
			Message: "время начала не может быть в прошлом",
		}
	}
	if e.End.Before(e.Start) {
		return &ErrInvalidEvent{
			Field:   "end",
			Message: "время окончания события должно быть после времени начала",
		}
	}
	if e.TimeBefore < 0 {
		return &ErrInvalidEvent{
			Field:   "timeBefore",
			Message: "время уведомления должно быть положительным",
		}
	}
	return nil
}

func (e Event) GetInterval() Interval {
	return Interval{e.ID, e.Start, e.End}
}

func (d DurationSeconds) MarshalJSON() ([]byte, error) {
	seconds := int64(time.Duration(d).Seconds())
	return json.Marshal(seconds)
}

func (d *DurationSeconds) UnmarshalJSON(data []byte) error {
	var seconds int64
	if err := json.Unmarshal(data, &seconds); err != nil {
		return err
	}
	*d = DurationSeconds(time.Duration(seconds) * time.Second)
	return nil
}

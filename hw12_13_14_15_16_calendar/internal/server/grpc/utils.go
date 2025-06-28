package grpcserver

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/api"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	server "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
	storage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertToEventProto(e storage.Event) *pb.Event {
	return &pb.Event{
		Id:          e.ID.String(),
		UserId:      e.UserID.String(),
		Title:       e.Title,
		StartTime:   timestamppb.New(e.Start),
		EndTime:     timestamppb.New(e.End),
		Description: e.Description,
		TimeBefore:  int64(e.TimeBefore.Seconds()),
	}
}

func getEventFromBody[T interface{ GetEvent() *pb.Event }](
	ctx context.Context,
	log *slog.Logger,
	req T,
) (storage.Event, error) {
	ctx = logger.WithLogMethod(ctx, "getEventFromBody")
	log.DebugContext(ctx, "attempting to extract event from request body")
	eventPB := req.GetEvent()
	if eventPB == nil {
		return storage.Event{}, logger.WrapError(ctx, server.ErrMissingEvent)
	}
	id, err := uuid.Parse(eventPB.Id)
	if err != nil {
		return storage.Event{}, logger.WrapError(ctx, server.ErrInvalidEventID)
	}

	userID, err := uuid.Parse(eventPB.UserId)
	if err != nil {
		return storage.Event{}, logger.WrapError(ctx, server.ErrInvalidUserID)
	}

	return storage.Event{
		ID:          id,
		UserID:      userID,
		Title:       eventPB.Title,
		Start:       eventPB.StartTime.AsTime(),
		End:         eventPB.EndTime.AsTime(),
		Description: eventPB.Description,
		TimeBefore:  time.Duration(eventPB.TimeBefore * int64(time.Second)),
	}, nil
}

func getEventIDFromBody[T interface{ GetId() string }](
	ctx context.Context,
	log *slog.Logger,
	req T,
) (uuid.UUID, error) {
	ctx = logger.WithLogMethod(ctx, "getEventIDFromBody")
	log.DebugContext(ctx, "attempting to extract event ID from request parameters")
	id := req.GetId()
	if id == "" {
		return uuid.Nil, logger.WrapError(ctx, server.ErrMissingEventID)
	}
	uuID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, logger.WrapError(ctx, server.ErrInvalidEventID)
	}
	ctx = logger.WithLogEventID(ctx, uuID)
	log.DebugContext(ctx, "event ID successfully extracted from request parameters")
	return uuID, nil
}

func getStartTime[T interface{ GetStart() *timestamppb.Timestamp }](
	ctx context.Context,
	log *slog.Logger,
	req T,
) (time.Time, error) {
	ctx = logger.WithLogMethod(ctx, "getStartTime")
	log.DebugContext(ctx, "attempting to extract start time from request parameters")
	startTimestamp := req.GetStart()
	if startTimestamp == nil {
		return time.Time{}, logger.WrapError(ctx, server.ErrInvalidStartPeriod)
	}
	start := startTimestamp.AsTime()
	ctx = logger.WithLogStart(ctx, start)
	log.InfoContext(ctx, "start time successfully extracted from request parameters")
	return start, nil
}

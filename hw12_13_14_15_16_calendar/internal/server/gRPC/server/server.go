package grpcserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/grpc/pb"
)

type application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error)
}

type CalendarServer struct {
	logger *slog.Logger
	pb.UnimplementedCalendarServer
	app application
	lis net.Listener
}

func NewServerGRPC(logger *slog.Logger, lis net.Listener, app application) *CalendarServer {
	return &CalendarServer{
		logger: logger,
		lis:    lis,
		app:    app,
	}
}

func convertToEventModel(e *pb.Event, logger *slog.Logger) (storage.Event, error) {
	if e == nil {
		logger.Error("пустое событие")
		return storage.Event{}, errors.New("invalid event")
	}
	id, err := uuid.Parse(e.Id)
	if err != nil {
		logger.Error("неверный формат ID события", "id", e.Id, "error", err)
		return storage.Event{}, err
	}

	userID, err := uuid.Parse(e.UserId)
	if err != nil {
		logger.Error("неверный формат ID пользователя", "id", e.UserId, "error", err)
		return storage.Event{}, err
	}

	return storage.Event{
		ID:          id,
		UserID:      userID,
		Title:       e.Title,
		Start:       e.StartTime.AsTime(),
		End:         e.EndTime.AsTime(),
		Description: e.Description,
		TimeBefore:  time.Duration(e.TimeBefore),
	}, nil
}

func convertToEventProto(e storage.Event) *pb.Event {
	return &pb.Event{
		Id:          e.ID.String(),
		UserId:      e.UserID.String(),
		Title:       e.Title,
		StartTime:   timestamppb.New(e.Start),
		EndTime:     timestamppb.New(e.End),
		Description: e.Description,
		TimeBefore:  int64(e.TimeBefore),
	}
}

func (s *CalendarServer) CreateEvent(ctx context.Context, req *pb.CreateEventReq) (*emptypb.Empty, error) {
	event, err := convertToEventModel(req.GetEvent(), s.logger)
	if err != nil {
		return &emptypb.Empty{}, server.ErrInvalidEventData
	}
	s.logger.Debug("попытка создать событие", "method", "CreateEvent",
		"eventID", event.ID, "userID", event.UserID)
	err = s.app.CreateEvent(ctx, event)
	if err != nil {
		s.logger.Error("ошибка при создании события", "method", "CreateEvent",
			"eventID", event.ID, "userID", event.UserID, "error", err)
		return &emptypb.Empty{}, server.ErrCreateEvent
	}
	s.logger.Info("событие успешно создано", "method", "CreateEvent",
		"eventID", event.ID, "userID", event.UserID)
	return &emptypb.Empty{}, nil
}

func (s *CalendarServer) UpdateEvent(ctx context.Context, req *pb.UpdateEventReq) (*emptypb.Empty, error) {
	event, err := convertToEventModel(req.GetEvent(), s.logger)
	if err != nil {
		return &emptypb.Empty{}, server.ErrInvalidEventData
	}
	uuID, err := getEventIDFromBody(s.logger, req)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	event.ID = uuID
	s.logger.Debug("попытка обновить событие", "method", "UpdateEvent",
		"eventID", event.ID, "userID", event.UserID)
	err = s.app.UpdateEvent(ctx, event.ID, event)
	if err != nil {
		s.logger.Error("ошибка при обновлении события", "method", "UpdateEvent",
			"eventID", event.ID, "userID", event.UserID, "error", err)
		return &emptypb.Empty{}, server.ErrUpdateEvent
	}
	s.logger.Info("событие успешно обновлено", "method", "UpdateEvent",
		"eventID", event.ID, "userID", event.UserID)
	return &emptypb.Empty{}, nil
}

func (s *CalendarServer) DeleteEvent(ctx context.Context, req *pb.DeleteEventReq) (*emptypb.Empty, error) {
	id, err := getEventIDFromBody(s.logger, req)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	s.logger.Debug("попытка удаления события", "method", "DeleteEvent",
		"eventID", id)
	err = s.app.DeleteEvent(ctx, id)
	if err != nil {
		s.logger.Error("ошибка при удалении события", "method", "DeleteEvent",
			"eventID", id, "error", err)
		return &emptypb.Empty{}, server.ErrDeleteEvent
	}
	s.logger.Info("событие успешно удалено", "method", "DeleteEvent",
		"eventID", id)
	return &emptypb.Empty{}, nil
}

func (s *CalendarServer) GetEventsDay(ctx context.Context, req *pb.GetEventsDayReq) (*pb.GetEventsDayResp, error) {
	start, err := GetStartTime(s.logger, req)
	if err != nil {
		s.logger.Error("ошибка при получении времени начала", "method", "GetEventsDay",
			"error", err)
		return nil, err
	}
	events, err := s.app.GetEventsDay(ctx, start)
	if err != nil {
		s.logger.Error("ошибка при получении событий", "method", "GetEventsDay",
			"start", start.Format(time.RFC3339), "error", err)
		return nil, server.ErrEventRetrieval
	}
	resp := &pb.GetEventsDayResp{}
	for _, e := range events {
		resp.Events = append(resp.Events, convertToEventProto(e))
	}
	s.logger.Info("события успешно получены", "method", "GetEventsDay",
		"start", start.Format(time.RFC3339))
	return resp, nil
}

func (s *CalendarServer) GetEventsWeek(ctx context.Context, req *pb.GetEventsWeekReq) (*pb.GetEventsWeekResp, error) {
	start, err := GetStartTime(s.logger, req)
	if err != nil {
		s.logger.Error("ошибка при получении времени начала", "method", "GetEventsWeek",
			"error", err)
		return nil, err
	}
	events, err := s.app.GetEventsWeek(ctx, start)
	if err != nil {
		s.logger.Error("ошибка при получении событий", "method", "GetEventsWeek",
			"start", start.Format(time.RFC3339), "error", err)
		return nil, server.ErrEventRetrieval
	}
	resp := &pb.GetEventsWeekResp{}
	for _, e := range events {
		resp.Events = append(resp.Events, convertToEventProto(e))
	}
	s.logger.Info("события успешно получены", "method", "GetEventsWeek",
		"start", start.Format(time.RFC3339))
	return resp, nil
}

func (s *CalendarServer) GetEventsMonth(ctx context.Context, req *pb.GetEventsMonthReq) (*pb.GetEventsMonthResp, error) {
	start, err := GetStartTime(s.logger, req)
	if err != nil {
		s.logger.Error("ошибка при получении времени начала", "method", "GetEventsMonth",
			"error", err)
		return nil, err
	}
	events, err := s.app.GetEventsMonth(ctx, start)
	if err != nil {
		s.logger.Error("ошибка при получении событий", "method", "GetEventsMonth",
			"start", start.Format(time.RFC3339), "error", err)
		return nil, server.ErrEventRetrieval
	}
	resp := &pb.GetEventsMonthResp{}
	for _, e := range events {
		resp.Events = append(resp.Events, convertToEventProto(e))
	}
	s.logger.Info("события успешно получены", "method", "GetEventsMonth",
		"start", start.Format(time.RFC3339))
	return resp, nil
}

func getEventIDFromBody[T interface{ GetId() string }](logger *slog.Logger, req T) (uuid.UUID, error) {
	logger.Debug("попытка извлечь ID события из параметров запроса", "method", "getEventIDFromBody")
	id := req.GetId()
	if id == "" {
		logger.Error(server.ErrMissingEventID.Error(), "id", nil)
		return uuid.Nil, server.ErrMissingEventID
	}
	uuID, err := uuid.Parse(id)
	if err != nil {
		logger.Error(server.ErrInvalidEventID.Error(), "id", id, "error", err)
		return uuid.Nil, server.ErrInvalidEventID
	}
	logger.Debug("успешно извлечён ID из параметров запроса", "eventID", uuID.String())
	return uuID, nil
}

func GetStartTime(logger *slog.Logger, req interface{ GetStart() *timestamppb.Timestamp }) (time.Time, error) {
	startTimestamp := req.GetStart()
	if startTimestamp == nil {
		return time.Time{}, server.ErrInvalidStartPeriod
	}
	start := startTimestamp.AsTime()
	return start, nil
}

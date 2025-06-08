package grpcserver

import (
	"context"
	"log/slog"
	"net"
	"time"

	pb "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/api"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	server "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
	storage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *CalendarServer) CreateEvent(ctx context.Context, req *pb.CreateEventReq) (*emptypb.Empty, error) {
	ctx = logger.WithLogMethod(ctx, "CreateEvent")
	event, err := getEventFromBody(ctx, s.logger, req)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return &emptypb.Empty{}, server.ErrInvalidEventData
	}
	ctx = logger.WithLogEventID(ctx, event.ID)

	s.logger.DebugContext(ctx, "попытка создать событие")
	err = s.app.CreateEvent(ctx, event)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return &emptypb.Empty{}, err
	}

	s.logger.InfoContext(ctx, "событие успешно создано")
	return &emptypb.Empty{}, nil
}

func (s *CalendarServer) UpdateEvent(ctx context.Context, req *pb.UpdateEventReq) (*emptypb.Empty, error) {
	ctx = logger.WithLogMethod(ctx, "UpdateEvent")
	event, err := getEventFromBody(ctx, s.logger, req)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return &emptypb.Empty{}, server.ErrInvalidEventData
	}
	uuID, err := getEventIDFromBody(ctx, s.logger, req)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return &emptypb.Empty{}, err
	}
	ctx = logger.WithLogEventID(ctx, uuID)
	event.ID = uuID
	s.logger.DebugContext(ctx, "попытка обновить событие")
	err = s.app.UpdateEvent(ctx, event.ID, event)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return &emptypb.Empty{}, err
	}
	s.logger.InfoContext(ctx, "событие успешно обновлено")
	return &emptypb.Empty{}, nil
}

func (s *CalendarServer) DeleteEvent(ctx context.Context, req *pb.DeleteEventReq) (*emptypb.Empty, error) {
	ctx = logger.WithLogMethod(ctx, "DeleteEvent")
	id, err := getEventIDFromBody(ctx, s.logger, req)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return &emptypb.Empty{}, err
	}
	ctx = logger.WithLogEventID(ctx, id)
	s.logger.DebugContext(ctx, "попытка удаления события")

	err = s.app.DeleteEvent(ctx, id)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return &emptypb.Empty{}, err
	}
	s.logger.InfoContext(ctx, "событие успешно удалено")
	return &emptypb.Empty{}, nil
}

func (s *CalendarServer) GetEventsDay(ctx context.Context, req *pb.GetEventsReq) (*pb.GetEventsResp, error) {
	return s.getEvents(ctx, "GetEventsDay", req, s.app.GetEventsDay)
}

func (s *CalendarServer) GetEventsWeek(ctx context.Context, req *pb.GetEventsReq) (*pb.GetEventsResp, error) {
	return s.getEvents(ctx, "GetEventsWeek", req, s.app.GetEventsWeek)
}

func (s *CalendarServer) GetEventsMonth(ctx context.Context, req *pb.GetEventsReq) (*pb.GetEventsResp, error) {
	return s.getEvents(ctx, "GetEventsMonth", req, s.app.GetEventsMonth)
}

func (s *CalendarServer) getEvents(
	ctx context.Context,
	methodName string,
	req *pb.GetEventsReq,
	getFunc func(context.Context, time.Time) ([]storage.Event, error),
) (*pb.GetEventsResp, error) {
	ctx = logger.WithLogMethod(ctx, methodName)

	start, err := getStartTime(ctx, s.logger, req)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return nil, err
	}
	ctx = logger.WithLogStart(ctx, start)

	events, err := getFunc(ctx, start)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return nil, server.ErrEventRetrieval
	}

	resp := &pb.GetEventsResp{}
	for _, e := range events {
		resp.Events = append(resp.Events, convertToEventProto(e))
	}
	s.logger.InfoContext(ctx, "события успешно получены")
	return resp, nil
}

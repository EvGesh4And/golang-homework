package grpcserver

import (
	"context"
	"net"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/gRPC/pb"
)

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error)
}

type CalendarServer struct {
	pb.UnimplementedCalendarServer
	App Application
	lis net.Listener
}

func NewServerGRPC(lis net.Listener, app Application) *CalendarServer {
	return &CalendarServer{
		lis: lis,
		App: app,
	}
}

func convertToEventModel(e *pb.Event) storage.Event {
	return storage.Event{
		ID:          uuid.MustParse(e.Id),
		UserID:      uuid.MustParse(e.UserId),
		Title:       e.Title,
		Start:       e.StartTime.AsTime(),
		End:         e.EndTime.AsTime(),
		Description: e.Description,
		TimeBefore:  time.Duration(e.TimeBefore),
	}
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

func (s *CalendarServer) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*emptypb.Empty, error) {
	event := convertToEventModel(req.Event)
	event.ID = uuid.MustParse(req.Id)
	return &emptypb.Empty{}, s.App.CreateEvent(ctx, event)
}

func (s *CalendarServer) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*emptypb.Empty, error) {
	event := convertToEventModel(req.Event)
	event.ID = uuid.MustParse(req.Id)
	return &emptypb.Empty{}, s.App.UpdateEvent(ctx, event.ID, event)
}

func (s *CalendarServer) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, s.App.DeleteEvent(ctx, id)
}

func (s *CalendarServer) GetEventsDay(ctx context.Context, req *pb.GetEventsDayRequest) (*pb.GetEventsDayResponse, error) {
	start := req.Start.AsTime()
	events, err := s.App.GetEventsDay(ctx, start)
	if err != nil {
		return nil, err
	}
	resp := &pb.GetEventsDayResponse{}
	for _, e := range events {
		resp.Events = append(resp.Events, convertToEventProto(e))
	}
	return resp, nil
}

func (s *CalendarServer) GetEventsWeek(ctx context.Context, req *pb.GetEventsWeekRequest) (*pb.GetEventsWeekResponse, error) {
	start := req.Start.AsTime()
	events, err := s.App.GetEventsWeek(ctx, start)
	if err != nil {
		return nil, err
	}
	resp := &pb.GetEventsWeekResponse{}
	for _, e := range events {
		resp.Events = append(resp.Events, convertToEventProto(e))
	}
	return resp, nil
}

func (s *CalendarServer) GetEventsMonth(ctx context.Context, req *pb.GetEventsMonthRequest) (*pb.GetEventsMonthResponse, error) {
	start := req.Start.AsTime()
	events, err := s.App.GetEventsMonth(ctx, start)
	if err != nil {
		return nil, err
	}
	resp := &pb.GetEventsMonthResponse{}
	for _, e := range events {
		resp.Events = append(resp.Events, convertToEventProto(e))
	}
	return resp, nil
}

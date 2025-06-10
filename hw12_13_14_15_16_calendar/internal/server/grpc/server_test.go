package grpcserver

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	pb "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/api"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	server "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
	storage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockApp struct {
	CreateEventFn    func(ctx context.Context, event storage.Event) error
	UpdateEventFn    func(ctx context.Context, id uuid.UUID, event storage.Event) error
	DeleteEventFn    func(ctx context.Context, id uuid.UUID) error
	GetEventsDayFn   func(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsWeekFn  func(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsMonthFn func(ctx context.Context, start time.Time) ([]storage.Event, error)
}

func (m *mockApp) CreateEvent(ctx context.Context, event storage.Event) error {
	return m.CreateEventFn(ctx, event)
}

func (m *mockApp) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error {
	return m.UpdateEventFn(ctx, id, event)
}

func (m *mockApp) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return m.DeleteEventFn(ctx, id)
}

func (m *mockApp) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return m.GetEventsDayFn(ctx, start)
}

func (m *mockApp) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return m.GetEventsWeekFn(ctx, start)
}

func (m *mockApp) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return m.GetEventsMonthFn(ctx, start)
}

func newTestServer(t *testing.T, app server.Application) (pb.CalendarClient, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	s := grpc.NewServer()
	log := logger.New("info", os.Stdout)
	pb.RegisterCalendarServer(s, NewServerGRPC(log, lis, app))

	// Канал для отслеживания ошибок сервера
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Serve(lis)
	}()

	// Устанавливаем соединение с сервером
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//nolint:staticcheck // SA1019: grpc.DialContext и grpc.WithBlock пока допустимы
	conn, err := grpc.DialContext(
		ctx,
		lis.Addr().String(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		s.Stop()
		t.Fatalf("Failed to create gRPC client: %v", err)
	}

	calendarClient := pb.NewCalendarClient(conn)

	cleanup := func() {
		s.Stop()
		conn.Close()
		if err := <-serverErr; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("server error: %v", err)
		}
	}

	return calendarClient, cleanup
}

func TestCreateEvent(t *testing.T) {
	called := false

	event := &pb.Event{
		Id:          uuid.New().String(),
		UserId:      uuid.New().String(),
		Title:       "Test",
		StartTime:   timestamppb.Now(),
		EndTime:     timestamppb.New(time.Now().Add(time.Hour)),
		Description: "desc",
		TimeBefore:  int64(time.Minute),
	}

	client, shutdown := newTestServer(t, &mockApp{
		CreateEventFn: func(ctx context.Context, e storage.Event) error {
			_ = ctx
			called = true
			assert.Equal(t, event.Title, e.Title)
			return nil
		},
	})

	defer shutdown()

	_, err := client.CreateEvent(context.Background(), &pb.CreateEventReq{Event: event})

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestUpdateEvent(t *testing.T) {
	called := false

	id := uuid.New()

	client, shutdown := newTestServer(t, &mockApp{
		UpdateEventFn: func(ctx context.Context, uid uuid.UUID, e storage.Event) error {
			_ = ctx
			_ = uid
			_ = e
			called = true
			assert.Equal(t, id, uid)
			return nil
		},
	})

	defer shutdown()

	_, err := client.UpdateEvent(context.Background(), &pb.UpdateEventReq{
		Id: id.String(),

		Event: &pb.Event{
			Id:          id.String(),
			UserId:      uuid.New().String(),
			Title:       "Update",
			StartTime:   timestamppb.Now(),
			EndTime:     timestamppb.New(time.Now().Add(time.Hour)),
			TimeBefore:  int64(time.Minute),
			Description: "update desc",
		},
	})

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestDeleteEvent(t *testing.T) {
	id := uuid.New()

	called := false

	client, shutdown := newTestServer(t, &mockApp{
		DeleteEventFn: func(ctx context.Context, uid uuid.UUID) error {
			_ = ctx
			called = true
			assert.Equal(t, id, uid)
			return nil
		},
	})

	defer shutdown()

	_, err := client.DeleteEvent(context.Background(), &pb.DeleteEventReq{Id: id.String()})

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestGetEventsDay(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	client, shutdown := newTestServer(t, &mockApp{
		GetEventsDayFn: func(ctx context.Context, start time.Time) ([]storage.Event, error) {
			_ = ctx
			assert.WithinDuration(t, now, start, time.Second)
			return []storage.Event{
				{ID: uuid.New(), Title: "Day Event", Start: now, End: now.Add(time.Hour), UserID: uuid.New()},
			}, nil
		},
	})

	defer shutdown()

	resp, err := client.GetEventsDay(context.Background(), &pb.GetEventsReq{
		Start: timestamppb.New(now),
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Events, 1)
	assert.Equal(t, "Day Event", resp.Events[0].Title)
}

func TestGetEventsWeek(t *testing.T) {
	now := time.Now()

	client, shutdown := newTestServer(t, &mockApp{
		GetEventsWeekFn: func(ctx context.Context, start time.Time) ([]storage.Event, error) {
			_ = ctx
			_ = start
			return []storage.Event{
				{ID: uuid.New(), Title: "Week Event", Start: now, End: now.Add(time.Hour), UserID: uuid.New()},
			}, nil
		},
	})

	defer shutdown()

	resp, err := client.GetEventsWeek(context.Background(), &pb.GetEventsReq{
		Start: timestamppb.New(now),
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Events, 1)
}

func TestGetEventsMonth(t *testing.T) {
	now := time.Now()

	client, shutdown := newTestServer(t, &mockApp{
		GetEventsMonthFn: func(ctx context.Context, start time.Time) ([]storage.Event, error) {
			_ = ctx
			_ = start
			return []storage.Event{
				{ID: uuid.New(), Title: "Month Event", Start: now, End: now.Add(time.Hour), UserID: uuid.New()},
			}, nil
		},
	})
	defer shutdown()
	resp, err := client.GetEventsMonth(context.Background(), &pb.GetEventsReq{
		Start: timestamppb.New(now),
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Events, 1)
}

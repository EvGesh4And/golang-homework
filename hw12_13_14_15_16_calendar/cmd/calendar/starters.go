package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	pb "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/api"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/app"
	grpcserver "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/grpc"
	internalhttp "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/http"
	memorystorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/sql"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func setupStorage(ctx context.Context, cfg Config, lg *slog.Logger) (app.Storage, io.Closer, error) {
	switch cfg.Storage.Mod {
	case "memory":
		log.Print("using in-memory storage")
		return memorystorage.New(lg), nil, nil

	case "sql":
		log.Print("initializing connection to PostgreSQL...")

		sqlStorage := sqlstorage.New(lg, cfg.Storage.DSN)
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := sqlStorage.Connect(ctx); err != nil {
			log.Printf("error connecting to PostgreSQL: %v", err)
			return nil, nil, err
		}

		log.Print("executing migrations...")
		if err := sqlStorage.Migrate(ctx, cfg.Storage.Migration); err != nil {
			log.Print(err)
			return nil, nil, err
		}

		log.Print("SQL storage successfully initialized and connected")
		return sqlStorage, sqlStorage, nil

	default:
		log.Printf("unknown storage type: %v", cfg.Storage.Mod)
		return nil, nil, fmt.Errorf("unknown storage type: %v", cfg.Storage.Mod)
	}
}

func startHTTPServer(
	ctx context.Context,
	g *errgroup.Group,
	cfg Config,
	lg *slog.Logger,
	calendar *app.App,
) {
	serverHTTP := internalhttp.NewServerHTTP(cfg.HTTP.Host, cfg.HTTP.Port, lg, calendar)
	log.Print("HTTP server created")

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	g.Go(func() error {
		log.Printf("HTTP server starting %s...", addr)
		if err := serverHTTP.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP start: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		log.Print("[shutdown] stopping HTTP server...")

		if err := serverHTTP.Stop(shutdownCtx); err != nil {
			log.Printf("[shutdown] error stopping HTTP server: %s", err)
		} else {
			log.Print("[shutdown] HTTP server stopped gracefully...")
		}
		return ctx.Err()
	})
}

func startGRPCServer(
	ctx context.Context,
	g *errgroup.Group,
	cfg Config,
	lg *slog.Logger,
	calendar *app.App,
) {
	addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		g.Go(func() error {
			return fmt.Errorf("gRPC failed to listen on port %s: %w", addr, err)
		})
		return
	}

	serverGRPC := grpcserver.NewServerGRPC(lg, lis, calendar)
	grpcSrv := grpc.NewServer()
	pb.RegisterCalendarServer(grpcSrv, serverGRPC)

	g.Go(func() error {
		log.Printf("gRPC server starting %s...", lis.Addr().String())
		if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			grpcSrv.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Print("[shutdown] gRPC server stopped gracefully...")
		case <-shutdownCtx.Done():
			log.Print("[shutdown] gRPC graceful shutdown timeout, calling Stop()")
			grpcSrv.Stop()
		}
		return ctx.Err()
	})
}

package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Shuba-Buba/usdt-rate-service/internal/config"
	grpcServer "github.com/Shuba-Buba/usdt-rate-service/internal/grpc"
	"github.com/Shuba-Buba/usdt-rate-service/internal/repository"
	"github.com/Shuba-Buba/usdt-rate-service/internal/service"
	"github.com/Shuba-Buba/usdt-rate-service/pkg/grinex"
	"github.com/Shuba-Buba/usdt-rate-service/pkg/logger"
)

func main() {
	cfg := config.New()

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := log.Sync(); err != nil {
			log.Fatal("Failed to sync logger", zap.Error(err))
		}
	}()

	log.Info("Starting USDT rate service",
		zap.String("grpc_port", cfg.GRPCPort),
		zap.String("grinex_url", cfg.GrinexURL),
	)

	db, err := sqlx.Connect("postgres", cfg.PostgresURL)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal("Failed to close connection", zap.Error(err))
		}
	}()

	log.Info("Connected to database", zap.String("url", cfg.PostgresURL))

	rateRepo := repository.NewRateRepository(db)

	grinexClient := grinex.NewClient(cfg.GrinexURL, log)

	rateService := service.NewRateService(rateRepo, grinexClient, log)

	grpcSrv := grpc.NewServer()
	grpcServer.Register(grpcSrv, rateService, log)

	reflection.Register(grpcSrv)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	go func() {
		log.Info("Starting gRPC server", zap.String("address", lis.Addr().String()))
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down gracefully...")

	grpcSrv.GracefulStop()

	if err := db.Close(); err != nil {
		log.Error("Error closing database connection", zap.Error(err))
	}

	log.Info("Service stopped")
}

package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/Shuba-Buba/usdt-rate-service/internal/domain"
	pb "github.com/Shuba-Buba/usdt-rate-service/proto/rate/v1"
)

type RateService interface {
	GetRates(ctx context.Context) (*domain.Rate, error)
	HealthCheck(ctx context.Context) error
}

type Server struct {
	pb.UnimplementedRateServiceServer
	service RateService
	logger  *zap.Logger
}

func NewServer(service RateService, logger *zap.Logger) *Server {
	return &Server{
		service: service,
		logger:  logger,
	}
}

func (s *Server) GetRates(ctx context.Context, req *pb.GetRatesRequest) (*pb.GetRatesResponse, error) {
	s.logger.Info("GetRates request received")

	rate, err := s.service.GetRates(ctx)
	if err != nil {
		s.logger.Error("Failed to get rates", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get rates: %v", err)
	}

	asks := make([]*pb.OrderBookEntry, len(rate.Asks))
	bids := make([]*pb.OrderBookEntry, len(rate.Bids))

	for i, ask := range rate.Asks {
		asks[i] = &pb.OrderBookEntry{
			Price:  ask.Price,
			Volume: ask.Volume,
			Amount: ask.Amount,
			Factor: ask.Factor,
			Type:   ask.Type,
		}
	}

	for i, bid := range rate.Bids {
		bids[i] = &pb.OrderBookEntry{
			Price:  bid.Price,
			Volume: bid.Volume,
			Amount: bid.Amount,
			Factor: bid.Factor,
			Type:   bid.Type,
		}
	}

	return &pb.GetRatesResponse{
		Asks:      asks,
		Bids:      bids,
		Timestamp: rate.Timestamp.Unix(),
	}, nil
}

func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	s.logger.Info("HealthCheck request received")

	if err := s.service.HealthCheck(ctx); err != nil {
		s.logger.Error("Health check failed", zap.Error(err))
		return &pb.HealthCheckResponse{Status: "NOT_SERVING"}, nil
	}

	return &pb.HealthCheckResponse{Status: "SERVING"}, nil
}

type Health struct {
	service RateService
}

func NewHealth(service RateService) *Health {
	return &Health{service: service}
}

func (h *Health) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if err := h.service.HealthCheck(ctx); err != nil {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *Health) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "health check watching is not implemented")
}

func (h *Health) List(ctx context.Context, req *grpc_health_v1.HealthListRequest) (*grpc_health_v1.HealthListResponse, error) {
	response := &grpc_health_v1.HealthListResponse{
		Statuses: make(map[string]*grpc_health_v1.HealthCheckResponse),
	}

	response.Statuses[""] = &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}

	response.Statuses["rate.v1.RateService"] = &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}

	return response, nil
}

func Register(grpcServer *grpc.Server, service RateService, logger *zap.Logger) {
	server := NewServer(service, logger)
	health := NewHealth(service)

	pb.RegisterRateServiceServer(grpcServer, server)
	grpc_health_v1.RegisterHealthServer(grpcServer, health)
}

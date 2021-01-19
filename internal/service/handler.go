package service

import (
	"context"
	"errors"

	pb "github.com/srcabl/protos/users"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler implments the sources service
type Handler struct {
	pb.UnimplementedUserServiceServer
}

// New creates the service handler
func New() (*Handler, error) {
	return &Handler{}, nil
}

// HealthCheck is the base healthcheck for the service
func (h *Handler) HealthCheck(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {

	return nil, nil
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

	return nil, errors.New("not implemented")
}

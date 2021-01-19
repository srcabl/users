package bootstrap

import (
	pb "github.com/srcabl/protos/users"
	"github.com/srcabl/users/internal/config"
	"github.com/srcabl/users/internal/service"
)

type Bootstrap struct {
	config  *config.Environment
	service pb.UserServiceServer
}

func New(cfg *config.Environment) (*Bootstrap, error) {

	srvc, err := service.New()
	if err != nil {
		return nil, err
	}

	return &Bootstrap{
		config:  cfg,
		service: srvc,
	}, nil
}

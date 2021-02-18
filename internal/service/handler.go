package service

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	pb "github.com/srcabl/protos/users"
	"github.com/srcabl/services/pkg/db/mysql"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler implments the users service
type Handler struct {
	pb.UnimplementedUsersServiceServer
	datarepo DataRepository
}

// New creates the service handler
func New(db *mysql.Client) (*Handler, error) {
	dataRepo, err := NewDataRepository(db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create data repo")
	}
	return &Handler{
		datarepo: dataRepo,
	}, nil
}

// HealthCheck is the base healthcheck for the service
func (h *Handler) HealthCheck(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {

	return nil, nil
}

// GetUser handles the login of users
func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	id, err := uuid.FromBytes(req.Uuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "uuid is not well formed").Error())
	}
	user, err := h.datarepo.GetUserByID(ctx, id.String())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "uuid is not well formed").Error())
	}
	pbUser, err := user.ToGRPC()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "uuid is not well formed").Error())
	}
	return &pb.GetUserResponse{User: pbUser}, nil
}

// Follow handles the adding of followers
func (h *Handler) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowResponse, error) {
	res, err := performFollow(ctx, req, h.datarepo.AddUserFollower, h.datarepo.AddSourceFollower)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "something happened when trying to follow").Error())
	}
	return res, nil
}

// UnFollow handles the removing of followers
func (h *Handler) UnFollow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowResponse, error) {
	res, err := performFollow(ctx, req, h.datarepo.RemoveUserFollower, h.datarepo.RemoveSourceFollower)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "something happened when trying to unfollow").Error())
	}
	return res, nil
}

type drFollowFunc func(context.Context, string, string) error

func performFollow(ctx context.Context, req *pb.FollowRequest, userFollowFunc, sourceFollowFunc drFollowFunc) (*pb.FollowResponse, error) {
	followerUUID, err := uuid.FromBytes(req.FollowerUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "uuid of follower is invalid").Error())
	}
	followedUUID, err := uuid.FromBytes(req.FollowedUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "uuid of followed is invalid").Error())
	}
	var followFunc drFollowFunc
	if req.Type == pb.FollowRequest_SOURCE {
		followFunc = sourceFollowFunc
	}
	if req.Type == pb.FollowRequest_USER {
		followFunc = userFollowFunc
	}
	if followFunc == nil {
		return nil, status.Error(codes.InvalidArgument, "follow type is not valid")
	}
	if err := followFunc(ctx, followerUUID.String(), followedUUID.String()); err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "Something happened").Error())
	}
	return &pb.FollowResponse{}, nil

}

// ValidateUserCredentials handles the login of users
func (h *Handler) ValidateUserCredentials(ctx context.Context, req *pb.ValidateUserCredentialsRequest) (*pb.ValidateUserCredentialsResponse, error) {
	var dbUser *DBUser
	if req.ValidateUserBy == pb.ValidateUserCredentialsRequest_EMAIL {
		emailUser, err := h.datarepo.GetUserByEmail(ctx, req.Email)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to get user by email").Error())
		}
		dbUser = emailUser
	}
	if req.ValidateUserBy == pb.ValidateUserCredentialsRequest_USERNAME {
		usernameUser, err := h.datarepo.GetUserByUsername(ctx, req.Username)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to get user by username").Error())
		}
		dbUser = usernameUser
	}
	if dbUser == nil {
		return nil, status.Error(codes.InvalidArgument, errors.New("user can only be validated by email or username").Error())
	}
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.HashedPassword), []byte(req.Password)); err != nil {
		return &pb.ValidateUserCredentialsResponse{
			User:    nil,
			IsValid: false,
		}, nil
	}

	fmt.Printf("\n\nUser Getting GRPC: %+v", dbUser)
	pbUser, err := dbUser.ToGRPC()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "uuid is not well formed").Error())
	}

	fmt.Printf("\n\npb validated: %+v", pbUser)
	return &pb.ValidateUserCredentialsResponse{User: pbUser, IsValid: true}, nil
}

// CreateUser handles the creation of users
func (h *Handler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	dbUser, err := HydrateModelForCreate(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to hydrate user for create").Error())
	}
	fmt.Printf("incoming User: %+v", dbUser)
	if isValid := h.datarepo.ValidateUserForCreate(ctx, dbUser); !isValid {
		fmt.Println("not valid")
		return nil, status.Error(codes.InvalidArgument, errors.New("failed to validate user for create").Error())
	}
	fmt.Println("isvalid")
	if err := h.datarepo.CreateUser(ctx, dbUser); err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to create user").Error())
	}
	hydratedPBUser, err := dbUser.ToGRPC()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "something ridiculos").Error())
	}
	return &pb.CreateUserResponse{
		User: hydratedPBUser,
	}, nil
}

// UpdateUser handles the updating of users
func (h *Handler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {

	return nil, errors.New("not implemented")
}

// DeleteUser handles the deletion of users
func (h *Handler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {

	return nil, errors.New("not implemented")
}

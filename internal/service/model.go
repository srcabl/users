package service

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	sharedpb "github.com/srcabl/protos/shared"
	userspb "github.com/srcabl/protos/users"
	"github.com/srcabl/services/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DBUser is the database user model
type DBUser struct {
	UUID           string
	Username       string
	Email          string
	HashedPassword string
	CreatedByUUID  string
	CreatedAt      int64
	UpdatedByUUID  sql.NullString
	UpdatedAt      sql.NullInt64
}

// CreatedByUUIDString satisfies the services helper to transform db auditfields to grpc auditfields
func (u *DBUser) CreatedByUUIDString() string {
	return u.CreatedByUUID
}

// CreatedAtUnixInt satisfies the services helper to transform db auditfields to grpc auditfields
func (u *DBUser) CreatedAtUnixInt() int64 {
	return u.CreatedAt
}

// UpdatedByUUIDNullString satisfies the services helper to transform db auditfields to grpc auditfields
func (u *DBUser) UpdatedByUUIDNullString() sql.NullString {
	return u.UpdatedByUUID
}

// UpdatedAtUnixNullInt satisfies the services helper to transform db auditfields to grpc auditfields
func (u *DBUser) UpdatedAtUnixNullInt() sql.NullInt64 {
	return u.UpdatedAt
}

// ToGRPC transforms the dbuser to proto user
func (u *DBUser) ToGRPC() (*sharedpb.User, error) {
	id, err := uuid.FromString(u.UUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to transform uuid: %s", u.UUID)
	}
	auditFields, err := proto.DBAuditFieldsToGRPC(u)
	if err != nil {
		return nil, errors.Wrap(err, "failed to transform auditfields")
	}
	return &sharedpb.User{
		Uuid:            id.Bytes(),
		Username:        u.Username,
		Email:           u.Email,
		HashedPasssword: u.HashedPassword,
		AuditFields:     auditFields,
	}, nil
}

// HydrateModelForCreate creates a db user from a proto user and fills in any missing data
func HydrateModelForCreate(req *userspb.CreateUserRequest) (*DBUser, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to generate uuid for user").Error())
	}
	now := time.Now().Unix()
	return &DBUser{
		UUID:           newUUID.String(),
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: req.HashedPasssword,
		CreatedByUUID:  newUUID.String(),
		CreatedAt:      now,
		UpdatedByUUID:  sql.NullString{Valid: true, String: newUUID.String()},
		UpdatedAt:      sql.NullInt64{Valid: true, Int64: now},
		//TODO display and description
	}, nil
}

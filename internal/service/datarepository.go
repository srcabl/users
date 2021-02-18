package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/srcabl/services/pkg/db/mysql"
)

// DataRepository specifies behavior of the data repo
type DataRepository interface {
	DataRepositoryGetter
	DataRepositoryCreator
	DataRepositoryUpdater
}

// DataRepositoryGetter specifies behavior of the data repo getters
type DataRepositoryGetter interface {
	GetUserByID(context.Context, string) (*DBUser, error)
	GetUserByUsername(context.Context, string) (*DBUser, error)
	GetUserByEmail(context.Context, string) (*DBUser, error)
}

// DataRepositoryCreator specifies the behavior of the data repo creators
type DataRepositoryCreator interface {
	ValidateUserForCreate(context.Context, *DBUser) bool
	CreateUser(context.Context, *DBUser) error
}

// DataRepositoryUpdater specifies the behavior of the data repo updaters
type DataRepositoryUpdater interface {
	AddUserFollower(context.Context, string, string) error
	RemoveUserFollower(context.Context, string, string) error
	AddSourceFollower(context.Context, string, string) error
	RemoveSourceFollower(context.Context, string, string) error
}

type dataRepository struct {
	db *mysql.Client
}

// NewDataRepository news up a data repo
func NewDataRepository(db *mysql.Client) (DataRepository, error) {
	return &dataRepository{
		db: db,
	}, nil
}

const getUserByQuery = `
SELECT
	uuid,
	username,
	email,
	hashed_password,
	created_by_uuid,
	created_at,
	updated_by_uuid,
	updated_at
FROM
	users

`

// GetUserByID gets user by the id
func (dr *dataRepository) GetUserByID(ctx context.Context, uuid string) (*DBUser, error) {
	getQuery := getUserByQuery + `WHERE uuid=?`
	user, err := dr.getUser(ctx, getQuery, uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find user with ID %s", uuid)
	}
	return user, nil
}

// GetUserByUsername gets user by the username
func (dr *dataRepository) GetUserByUsername(ctx context.Context, username string) (*DBUser, error) {
	getQuery := getUserByQuery + `WHERE username=?`
	user, err := dr.getUser(ctx, getQuery, username)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find user with username %s", username)
	}
	return user, nil
}

// GetUserByEmail gets user by the email
func (dr *dataRepository) GetUserByEmail(ctx context.Context, email string) (*DBUser, error) {
	getQuery := getUserByQuery + `WHERE email=?`
	user, err := dr.getUser(ctx, getQuery, email)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find user with email %s", email)
	}
	return user, nil
}

func (dr *dataRepository) getUser(ctx context.Context, query string, param string) (*DBUser, error) {
	user := &DBUser{}
	err := dr.db.DB.QueryRow(query, param).Scan(
		&user.UUID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,
		&user.CreatedByUUID,
		&user.CreatedAt,
		&user.UpdatedByUUID,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find user with param %s", param)
	}
	return user, nil
}

const addUserFollowerStatement = `
INSERT INTO
	user_user_follows (
		follower,
		followed
	)
VALUES
	(?, ?)
`

// AddUserFollower adds a user follow relationship
func (dr *dataRepository) AddUserFollower(ctx context.Context, follower, followed string) error {
	if err := dr.performFollowStatement(ctx, addUserFollowerStatement, follower, followed); err != nil {
		return errors.Wrap(err, "failed to add user follower")
	}
	return nil
}

const removeUserFollowerStatement = `
DELETE FROM
	user_user_follows
WHERE
	follower=? AND followed=?
`

// RevomeUserFollower adds a user follow relationship
func (dr *dataRepository) RemoveUserFollower(ctx context.Context, follower, followed string) error {
	if err := dr.performFollowStatement(ctx, removeUserFollowerStatement, follower, followed); err != nil {
		return errors.Wrap(err, "failed to remove user follower")
	}
	return nil
}

const addSourceFollowerStatement = `
INSERT INTO
	user_source_follows (
		follower,
		followed
	)
VALUES
	(?, ?)
`

// AddSourceFollower adds a user follow relationship
func (dr *dataRepository) AddSourceFollower(ctx context.Context, follower, followed string) error {
	if err := dr.performFollowStatement(ctx, addSourceFollowerStatement, follower, followed); err != nil {
		return errors.Wrap(err, "failed to add source follower")
	}
	return nil
}

const removeSourceFollowerStatement = `
DELETE FROM
	user_source_follows
WHERE
	follower=? AND followed=?
`

// RemoveSourceFollower adds a user follow relationship
func (dr *dataRepository) RemoveSourceFollower(ctx context.Context, follower, followed string) error {
	if err := dr.performFollowStatement(ctx, removeSourceFollowerStatement, follower, followed); err != nil {
		return errors.Wrap(err, "failed to remove source follower")
	}
	return nil
}

func (dr *dataRepository) performFollowStatement(ctx context.Context, statement, follower, followed string) error {
	tx, err := dr.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to begin transaction")
	}
	stm, err := tx.PrepareContext(ctx, addUserFollowerStatement)
	if err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to perform follow %s-%s", follower, followed)
		}
		return errors.Wrapf(err, "failed to prepare statement to perform follow %s-%s", follower, followed)
	}
	_, err = stm.ExecContext(ctx,
		follower,
		followed,
	)
	if err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to perform follow %s-%s", follower, followed)
		}
		return errors.Wrapf(err, "failed to execute statment to perform follow %s-%s", follower, followed)
	}
	if err := tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to perform follow %s-%s", follower, followed)
		}
		return errors.Wrapf(err, "failed to perform follow %s-%s", follower, followed)
	}
	return nil
}

// ValidateUserForCreate validates user fields against the data repo for create
func (dr *dataRepository) ValidateUserForCreate(ctx context.Context, user *DBUser) bool {
	if checkUser, _ := dr.GetUserByEmail(ctx, user.Email); checkUser != nil {
		fmt.Printf("user email exists\n")
		return false
	}
	if checkUser, _ := dr.GetUserByUsername(ctx, user.Username); checkUser != nil {
		fmt.Printf("user username exists\n")
		return false
	}
	return true
}

const createUserStatement = `
INSERT INTO
	users (
		uuid,
		username,
		email,
		hashed_password,
		created_by_uuid,
		created_at,
		updated_by_uuid,
		updated_at
	)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?)
`

//CreateUser creates a user
func (dr *dataRepository) CreateUser(ctx context.Context, user *DBUser) error {
	tx, err := dr.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to begin transaction")
	}
	stm, err := tx.PrepareContext(ctx, createUserStatement)
	if err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to create user %+v", user)
		}
		return errors.Wrapf(err, "failed to prepare statement to create user %+v", user)
	}
	_, err = stm.ExecContext(ctx,
		user.UUID,
		user.Username,
		user.Email,
		user.HashedPassword,
		user.CreatedByUUID,
		user.CreatedAt,
		user.UpdatedByUUID.String,
		user.UpdatedAt.Int64,
	)
	if err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to create user %+v", user)
		}
		return errors.Wrapf(err, "failed to execute statment to create user %+v", user)
	}
	if err := tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to create user %+v", user)
		}
		return errors.Wrapf(err, "failed to create user %+v", user)
	}
	return nil
}

package user

import (
	"context"
	"time"

	"github.com/xhd2015/kool/tools/create/server_go_db_template/dao/user/t_user"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/dao/user/t_user_auth_token"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/dao/user/t_user_token"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/server_errors"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

// QueryUserByAuthToken queries user ID by auth token
// Returns 0 if token is not found or invalid
func QueryUserByAuthToken(ctx context.Context, token string) (types.UserID, error) {
	if token == "" {
		return 0, nil
	}

	userAuthToken, err := t_user_auth_token.ORM.SelectAll().Where(t_user_auth_token.Token.Eq(token)).QueryOne(ctx)
	if err != nil {
		return 0, err
	}
	if userAuthToken == nil {
		return 0, nil
	}

	return userAuthToken.UserId, nil
}

// InsertUser inserts a new user into the database
func InsertUser(ctx context.Context, name, pwdHash string) error {
	_, err := t_user.ORM.Insert(ctx, &t_user.User{
		Name:    name,
		PwdHash: pwdHash,
	})
	return err
}

// FindUserPasswordByName finds a user by their name (includes password hash)
func FindUserPasswordByName(ctx context.Context, name string) (*t_user.User, error) {
	return t_user.ORM.SelectAll().Where(t_user.Name.Eq(name)).QueryOne(ctx)
}

// GetUserByID gets a user by their ID
func GetUserByID(ctx context.Context, userID types.UserID) (*t_user.User, error) {
	user, err := t_user.ORM.SelectAll().Where(t_user.ID.Eq(int64(userID))).QueryOne(ctx)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, server_errors.ErrUserNotFound
	}
	return user, nil
}

// CountUsersByName counts users with the given name
func CountUsersByName(ctx context.Context, name string) (int64, error) {
	return t_user.ORM.Count().Where(t_user.Name.Eq(name)).Query(ctx)
}

// InsertUserToken inserts a new token for a user
func InsertUserToken(ctx context.Context, userID types.UserID, token string, expireTime time.Time) error {
	_, err := t_user_token.ORM.Insert(ctx, &t_user_token.UserToken{
		UserId:     userID,
		Token:      token,
		ExpireTime: expireTime,
	})
	return err
}

// FindUserByToken finds a user by their token
// Uses two ORM queries: first finds the token, then finds the user
// Note: arc-orm doesn't support joins, so we use two separate queries
func FindUserByToken(ctx context.Context, token string) (*t_user.User, error) {
	// First, find the token and check if it's not expired
	userToken, err := t_user_token.ORM.SelectAll().
		Where(t_user_token.Token.Eq(token)).
		QueryOne(ctx)
	if err != nil {
		return nil, err
	}
	if userToken == nil {
		return nil, nil
	}

	// Check if token is expired
	if userToken.ExpireTime.Before(time.Now()) {
		return nil, nil
	}

	// Then, find the user by ID
	return t_user.ORM.SelectAll().Where(t_user.ID.Eq(int64(userToken.UserId))).QueryOne(ctx)
}

// DeleteUserToken deletes a token
func DeleteUserToken(ctx context.Context, token string) error {
	return t_user_token.ORM.DeleteWhere(ctx, t_user_token.Token.Eq(token))
}

// GetAllUsers retrieves all users from the database
func GetAllUsers(ctx context.Context) ([]*t_user.User, error) {
	return t_user.ORM.SelectAll().Query(ctx)
}

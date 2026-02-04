package model

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

// User represents a user in the system
type User struct {
	Id       types.UserID `json:"id"`
	Name     string       `json:"name"`
	CreateAt time.Time    `json:"create_time"`
	UpdateAt time.Time    `json:"update_time"`
}

type UserPassword struct {
	Id       int64     `json:"id"`
	Name     string    `json:"name"`
	PwdHash  string    `json:"pwd_hash"`
	CreateAt time.Time `json:"create_time"`
	UpdateAt time.Time `json:"update_time"`
}

// UserToken represents a user's authentication token
type UserToken struct {
	Id         int64     `json:"id"`
	UserId     int64     `json:"user_id"`
	Token      string    `json:"token"`
	ExpireTime time.Time `json:"expire_time"`
	CreateAt   time.Time `json:"create_time"`
	UpdateAt   time.Time `json:"update_time"`
}

// LoginRequest represents the request body for login/register
type LoginRequest struct {
	Name string `json:"name"`

	X string `json:"x"` // base64 encoded password

	From string `json:"from"` // from: web, ios, android, etc.
}

// DecodePassword decodes the base64 encoded password
func (r *LoginRequest) DecodePassword() (string, error) {
	if r.X == "" {
		return "", fmt.Errorf("password is required")
	}
	decoded, err := base64.StdEncoding.DecodeString(r.X)
	if err != nil {
		return "", fmt.Errorf("invalid base64 encoded password: %w", err)
	}
	return string(decoded), nil
}

// LoginResponse represents the response for login/register
type LoginResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
}

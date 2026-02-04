package t_user_auth_token

import (
	"time"

	"github.com/xhd2015/arc-orm/orm"
	"github.com/xhd2015/arc-orm/table"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/dao/engine"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

/*
CREATE TABLE `t_user_auth_token` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `user_id` bigint(20) NOT NULL DEFAULT 0 COMMENT 'user id',
    `token` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'auth token value',
    `token_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'token type (bearer, etc)',
    `expires_at` datetime NOT NULL DEFAULT current_timestamp() COMMENT 'token expiration time',
    `create_time` datetime NOT NULL DEFAULT current_timestamp(),
    `update_time` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
    PRIMARY KEY (`id`),
    INDEX `idx_user_id`(`user_id`),
    INDEX `idx_token`(`token`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'user authentication tokens';
*/
/*
  ALTER TABLE `t_user_auth_token` ADD COLUMN `token_type` VARCHAR(64) NOT NULL DEFAULT "" COMMENT "token type (bearer, etc)"";
*/
var Table = table.New("t_user_auth_token")

var (
	ID         = Table.Int64("id")
	UserID     = Table.Int64("user_id")
	Token      = Table.String("token")
	TokenType  = Table.String("token_type")
	ExpiresAt  = Table.Time("expires_at")
	CreateTime = Table.Time("create_time")
	UpdateTime = Table.Time("update_time")
)

var ORM = orm.Bind[UserAuthToken, UserAuthTokenOptional](engine.Engine, Table)

type UserAuthToken struct {
	Id         int64        `json:"id"`
	UserId     types.UserID `json:"user_id"`
	Token      string       `json:"token"`
	TokenType  string       `json:"token_type"`
	ExpiresAt  time.Time    `json:"expires_at"`
	CreateTime time.Time    `json:"create_time"`
	UpdateTime time.Time    `json:"update_time"`
}

type UserAuthTokenOptional struct {
	Id         *int64        `json:"id"`
	UserId     *types.UserID `json:"user_id"`
	Token      *string       `json:"token"`
	TokenType  *string       `json:"token_type"`
	ExpiresAt  *time.Time    `json:"expires_at"`
	CreateTime *time.Time    `json:"create_time"`
	UpdateTime *time.Time    `json:"update_time"`
}

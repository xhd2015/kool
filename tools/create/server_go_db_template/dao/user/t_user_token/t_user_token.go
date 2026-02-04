package t_user_token

import (
	"time"

	"github.com/xhd2015/arc-orm/orm"
	"github.com/xhd2015/arc-orm/table"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/dao/engine"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

/*
CREATE TABLE `t_user_token` (

	`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
	`user_id` bigint(20) NOT NULL DEFAULT 0 COMMENT 'user id',
	`token` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'auth token',
	`expire_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT 'token expiration time',
	`create_time` datetime NOT NULL DEFAULT current_timestamp(),
	`update_time` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
	PRIMARY KEY (`id`),
	INDEX `idx_user_id`(`user_id`),
	INDEX `idx_token`(`token`)

) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'user tokens';
*/
var Table = table.New("t_user_token")

var (
	ID         = Table.Int64("id")
	UserID     = Table.Int64("user_id")
	Token      = Table.String("token")
	ExpireTime = Table.Time("expire_time")
	CreateTime = Table.Time("create_time")
	UpdateTime = Table.Time("update_time")
)

var ORM = orm.Bind[UserToken, UserTokenOptional](engine.Engine, Table)

type UserToken struct {
	Id         int64        `json:"id"`
	UserId     types.UserID `json:"user_id"`
	Token      string       `json:"token"`
	ExpireTime time.Time    `json:"expire_time"`
	CreateTime time.Time    `json:"create_time"`
	UpdateTime time.Time    `json:"update_time"`
}

type UserTokenOptional struct {
	Id         *int64        `json:"id"`
	UserId     *types.UserID `json:"user_id"`
	Token      *string       `json:"token"`
	ExpireTime *time.Time    `json:"expire_time"`
	CreateTime *time.Time    `json:"create_time"`
	UpdateTime *time.Time    `json:"update_time"`
}

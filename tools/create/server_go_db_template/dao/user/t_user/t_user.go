package t_user

import (
	"time"

	"github.com/xhd2015/arc-orm/orm"
	"github.com/xhd2015/arc-orm/table"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/dao/engine"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

/*
CREATE TABLE `t_user` (

	`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
	`name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'user name',
	`pwd_hash` VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'password hash',
	`create_time` datetime NOT NULL DEFAULT current_timestamp(),
	`update_time` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
	PRIMARY KEY (`id`),
	UNIQUE INDEX `idx_name`(`name`)

) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'users';
*/
var Table = table.New("t_user")

var (
	ID         = Table.Int64("id")
	Name       = Table.String("name")
	PwdHash    = Table.String("pwd_hash")
	CreateTime = Table.Time("create_time")
	UpdateTime = Table.Time("update_time")
)

var ORM = orm.Bind[User, UserOptional](engine.Engine, Table)

type User struct {
	Id         types.UserID `json:"id"`
	Name       string       `json:"name"`
	PwdHash    string       `json:"pwd_hash"`
	CreateTime time.Time    `json:"create_time"`
	UpdateTime time.Time    `json:"update_time"`
}

type UserOptional struct {
	Id         *types.UserID `json:"id"`
	Name       *string       `json:"name"`
	PwdHash    *string       `json:"pwd_hash"`
	CreateTime *time.Time    `json:"create_time"`
	UpdateTime *time.Time    `json:"update_time"`
}

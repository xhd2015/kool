package dao

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	engine_dao "github.com/xhd2015/kool/tools/create/server_go_db_template/dao/engine"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/env"
	"xorm.io/xorm"
)

var Engine *xorm.Engine

var DB *sql.DB

func Init() {
	database := env.MySQLDatabase()
	if database == "" {
		database = "app_db"
	}
	port := env.MySQLPort()
	if port == "" {
		port = "3306"
	}
	user := env.MySQLUser()
	if user == "" {
		user = "root"
	}
	password := env.MySQLPassword()
	host := env.MySQLHost()
	if host == "" {
		host = "localhost"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True", user, password, host, port, database)

	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		log.Fatalf("failed opening xorm connection to mysql: %v", err)
	}
	engine.DatabaseTZ = time.Local
	engine.TZLocation = time.Local

	engine.ShowSQL(true)

	Engine = engine

	engine_dao.Init(ormAdaptor{})
}

func Destroy() error {
	return Engine.Close()
}

func Session() *xorm.Session {
	return Engine.NoAutoCondition().NoAutoTime()
}

func QuerySQL(ctx context.Context, sql string, args []interface{}, result interface{}) error {
	return Session().Context(ctx).SQL(sql, args...).Find(result)
}

func QueryOneSQL(ctx context.Context, sql string, args []interface{}, result interface{}) error {
	rv := reflect.ValueOf(result)
	typ := rv.Type()
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}
	if typ.Kind() == reflect.Slice {
		return fmt.Errorf("result must be a pointer to a single object")
	}
	elem := typ.Elem()
	list := reflect.New(reflect.SliceOf(elem))
	err := Session().Context(ctx).SQL(sql, args...).Find(list.Interface())
	if err != nil {
		return err
	}
	if list.Elem().Len() == 0 {
		return nil
	}
	rv.Elem().Set(list.Elem().Index(0))
	return nil
}

func ExecSQL(ctx context.Context, sql string, args []interface{}) error {
	_, err := Session().Context(ctx).Exec(append([]interface{}{sql}, args...)...)
	return err
}

func InsertSQL(ctx context.Context, sql string, args []interface{}) (int64, error) {
	res, err := Session().Context(ctx).Exec(append([]interface{}{sql}, args...)...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

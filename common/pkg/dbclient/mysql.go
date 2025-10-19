package dbclient

import (
	"database/sql"
	"fmt"
	"pulse/common/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var _defaultDB *gorm.DB

func GetMysqlDB() *gorm.DB {
	if _defaultDB == nil {
		logger.GetLogger().Error("mysql database is not initialized")
		return nil
	}
	return _defaultDB
}

// SetMysqlDB 仅用于测试，用于注入一个mock的DB实例
func SetMysqlDB(mockDB *gorm.DB) {
	_defaultDB = mockDB
}

// Init 函数负责初始化数据库连接
func Init(dsn, logMode string, maxIdleConns, maxOpenConns int) (*gorm.DB, error) {
	mysqlConfig := mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		SkipInitializeWithVersion: false,
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), setConfig(logMode)); err != nil {
		return nil, err
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		_defaultDB = db
		return db, nil
	}
}

// 辅助函数, 用于执行创建数据库的SQL语句
func CreateDatabase(dsn string, driver string, createSql string) error {
	// 验证参数
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}

	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(db)

	if err = db.Ping(); err != nil {
		return err
	}

	// 执行传入sql语句
	_, err = db.Exec(createSql)
	return err
}

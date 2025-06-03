package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	gormMysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chaihaobo/gocommon/tls"
)

const (
	DefaultMaxOpen     = 10
	DefaultMaxIdle     = 10
	DefaultMaxLifetime = 3
)

func dataSourceName(config Config) string {
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.User, config.Password, config.Host, config.Port, config.Name)
	val := url.Values{}
	if config.ParseTime {
		val.Add("parseTime", "1")
	}
	if len(config.Location) > 0 {
		val.Add("loc", config.Location)
	}
	if config.CA != nil {
		val.Add("tls", "custom")
	}

	if len(val) == 0 {
		return connection
	}
	return fmt.Sprintf("%s?%s", connection, val.Encode())
}

// DB return new sql db
func DB(config Config) (*sql.DB, error) {
	if config.CA != nil && len(config.ServerName) != 0 {
		if err := mysql.RegisterTLSConfig("custom", tls.WithServerAndCA(config.ServerName, config.CA)); err != nil {
			log.Println(err)
			return nil, err
		}
	} else if config.CA != nil {
		if err := mysql.RegisterTLSConfig("custom", tls.WithCA(config.CA)); err != nil {
			log.Println(err)
			return nil, err
		}
	}
	db, err := sql.Open("mysql", dataSourceName(config))
	if err != nil {
		return nil, err
	}

	if config.MaxOpen > 0 {
		db.SetMaxOpenConns(config.MaxOpen)
	} else {
		db.SetMaxOpenConns(DefaultMaxOpen)
	}

	if config.MaxIdle > 0 {
		db.SetMaxIdleConns(config.MaxIdle)
	} else {
		db.SetMaxIdleConns(DefaultMaxIdle)
	}

	if config.MaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Minute)
	} else {
		db.SetConnMaxLifetime(time.Duration(DefaultMaxLifetime) * time.Minute)
	}

	if config.MaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(config.MaxIdleTime) * time.Minute)
	}

	return db, nil
}

// GormDB return new gorm db
func GormDB(config Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	rawDB, err := DB(config)
	if err != nil {
		return nil, err
	}
	if gormConfig == nil {
		gormConfig = defaultGormConfig()
	}
	gormDB, err := gorm.Open(gormMysqlDriver.New(gormMysqlDriver.Config{
		Conn: rawDB,
	}), gormConfig)
	return gormDB, nil
}

// defaultGormConfig return default gorm config
func defaultGormConfig() *gorm.Config {
	defaultGormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	}
	return defaultGormConfig
}

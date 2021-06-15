package db

import (
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"

	"agones-minecraft/config"
	gamev1Model "agones-minecraft/models/v1/game"
	twitchv1Model "agones-minecraft/models/v1/twitch"
	userv1Model "agones-minecraft/models/v1/user"
)

const (
	DefaultThreshold time.Duration = time.Millisecond * 500
)

var db *gorm.DB

func New() *bun.DB {
	dbconfig := config.GetDBConfig()

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbconfig.User,
		dbconfig.Password,
		dbconfig.Hostname,
		dbconfig.Port,
		dbconfig.Name,
	)

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	// disable prepared statements
	config.PreferSimpleProtocol = true

	sqldb := stdlib.OpenDB(*config)

	if err != nil {
		log.Fatal("error connecting to datbase")
	}

	return bun.NewDB(sqldb, pgdialect.New())
}

func Init() {
	var err error
	logger := zapgorm2.New(zap.L())
	logger.SlowThreshold = DefaultThreshold

	if config.GetEnv() == config.Production {
		logger.IgnoreRecordNotFoundError = true
	} else {
		logger.LogLevel = gormlogger.Info
	}

	dbconfig := config.GetDBConfig()

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		dbconfig.User,
		dbconfig.Password,
		dbconfig.Hostname,
		dbconfig.Port,
		dbconfig.Name,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger})
	if err != nil {
		zap.L().Fatal("error connecting to db", zap.Error(err))
	}

	if config.GetEnv() == config.Development {
		if err = db.AutoMigrate(
			&userv1Model.User{},
			&twitchv1Model.TwitchToken{},
			&gamev1Model.Game{},
		); err != nil {
			zap.L().Fatal("error auto-migrating db", zap.Error(err))
		}
	}

}

func DB() *gorm.DB {
	return db
}

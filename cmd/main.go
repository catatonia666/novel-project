package main

import (
	"context"
	"dialogue/internal/models"
	"html/template"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	dialogues     *models.DialogueModel
	users         *models.UserModel
	templateCache map[string]*template.Template
	redisClient   *redis.Client
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)

	dsn := DefaultPostgresConfig().ConnectionInfo()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	db.Migrator().DropTable("first_blocks", "blocks")
	db.AutoMigrate(&models.FirstBlock{}, &models.Block{}, &models.User{})

	redisClient := redis.NewClient(DefaultRedisConfig())

	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		dialogues:     &models.DialogueModel{DB: db},
		users:         &models.UserModel{DB: db},
		templateCache: templateCache,
		redisClient:   redisClient,
	}

	app.routes().Run(srvArrd)
}

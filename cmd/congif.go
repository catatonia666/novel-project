package main

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5431,
		User:     "postgres",
		Password: "world555",
		Name:     "rpg",
	}
}

func (c PostgresConfig) ConnectionInfo() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s "+
			"sslmode=disable", c.Host, c.Port, c.User, c.Name)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s "+
		"dbname=%s sslmode=disable", c.Host, c.Port, c.User,
		c.Password, c.Name)
}

func DefaultRedisConfig() *redis.Options {
	return &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
}

const srvArrd = ":3000"

package database

import (
	"fmt"

	"github.com/gopl-dev/server/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (db *gorm.DB, err error) {
	c := config.Get().DB
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable TimeZone=Asia/Shanghai",
		c.Host,
		c.Port,
		c.Name,
		c.User,
		c.Password,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		err = fmt.Errorf("db connect: %w", err)
	}

	return
}

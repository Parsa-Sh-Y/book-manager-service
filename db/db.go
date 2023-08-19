package db

import (
	"fmt"

	"github.com/Parsa-Sh-Y/book-manager-service/config"
	"github.com/Parsa-Sh-Y/book-manager-service/db/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type gormDB struct {
	db *gorm.DB
}

func CreateNewGormDB(config config.Config) (*gormDB, error) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &gormDB{
		db: db,
	}, nil

}

func (gdb *gormDB) CreateSchema() error {

	err := gdb.db.AutoMigrate(&models.User{}, &models.Book{}, &models.Content{})

	if err != nil {
		return err
	}

	return nil

}

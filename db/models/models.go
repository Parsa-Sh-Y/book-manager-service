package models

import (
	"time"
)

type User struct {
	ID          uint
	Username    string `gorm:"type:varchar(50)"`
	Email       string `gorm:"type:varchar(50)"`
	Password    string `gorm:"type:varchar(255)"`
	Firstname   string `gorm:"type:varchar(50)"`
	Lastname    string `gorm:"type:varchar(50)"`
	PhoneNumber string `gorm:"type:char(11)"`
	Gender      string `gorm:"type:varchar(50)"`
	Books       []Book
}

type Content struct {
	ID          uint
	ContentName string `gorm:"type:varchar(255)"`
	BookId      uint
}

type Book struct {
	ID                uint
	Name              string    `gorm:"type:varchar(255)"`
	Category          string    `gorm:"type:varchar(255)"`
	Volumn            int       `gorm:"type:integer"`
	PublishedAt       time.Time `gorm:"type:date"`
	TableOfContents   []Content `gorm:"constraint:onUpdate:CASCADE,onDelete:CASCADE"`
	Summary           string    `gorm:"type:text"`
	Publisher         string    `gorm:"type:varchar(255)"`
	AuthorFirstName   string    `gorm:"type:varchar(50)"`
	AuthorLastName    string    `gorm:"type:varchar(50)"`
	AuthorBirthday    time.Time `gorm:"type:date"`
	AuthorNationality string    `gorm:"type:varchar(50)"`
	UserID            uint
}

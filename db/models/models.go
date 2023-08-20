package models

import (
	"time"
)

type User struct {
	ID          uint
	Username    string `gorm:"type:varchar(50)" json:"user_name"`
	Email       string `gorm:"type:varchar(50)" json:"email"`
	Password    string `gorm:"type:varchar(255)" json:"password"`
	Firstname   string `gorm:"type:varchar(50)" json:"first_name"`
	Lastname    string `gorm:"type:varchar(50)" json:"last_name"`
	PhoneNumber string `gorm:"type:char(11)"  json:"phone_number"`
	Gender      string `gorm:"type:varchar(50)"  json:"gender"`
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

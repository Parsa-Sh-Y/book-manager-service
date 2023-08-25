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
	ID                  uint
	Name                string    `gorm:"type:varchar(255)" json:"name"`
	Category            string    `gorm:"type:varchar(255)" json:"category"`
	Volumn              int       `gorm:"type:integer" json:"volumn"`
	PublishedAt         time.Time `gorm:"type:date" json:"published_at"`
	TableOfContents     []Content `gorm:"constraint:onUpdate:CASCADE,onDelete:CASCADE"`
	TableOfContentsJson []string  `gorm:"-:all" json:"table_of_contents"` // only for json puposes, no such field would be created in the database
	Summary             string    `gorm:"type:text" json:"summary"`
	Publisher           string    `gorm:"type:varchar(255)" json:"publisher"`
	Author              author    `gorm:"embedded" json:"author"`
	UserID              uint      `json:"-"`
}

type author struct {
	AuthorFirstName   string    `gorm:"type:varchar(50)" json:"first_name"`
	AuthorLastName    string    `gorm:"type:varchar(50)" json:"last_name"`
	AuthorBirthday    time.Time `gorm:"type:date" json:"birthday"`
	AuthorNationality string    `gorm:"type:varchar(50)" json:"nationality"`
}

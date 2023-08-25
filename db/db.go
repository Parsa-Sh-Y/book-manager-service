package db

import (
	"errors"
	"fmt"

	"github.com/Parsa-Sh-Y/book-manager-service/config"
	"github.com/Parsa-Sh-Y/book-manager-service/db/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// ErrEmailIsInUse There exists another account with the same email
	ErrEmailIsInUse = errors.New("email is in use by another account")
	// ErrUsernameIsInUse There exists another account with the same username
	ErrUsernameIsInUse = errors.New("username is in use by another account")
	// ErrPhoneNumberIsInUse There exists another account with the same phone number
	ErrPhoneNumberIsInUse = errors.New("phone number is in use by another account")
	// ErrUserNotFound Could not find the user
	ErrUserNotFound = errors.New("no user was found or multiple users were found")
	// ErrBookNotFound No such book exists
	ErrBookNotFound = errors.New("no such book exists")
	// ErrPermissionDenied Permission Denied
	ErrPermissionDenied = errors.New("permission Denied")
)

type GormDB struct {
	db *gorm.DB
}

func CreateNewGormDB(config config.Config) (*GormDB, error) {

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

	return &GormDB{
		db: db,
	}, nil

}

func (gdb *GormDB) CreateSchema() error {

	err := gdb.db.AutoMigrate(&models.User{}, &models.Book{}, &models.Content{})

	if err != nil {
		return err
	}

	return nil

}

func (gdb *GormDB) CreateUser(user *models.User) error {

	// Check if no other account with the same username exists
	var count int64
	gdb.db.Model(&models.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return ErrUsernameIsInUse
	}

	// Check if no other account with the same email exists
	gdb.db.Model(&models.User{}).Where("email = ?", user.Email).Count(&count)
	if count > 0 {
		return ErrEmailIsInUse
	}

	// Check if no other account with the same phone number exists
	gdb.db.Model(&models.User{}).Where("phone_number = ?", user.PhoneNumber).Count(&count)
	if count > 0 {
		return ErrPhoneNumberIsInUse
	}

	if pw, err := bcrypt.GenerateFromPassword([]byte(user.Password), 4); err != nil {
		return err
	} else {
		user.Password = string(pw)
	}

	result := gdb.db.Create(user)
	return result.Error
}

func (gdb *GormDB) CreateBook(book *models.Book) error {

	return gdb.db.Create(&book).Error
}

func (gdb *GormDB) GetBook(id int) (*models.Book, error) {

	var book models.Book
	err := gdb.db.Where("id = ?", id).First(&book).Error

	if err != nil {
		return nil, err
	} else {
		return &book, nil
	}

}

func (gdb *GormDB) DeleteBook(id uint) error {

	return gdb.db.Delete(&models.Book{}, id).Error
}

func (gdb *GormDB) DeleteUserBook(username string, bookId uint) error {

	// find the book
	var book models.Book
	result := gdb.db.Model(models.Book{}).Where("id = ?", bookId).Find(&book)
	if result.RowsAffected == 0 {
		return ErrBookNotFound
	} else if result.Error != nil {
		return result.Error
	}

	// find the user who owns the book
	var user models.User
	result = gdb.db.Model(models.User{}).Where("username = ?", username).Find(&user)
	if result.RowsAffected != 1 {
		return ErrUserNotFound
	} else if result.Error != nil {
		return result.Error
	}

	// delete the book if the user owns it
	if user.ID == book.UserID {
		return gdb.DeleteBook(bookId)
	} else {
		return ErrPermissionDenied
	}
}

func (gdb *GormDB) UpdateBook(id uint, name string, category string) error {
	return gdb.db.Model(models.Book{}).Where("id = ?", id).Update("name", name).Update("category", category).Error
}

func (gdb *GormDB) UpdateUserBook(username string, bookId uint, name string, category string) error {
	// find the book
	var book models.Book
	result := gdb.db.Model(models.Book{}).Where("id = ?", bookId).Find(&book)
	if result.RowsAffected == 0 {
		return ErrBookNotFound
	} else if result.Error != nil {
		return result.Error
	}

	// find the user who owns the book
	var user models.User
	result = gdb.db.Model(models.User{}).Where("username = ?", username).Find(&user)
	if result.RowsAffected != 1 {
		return ErrUserNotFound
	} else if result.Error != nil {
		return result.Error
	}

	// update the book if the user owns it
	if user.ID == book.UserID {
		return gdb.UpdateBook(bookId, name, category)
	} else {
		return ErrPermissionDenied
	}
}

// The boolean returned is flase when there is an error
func (gdb *GormDB) IsUsernamePresent(username string) (bool, error) {

	var count int64
	err := gdb.db.Model(models.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}

	if count == 1 {
		return true, nil
	} else {
		return false, nil
	}

}

// When there is an error nil is return instead of a user
func (gdb *GormDB) GetUserByUsername(username string) (*models.User, error) {

	var user models.User
	result := gdb.db.Model(models.User{}).Where("username = ?", username).Find(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 1 {
		return &user, nil
	} else {
		return nil, ErrUserNotFound
	}

}

func (gdb *GormDB) GetAllBooks() (*[]models.Book, error) {

	var books []models.Book

	err := gdb.db.Model(models.Book{}).Find(&books).Error
	if err != nil {
		return nil, err
	}

	return &books, nil

}

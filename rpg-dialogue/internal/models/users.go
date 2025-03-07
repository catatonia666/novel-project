package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID             int    `gorm:"primary_key"`
	NickName       string `gorm:"type:text"`
	Email          string `gorm:"uniqueIndex"`
	HashedPassword []byte `gorm:"type:varchar(100)"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type UserModel struct {
	DB *gorm.DB
}

// Insert insets a new user into the database.
func (um *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (nick_name, email, hashed_password)
    VALUES(?, ?, ?)`
	if result := um.DB.Exec(stmt, name, email, hashedPassword); result.Error != nil {
		return ErrDuplicateEmail
	}
	return nil
}

// Authenticate authenticates a user with given data.
func (um *UserModel) Authenticate(email, password string) (int, error) {

	// Retrieve ID and hashed password associated with the given email. If no matching email exists then return an error.
	var userInfo User
	result := um.DB.Model(&User{}).Where("email = ?", email).Find(&userInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, result.Error
		}
	}

	// Check whether the hashed password and plain-text password provided match. If correct, return user ID.
	err := bcrypt.CompareHashAndPassword(userInfo.HashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	return userInfo.ID, nil
}

// Exists checks if user with provided ID exists in the database.
func (um *UserModel) Exists(id int) (bool, error) {
	var exists bool
	err := um.DB.Model(&User{}).Select("count(*) > 0").Where("id = ?", id).Find(&exists).Error
	return exists, err
}

// GetUser gets user with provided ID if exists.
func (um *UserModel) GetUser(id int) (*User, error) {
	var user User
	err := um.DB.Model(&User{}).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return &user, nil
}

// PasswordUpdate updates user's password.
func (um *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	var currentHashedPassword User
	err := um.DB.Model(&User{}).Select("hashed_password").Where("id = ?", id).Scan(&currentHashedPassword).Error
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword(currentHashedPassword.HashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	err = um.DB.Model(&User{}).Where("id = ?", id).Update("hashed_password", &newHashedPassword).Error
	return err
}

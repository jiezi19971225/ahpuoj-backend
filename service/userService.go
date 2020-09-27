package service

import "gorm.io/gorm"

type UserService struct {
	*gorm.DB
}

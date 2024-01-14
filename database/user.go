package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID    int
	Name  string
	Email string
}

package database

//
//import (
//	"fmt"
//	"log"
//
//	"gorm.io/driver/mysql"
//	"gorm.io/gorm"
//)
//
//const (
//	username = "usercache_user"
//	password = "password"
//	host     = "localhost"
//	port     = 3306
//	dbName   = "cache_map"
//)
//
//var db *gorm.DB
//
//func init() {
//	ConnectDB()
//}
//
//type User struct {
//	gorm.Model
//	ID    int
//	Name  string
//	Email string
//}
//
//func ConnectDB() {
//	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
//		username, password, host, port, dbName)
//
//	var err error
//	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Connected to the database")
//
//}
//func GetUserByID(userID uint) (*User, error) {
//	var user User
//	result := db.First(&user, userID)
//	if result.Error != nil {
//		log.Printf("Error retrieving user with ID %d: %v", userID, result.Error)
//		return nil, result.Error
//	}
//	log.Printf("User with ID %d retrieved from the database", userID)
//	return &user, nil
//}
//
//func createUsers() {
//	for i := 1; i <= 100; i++ {
//		user := User{
//			Name:  fmt.Sprintf("User%d", i),
//			Email: fmt.Sprintf("user%d@email.com", i),
//		}
//
//		db.Create(&user)
//	}
//	log.Println("Created 100 users in the database")
//}

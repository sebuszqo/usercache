package database

import (
	"log"
	"sync"
	"sync/atomic"
)

var (
	getUserCounterAtomic int64
	slowQueryErrors      int
	slowQueryErrorsMu    sync.Mutex

	getUserCounterMutex sync.Mutex
	getUserCounter      map[uint]int
)

func init() {
	getUserCounter = make(map[uint]int)
}

func IncrementGetUserCounterAtomic() {
	atomic.AddInt64(&getUserCounterAtomic, 1)
}

func GetGetUserCounterAtomic() int64 {
	return atomic.LoadInt64(&getUserCounterAtomic)
}
func IncrementGetUserCounter(userID uint) {
	getUserCounterMutex.Lock()
	defer getUserCounterMutex.Unlock()
	getUserCounter[userID]++
}

func GetUserCounter() map[uint]int {
	getUserCounterMutex.Lock()
	defer getUserCounterMutex.Unlock()

	return getUserCounter
}

func IncrementSlowQueryErrors() {
	slowQueryErrorsMu.Lock()
	slowQueryErrors++
	slowQueryErrorsMu.Unlock()
}

func GetSlowQueryErrors() int {
	slowQueryErrorsMu.Lock()
	defer slowQueryErrorsMu.Unlock()
	return slowQueryErrors
}

func GetUserByID(userID uint) (*User, error) {
	var user User

	result := db.First(&user, userID)

	IncrementGetUserCounter(userID)
	IncrementGetUserCounterAtomic()
	if result.Error != nil {
		IncrementSlowQueryErrors()
		log.Printf("Error retrieving user with ID %d: %v", userID, result.Error)
		return nil, result.Error
	}
	log.Printf("User with ID %d retrieved from the database", userID)
	return &user, nil

}

func CreateUser(user *User) error {
	result := db.Create(user)
	if result.Error != nil {
		log.Printf("Error creating user: %v", result.Error)
		return result.Error
	}
	log.Printf("User with ID %d created in the database", user.ID)
	return nil
}

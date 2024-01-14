package test

import (
	"fmt"
	cachebucket "github.com/sebuszqo/usercache/cache_bucket"
	"github.com/sebuszqo/usercache/database"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestUserCache_GetUser(t *testing.T) {

	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: "localhost:6379",
	//	DB:   0,
	//})
	//uc := cacheredis.NewUserCache(redisClient)

	//uc := cachemap.NewUserCache()

	uc := cachebucket.NewUserCache(50)

	//redisClient := redis.NewClusterClient(&redis.ClusterOptions{
	//	Addrs: []string{"127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7002"},
	//})
	//uc := cacheredis.NewUserCache(redisClient)

	var wg sync.WaitGroup

	requests := 10000
	usersWithId := 200
	processedCount := int64(0)
	userChannel := make(chan *database.User, requests)

	startTime := time.Now()
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(userID uint) {
			defer wg.Done()

			user, err := uc.GetUser(userID)
			// normal database requests without cache
			// user, err := database.GetUserByID(userID)
			atomic.AddInt64(&processedCount, 1)

			if err != nil {
				//fmt.Printf("Error during getting user: %v\n", err)
				userChannel <- nil
				return
			}
			if user == nil {
				//fmt.Printf("User with ID %d couldn't be found\n", userID)
				userChannel <- nil
				return
			}
			userChannel <- user
		}(uint(i % usersWithId))
	}

	wg.Wait()
	close(userChannel)

	fmt.Println(database.GetUserCounter())
	fmt.Println(database.GetGetUserCounterAtomic())
	fmt.Println("errors:: ", database.GetSlowQueryErrors())
	// retrieving data from channel
	//for user := range userChannel {
	//	if user != nil {
	//		fmt.Printf("Getting user with ID: %+v\n", user)
	//	}
	//}
	elapsedTime := time.Since(startTime)
	fmt.Printf("Total time: %s\n", elapsedTime)
	counter := strconv.FormatInt(processedCount, 10)
	fmt.Printf("count: %v", counter)
}

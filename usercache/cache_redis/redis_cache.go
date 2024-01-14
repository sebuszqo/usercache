package cacheredis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sebuszqo/usercache/database"
	"log"
	"sync"
	"time"
)

var ctx = context.Background()

type UserCache struct {
	redisClient *redis.ClusterClient
	mu          sync.RWMutex
	pendingIDs  map[uint]*sync.WaitGroup
}

func NewUserCache(redisClient *redis.ClusterClient) *UserCache {
	return &UserCache{
		redisClient: redisClient,
		mu:          sync.RWMutex{},
		pendingIDs:  make(map[uint]*sync.WaitGroup),
	}
}

func (uc *UserCache) GetUser(userID uint) (*database.User, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)
	uc.mu.RLock()
	cachedData, err := uc.redisClient.Get(ctx, cacheKey).Bytes()
	uc.mu.RUnlock()

	var userToJson database.User
	err = json.Unmarshal(cachedData, &userToJson)
	if err == nil {
		log.Printf("User with ID: %d found in cache", userID)
		return &userToJson, nil
	}

	uc.mu.Lock()
	wg, pending := uc.pendingIDs[userID]
	if pending {
		uc.mu.Unlock()
		wg.Wait()
		log.Printf("User with ID: %d not found in cache, retrying after completion", userID)
		return uc.GetUser(userID) // Retry after completion
	}

	newWg := &sync.WaitGroup{}
	newWg.Add(1)
	uc.pendingIDs[userID] = newWg
	uc.mu.Unlock()

	log.Printf("User with ID: %d not found in cache, fetching from the database", userID)
	user, err := database.GetUserByID(userID)
	uc.mu.Lock()

	defer func() {
		newWg.Done()
		delete(uc.pendingIDs, userID)
		uc.mu.Unlock()
	}()

	if err != nil {
		jsonData, err := json.Marshal(database.User{})
		if err != nil {
			return nil, err
		}
		uc.redisClient.Set(ctx, cacheKey, jsonData, time.Second*10)
		return nil, err
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	uc.redisClient.Set(ctx, cacheKey, jsonData, time.Second*10)
	log.Printf("User %d fetched from the database and added to cache", userID)

	return user, nil
}

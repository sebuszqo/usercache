package cachebucket

import (
	"github.com/sebuszqo/usercache/database"
	"log"
	"sync"
)

// UserCache represents a cache with multiple buckets.
type UserCache struct {
	cacheBuckets []*cacheBucket
}

// cacheBucket represents an individual cache bucket.
type cacheBucket struct {
	cache      map[uint]*database.User
	mu         sync.RWMutex
	pendingIDs map[uint]chan struct{}
}

// NewUserCache initializes a new UserCache instance with the specified number of cache buckets.
func NewUserCache(cacheBucketCount int) *UserCache {
	cache := &UserCache{}
	cache.cacheBuckets = make([]*cacheBucket, cacheBucketCount)
	for i := 0; i < cacheBucketCount; i++ {
		cache.cacheBuckets[i] = &cacheBucket{
			cache:      make(map[uint]*database.User),
			mu:         sync.RWMutex{},
			pendingIDs: make(map[uint]chan struct{}),
		}
	}
	return cache
}

// GetUser retrieves user information either from the cache or the database using cache buckets.
func (uc *UserCache) GetUser(userID uint) (*database.User, error) {
	// Determine the bucket index based on the userID
	bucketIndex := userID % uint(len(uc.cacheBuckets))
	bucket := uc.cacheBuckets[bucketIndex]

	// RLock for concurrent read
	bucket.mu.Lock()
	user, ok := bucket.cache[userID]
	bucket.mu.Unlock()

	if ok {
		log.Printf("User with ID: %d found in cache\n", userID)
		return user, nil
	}

	bucket.mu.Lock()
	ch, pending := bucket.pendingIDs[userID]
	if pending {
		bucket.mu.Unlock()
		<-ch
		log.Printf("User with ID: %d not found in cache, retrying after completion", userID)
		return uc.GetUser(userID) // Retry after completion
	}

	// Create a channel for the pending request
	ch = make(chan struct{})
	bucket.pendingIDs[userID] = ch
	bucket.mu.Unlock()

	// Fetch the user from the database
	log.Printf("User with ID: %d not found in cache, fetching from the database", userID)
	user, err := database.GetUserByID(userID)
	if err != nil {
		bucket.mu.Lock()
		close(ch)
		delete(bucket.pendingIDs, userID)
		bucket.cache[userID] = nil
		bucket.mu.Unlock()
		return nil, err
	}

	bucket.mu.Lock()
	bucket.cache[userID] = user
	log.Printf("User %d fetched from the database and added to cache", userID)
	close(ch)
	delete(bucket.pendingIDs, userID)
	bucket.mu.Unlock()

	return user, nil
}

package cachemap

import (
	"github.com/sebuszqo/usercache/database"
	"log"
	"sync"
)

// UserCache represents a cache for storing user information with concurrent access support.
type UserCache struct {
	cache      map[uint]*database.User // Cache for storing user information
	mu         sync.RWMutex            // Mutex for protecting cache and pendingIDs maps
	pendingIDs map[uint]chan struct{}  // Map to track pending user requests
}

// NewUserCache creates a new instance of UserCache with initialized fields.
func NewUserCache() *UserCache {
	return &UserCache{
		cache:      make(map[uint]*database.User),
		mu:         sync.RWMutex{},
		pendingIDs: make(map[uint]chan struct{}),
	}
}

// GetUser retrieves user information either from the cache or the database using sync.RWMutex.
func (uc *UserCache) GetUser(userID uint) (*database.User, error) {
	// Try to get the user from the cache without blocking other readers
	uc.mu.Lock()
	entry, found := uc.cache[userID]
	uc.mu.Unlock()

	if found {
		log.Printf("User with ID: %d found in cache.", userID)
		return entry, nil
	}

	// User not found in cache, check if another goroutine is fetching the user
	uc.mu.Lock()
	ch, pending := uc.pendingIDs[userID]
	if pending {
		uc.mu.Unlock()
		<-ch
		log.Printf("User with ID: %d not found in cache, retrying after completion", userID)
		return uc.GetUser(userID) // Retry after completion
	}

	// Creating a WaitGroup for the pending request
	ch = make(chan struct{})
	uc.pendingIDs[userID] = ch
	uc.mu.Unlock()

	log.Printf("User with ID: %d not found in cache, fetching from the database", userID)
	user, err := database.GetUserByID(userID)

	if err != nil {
		// Error fetching user from the database, cleanup and return error
		uc.mu.Lock()
		uc.cache[userID] = nil
		delete(uc.pendingIDs, userID)
		close(ch)
		uc.mu.Unlock()
		return nil, err
	}

	// User successfully fetched, update cache and cleanup
	uc.mu.Lock()
	uc.cache[userID] = user
	log.Printf("User %d fetched from the database and added to cache", userID)
	delete(uc.pendingIDs, userID)
	close(ch)
	uc.mu.Unlock()

	return user, nil
}

// Solution with sync.Map - that works well
//// UserCache represents a concurrent-safe user cache using sync.Map.
//type UserCache struct {
//	cache      sync.Map
//	pendingIDs map[uint]*sync.WaitGroup
//	mu         sync.Mutex
//}
//
//// NewUserCache creates a new UserCache instance with initialized fields.
//func NewUserCache() *UserCache {
//	return &UserCache{
//		pendingIDs: make(map[uint]*sync.WaitGroup),
//	}
//}
//
//// GetUser retrieves user information either from the cache or the database using sync.Map.
//func (uc *UserCache) GetUser(userID uint) (*database.User, error) {
//	// Check if the user is in the cache
//	if user, ok := uc.cache.Load(userID); ok {
//		log.Printf("User with ID: %d found in cache", userID)
//		return user.(*database.User), nil
//	}
//
//	// Check if another goroutine is fetching the user
//	uc.mu.Lock()
//	wg, pending := uc.pendingIDs[userID]
//	if pending {
//		uc.mu.Unlock()
//		wg.Wait()
//		return uc.GetUser(userID) // Retry after completion
//	}
//
//	// Create a WaitGroup for the pending request
//	wg = &sync.WaitGroup{}
//	wg.Add(1)
//	uc.pendingIDs[userID] = wg
//	uc.mu.Unlock()
//
//	// Fetch the user from the database
//	log.Printf("User with ID: %d not found in cache, fetching from the database", userID)
//	user, err := database.GetUserByID(userID)
//	if err != nil {
//		uc.mu.Lock()
//		delete(uc.pendingIDs, userID)
//		uc.mu.Unlock()
//		wg.Done() // Decrease the number of pending requests on error
//		return nil, err
//	}
//
//	// Add the user to the cache after successful database retrieval
//	uc.cache.Store(userID, user)
//	uc.mu.Lock()
//	delete(uc.pendingIDs, userID)
//	wg.Done() // Decrease the number of pending requests on success
//	uc.mu.Unlock()
//
//	return user, nil
//}

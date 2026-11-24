package services

import (
	"encoding/json"
	"errors"
	"go-microservice/models"
	"sort"
	"sync"
	"sync/atomic"
)

type UserService struct {
	mu     sync.RWMutex
	users  map[int]models.User
	nextID int
	cache  atomic.Value
}

func NewUserService() *UserService {
	s := &UserService{
		users:  make(map[int]models.User),
		nextID: 1,
	}
	s.rebuildCacheLocked()
	return s
}

func (s *UserService) rebuildCacheLocked() {
	users := make([]models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	body, err := json.Marshal(users)
	if err != nil {
		body = []byte("[]")
	}

	s.cache.Store(append(body, '\n'))
}

func (s *UserService) GetAllJSON() []byte {
	return s.cache.Load().([]byte)
}

func (s *UserService) GetAll() []models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.User, 0, len(s.users))
	for _, user := range s.users {
		result = append(result, user)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

func (s *UserService) GetByID(id int) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return models.User{}, errors.New("user not found")
	}

	return user, nil
}

func (s *UserService) Create(user models.User) models.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user.ID = s.nextID
	s.nextID++
	s.users[user.ID] = user
	s.rebuildCacheLocked()

	return user
}

func (s *UserService) Update(id int, updated models.User) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return models.User{}, errors.New("user not found")
	}

	updated.ID = id
	s.users[id] = updated
	s.rebuildCacheLocked()

	return updated, nil
}

func (s *UserService) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return errors.New("user not found")
	}

	delete(s.users, id)
	s.rebuildCacheLocked()

	return nil
}

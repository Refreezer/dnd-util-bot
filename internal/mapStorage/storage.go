package mapStorage

import (
	"fmt"
	"github.com/Refreezer/dnd-util-bot/api"
	"sync"
)

type MapStorage struct {
	rwMutex          *sync.RWMutex
	userNameToUserId map[string]int64
	userIdToBalance  map[int64]uint
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		rwMutex:          new(sync.RWMutex),
		userNameToUserId: make(map[string]int64),
		userIdToBalance:  make(map[int64]uint),
	}
}

func (m *MapStorage) MoveMoneyFromUserToUser(fromId int64, toId int64, amount uint) error {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	fromBalance, ok := m.userIdToBalance[fromId]
	if !ok {
		return api.ErrorNotRegistered
	}

	if fromBalance < amount {
		return api.ErrorInsufficientMoney
	}

	toBalance, ok := m.userIdToBalance[toId]
	if !ok {
		return api.ErrorNotRegistered
	}

	m.userIdToBalance[fromId] = fromBalance - amount
	m.userIdToBalance[toId] = amount + toBalance
	return nil
}

func (m *MapStorage) SetUserBalance(userId int64, amount uint) error {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	m.userIdToBalance[userId] = amount
	return nil
}

func (m *MapStorage) GetUserBalance(userId int64) (uint, error) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	balance, ok := m.userIdToBalance[userId]
	if !ok {
		return 0, fmt.Errorf("no wallet for %d", userId)
	}

	return balance, nil
}

func (m *MapStorage) GetIdByUserName(userName string) (userId int64, ok bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	userId, ok = m.userNameToUserId[userName]
	return userId, ok
}

func (m *MapStorage) SaveUserNameToUserIdMapping(userName string, userId int64) error {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	m.userNameToUserId[userName] = userId
	return nil
}

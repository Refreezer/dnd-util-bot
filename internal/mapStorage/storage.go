package mapStorage

import (
	"fmt"
	"github.com/Refreezer/dnd-util-bot/api"
	"sync"
)

type balanceBucketKey struct {
	chatId int64
	userId int64
}

type MapStorage struct {
	rwMutex               *sync.RWMutex
	userNameToUserId      map[string]int64
	chatIdUserIdToBalance map[balanceBucketKey]uint
}

func (m *MapStorage) IsRegistered(chatId int64, userId int64) (bool, error) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	_, ok := m.chatIdUserIdToBalance[balanceBucketKey{chatId, userId}]
	return ok, nil
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		rwMutex:               new(sync.RWMutex),
		userNameToUserId:      make(map[string]int64),
		chatIdUserIdToBalance: make(map[balanceBucketKey]uint),
	}
}

func (m *MapStorage) MoveMoneyFromUserToUser(chatId int64, fromId int64, toId int64, amount uint) error {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	fromKey := balanceBucketKey{chatId, fromId}
	fromBalance, ok := m.chatIdUserIdToBalance[fromKey]
	if !ok {
		return api.ErrorNotRegistered
	}

	if fromBalance < amount {
		return api.ErrorInsufficientMoney
	}

	toKey := balanceBucketKey{chatId, toId}
	toBalance, ok := m.chatIdUserIdToBalance[toKey]
	if !ok {
		return api.ErrorNotRegistered
	}

	m.chatIdUserIdToBalance[fromKey] = fromBalance - amount
	m.chatIdUserIdToBalance[toKey] = amount + toBalance
	return nil
}

func (m *MapStorage) SetUserBalance(chatId int64, userId int64, amount uint) error {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	m.chatIdUserIdToBalance[balanceBucketKey{chatId, userId}] = amount
	return nil
}

func (m *MapStorage) GetUserBalance(chatId int64, userId int64) (uint, error) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	balance, ok := m.chatIdUserIdToBalance[balanceBucketKey{chatId, userId}]
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

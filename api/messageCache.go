package api

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

type MessageCache struct {
	sync.RWMutex
	m map[KeyValue[string, int64]]tgbotapi.Chattable
}

func NewMessageCache() *MessageCache {
	return &MessageCache{m: make(map[KeyValue[string, int64]]tgbotapi.Chattable)}
}

func (mc *MessageCache) Put(messageKey string, chatKey int64, cached tgbotapi.Chattable) {
	mc.Lock()
	defer mc.Unlock()
	mc.m[*NewKeyValue(messageKey, chatKey)] = cached
}

func (mc *MessageCache) Get(messageKey string, chatKey int64) (cached tgbotapi.Chattable, ok bool) {
	mc.RLock()
	defer mc.RUnlock()
	cached, ok = mc.m[*NewKeyValue(messageKey, chatKey)]
	return cached, ok
}

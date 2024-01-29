package boltStorage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Refreezer/dnd-util-bot/api"
	"github.com/boltdb/bolt"
	"github.com/op/go-logging"
	"math"
	"time"
)

var (
	userNameToUserIdBucketKey = []byte("userNameToUserId")
	userIdToBalanceBucketKey  = []byte("userIdToBalance")
	bucketsKeys               = [][]byte{
		userNameToUserIdBucketKey,
		userIdToBalanceBucketKey,
	}
)

type BoltStorage struct {
	db     *bolt.DB
	logger *logging.Logger
}

func (b *BoltStorage) IsRegistered(chatId int64, userId int64) (bool, error) {
	var ok bool
	err := b.db.View(func(tx *bolt.Tx) error {
		balance := tx.Bucket(userIdToBalanceBucketKey).Get(balanceBucketKey(chatId, userId))
		ok = balance != nil
		return nil
	})

	return ok, err
}

func NewBoltStorage(provider api.LoggerProvider, dbName string) (storage *BoltStorage, close func()) {
	logger := provider.MustGetLogger("boltStorage")
	logger.Debugf("Db path is %s", dbName)
	db, err := bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	Init(db, logger)
	if err != nil {
		logger.Fatalf("error while opening db connection %s", err)
	}

	return &BoltStorage{
			db:     db,
			logger: logger,
		},
		func() {
			err := db.Close()
			if err != nil {
				logger.Fatalf("error while closing db connection %s", err)
			}
		}
}

func Init(db *bolt.DB, logger *logging.Logger) {
	err := db.Update(func(tx *bolt.Tx) error {
		for _, key := range bucketsKeys {
			err := initBucket(tx, key)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if err != nil {
			logger.Fatalf("error during db initialization")
		}
	}
}

func initBucket(tx *bolt.Tx, bucketKey []byte) error {
	b := tx.Bucket(bucketKey)
	if b != nil {
		return nil
	}

	_, err := tx.CreateBucket(bucketKey)
	return err
}

func balanceBucketKey(chatId, userId int64) []byte {
	bytes := make([]byte, 0)
	bytes = binary.LittleEndian.AppendUint64(bytes, uint64(chatId))
	return binary.LittleEndian.AppendUint64(bytes, uint64(userId))
}

func int64ToByteArr(value int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(value))
	return b
}

func uintToByteArr(value uint) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(value))
	return b
}

func uintFromByteArr(arr []byte) uint {
	return uint(binary.LittleEndian.Uint32(arr))
}

func int64FromByteArr(arr []byte) int64 {
	return int64(binary.LittleEndian.Uint64(arr))
}

func (b *BoltStorage) MoveMoneyFromUserToUser(chatId int64, fromId int64, toId int64, amount uint) error {
	if toId == fromId {
		return api.ErrorInvalidTransactionParameters
	}

	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userIdToBalanceBucketKey)
		fromKey := balanceBucketKey(chatId, fromId)
		fromBalanceBytes := bucket.Get(fromKey)
		if fromBalanceBytes == nil {
			return fmt.Errorf("error while MoveMoneyFromUserToUser (sender) %w", api.ErrorNotRegistered)
		}

		toKey := balanceBucketKey(chatId, fromId)
		toBalanceBytes := bucket.Get(toKey)
		if toBalanceBytes == nil {
			return fmt.Errorf("error while MoveMoneyFromUserToUser (recepient) %w", api.ErrorNotRegistered)
		}

		fromBalance := uintFromByteArr(fromBalanceBytes)
		toBalance := uintFromByteArr(toBalanceBytes)
		if fromBalance < amount {
			return api.ErrorInsufficientMoney
		}

		newFromBalanceBytes := uintToByteArr(fromBalance - amount)
		err := bucket.Put(fromKey, newFromBalanceBytes)
		if err != nil {
			return err
		}

		if amount > math.MaxUint32-toBalance {
			return api.ErrorBalanceOverflow
		}

		err = bucket.Put(toKey, uintToByteArr(amount+toBalance))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		b.logger.Errorf("error while MoveMoneyFromUserToUser %s", err)
	}

	return err
}

func (b *BoltStorage) SetUserBalance(chatId int64, userId int64, amount uint) error {
	if amount < 0 {
		return api.ErrorInsufficientMoney
	}

	if amount > math.MaxUint32 {
		return api.ErrorBalanceOverflow
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userIdToBalanceBucketKey)
		userIdKey := balanceBucketKey(chatId, userId)
		return bucket.Put(userIdKey, uintToByteArr(amount))
	})
}

func (b *BoltStorage) GetUserBalance(chatId int64, userId int64) (uint, error) {
	var balance uint
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userIdToBalanceBucketKey)
		userKey := balanceBucketKey(chatId, userId)
		balanceBytes := bucket.Get(userKey)
		if balanceBytes == nil {
			return fmt.Errorf("error while GetUserBalance %w", api.ErrorNotRegistered)
		}

		balance = uintFromByteArr(balanceBytes)
		return nil
	})

	if err != nil && !errors.Is(err, api.ErrorNotRegistered) {
		b.logger.Errorf("error while GetUserBalance: %s", err)
	}

	return balance, err
}

func (b *BoltStorage) GetIdByUserName(userName string) (userId int64, ok bool) {
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userNameToUserIdBucketKey)
		userIdBytes := bucket.Get([]byte(userName))
		if userIdBytes == nil {
			return fmt.Errorf("error while GetIdByUserName %w", api.ErrorNotRegistered)
		}

		userId = int64FromByteArr(userIdBytes)
		ok = true
		return nil
	})

	if err != nil && !errors.Is(err, api.ErrorNotRegistered) {
		b.logger.Errorf("error while GetIdByUserName: %s", err)
		return 0, false
	}

	return userId, ok
}

func (b *BoltStorage) SaveUserNameToUserIdMapping(userName string, id int64) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userNameToUserIdBucketKey)
		return bucket.Put([]byte(userName), int64ToByteArr(id))
	})

	if err != nil {
		b.logger.Errorf("error while SaveUserNameToUserIdMapping: %s", err)
	}

	return err
}

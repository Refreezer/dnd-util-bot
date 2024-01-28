package boltStorage

import (
	"encoding/binary"
	"errors"
	"github.com/Refreezer/dnd-util-bot/api"
	"github.com/Refreezer/dnd-util-bot/internal"
	"github.com/boltdb/bolt"
	"github.com/op/go-logging"
	"math"
	"os"
	"time"
)

var (
	DBName = os.Getenv(string(internal.DndUtilDbPath))
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

func (b *BoltStorage) IsRegistered(userId int64) (bool, error) {
	var ok bool
	err := b.db.View(func(tx *bolt.Tx) error {
		balance := tx.Bucket(userIdToBalanceBucketKey).Get(int64ToByteArr(userId))
		ok = balance == nil
		return nil
	})

	return ok, err
}

func NewBoltStorage(provider api.LoggerProvider) (storage *BoltStorage, close func()) {
	logger := provider.MustGetLogger("boltStorage")
	logger.Infof("Db path is %s", DBName)
	db, err := bolt.Open(DBName, 0600, &bolt.Options{Timeout: 1 * time.Second})
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

func (b *BoltStorage) MoveMoneyFromUserToUser(fromId int64, toId int64, amount uint) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userIdToBalanceBucketKey)
		fromIdKey := int64ToByteArr(fromId)
		fromBalanceBytes := bucket.Get(fromIdKey)
		if fromBalanceBytes == nil {
			return api.ErrorNotRegistered
		}

		toIdKey := int64ToByteArr(toId)
		toBalanceBytes := bucket.Get(toIdKey)
		if toBalanceBytes == nil {
			return api.ErrorNotRegistered
		}

		fromBalance := uintFromByteArr(fromBalanceBytes)
		toBalance := uintFromByteArr(toBalanceBytes)
		if fromBalance < amount {
			return api.ErrorInsufficientMoney
		}

		newFromBalanceBytes := uintToByteArr(fromBalance - amount)
		err := bucket.Put(fromIdKey, newFromBalanceBytes)
		if err != nil {
			return err
		}

		if amount > math.MaxUint32-toBalance {
			return api.ErrorBalanceOverflow
		}

		err = bucket.Put(toIdKey, uintToByteArr(amount+toBalance))
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

func (b *BoltStorage) SetUserBalance(userId int64, amount uint) error {
	if amount < 0 {
		return api.ErrorInsufficientMoney
	}

	if amount > math.MaxUint32 {
		return api.ErrorBalanceOverflow
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userIdToBalanceBucketKey)
		userIdKey := int64ToByteArr(userId)
		return bucket.Put(userIdKey, uintToByteArr(amount))
	})
}

func (b *BoltStorage) GetUserBalance(userId int64) (uint, error) {
	var balance uint
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userIdToBalanceBucketKey)
		userIdKey := int64ToByteArr(userId)
		balanceBytes := bucket.Get(userIdKey)
		if balanceBytes == nil {
			return api.ErrorNotRegistered
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
			return api.ErrorNotRegistered
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

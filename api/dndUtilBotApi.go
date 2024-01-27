package api

type (
	API interface {
		send(amount int64, uid int64)
		getBalance(uid int64)
		dice()
	}
)

package api

import "fmt"

var (
	ErrorInvalidParameters            = fmt.Errorf("inalid command parameters")
	ErrorInvalidIntegerParameter      = fmt.Errorf("invalid integer parameter")
	ErrorInvalidTransactionParameters = fmt.Errorf("invalid transaction parameters")
	ErrorBalanceOverflow              = fmt.Errorf("balance has exceeded uint32")
	ErrorInsufficientMoney            = fmt.Errorf("insufficient pounds")
	ErrorNotRegistered                = fmt.Errorf("not registered error")
)

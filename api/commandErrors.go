package api

import "fmt"

var (
	ErrorInvalidParameters            = fmt.Errorf("inalid command parameters")
	ErrorInvalidIntegerParameter      = fmt.Errorf("invalid integer parameter")
	ErrorInvalidTransactionParameters = fmt.Errorf("invalid transaction parameters")
)

package internal

import "fmt"

var (
	ErrorInsufficientMoney = fmt.Errorf("insufficient pounds")
	ErrorNotRegistered     = fmt.Errorf("not registered error")
)

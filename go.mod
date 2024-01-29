module github.com/Refreezer/dnd-util-bot

go 1.21.0

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
)

require (
	github.com/boltdb/bolt v1.3.1 // indirect
	golang.org/x/sys v0.16.0 // indirect
)

replace github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1 => github.com/Refreezer/telegram-bot-api/v5 v5.0.0-20240108230938-63e5c59035bf

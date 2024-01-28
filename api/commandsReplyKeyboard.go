package api

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var (
	buttonStart                   = tgbotapi.NewKeyboardButton(commandStartLabel)
	buttonThrowDice               = tgbotapi.NewKeyboardButton(commandThrowDiceLabel)
	buttonGetBalance              = tgbotapi.NewKeyboardButton(commandGetBalanceLabel)
	buttonSendMoney               = tgbotapi.NewKeyboardButton(commandSendMoneyLabel)
	buttonGetUserBalance          = tgbotapi.NewKeyboardButton(commandGetUserBalanceLabel)
	buttonSetUserBalance          = tgbotapi.NewKeyboardButton(commandSetUserBalanceLabel)
	buttonMoveMoneyFromUserToUser = tgbotapi.NewKeyboardButton(commandMoveMoneyFromUserToUserLabel)
)

func mainMenu(api *dndUtilBotApi, upd *tgbotapi.Update) *tgbotapi.ReplyKeyboardMarkup {
	userName := upd.SentFrom().UserName
	isRegistered := api.isUserRegistered(userName)
	var kb tgbotapi.ReplyKeyboardMarkup
	if upd.FromChat().Type == ChatTypePrivate && !isRegistered {
		kb = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(buttonStart),
		)

		return &kb
	}

	var firstRow []tgbotapi.KeyboardButton
	if !isRegistered {
		firstRow = tgbotapi.NewKeyboardButtonRow(buttonStart, buttonThrowDice)
	} else {
		firstRow = tgbotapi.NewKeyboardButtonRow(buttonThrowDice)
	}
	kb = tgbotapi.NewReplyKeyboard(
		firstRow,
		tgbotapi.NewKeyboardButtonRow(buttonGetBalance, buttonSendMoney),
	)

	return &kb
}

package api

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var (
	buttonStart           = tgbotapi.NewKeyboardButton(commandStartLabel)
	buttonThrowDice       = tgbotapi.NewKeyboardButton(commandThrowDiceLabel)
	buttonGetBalance      = tgbotapi.NewKeyboardButton(commandGetBalanceLabel)
	buttonSendMoneyPrompt = tgbotapi.NewKeyboardButton(commandSendMoneyPromptLabel)
	//buttonGetUserBalance  = tgbotapi.NewKeyboardButton(commandGetUserBalanceLabel)
	//buttonSetUserBalance          = tgbotapi.NewKeyboardButton(commandSetUserBalanceLabel)
	//buttonMoveMoneyFromUserToUser = tgbotapi.NewKeyboardButton(commandMoveMoneyFromUserToUserLabel)
)

func mainMenu(api *dndUtilBotApi, upd *tgbotapi.Update) *tgbotapi.ReplyKeyboardMarkup {
	var kb tgbotapi.ReplyKeyboardMarkup
	if upd.FromChat().Type == ChatTypePrivate {
		kb = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(buttonStart),
		)

		return &kb
	}

	kb = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(buttonThrowDice),
		tgbotapi.NewKeyboardButtonRow(buttonGetBalance, buttonSendMoneyPrompt),
	)

	return &kb
}

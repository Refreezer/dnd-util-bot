package api

import "fmt"

const (
	commandMoveMoneyFromUserToUserLabel = "(Админ) Перевод между игроками"
	commandSetUserBalanceLabel          = "(Админ) Задать баланс игрока"
	commandGetUserBalanceLabel          = "(Админ) Получить баланс юзера"
	commandThrowDiceLabel               = "Кинуть кубик"
	commandGetBalanceLabel              = "Получить баланс"
	commandSendMoneyPromptLabel         = "Послать денежку"
	commandStartLabel                   = "Начать"

	usageMoveMoneyFromUserToUser = "%s @sender @recipient amount"
	usageSetUserBalance          = "%s @username amount"
	usageGetUserBalance          = "%s @username"
	usageSendMoney               = "%s @username amount"
)

var (
	commandMoveMoneyFromUserToUser = &command{
		handler:          handlerMoveMoneyFromUserToUser.setReplyMarkup(mainMenu),
		needsAdminRights: true,
		label:            commandMoveMoneyFromUserToUserLabel,
		usage:            fmt.Sprintf(usageMoveMoneyFromUserToUser, addSlash(commandKeyMoveMoneyFromUserToUser)),
	}
	commandSetUserBalance = &command{
		handler:          handlerSetUserBalance.setReplyMarkup(mainMenu),
		needsAdminRights: true,
		label:            commandSetUserBalanceLabel,
		usage:            fmt.Sprintf(usageSetUserBalance, addSlash(commandKeySetUserBalance)),
	}
	commandGetUserBalance = &command{
		handler:          handlerGetUserBalance.setReplyMarkup(mainMenu),
		needsAdminRights: true,
		label:            commandGetUserBalanceLabel,
		usage:            fmt.Sprintf(usageGetUserBalance, addSlash(commandKeyGetUserBalance)),
	}
	commandThrowDice = &command{
		handler: handlerThrowDice.setReplyMarkup(mainMenu).setReplyToMessageID(),
		label:   commandThrowDiceLabel,
	}
	commandGetBalance = &command{
		handler: handlerGetBalance.setReplyMarkup(mainMenu),
		label:   commandGetBalanceLabel,
	}
	commandSendMoneyPrompt = &command{
		handler: handlerSendMoneyPrompt.setReplyMarkup(mainMenu).setReplyToMessageID(),
		label:   commandSendMoneyPromptLabel,
	}

	commandSendMoney = &command{
		handler: handlerSendMoney.setReplyMarkup(mainMenu).setReplyToMessageID(),
		usage:   fmt.Sprintf(usageSendMoney, addSlash(commandKeySendMoney)),
	}
	commandStart = &command{
		handler: handlerStart.setReplyMarkup(mainMenu),
		label:   commandStartLabel,
	}
	commandNotImplemented = &command{
		handler: handlerNotImplemented.setReplyMarkup(mainMenu),
	}
	commandRightsViolation = &command{
		handler: handlerRightsViolation.setReplyMarkup(mainMenu),
	}
	commandCanNotResolve = &command{
		handler: handlerCantResolve.setReplyMarkup(mainMenu),
	}
)

func addSlash(s string) string {
	return fmt.Sprintf("/%s", s)
}

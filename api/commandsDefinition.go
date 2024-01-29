package api

import "fmt"

const (
	commandMoveMoneyFromUserToUserLabel = "(Админ) Перевод между игроками"
	commandSetUserBalanceLabel          = "(Админ) Задать баланс игрока"
	commandGetUserBalanceLabel          = "(Админ) Получить баланс юзера"
	commandThrowDiceLabel               = "Бросок d20"
	commandGetBalanceLabel              = "Мой кошель"
	commandSendMoneyPromptLabel         = "Передать монеты"
	commandStartLabel                   = "Начать"
	commandEmptyLabel                   = "-"

	usageMoveMoneyFromUserToUser = "%s @sender @recipient 123"
	usageSetUserBalance          = "%s @username 123"
	usageGetUserBalance          = "%s @username"
	usageSendMoney               = "%s @recipient 123"
)

var (
	commandMoveMoneyFromUserToUser = &command{
		handler:          handlerMoveMoneyFromUserToUser.setReplyToMessageID(),
		needsAdminRights: true,
		label:            commandMoveMoneyFromUserToUserLabel,
		usage:            fmt.Sprintf(usageMoveMoneyFromUserToUser, addSlash(commandKeyMoveMoneyFromUserToUser)),
	}
	commandSetUserBalance = &command{
		handler:          handlerSetUserBalance.setReplyToMessageID(),
		needsAdminRights: true,
		label:            commandSetUserBalanceLabel,
		usage:            fmt.Sprintf(usageSetUserBalance, addSlash(commandKeySetUserBalance)),
	}
	commandGetUserBalance = &command{
		handler:          handlerGetUserBalance.setReplyToMessageID(),
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
		label:   commandEmptyLabel,
	}
	commandStart = &command{
		handler: handlerStart.setReplyMarkup(mainMenu),
		label:   commandStartLabel,
	}
	commandNotImplemented = &command{
		handler: handlerNotImplemented.setReplyMarkup(mainMenu),
		label:   commandEmptyLabel,
	}
	commandRightsViolation = &command{
		handler: handlerRightsViolation.setReplyMarkup(mainMenu),
		label:   commandEmptyLabel,
	}
	commandCanNotResolve = &command{
		handler: handlerCantResolve.setReplyMarkup(mainMenu),
	}
)

func addSlash(s string) string {
	return fmt.Sprintf("/%s", s)
}

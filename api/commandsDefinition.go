package api

const (
	commandMoveMoneyFromUserToUserLabel = "(Админ) Перевод между игроками"
	commandSetUserBalanceLabel          = "(Админ) Задать баланс игрока"
	commandGetUserBalanceLabel          = "(Админ) Получить баланс юзера"
	commandThrowDiceLabel               = "Кинуть кубик"
	commandGetBalanceLabel              = "Получить баланс"
	commandSendMoneyLabel               = "Послать денежку"
	commandStartLabel                   = "Начать"
)

var (
	commandMoveMoneyFromUserToUser = &command{
		handler:          handlerMoveMoneyFromUserToUser.setReplyMarkup(mainMenu),
		needsAdminRights: true,
		label:            "(Админ) Перевод между игроками",
	}
	commandSetUserBalance = &command{
		handler:          handlerSetUserBalance.setReplyMarkup(mainMenu),
		needsAdminRights: true,
		label:            "(Админ) Задать баланс игрока",
	}
	commandGetUserBalance = &command{
		handler:          handlerGetUserBalance.setReplyMarkup(mainMenu),
		needsAdminRights: true,
		label:            "(Админ) Получить баланс юзера",
	}
	commandThrowDice = &command{
		handler: handlerThrowDice.setReplyMarkup(mainMenu).setReplyToMessageID(),
		label:   "Кинуть кубик",
	}
	commandGetBalance = &command{
		handler: handlerGetBalance.setReplyMarkup(mainMenu),
		label:   "Получить баланс",
	}
	commandSendMoney = &command{
		handler: handlerSendMoney.setReplyMarkup(mainMenu).setReplyToMessageID(),
		label:   "Послать денежку",
	}
	commandStart = &command{
		handler: handlerStart.setReplyMarkup(mainMenu),
		label:   "Начать",
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

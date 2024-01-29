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

	usageMoveMoneyFromUserToUser = "`%s @sender @recipient 123`"
	usageSetUserBalance          = "`%s @username 123`"
	usageGetUserBalance          = "`%s @username`"
	usageSendMoney               = "`%s @recipient 123`"
)

var (
	groupCommandsMap = map[string]*command{
		commandKeyStart:                   commandStart,
		commandKeySendMoneyPrompt:         commandSendMoneyPrompt,
		commandKeySendMoney:               commandSendMoney,
		commandKeyGetBalance:              commandGetBalance,
		commandKeyThrowDice:               commandThrowDice,
		commandKeyGetUserBalance:          commandGetUserBalance,
		commandKeySetUserBalance:          commandSetUserBalance,
		commandKeyMoveMoneyFromUserToUser: commandMoveMoneyFromUserToUser,
		commandKeyHelp:                    commandHelp,
	}

	privateCommandsMap = map[string]*command{
		commandKeyStart: commandStart,
		commandKeyHelp:  commandHelp,
	}

	chatTypeToCommandMap = map[string]map[string]*command{
		ChatTypeGroup:      groupCommandsMap,
		ChatTypeSuperGroup: groupCommandsMap,
		ChatTypePrivate:    privateCommandsMap,
	}
)

var (
	commandMoveMoneyFromUserToUser = &command{
		handler:          handlerMoveMoneyFromUserToUser.setReplyToMessageID(),
		needsAdminRights: true,
		label:            commandMoveMoneyFromUserToUserLabel,
		usage:            fmt.Sprintf(usageMoveMoneyFromUserToUser, addSlash(commandKeyMoveMoneyFromUserToUser)),
		description:      "перевести деньги от игрока к игроку",
	}
	commandSetUserBalance = &command{
		handler:          handlerSetUserBalance.setReplyToMessageID(),
		needsAdminRights: true,
		label:            commandSetUserBalanceLabel,
		usage:            fmt.Sprintf(usageSetUserBalance, addSlash(commandKeySetUserBalance)),
		description:      "задать баланс игрока",
	}
	commandGetUserBalance = &command{
		handler:          handlerGetUserBalance.setReplyToMessageID(),
		needsAdminRights: true,
		label:            commandGetUserBalanceLabel,
		usage:            fmt.Sprintf(usageGetUserBalance, addSlash(commandKeyGetUserBalance)),
		description:      "посмотреть баланс игрока",
	}
	commandThrowDice = &command{
		handler:     handlerThrowDice.setReplyMarkup(mainMenu).setReplyToMessageID(),
		label:       commandThrowDiceLabel,
		description: "бросок d20",
	}
	commandGetBalance = &command{
		handler:     handlerGetBalance.setReplyMarkup(mainMenu),
		label:       commandGetBalanceLabel,
		description: "посмотреть свой баланс",
	}
	commandSendMoneyPrompt = &command{
		handler:     handlerSendMoneyPrompt.setReplyMarkup(mainMenu).setReplyToMessageID(),
		label:       commandSendMoneyPromptLabel,
		description: "посмотреть команду для перевода",
	}
	commandSendMoney = &command{
		handler:     handlerSendMoney.setReplyMarkup(mainMenu).setReplyToMessageID(),
		usage:       fmt.Sprintf(usageSendMoney, addSlash(commandKeySendMoney)),
		label:       commandEmptyLabel,
		description: "перевести деньги игроку",
	}
	commandStart = &command{
		handler: handlerStart.setReplyMarkup(mainMenu),
		label:   commandStartLabel,
	}
	commandHelp = &command{
		handler:          handlerHelp,
		needsAdminRights: true,
		description:      "посмотреть команды",
		label:            commandEmptyLabel,
	}

	// service commands
	commandNotImplemented = &command{
		handler: handlerNotImplemented.setReplyMarkup(mainMenu).setReplyToMessageID(),
		label:   commandEmptyLabel,
	}
	commandRightsViolation = &command{
		handler: handlerRightsViolation.setReplyMarkup(mainMenu).setReplyToMessageID(),
		label:   commandEmptyLabel,
	}
	commandCanNotResolve = &command{
		handler: handlerCantResolve.setReplyMarkup(mainMenu),
	}
)

func addSlash(s string) string {
	return fmt.Sprintf("/%s", s)
}

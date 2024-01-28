package api

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	commandKeyStart                   = "start"
	commandKeySendMoney               = "send_money"
	commandKeyGetBalance              = "get_balance"
	commandKeyThrowDice               = "throw_dice"
	commandKeyGetUserBalance          = "get_user_balance"
	commandKeySetUserBalance          = "set_user_balance"
	commandKeyMoveMoneyFromUserToUser = "move_money_from_user_to_user"
)

var (
	groupCommandsMap = map[string]*command{
		commandKeyStart:                   {handlerStart, false},
		commandKeySendMoney:               {handlerSendMoney, false},
		commandKeyGetBalance:              {handlerGetBalance, false},
		commandKeyThrowDice:               {handlerThrowDice, false},
		commandKeyGetUserBalance:          {handlerGetUserBalance, true},
		commandKeySetUserBalance:          {handlerSetUserBalance, true},
		commandKeyMoveMoneyFromUserToUser: {handlerMoveMoneyFromUserToUser, true},
	}

	privateCommandsMap = map[string]*command{
		commandKeyStart: {handlerStart, false},
	}

	chatTypeToCommandMap = map[string]map[string]*command{
		ChatTypeGroup:      groupCommandsMap,
		ChatTypeSuperGroup: groupCommandsMap,
		ChatTypePrivate:    privateCommandsMap,
	}

	commandNotImplemented  = &command{handlerNotImplemented, false}
	commandRightsViolation = &command{handlerRightsViolation, false}
	commandCanNotResolve   = &command{handlerCantResolve, false}
)

type (
	commands struct {
		api *dndUtilBotApi
	}

	command struct {
		handler          func(api *dndUtilBotApi, upd *tgbotapi.Update) error
		needsAdminRights bool
	}

	builtUpCommand struct {
		handler func(upd *tgbotapi.Update) error
	}
)

func (c *command) isImplemented() bool {
	return c != commandNotImplemented
}

func (c *command) isAuthorized() bool {
	return c != commandRightsViolation
}

func newCommands(api *dndUtilBotApi) *commands {
	return &commands{
		api: api,
	}
}

func (c *commands) resolve(upd *tgbotapi.Update) *command {
	commandKey := upd.Message.Command()
	commandsMap, ok := chatTypeToCommandMap[upd.FromChat().Type]
	if !ok {
		return commandNotImplemented
	}

	cmd, ok := commandsMap[commandKey]
	if !ok {
		return commandNotImplemented
	}

	if cmd.needsAdminRights {
		isRelatedMemberAdmin, err := c.api.isRelatedMemberAdmin(upd)
		if err != nil {
			return commandCanNotResolve
		}

		if !isRelatedMemberAdmin {
			return commandRightsViolation
		}
	}

	return cmd
}

func (c *command) build(api *dndUtilBotApi) *builtUpCommand {
	return c.newBuiltUpCommand(api)
}

func (buc *builtUpCommand) execute(upd *tgbotapi.Update) error {
	err := buc.handler(upd)
	if err != nil {
		return err
	}

	return nil
}

func (c *command) newBuiltUpCommand(api *dndUtilBotApi) *builtUpCommand {
	return &builtUpCommand{
		handler: func(upd *tgbotapi.Update) error {
			err := c.handler(api, upd)
			if err != nil {
				api.logger.Errorf("error on executing command handler %s", err)
			}
			return err
		}}
}

func handlerMoveMoneyFromUserToUser(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.moveMoneyFromUserToUser(upd)
}

func handlerSetUserBalance(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.setUserBalance(upd)
}

func handlerGetUserBalance(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.getUserBalance(upd)
}

func handlerThrowDice(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.throwDice(upd)
}

func handlerGetBalance(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.getBalance(upd)
}

func handlerSendMoney(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.sendMoney(upd)
}

func handlerNotImplemented(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.notImplemented(upd)
}

func handlerRightsViolation(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.rightsViolation(upd)
}

func handlerCantResolve(_ *dndUtilBotApi, _ *tgbotapi.Update) error {
	return nil
}

func handlerStart(api *dndUtilBotApi, upd *tgbotapi.Update) error {
	return api.start(upd)
}

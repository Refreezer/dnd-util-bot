package api

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	commandKeyStart                   = "start"
	commandKeySendMoney               = "send_money"
	commandKeySendMoneyPrompt         = "send_money_prompt"
	commandKeyGetBalance              = "get_balance"
	commandKeyThrowDice               = "throw_dice"
	commandKeyGetUserBalance          = "get_user_balance"
	commandKeySetUserBalance          = "set_user_balance"
	commandKeyMoveMoneyFromUserToUser = "move_money_from_user_to_user"
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
	}

	privateCommandsMap = map[string]*command{
		commandKeyStart: commandStart,
	}

	chatTypeToCommandMap = map[string]map[string]*command{
		ChatTypeGroup:      groupCommandsMap,
		ChatTypeSuperGroup: groupCommandsMap,
		ChatTypePrivate:    privateCommandsMap,
	}
)

type (
	commands struct {
		api *dndUtilBotApi
	}

	commandHandler func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error)
	command        struct {
		handler          commandHandler
		needsAdminRights bool
		label            string
		usage            string
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
		cmd, ok = c.findCommandByLabel(upd.Message.Text, commandsMap)
		if !ok {
			return commandCanNotResolve
		}
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

func (c *commands) findCommandByLabel(messageText string, commandsMap map[string]*command) (*command, bool) {
	for _, command := range commandsMap {
		if command.label == messageText {
			return command, true
		}
	}
	return nil, false
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
			chatable, err := c.handler(api, upd)
			if err != nil {
				api.logger.Errorf("error on executing command handler: %s", err)
			}

			if chatable != nil {
				api.sendToChat(chatable)
			}

			return err
		}}
}

package api

import (
	"cmp"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"slices"
	"strings"
	"text/tabwriter"
)

var (
	administrativeSeparator = []string{
		"", administrativeCommandsSeparatorString, "",
	}
)

const (
	commandKeyHelp                    = "help"
	commandKeyStart                   = "start"
	commandKeySendMoney               = "send"
	commandKeySendMoneyPrompt         = "send_prompt"
	commandKeyGetBalance              = "balance"
	commandKeyThrowDice               = "dice"
	commandKeyGetUserBalance          = "get_balance"
	commandKeySetUserBalance          = "set_balance"
	commandKeyMoveMoneyFromUserToUser = "transaction"
)

type (
	commands struct {
		api          *dndUtilBotApi
		commandMap   map[string]map[string]*command
		messageCache *MessageCache
	}

	commandHandler func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error)
	command        struct {
		commandKey       string
		handler          commandHandler
		needsAdminRights bool
		label            string
		usage            string
		description      string
		messageCache     *MessageCache
	}

	builtUpCommand struct {
		handler func(upd *tgbotapi.Update) error
	}

	KeyValue[TKey any, TValue any] struct {
		key   TKey
		value TValue
	}
)

func NewKeyValue[TKey any, TValue any](key TKey, value TValue) *KeyValue[TKey, TValue] {
	return &KeyValue[TKey, TValue]{key: key, value: value}
}

func newCommands(api *dndUtilBotApi, commandMap map[string]map[string]*command) *commands {
	newCommands := commands{
		api:          api,
		commandMap:   commandMap,
		messageCache: NewMessageCache(),
	}

	initialize(&newCommands)
	return &newCommands
}

func initialize(newCommands *commands) {
	for _, chatCommandMap := range newCommands.commandMap {
		for commandKey, command := range chatCommandMap {
			command.commandKey = commandKey
			command.messageCache = newCommands.messageCache
		}
	}
}

func (c *commands) findCommandByLabel(messageText string, commandsMap map[string]*command) (*command, bool) {
	for _, command := range commandsMap {
		if command.label == messageText {
			return command, true
		}
	}
	return nil, false
}

func (c *commands) printHelp(upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
	return c.newHelpMessage(upd), nil
}

func (c *commands) newHelpMessage(upd *tgbotapi.Update) *tgbotapi.MessageConfig {
	chat := upd.FromChat()
	commandsMap, ok := c.commandMap[chat.Type]
	notImplMessage, _ := notImplemented(upd)
	if !ok {
		return notImplMessage
	}

	commandUsagesWithRights := make([]*KeyValue[string, bool], 0, len(commandsMap))
	for commandKey, command := range commandsMap {
		var usage string
		if command.usage == "" {
			usage = fmt.Sprintf("`/%s`", commandKey)
		} else {
			usage = command.usage
		}

		if command.description != "" {
			usage = fmt.Sprintf("%s \\- %s", usage, command.description)
		}

		commandUsagesWithRights = append(
			commandUsagesWithRights,
			NewKeyValue(usage, command.needsAdminRights),
		)
	}

	slices.SortStableFunc(commandUsagesWithRights, func(a, b *KeyValue[string, bool]) int {
		return cmp.Compare(len(a.key), len(b.key))
	})

	slices.SortStableFunc(commandUsagesWithRights, func(a, b *KeyValue[string, bool]) int {
		return cmpBool(a.value, b.value)
	})

	commandUsagesWithRights = insertAdministrativeSeparator(commandUsagesWithRights)
	messageText := getUsageMessageText(commandUsagesWithRights)
	msg := plainMessage(chat.ID, messageText)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	c.messageCache.Put(commandKeyHelp, chat.ID, msg)
	return msg
}

func insertAdministrativeSeparator(commandUsagesWithRights []*KeyValue[string, bool]) []*KeyValue[string, bool] {
	firstAdminIdx := 0
	for i, val := range commandUsagesWithRights {
		if val.value {
			firstAdminIdx = i
			break
		}

		i += 1
	}

	for _, sep := range administrativeSeparator {
		commandUsagesWithRights = slices.Insert(
			commandUsagesWithRights,
			firstAdminIdx,
			&KeyValue[string, bool]{key: sep},
		)
	}

	return commandUsagesWithRights
}

func cmpBool(aBool bool, bBool bool) int {
	if aBool == bBool {
		return 0
	}

	if aBool && !bBool {
		return 1
	}

	return -1
}

func getUsageMessageText(commandUsageStrings []*KeyValue[string, bool]) string {
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	for _, s := range commandUsageStrings {
		fmt.Fprintln(tw, s.key)
	}

	tw.Flush()
	messageText := sb.String()
	return messageText
}

func (c *commands) Resolve(upd *tgbotapi.Update) *command {
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

func (c *command) Build(api *dndUtilBotApi) *builtUpCommand {
	return c.newBuiltUpCommand(api)
}

func (buc *builtUpCommand) Execute(upd *tgbotapi.Update) error {
	err := buc.handler(upd)
	if err != nil {
		return err
	}

	return nil
}

func (c *command) newBuiltUpCommand(api *dndUtilBotApi) *builtUpCommand {
	return &builtUpCommand{
		handler: func(upd *tgbotapi.Update) error {
			cached, ok := c.messageCache.Get(c.commandKey, upd.FromChat().ID)
			if ok {
				api.sendToChat(cached)
				return nil
			}

			chattable, err := c.handler(api, upd)
			if err != nil {
				api.logger.Errorf("error on executing command handler: %s", err)
			}

			if chattable != nil {
				api.sendToChat(chattable)
			}

			return err
		}}
}

func (c *command) isImplemented() bool {
	return c != commandNotImplemented
}

func (c *command) isAuthorized() bool {
	return c != commandRightsViolation
}

func notImplemented(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	return newMessageNotImplemented(upd), nil
}

func rightsViolation(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	return newMessageRightsViolation(upd), nil
}

func newMessageRightsViolation(upd *tgbotapi.Update) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, messageRejectedRightsViolation)
	return &msg
}

func newMessageNotImplemented(upd *tgbotapi.Update) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, messageNotImplemented)
	return &msg
}

package api

import (
	"context"
	"fmt"
	"github.com/Refreezer/dnd-util-bot/api/listener"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/op/go-logging"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	ChatTypeGroup                 = "group"
	ChatTypeSuperGroup            = "supergroup"
	ChatTypePrivate               = "private"
	ChatChannel                   = "channel"
	ChatMemberCreator             = "creator"
	ChatMemberStatusAdministrator = "administrator"
)

type (
	DndUtilApi interface {
		listener.UpdateHandler
	}

	Storage interface {
		MoveMoneyFromUserToUser(fromId int64, toId int64, amount uint) error
		SetUserBalance(userId int64, amount uint) error
		GetUserBalance(userId int64) (uint, error)
		GetIdByUserName(userName string) (userId int64, ok bool)
		SaveUserNameToUserIdMapping(name string, id int64) error
	}

	LoggerProvider interface {
		MustGetLogger(moduleName string) *logging.Logger
	}

	dndUtilBotApi struct {
		tgBotApi   *tgbotapi.BotAPI
		logger     *logging.Logger
		commands   *commands
		storage    Storage
		randomizer *rand.Rand
		botName    string
	}
)

func NewDndUtilApi(
	tgBotApi *tgbotapi.BotAPI,
	loggerProvider LoggerProvider,
	storage Storage,
) DndUtilApi {
	return newDndUtilApi(tgBotApi, loggerProvider, storage)
}

func newDndUtilApi(
	tgBotApi *tgbotapi.BotAPI,
	loggerProvider LoggerProvider,
	storage Storage,
) *dndUtilBotApi {
	api := &dndUtilBotApi{
		tgBotApi:   tgBotApi,
		logger:     loggerProvider.MustGetLogger("dndUtilBotApi"),
		storage:    storage,
		randomizer: rand.New(rand.NewSource(time.Now().Unix())),
		botName:    "DnDUtilTest_bot",
	}

	api.commands = newCommands(api)
	return api
}

func (api *dndUtilBotApi) Handle(ctx context.Context, upd *tgbotapi.Update) {
	if upd.Message != nil {
		api.handleMessage(upd)
	}
}

func (api *dndUtilBotApi) handleMessage(upd *tgbotapi.Update) {
	if upd.Message == nil || upd.Message.Chat == nil || upd.Message.From == nil || upd.Message.Text == "" {
		return
	}

	from := upd.SentFrom()
	api.registerWallet(from)
	switch upd.Message.Chat.Type {
	case ChatTypeGroup:
		api.executeCommand(upd)
		break
	case ChatTypeSuperGroup:
		api.executeCommand(upd)
		break
	case ChatTypePrivate:
		api.executeCommand(upd)
		break
	case ChatChannel: // not supported
	default:
		break
	}
}

func (api *dndUtilBotApi) registerWallet(from *tgbotapi.User) {
	if _, ok := api.storage.GetIdByUserName(from.UserName); ok {
		return
	}

	err := api.storage.SaveUserNameToUserIdMapping(from.UserName, from.ID)
	if err != nil {
		api.logger.Errorf("couldn't save user id mapping for %v", from)
	}

	err = api.storage.SetUserBalance(from.ID, 0)
	if err != nil {
		api.logger.Errorf("couldn't set balance for %v", from)
	}
}

func (api *dndUtilBotApi) executeCommand(upd *tgbotapi.Update) {
	err := api.commands.resolve(upd).build(api).execute(upd)
	if err != nil {
		api.logger.Errorf("couldn't execute command %s", err)
	}
}

func (api *dndUtilBotApi) replyWithRightsViolation(upd *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, messageRejectedRightsViolation)
	api.replyWithMessage(upd, &msg)
}

func (api *dndUtilBotApi) replyWithNotImplemented(upd *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, messageNotImplemented)
	api.replyWithMessage(upd, &msg)
}

func (api *dndUtilBotApi) replyWithMessage(upd *tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	msg.ReplyToMessageID = upd.Message.MessageID
	_, err := api.tgBotApi.Send(msg)
	if err != nil {
		api.logger.Errorf("can't reply with message %s error: %s", msg, err)
		return
	}
}

func (api *dndUtilBotApi) isRelatedMemberAdmin(upd *tgbotapi.Update) (bool, error) {
	member, err := api.getMember(upd.Message.Chat.ID, upd.Message.From.ID)
	if err != nil {
		return false, err
	}

	return isAdmin(&member), nil
}

func isAdmin(member *tgbotapi.ChatMember) bool {
	return member.Status == ChatMemberStatusAdministrator || member.Status == ChatMemberCreator
}

func (api *dndUtilBotApi) getMember(chatID int64, userID int64) (tgbotapi.ChatMember, error) {
	member, err := api.tgBotApi.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	})

	return member, err
}

func (api *dndUtilBotApi) userIdByUserNameAndReplyIfCant(userName string, upd *tgbotapi.Update) (int64, bool) {
	sanitizedUserName := strings.Replace(userName, "@", "", 1)
	uid, ok := api.storage.GetIdByUserName(sanitizedUserName)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, userName),
		)

		api.replyWithMessage(upd, &msg)
	}

	return uid, ok
}

func (api *dndUtilBotApi) moveMoneyFromUserToUser(upd *tgbotapi.Update) error {
	params := api.getParams(upd.Message.Text)
	if len(params) < 4 {
		return fmt.Errorf("inalid MoveMoneyFromUserToUser command parameters")
	}

	amount, err := strconv.Atoi(params[3])
	if err != nil {
		return fmt.Errorf("can't convert money amount toId integer %w", err)
	}

	fromId, ok := api.userIdByUserNameAndReplyIfCant(params[1], upd)
	if !ok {
		return nil
	}

	toId, ok := api.userIdByUserNameAndReplyIfCant(params[2], upd)
	if !ok {
		return nil
	}

	if fromId == toId {
		return fmt.Errorf("can't make transaction from sender acc to itself")
	}

	err = api.storage.MoveMoneyFromUserToUser(fromId, toId, uint(amount))
	if err != nil {
		return fmt.Errorf("error during MoveMoneyFromUserToUser %w", err)
	}

	return nil
}

func (api *dndUtilBotApi) getParams(text string) []string {
	params := strings.Split(text, " ")
	params = slices.DeleteFunc(params, func(s string) bool {
		return strings.HasSuffix(s, api.botName)
	})

	return params
}

func (api *dndUtilBotApi) setUserBalance(upd *tgbotapi.Update) error {
	params := api.getParams(upd.Message.Text)
	if len(params) < 3 {
		return fmt.Errorf("inalid setUserBalance command parameters")
	}

	amount, err := strconv.Atoi(params[2])
	if err != nil {
		return fmt.Errorf("can't convert money amount to integer %w", err)
	}

	userName := params[1]
	userId, ok := api.userIdByUserNameAndReplyIfCant(userName, upd)
	if !ok {
		return nil
	}

	err = api.storage.SetUserBalance(userId, uint(amount))
	if err != nil {
		return fmt.Errorf("error during setUserBalance %w", err)
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, fmt.Sprintf(messageSetUserBalanceSuccess, userName, amount))
	api.replyWithMessage(upd, &msg)
	return err
}

func (api *dndUtilBotApi) getUserBalance(upd *tgbotapi.Update) error {
	params := api.getParams(upd.Message.Text)
	if len(params) < 2 {
		return fmt.Errorf("inalid GetUserBalance command parameters")
	}

	userName := params[1]
	userId, ok := api.userIdByUserNameAndReplyIfCant(userName, upd)
	if !ok {
		return nil
	}

	balance, err := api.storage.GetUserBalance(userId)
	if err != nil {
		return fmt.Errorf("error during getting balance from storage %w", err)
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, fmt.Sprintf(messageGetUserBalanceSuccess, userName, balance))
	api.replyWithMessage(upd, &msg)
	return nil
}

func (api *dndUtilBotApi) throwDice(upd *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(
		upd.Message.Chat.ID,
		fmt.Sprintf(messageThrowDice, api.randomizer.Intn(20)+1),
	)

	api.replyWithMessage(upd, &msg)
	return nil
}

func (api *dndUtilBotApi) getBalance(upd *tgbotapi.Update) error {
	balance, err := api.storage.GetUserBalance(upd.SentFrom().ID)
	if err != nil {
		return fmt.Errorf("error during getBalance from storage %w", err)
	}

	msg := tgbotapi.NewMessage(
		upd.Message.Chat.ID,
		fmt.Sprintf(messageGetUserBalanceSuccess, upd.SentFrom().UserName, balance),
	)

	api.replyWithMessage(upd, &msg)
	return nil
}

func (api *dndUtilBotApi) sendMoney(upd *tgbotapi.Update) error {
	params := api.getParams(upd.Message.Text)
	if len(params) < 3 {
		return fmt.Errorf("inalid MoveMoneyFromUserToUser command parameters")
	}

	amount, err := strconv.Atoi(params[2])
	if err != nil {
		return fmt.Errorf("can't convert money amount toId integer %w", err)
	}

	fromId := upd.SentFrom().ID
	toId, ok := api.userIdByUserNameAndReplyIfCant(params[1], upd)
	if !ok {
		return nil
	}

	err = api.storage.MoveMoneyFromUserToUser(fromId, toId, uint(amount))
	if err != nil {
		return fmt.Errorf("error during MoveMoneyFromUserToUser %w", err)
	}

	return nil
}

func (api *dndUtilBotApi) notImplemented(upd *tgbotapi.Update) error {
	api.replyWithNotImplemented(upd)
	return nil
}

func (api *dndUtilBotApi) rightsViolation(upd *tgbotapi.Update) error {
	api.replyWithRightsViolation(upd)
	return nil
}

func (api *dndUtilBotApi) start(upd *tgbotapi.Update) error {
	chat := upd.FromChat()
	if chat.Type == ChatTypePrivate {
		balance, err := api.storage.GetUserBalance(upd.SentFrom().ID)
		if err != nil {
			return err
		}
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf(messageStart, balance))
		api.replyWithMessage(upd, &msg)
	}

	// pass for groups, because it is excessive
	return nil
}

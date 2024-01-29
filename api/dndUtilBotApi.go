package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/Refreezer/dnd-util-bot/api/listener"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/op/go-logging"
	"math"
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
	D20StickerSetname             = "D20STUMP"
)

var (
	d20NumToEmojiMap = map[int]string{
		1:  "1️⃣",
		2:  "2️⃣",
		3:  "3️⃣",
		4:  "4️⃣",
		5:  "5️⃣",
		6:  "6️⃣",
		7:  "7️⃣",
		8:  "8️⃣",
		9:  "9️⃣",
		10: "🔟",
		11: "🆙",
		12: "🆗",
		13: "🔡",
		14: "🔢",
		15: "🔤",
		16: "🅰️",
		17: "🆎",
		18: "🅱️",
		19: "🆑",
		20: "🆒",
	}
)

type (
	DndUtilApi interface {
		listener.UpdateHandler
	}

	ResourceProvider interface {
		get(uri string) ([]byte, error)
	}

	Storage interface {
		MoveMoneyFromUserToUser(chatId int64, fromId int64, toId int64, amount uint) error
		SetUserBalance(chatId int64, userId int64, amount uint) error
		GetUserBalance(chatId int64, userId int64) (uint, error)
		GetIdByUserName(userName string) (userId int64, ok bool)
		SaveUserNameToUserIdMapping(name string, id int64) error
		IsRegistered(chatId int64, userId int64) (bool, error)
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
		//resourceProvider ResourceProvider
		botName string
	}
)

func NewDndUtilApi(
	tgBotApi *tgbotapi.BotAPI,
	loggerProvider LoggerProvider,
	storage Storage,
	botName string,
) DndUtilApi {
	return newDndUtilApi(
		tgBotApi,
		loggerProvider,
		storage,
		//resourceProvider,
		botName,
	)
}

func newDndUtilApi(
	tgBotApi *tgbotapi.BotAPI,
	loggerProvider LoggerProvider,
	storage Storage,
	botName string,
) *dndUtilBotApi {
	api := &dndUtilBotApi{
		tgBotApi:   tgBotApi,
		logger:     loggerProvider.MustGetLogger("dndUtilBotApi"),
		storage:    storage,
		randomizer: rand.New(rand.NewSource(time.Now().Unix())),
		botName:    botName,
		//resourceProvider: resourceProvider,
	}

	api.commands = newCommands(api, chatTypeToCommandMap)
	return api
}

func (api *dndUtilBotApi) HandleUpdate(ctx context.Context, upd *tgbotapi.Update) {
	if upd.Message != nil {
		api.handleUpdate(upd)
	}
}

func validateUsernameIsNotHidden(upd *tgbotapi.Update) error {
	if upd.SentFrom().UserName == "" {
		return ErrorUsernameHidden
	}

	return nil
}

func (api *dndUtilBotApi) handleUpdate(upd *tgbotapi.Update) {
	if upd.Message == nil || upd.Message.Chat == nil || upd.Message.From == nil {
		return
	}

	from := upd.SentFrom()
	api.registerWalletIfNeeded(upd.FromChat().ID, from)
	switch upd.Message.Chat.Type {
	case ChatTypeGroup, ChatTypeSuperGroup, ChatTypePrivate:
		api.executeCommand(upd)
		break
	case ChatChannel: // not supported
	default:
		break
	}
}

func (api *dndUtilBotApi) registerWalletIfNeeded(chatId int64, from *tgbotapi.User) {
	_, mappingRegistered := api.getIdByUserNameSanitized(from.UserName)
	if !mappingRegistered {
		err := api.storage.SaveUserNameToUserIdMapping(from.UserName, from.ID)
		if err != nil {
			api.logger.Errorf("couldn't save user id mapping for %+v", from)
		}
	}

	isRegistered, err := api.storage.IsRegistered(chatId, from.ID)
	if err != nil {
		api.logger.Errorf("couldn't know if user is registered for chatID=%d username=%s", chatId, from.UserName)
	}

	if isRegistered {
		return
	}

	err = api.storage.SetUserBalance(chatId, from.ID, 0)
	if err != nil {
		api.logger.Errorf("couldn't set balance for %v", from)
	}
}

func (api *dndUtilBotApi) executeCommand(upd *tgbotapi.Update) {
	cmd := api.commands.Resolve(upd)
	err := cmd.Build(api).Execute(upd)
	if err == nil {
		return
	}

	var msg *tgbotapi.MessageConfig
	chatID := upd.FromChat().ID
	messageId := upd.Message.MessageID
	if errors.Is(err, ErrorInvalidParameters) {
		msg = markdownMessage(chatID, messageId, fmt.Sprintf(errorMessageInvalidParametersFormat, cmd.usage))
	} else if errors.Is(err, ErrorInvalidIntegerParameter) {
		msg = markdownMessage(chatID, messageId, errorMessageInvalidIntegerParameter)
	} else if errors.Is(err, ErrorInvalidTransactionParameters) {
		msg = markdownMessage(chatID, messageId, errorMessageInvalidTransactionParameters)
	} else if errors.Is(err, ErrorInsufficientMoney) {
		msg = markdownMessage(chatID, messageId, errorMessageInsufficientPounds)
	} else if errors.Is(err, ErrorBalanceOverflow) {
		msg = markdownMessage(chatID, messageId, errorMessageBalanceOverflow)
	} else if errors.Is(err, ErrorUsernameHidden) {
		msg = markdownMessage(chatID, messageId, messageUsernameHidden)
	}

	if msg == nil {
		return
	}

	_, err = api.tgBotApi.Send(msg)
	if err != nil {
		api.logger.Errorf("couldn't send reply error message %s", err)
	}
}

func markdownMessage(chatId int64, messageId int, text string) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyToMessageID = messageId
	return &msg
}

func (api *dndUtilBotApi) sendToChat(chatable tgbotapi.Chattable) {
	_, err := api.tgBotApi.Send(chatable)
	if err != nil {
		api.logger.Errorf("can't reply with message %s error: %s", chatable, err)
		return
	}
}

func (api *dndUtilBotApi) isRelatedMemberAdmin(upd *tgbotapi.Update) (bool, error) {
	if upd.FromChat().Type == ChatTypePrivate {
		return true, nil
	}

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

func (api *dndUtilBotApi) userIdByUserName(userName string) (int64, bool) {
	uid, ok := api.getIdByUserNameSanitized(userName)
	return uid, ok
}

func (api *dndUtilBotApi) getIdByUserNameSanitized(userName string) (int64, bool) {
	sanitizedUserName := strings.Replace(userName, "@", "", 1)
	uid, ok := api.storage.GetIdByUserName(sanitizedUserName)
	return uid, ok
}

func (api *dndUtilBotApi) Transaction(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	params := api.getParams(upd.Message.Text)
	if len(params) < 4 {
		return nil, ErrorInvalidParameters
	}

	amount, err := strconv.Atoi(params[3])
	if err != nil || amount <= 0 {
		return nil, ErrorInvalidIntegerParameter
	}

	from := params[1]
	fromId, ok := api.userIdByUserName(from)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, from),
		)

		return &msg, nil
	}

	to := params[2]
	toId, ok := api.userIdByUserName(to)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, to),
		)

		return &msg, nil
	}

	if fromId == toId {
		return nil, ErrorInvalidTransactionParameters
	}

	chatId := upd.FromChat().ID
	fromBalance, err := api.storage.GetUserBalance(chatId, fromId)
	if err == nil && fromBalance < uint(amount) {
		return markdownMessage(
			chatId,
			upd.Message.MessageID,
			fmt.Sprintf(
				errorMessageInsufficientPoundsInUserWallet,
				from,
			),
		), nil
	}

	toBalance, err := api.storage.GetUserBalance(chatId, toId)
	if err == nil && toBalance > math.MaxUint32-uint(amount) {
		return markdownMessage(chatId, upd.Message.MessageID, errorMessageBalanceOverflow), nil
	}

	err = api.storage.MoveMoneyFromUserToUser(chatId, fromId, toId, uint(amount))
	if err != nil {
		return nil, fmt.Errorf("error during MoveMoneyFromUserToUser %w", err)
	}

	return api.messageSendMoney(upd, amount, from, to), nil
}

func (api *dndUtilBotApi) getParams(text string) []string {
	params := strings.Split(text, " ")
	params = slices.DeleteFunc(params, func(s string) bool {
		return api.botName == s ||
			(len(s) > 0 && s[0] == '@' && s[1:] == api.botName) ||
			s == "" || s == " "
	})

	return params
}

func (api *dndUtilBotApi) setUserBalance(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	params := api.getParams(upd.Message.Text)
	if len(params) < 3 {
		return nil, ErrorInvalidParameters
	}

	amount, err := strconv.Atoi(params[2])
	if err != nil || amount < 0 {
		return nil, ErrorInvalidIntegerParameter
	}

	userName := params[1]
	userId, ok := api.userIdByUserName(userName)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, userName),
		)

		return &msg, nil
	}

	err = api.storage.SetUserBalance(upd.FromChat().ID, userId, uint(amount))
	if err != nil {
		return nil, fmt.Errorf("error during setUserBalance %w", err)
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, fmt.Sprintf(messageSetUserBalanceSuccess, userName, amount))
	return &msg, err
}

func (api *dndUtilBotApi) getUserBalance(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	params := api.getParams(upd.Message.Text)
	if len(params) < 2 {
		return nil, ErrorInvalidParameters
	}

	userName := params[1]
	userId, ok := api.userIdByUserName(userName)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, userName),
		)

		return &msg, nil
	}

	balance, err := api.storage.GetUserBalance(upd.FromChat().ID, userId)
	if err != nil {
		return nil, fmt.Errorf("error during getting balance from storage %w", err)
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, fmt.Sprintf(messageGetUserBalanceSuccess, userName, balance))
	return &msg, nil
}

func (api *dndUtilBotApi) throwDice(upd *tgbotapi.Update) (*tgbotapi.StickerConfig, error) {
	return api.stickerThrowDice(upd)
}

func (api *dndUtilBotApi) stickerThrowDice(upd *tgbotapi.Update) (*tgbotapi.StickerConfig, error) {
	d20 := api.randomizer.Intn(20) + 1
	emoji, ok := d20NumToEmojiMap[d20]
	if !ok {
		return nil, fmt.Errorf("error getting d20 emoji mapping")
	}

	set, err := api.tgBotApi.GetStickerSet(tgbotapi.GetStickerSetConfig{
		Name: D20StickerSetname,
	})
	if err != nil {
		return nil, err
	}

	for _, sticker := range set.Stickers {
		if sticker.Emoji != emoji {
			continue
		}

		stickerConfig := tgbotapi.NewSticker(upd.FromChat().ID, tgbotapi.FileID(sticker.FileID))
		return &stickerConfig, nil
	}

	return nil, fmt.Errorf("error getting d20 sticker from emoji mapping %d - %s", d20, emoji)
}

func (api *dndUtilBotApi) getBalance(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	err := validateUsernameIsNotHidden(upd)
	if err != nil {
		return nil, err
	}

	balance, err := api.storage.GetUserBalance(upd.FromChat().ID, upd.SentFrom().ID)
	if err != nil {
		return nil, fmt.Errorf("error during getBalance from storage %w", err)
	}

	return api.messageGetUserBalanceSuccess(upd, balance), nil
}

func (api *dndUtilBotApi) messageGetUserBalanceSuccess(upd *tgbotapi.Update, balance uint) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(
		upd.Message.Chat.ID,
		fmt.Sprintf(messageGetUserBalanceSuccess, upd.SentFrom().UserName, balance),
	)

	return &msg
}

func (api *dndUtilBotApi) sendMoney(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	err := validateUsernameIsNotHidden(upd)
	if err != nil {
		return nil, err
	}

	params := api.getParams(upd.Message.Text)
	if len(params) < 3 {
		return nil, ErrorInvalidParameters
	}

	amount, err := strconv.Atoi(params[2])
	if err != nil || amount <= 0 {
		return nil, ErrorInvalidIntegerParameter
	}

	from := upd.SentFrom()
	fromId := from.ID
	toUserName := params[1]
	toId, ok := api.userIdByUserName(toUserName)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, toUserName),
		)

		return &msg, nil
	}

	if fromId == toId {
		return nil, ErrorInvalidTransactionParameters
	}

	err = api.storage.MoveMoneyFromUserToUser(upd.FromChat().ID, fromId, toId, uint(amount))
	if err != nil {
		return nil, fmt.Errorf("error during MoveMoneyFromUserToUser %w", err)
	}

	return api.messageSendMoney(upd, amount, from.UserName, toUserName), nil
}

func (api *dndUtilBotApi) messageSendMoney(upd *tgbotapi.Update, amount int, fromUserName string, toUserName string) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(
		upd.FromChat().ID,
		fmt.Sprintf(
			messageSendMoney,
			amount,
			fmt.Sprintf("@%s", fromUserName),
			toUserName,
		),
	)

	return &msg
}

func (api *dndUtilBotApi) start(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	err := validateUsernameIsNotHidden(upd)
	if err != nil {
		return nil, err
	}

	chat := upd.FromChat()
	balance, err := api.storage.GetUserBalance(upd.FromChat().ID, upd.SentFrom().ID)
	if err != nil {
		return nil, fmt.Errorf("error while start: %w", err)
	}

	msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf(messageStart, balance))
	return &msg, nil
}

func (api *dndUtilBotApi) sendMoneyPrompt(upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
	msg := tgbotapi.NewMessage(upd.FromChat().ID, fmt.Sprintf(messageSendMoneyPrompt, commandSendMoney.usage))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	return &msg, nil
}

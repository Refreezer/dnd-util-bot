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

var d20NumToEmojiMap = map[int]string{
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

type (
	DndUtilApi interface {
		listener.UpdateHandler
	}

	ResourceProvider interface {
		get(uri string) ([]byte, error)
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
		//resourceProvider ResourceProvider
		botName string
	}
)

func NewDndUtilApi(
	tgBotApi *tgbotapi.BotAPI,
	loggerProvider LoggerProvider,
	storage Storage,
	// resourceProvider ResourceProvider,
) DndUtilApi {
	return newDndUtilApi(
		tgBotApi,
		loggerProvider,
		storage,
		//resourceProvider,
	)
}

func newDndUtilApi(
	tgBotApi *tgbotapi.BotAPI,
	loggerProvider LoggerProvider,
	storage Storage,
	// resourceProvider ResourceProvider,
) *dndUtilBotApi {
	api := &dndUtilBotApi{
		tgBotApi:   tgBotApi,
		logger:     loggerProvider.MustGetLogger("dndUtilBotApi"),
		storage:    storage,
		randomizer: rand.New(rand.NewSource(time.Now().Unix())),
		botName:    "DnDUtilTest_bot",
		//resourceProvider: resourceProvider,
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
	if upd.Message == nil || upd.Message.Chat == nil || upd.Message.From == nil {
		return
	}

	from := upd.SentFrom()
	api.registerWallet(from)
	switch upd.Message.Chat.Type {
	case ChatTypeGroup, ChatTypeSuperGroup, ChatTypePrivate:
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

func (api *dndUtilBotApi) messageRightsViolation(upd *tgbotapi.Update) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, messageRejectedRightsViolation)
	return &msg
}

func (api *dndUtilBotApi) messageNotImplemented(upd *tgbotapi.Update) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, messageNotImplemented)
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

func (api *dndUtilBotApi) isUserRegistered(userName string) bool {
	_, ok := api.getIdByUserNameSanitized(userName)
	return ok
}

func (api *dndUtilBotApi) userIdByUserName(userName string, upd *tgbotapi.Update) (int64, bool) {
	uid, ok := api.getIdByUserNameSanitized(userName)
	return uid, ok
}

func (api *dndUtilBotApi) getIdByUserNameSanitized(userName string) (int64, bool) {
	sanitizedUserName := strings.Replace(userName, "@", "", 1)
	uid, ok := api.storage.GetIdByUserName(sanitizedUserName)
	return uid, ok
}

func (api *dndUtilBotApi) moveMoneyFromUserToUser(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	params := api.getParams(upd.Message.Text)
	if len(params) < 4 {
		return nil, fmt.Errorf("inalid MoveMoneyFromUserToUser command parameters")
	}

	amount, err := strconv.Atoi(params[3])
	if err != nil {
		return nil, fmt.Errorf("can't convert money amount toId integer %w", err)
	}

	userName := params[1]
	fromId, ok := api.userIdByUserName(userName, upd)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, userName),
		)

		return &msg, nil
	}

	to := params[2]
	toId, ok := api.userIdByUserName(to, upd)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, to),
		)

		return &msg, nil
	}

	if fromId == toId {
		return nil, fmt.Errorf("can't make transaction from sender acc to itself")
	}

	err = api.storage.MoveMoneyFromUserToUser(fromId, toId, uint(amount))
	if err != nil {
		return nil, fmt.Errorf("error during MoveMoneyFromUserToUser %w", err)
	}

	return nil, nil
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
		return nil, fmt.Errorf("inalid setUserBalance command parameters")
	}

	amount, err := strconv.Atoi(params[2])
	if err != nil {
		return nil, fmt.Errorf("can't convert money amount to integer %w", err)
	}

	userName := params[1]
	userId, ok := api.userIdByUserName(userName, upd)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, userName),
		)

		return &msg, nil
	}

	err = api.storage.SetUserBalance(userId, uint(amount))
	if err != nil {
		return nil, fmt.Errorf("error during setUserBalance %w", err)
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, fmt.Sprintf(messageSetUserBalanceSuccess, userName, amount))
	return &msg, err
}

func (api *dndUtilBotApi) getUserBalance(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	params := api.getParams(upd.Message.Text)
	if len(params) < 2 {
		return nil, fmt.Errorf("inalid GetUserBalance command parameters")
	}

	userName := params[1]
	userId, ok := api.userIdByUserName(userName, upd)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, userName),
		)

		return &msg, nil
	}

	balance, err := api.storage.GetUserBalance(userId)
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
		Name: "D20STUMP",
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
	balance, err := api.storage.GetUserBalance(upd.SentFrom().ID)
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
	params := api.getParams(upd.Message.Text)
	if len(params) < 3 {
		return nil, fmt.Errorf("inalid MoveMoneyFromUserToUser command parameters")
	}

	amount, err := strconv.Atoi(params[2])
	if err != nil {
		return nil, fmt.Errorf("can't convert money amount toId integer %w", err)
	}

	from := upd.SentFrom()
	fromId := from.ID
	to := params[1]
	toId, ok := api.userIdByUserName(to, upd)
	if !ok {
		msg := tgbotapi.NewMessage(
			upd.Message.Chat.ID,
			fmt.Sprintf(messageNotRegistered, to),
		)

		return &msg, nil
	}

	err = api.storage.MoveMoneyFromUserToUser(fromId, toId, uint(amount))
	if err != nil {
		return nil, fmt.Errorf("error during MoveMoneyFromUserToUser %w", err)
	}

	return api.messageSendMoney(upd, amount, from, to), nil
}

func (api *dndUtilBotApi) messageSendMoney(upd *tgbotapi.Update, amount int, from *tgbotapi.User, to string) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(
		upd.FromChat().ID,
		fmt.Sprintf(messageSendMoney, amount, fmt.Sprintf("@%s", from), to),
	)
	return &msg
}

func (api *dndUtilBotApi) notImplemented(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	return api.messageNotImplemented(upd), nil
}

func (api *dndUtilBotApi) rightsViolation(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	return api.messageRightsViolation(upd), nil
}

func (api *dndUtilBotApi) start(upd *tgbotapi.Update) (*tgbotapi.MessageConfig, error) {
	chat := upd.FromChat()
	balance, err := api.storage.GetUserBalance(upd.SentFrom().ID)
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf(messageStart, balance))
	return &msg, nil
}

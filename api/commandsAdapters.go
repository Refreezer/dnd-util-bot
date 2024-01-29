package api

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type markupProvider func(api *dndUtilBotApi, upd *tgbotapi.Update) *tgbotapi.ReplyKeyboardMarkup
type baseChatModifier func(c *tgbotapi.BaseChat, api *dndUtilBotApi, upd *tgbotapi.Update)

func wrapHandler(handler commandHandler, modifier baseChatModifier) commandHandler {
	return func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		msg, err := handler(api, upd)
		if err != nil {
			return nil, err
		}

		switch v := msg.(type) {
		case *tgbotapi.MessageConfig:
			modifier(&v.BaseChat, api, upd)
		case *tgbotapi.StickerConfig:
			modifier(&v.BaseChat, api, upd)
		}
		return msg, nil
	}
}

func (handler commandHandler) setReplyMarkup(markupProvider markupProvider) commandHandler {
	return wrapHandler(handler, func(c *tgbotapi.BaseChat, api *dndUtilBotApi, upd *tgbotapi.Update) {
		c.ReplyMarkup = markupProvider(api, upd)
	})
}

func (handler commandHandler) setReplyToMessageID() commandHandler {
	return wrapHandler(handler, func(c *tgbotapi.BaseChat, api *dndUtilBotApi, upd *tgbotapi.Update) {
		c.ReplyParameters.MessageID = upd.Message.MessageID
	})
}

func (handler commandHandler) setThreadIdForSuperGroup() commandHandler {
	return wrapHandler(handler, func(c *tgbotapi.BaseChat, api *dndUtilBotApi, upd *tgbotapi.Update) {
		chatType := upd.FromChat().Type
		if chatType != ChatTypeSuperGroup {
			return
		}

		c.MessageThreadID = upd.Message.MessageThreadID
	})
}

var (
	handlerMoveMoneyFromUserToUser commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.Transaction(upd)
	}

	handlerSetUserBalance commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.setUserBalance(upd)
	}

	handlerGetUserBalance commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.getUserBalance(upd)
	}

	handlerThrowDice commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.throwDice(upd)
	}

	handlerGetBalance commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.getBalance(upd)
	}

	handlerSendMoneyPrompt commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.sendMoneyPrompt(upd)
	}

	handlerSendMoney commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.sendMoney(upd)
	}

	handlerNotImplemented commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return notImplemented(upd)
	}

	handlerRightsViolation commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return rightsViolation(upd)
	}

	handlerCantResolve commandHandler = func(_ *dndUtilBotApi, _ *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return nil, nil
	}

	handlerStart commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.start(upd)
	}

	handlerHelp commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.commands.printHelp(upd)
	}
)

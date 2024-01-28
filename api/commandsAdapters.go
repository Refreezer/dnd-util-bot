package api

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type markupProvider func(api *dndUtilBotApi, upd *tgbotapi.Update) *tgbotapi.ReplyKeyboardMarkup

func (handler commandHandler) setReplyMarkup(markupProvider markupProvider) commandHandler {
	return func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		msg, err := handler(api, upd)
		if err != nil {
			return nil, err
		}

		switch v := msg.(type) {
		case *tgbotapi.MessageConfig:
			v.ReplyMarkup = markupProvider(api, upd)
		case *tgbotapi.StickerConfig:
			v.ReplyMarkup = markupProvider(api, upd)
		}
		return msg, nil
	}
}

func (handler commandHandler) setReplyToMessageID() commandHandler {
	return func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		msg, err := handler(api, upd)
		if err != nil {
			return nil, err
		}

		switch v := msg.(type) {
		case *tgbotapi.MessageConfig:
			v.ReplyToMessageID = upd.Message.MessageID
		case *tgbotapi.StickerConfig:
			v.ReplyToMessageID = upd.Message.MessageID
		}

		return msg, nil
	}
}

var (
	handlerMoveMoneyFromUserToUser commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.moveMoneyFromUserToUser(upd)
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
		return api.notImplemented(upd)
	}

	handlerRightsViolation commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.rightsViolation(upd)
	}

	handlerCantResolve commandHandler = func(_ *dndUtilBotApi, _ *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return nil, nil
	}

	handlerStart commandHandler = func(api *dndUtilBotApi, upd *tgbotapi.Update) (tgbotapi.Chattable, error) {
		return api.start(upd)
	}
)

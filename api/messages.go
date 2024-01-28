package api

const (
	messageSendMoneyPrompt         = "Чтобы передать золотые монеты 🟡 игроку, напиши:\n`/sendMoney @username 123`"
	messageSendMoney               = " %d 🟡 золотых монет %s передал %s"
	messageStart                   = "Доброго тебе дня, путник! Я - ролевой бот помощник. Я умею кидать Д20. Кстати, а у тебя теперь есть свой кошель 💰. У тебя %d золотых монет. Выполняй задания Гильдий и их будет больше! Успехов в твоем приключении 💚"
	messageNotImplemented          = "Кажется, я не совсем понял тебя, путник. Эти знания для меня недоступны...🍃"
	messageRejectedRightsViolation = "А ты хитёр... Но так сделать нельзя, путник 👿"
	messageGetUserBalanceSuccess   = "💰 Кошель %s - %d золотых монет 🟡"
	messageSetUserBalanceSuccess   = "💰 Кошель %s теперь %d золотых монет 🟡"
	messageNotRegistered           = "Кажется путник %s еще не зарегистрировался в Гильдии Приключений, так что я не могу это сделать 😓"

	errorMessageInvalidIntegerParameter      = "Путник, кажется твоё число неправильное 🤨\\. Попробуй иначе!\\."
	errorMessageInvalidTransactionParameters = "Думаешь, что перехитрил меня 😠? Чтобы я такого больше не видел!\\."
	errorMessageInvalidParametersFormat      = "Путник, кажется твои параметры неправильные ☹️\\. Смотри как надо:\n`%s`"
)

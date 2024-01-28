package api

const (
	messageSendMoneyPrompt         = "Чтобы перевести денежку игроку, напиши:\n`/sendMoney @username 123`"
	messageSendMoney               = "Денежка в размере %d улетела от %s к %s"
	messageStart                   = "Привет! Я - бот помощник для ДнД. Я зарегистрировал для тебя кошелек. У тебя %d денег."
	messageNotImplemented          = "Привет! Я - бот помощник для ДнД. Я пока этого не умею, но я рад тебя видеть!😀"
	messageRejectedRightsViolation = "Эй, тормози, тебе кто разрешил такое делать?🚨"
	messageGetUserBalanceSuccess   = "Баланс %s - %d💰"
	messageSetUserBalanceSuccess   = "Баланс %s теперь %d💰"
	messageNotRegistered           = "Игрок %s еще не зарегался, так что залупу тебе за воротник а не транзакцию, солнышко :)"

	errorMessageInvalidIntegerParameter      = "Мда, число у тебя неаправильное\\. Подумай получше\\."
	errorMessageInvalidTransactionParameters = "Это по твоему транзакция? Чтобы я такого больше не видел\\."
	errorMessageInvalidParametersFormat      = "Мда, параметры у тебя неправильные\\. Смотри как надо:\n`%s`"
)

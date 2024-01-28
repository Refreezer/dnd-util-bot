package listener

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/op/go-logging"
	"sync"
)

type (
	ShutDown    func()
	BotListener interface {
		ListenForUpdates(ctx context.Context) (waitForShutDown ShutDown, err error)
	}

	UpdateHandler interface {
		Handle(ctx context.Context, upd *tgbotapi.Update)
	}

	Config struct {
		Timeout        int
		AllowedUpdates []string
		UpdateHandler  UpdateHandler
	}

	dndUtilBotListener struct {
		conf           *Config
		tgBotApi       *tgbotapi.BotAPI
		lastUpdateId   int
		timeout        int
		updatesChannel tgbotapi.UpdatesChannel
		done           chan struct{}
		logger         *logging.Logger
	}
)

type LoggerProvider interface {
	MustGetLogger(moduleName string) *logging.Logger
}

func NewBotListener(
	tgBotApi *tgbotapi.BotAPI,
	conf *Config,
	loggerProvider LoggerProvider,
) BotListener {
	logger := loggerProvider.MustGetLogger("botListenerApi")
	return &dndUtilBotListener{
		conf:         conf,
		tgBotApi:     tgBotApi,
		lastUpdateId: 0,
		logger:       logger,
	}
}

func (l *dndUtilBotListener) ListenForUpdates(ctx context.Context) (ShutDown, error) {
	updates := l.tgBotApi.GetUpdatesChan(tgbotapi.UpdateConfig{
		Offset:         l.lastUpdateId + 1,
		Timeout:        l.conf.Timeout,
		AllowedUpdates: l.conf.AllowedUpdates,
	})

	l.updatesChannel = updates
	go l.eventLoop(ctx)
	return l.waitForShutDown(), nil
}

func (l *dndUtilBotListener) waitForShutDown() func() {
	return func() {
		<-l.done
	}
}

func (l *dndUtilBotListener) eventLoop(ctx context.Context) {
	defer func() {
		l.done <- struct{}{}
	}()

	wg := sync.WaitGroup{}
	for {
		select {
		case <-ctx.Done():
			l.tgBotApi.StopReceivingUpdates()
			l.logger.Info("exiting listen for updates loop due to canceled context")
			wg.Wait()
			l.logger.Info("Shutting down gracefully")
			return
		case update, ok := <-l.updatesChannel:
			if !ok {
				continue
			}

			l.lastUpdateId = max(update.UpdateID, l.lastUpdateId)
			wg.Add(1)
			go func() {
				defer wg.Done()
				l.conf.UpdateHandler.Handle(ctx, &update)
			}()
			continue
		default:
			// l.logger.Debugf("idle state")
			continue
		}
	}
}

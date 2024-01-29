package listener

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/op/go-logging"
	"runtime"
	"sync"
	"time"
)

type (
	ShutDown    func()
	BotListener interface {
		ListenForUpdates(ctx context.Context) (waitForShutDown ShutDown, err error)
	}

	UpdateHandler interface {
		HandleUpdate(ctx context.Context, upd *tgbotapi.Update)
	}

	Config struct {
		RateLimitRps   int
		TgTimeout      int
		AllowedUpdates []string
		UpdateHandler  UpdateHandler
	}

	dndUtilBotListener struct {
		conf           *Config
		tgBotApi       *tgbotapi.BotAPI
		lastUpdateId   int
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
		done:         make(chan struct{}),
	}
}

func (l *dndUtilBotListener) ListenForUpdates(ctx context.Context) (ShutDown, error) {
	updates := l.tgBotApi.GetUpdatesChan(tgbotapi.UpdateConfig{
		Offset:         l.lastUpdateId + 1,
		Timeout:        l.conf.TgTimeout,
		AllowedUpdates: l.conf.AllowedUpdates,
	})

	l.updatesChannel = updates
	tasks := make(chan *tgbotapi.Update, runtime.NumCPU())
	go l.eventLoop(ctx, tasks)
	rl := l.rateLimit(ctx)
	l.startWorkers(ctx, tasks, rl)
	return l.waitForShutDown(), nil
}

func (l *dndUtilBotListener) waitForShutDown() func() {
	return func() {
		<-l.done
	}
}

func (l *dndUtilBotListener) eventLoop(ctx context.Context, tasks chan<- *tgbotapi.Update) {
	defer func() {
		l.done <- struct{}{}
		close(tasks)
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
			tasks <- &update
			continue
		}
	}
}

func (l *dndUtilBotListener) startWorkers(ctx context.Context, tasks <-chan *tgbotapi.Update, rl <-chan struct{}) {
	for i := 0; i <= runtime.NumCPU(); i++ {
		go l.worker(ctx, tasks, rl)
	}
}

func (l *dndUtilBotListener) worker(ctx context.Context, tasks <-chan *tgbotapi.Update, rl <-chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			l.logger.Info("worker exits due to canceled context")
			return
		case update, ok := <-tasks:
			if _, ok := <-rl; !ok { // rate limit
				l.logger.Info("worker exits due to rateLimit channel was closed")
				return
			}
			if !ok {
				l.logger.Info("worker exits due to input channel was closed")
				return
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						l.logger.Warning("recovered after update handler failed %s", r)
					}
				}()

				l.conf.UpdateHandler.HandleUpdate(ctx, update)
			}()
			continue
		}
	}
}

func (l *dndUtilBotListener) rateLimit(ctx context.Context) <-chan struct{} {
	tokenRateMs := 1000 / l.conf.RateLimitRps
	rl := make(chan struct{}, l.conf.RateLimitRps)
	go func(ctx context.Context, tokens chan<- struct{}) {
		defer close(tokens)
		for {
			select {
			case <-ctx.Done():
				return
			case tokens <- struct{}{}:
				time.Sleep(time.Millisecond * time.Duration(tokenRateMs))
				continue
			}
		}

	}(ctx, rl)

	return rl
}

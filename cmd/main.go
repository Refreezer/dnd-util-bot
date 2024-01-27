package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Refreezer/dnd-util-bot/api"
	"github.com/Refreezer/dnd-util-bot/api/listener"
	. "github.com/Refreezer/dnd-util-bot/internal"
	"github.com/Refreezer/dnd-util-bot/internal/mapStorage"
	. "github.com/Refreezer/dnd-util-bot/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/op/go-logging"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type loggerProvider struct {
	Debug bool
}

func (lp *loggerProvider) MustGetLogger(moduleName string) *logging.Logger {
	logger := logging.MustGetLogger(moduleName)
	InitLogger(lp.Debug, logger)
	return logger
}

func main() {
	tgApiKey, timeout := parseEnvironmentVariables()
	debug := parseFlags()
	validateConfiguration(debug, timeout)

	tgBotApi, err := tgbotapi.NewBotAPI(tgApiKey)
	if err != nil {
		Logger.Fatalf("error while initializing telegram bot api %s", err)
	}

	loggerProvider := &loggerProvider{
		Debug: *debug,
	}
	botListener := listener.NewBotListener(
		tgBotApi,
		&listener.Config{
			Timeout:        2,
			AllowedUpdates: []string{tgbotapi.UpdateTypeMessage},
			UpdateHandler: api.NewDndUtilApi(
				tgBotApi,
				loggerProvider,
				mapStorage.NewMapStorage(),
			),
		},
		loggerProvider,
	)

	ctx, cancel := context.WithCancel(context.Background())
	waitForShutDown, err := botListener.ListenForUpdates(ctx)
	listenForCancelContextRequest(cancel)
	if err != nil {
		Logger.Fatalf("error while starting bot listener %s", err)
	}

	defer waitForShutDown()
}

func listenForCancelContextRequest(cancel context.CancelFunc) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go listenOsSignals(signalChannel, cancel)
}

func listenOsSignals(signalChannel chan os.Signal, cancel context.CancelFunc) {
	func(sigs <-chan os.Signal, cancelFunc context.CancelFunc) {
		for {
			sig := <-sigs
			switch sig {
			case os.Interrupt:
			case syscall.SIGTERM:
				fmt.Println("Shutting down")
				cancelFunc()
			default:
				fmt.Println("Unknown sig received.")
			}
		}
	}(signalChannel, cancel)
}

func validateConfiguration(debug *bool, timeout int) {
	if !*debug && timeout == 0 {
		Logger.Fatal("timeout can't be zero in production mode")
	}
}

func parseFlags() *bool {
	debug := flag.Bool("d", false, "Debug mode")
	flag.Parse()
	InitGlobalLogger(*debug)
	return debug
}

func parseEnvironmentVariables() (string, int) {
	tgApiKey := mustGetEnv(DndUtilTgApiKey)
	timeoutStr := mustGetEnv(DndUtilLongPollingTimeout)
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		Logger.Fatalf("timeout environment variable is invalid %s", timeoutStr)
	}
	return tgApiKey, timeout
}

func mustGetEnv(key EnvKey) string {
	env := os.Getenv(string(key))
	if env == EmptyString {
		Logger.Fatalf("%s not set", key)
	}

	return env
}

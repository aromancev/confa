package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aromancev/confa/cmd/sender/queue"
	"github.com/aromancev/confa/internal/platform/email/mailersend"
	"github.com/aromancev/confa/internal/proto/iam"
	"github.com/aromancev/confa/sender"
	"github.com/aromancev/confa/sender/email"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	config := Config{}.WithEnv()
	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}

	if config.LogFormat == LogConsole {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		log.Logger = zerolog.New(os.Stdout)
	}
	log.Logger = log.Logger.With().Timestamp().Caller().Logger()
	switch config.LogLevel {
	case LevelDebug:
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	case LevelWarn:
		log.Logger = log.Logger.Level(zerolog.WarnLevel)
	case LevelError:
		log.Logger = log.Logger.Level(zerolog.ErrorLevel)
	default:
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}
	ctx = log.Logger.WithContext(ctx)

	consumer, err := beanstalk.NewConsumer(config.Beanstalk.ParsePool(), []string{config.Beanstalk.TubeSend}, beanstalk.Config{
		Multiply:         1,
		NumGoroutines:    10,
		ReserveTimeout:   time.Second,
		ReconnectTimeout: 3 * time.Second,
		InfoFunc: func(message string) {
			log.Info().Msg(message)
		},
		ErrorFunc: func(err error, message string) {
			log.Err(err).Msg(message)
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to beanstalk.")
	}

	iamClient := iam.NewIAMProtobufClient("http://"+config.IAMRPCAddress, &http.Client{})

	jobHandler := queue.NewHandler(
		sender.NewSender(
			email.NewSender(
				mailersend.NewSender(
					&http.Client{},
					config.Email.MailersendBaseURL,
					config.Email.MailersendToken,
				),
				config.Email.MailersendFromEmail,
			),
			iamClient,
		),
		queue.Tubes{
			Send: config.Beanstalk.TubeSend,
		},
	)

	var done sync.WaitGroup
	done.Add(1)
	go func() {
		consumer.Receive(ctx, jobHandler.ServeJob)
		done.Done()
	}()

	<-ctx.Done()
	cancel()
	log.Info().Msg("Shutting down")
	done.Wait()
}

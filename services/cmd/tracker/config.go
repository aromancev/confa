package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

const (
	LogConsole = "console"

	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelError = "error"
	LevelWarn  = "warn"
)

type Config struct {
	ListenRPCAddress string `envconfig:"LISTEN_RPC_ADDRESS"`
	LogFormat        string `envconfig:"LOG_FORMAT"`
	LogLevel         string `envconfig:"LOG_LEVEL"`
	PublicKey        string `envconfig:"PUBLIC_KEY"`
	TmpDir           string `envconfig:"TMP_DIR"`
	Beanstalk        BeanstalkConfig
	Storage          StorageConfig
	Livekit          LiveKitConfig
}

func (c Config) WithEnv() Config {
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to process env")
	}

	pk, err := base64.StdEncoding.DecodeString(c.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode PUBLIC_KEY (expected base64)")
	}
	c.PublicKey = string(pk)
	return c
}

func (c Config) Validate() error {
	switch c.LogLevel {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
	default:
		return errors.New("LOG_LEVEL is not valid")
	}
	if c.ListenRPCAddress == "" {
		return errors.New("LISTEN_RPC_ADDRESS not set")
	}
	if c.PublicKey == "" {
		return errors.New("PUBLIC_KEY not set")
	}
	if c.TmpDir == "" {
		return errors.New("TMP_DIR not set")
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("invalid storage config: %w", err)
	}
	if err := c.Beanstalk.Validate(); err != nil {
		return fmt.Errorf("invalid beanstalk config: %w", err)
	}
	if err := c.Livekit.Validate(); err != nil {
		return fmt.Errorf("invalid livekit config: %w", err)
	}
	return nil
}

type BeanstalkConfig struct {
	Pool                     string `envconfig:"BEANSTALK_POOL"`
	TubeProcessTrack         string `envconfig:"BEANSTALK_TUBE_PROCESS_TRACK"`
	TubeStoreEvent           string `envconfig:"BEANSTALK_TUBE_STORE_EVENT"`
	TubeUpdateRecordingTrack string `envconfig:"BEANSTALK_TUBE_UPDATE_RECORDING_TRACK"`
}

func (c BeanstalkConfig) Validate() error {
	if c.Pool == "" {
		return errors.New("BEANSTALK_POOL not set")
	}
	if c.TubeProcessTrack == "" {
		return errors.New("BEANSTALK_TUBE_PROCESS_TRACK not set")
	}
	if c.TubeStoreEvent == "" {
		return errors.New("BEANSTALK_TUBE_STORE_EVENT not set")
	}
	if c.TubeUpdateRecordingTrack == "" {
		return errors.New("BEANSTALK_TUBE_UPDATE_RECORDING_TRACK not set")
	}
	return nil
}

func (c BeanstalkConfig) ParsePool() []string {
	return strings.Split(c.Pool, ",")
}

type StorageConfig struct {
	Host               string `envconfig:"STORAGE_HOST"`
	Scheme             string `envconfig:"STORAGE_SCHEME"`
	Region             string `envconfig:"STORAGE_REGION"`
	AccessKey          string `envconfig:"STORAGE_ACCESS_KEY"`
	SecretKey          string `envconfig:"STORAGE_SECRET_KEY"`
	BucketTrackRecords string `envconfig:"STORAGE_BUCKET_TRACK_RECORDS"`
}

func (c StorageConfig) Validate() error {
	if c.Host == "" {
		return errors.New("STORAGE_HOST not set")
	}
	if c.Scheme == "" {
		return errors.New("STORAGE_SCHEME not set")
	}
	if c.Region == "" {
		return errors.New("STORAGE_REGION not set")
	}
	if c.AccessKey == "" {
		return errors.New("STORAGE_ACCESS_KEY not set")
	}
	if c.SecretKey == "" {
		return errors.New("STORAGE_SECRET_KEY not set")
	}
	if c.BucketTrackRecords == "" {
		return errors.New("STORAGE_BUCKET_TRACK_RECORDS not set")
	}
	return nil
}

type LiveKitConfig struct {
	URL    string `envconfig:"LIVEKIT_URL"`
	Key    string `envconfig:"LIVEKIT_KEY"`
	Secret string `envconfig:"LIVEKIT_SECRET"`
}

func (c LiveKitConfig) Validate() error {
	if c.URL == "" {
		return errors.New("LIVEKIT_URL not set")
	}
	if c.Key == "" {
		return errors.New("LIVEKIT_KEY not set")
	}
	if c.Secret == "" {
		return errors.New("LIVEKIT_KEY not set")
	}
	return nil
}

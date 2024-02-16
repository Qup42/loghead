package types

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	Addr        string
	FileLogger  FileLoggerConfig
	Log         LogConfig
	Processors  ProcessorConfig
	SSHRecorder SSHRecorderConfig
}

type FileLoggerConfig struct {
	Dir string
}

type LogConfig struct {
	Level  zerolog.Level
	Format string
}

type ProcessorConfig struct {
	FileLogger bool
	Metrics    bool
	Hostinfo   bool
}

type SSHRecorderConfig struct {
	Addr string
	Dir  string
}

const (
	JSONLogFormat = "json"
	TextLogFormat = "text"
)

func GetProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		FileLogger: viper.GetBool("processors.filelogger"),
		Metrics:    viper.GetBool("processors.metrics"),
		Hostinfo:   viper.GetBool("processors.hostinfo"),
	}
}

func GetSSHRecorderConfig() SSHRecorderConfig {
	return SSHRecorderConfig{
		Addr: viper.GetString("ssh_recorder.listen_addr"),
		Dir:  viper.GetString("ssh_recorder.dir"),
	}
}

func GetFileLoggerConfig() FileLoggerConfig {
	return FileLoggerConfig{
		Dir: viper.GetString("filelogger.dir"),
	}
}

func GetLogConfig() LogConfig {
	logLevelStr := viper.GetString("log.level")
	logLevel, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = zerolog.DebugLevel
	}

	logFormatOpt := viper.GetString("log.format")
	var logFormat string
	switch logFormatOpt {
	case "json":
		logFormat = JSONLogFormat
	case "text":
		logFormat = TextLogFormat
	case "":
		logFormat = TextLogFormat
	default:
		log.Error().Str("func", "GetLogConfig").
			Msgf("Could not parse log format: %s. Valid choices are 'json' or 'text'", logFormatOpt)
	}

	return LogConfig{
		Level:  logLevel,
		Format: logFormat,
	}
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/loghead/")
	viper.AddConfigPath("$HOME/.loghead/")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("loghead")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("listen_addr", "0.0.0.0:5678")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", TextLogFormat)

	viper.SetDefault("processor.filelogger", true)
	viper.SetDefault("filelogger.dir", "./logs")
	viper.SetDefault("processor.metrics", false)
	viper.SetDefault("processor.hostinfo", false)

	viper.SetDefault("ssh_recorder.listen_addr", "0.0.0.0:5679")
	viper.SetDefault("ssh_recorder.dir", "./recordings")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("Failed to read config")
	}

	return Config{
		Addr:        viper.GetString("listen_addr"),
		Log:         GetLogConfig(),
		FileLogger:  GetFileLoggerConfig(),
		Processors:  GetProcessorConfig(),
		SSHRecorder: GetSSHRecorderConfig(),
	}
}

package types

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	Listener    ListenerConfig
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
	Forward    string
}

type ListenerConfig struct {
	Type           string
	Addr           string
	Port           string
	TS_AuthKey     string
	TS_ControllURL string
}

type SSHRecorderConfig struct {
	Dir      string
	Listener ListenerConfig
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
		Forward:    viper.GetString("processors.forward"),
	}
}

func GetListenerConfig(base string) ListenerConfig {
	return ListenerConfig{
		Type:           viper.GetString(base + ".listener.type"),
		Addr:           viper.GetString(base + ".listener.addr"),
		Port:           viper.GetString(base + ".listener.port"),
		TS_AuthKey:     viper.GetString(base + ".listener.tsnet.authKey"),
		TS_ControllURL: viper.GetString(base + ".listener.tsnet.controllURL"),
	}
}

func GetSSHRecorderConfig() SSHRecorderConfig {
	return SSHRecorderConfig{
		Dir:      viper.GetString("ssh_recorder.dir"),
		Listener: GetListenerConfig("ssh_recorder"),
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

	viper.SetDefault("ssh_recorder.listener.type", "tsnet")
	viper.SetDefault("ssh_recorder.listener.addr", "0.0.0.0")
	viper.SetDefault("ssh_recorder.listener.port", "80")
	viper.SetDefault("ssh_recorder.tsnet.controllURL", "https://controlplane.tailscale.com")
	viper.SetDefault("ssh_recorder.dir", "./recordings")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("Failed to read config")
	}

	var errorText string
	if (viper.GetString("ssh_recorder.listener.type") == "tsnet") && (!viper.IsSet("ssh_recorder.listener.tsnet.authKey")) {
		errorText += "Fatal config error: when using a tsnet listener, authkey must be provided\n"
	}
	if (viper.GetString("loghead.listener.type") == "tsnet") && (!viper.IsSet("loghead.listener.tsnet.authKey")) {
		errorText += "Fatal config error: when using a tsnet listener, authkey must be provided\n"
	}

	if errorText != "" {
		log.Error().Msg(strings.TrimSuffix(errorText, "\n"))
	}

	return Config{
		Listener:    GetListenerConfig("loghead"),
		Log:         GetLogConfig(),
		FileLogger:  GetFileLoggerConfig(),
		Processors:  GetProcessorConfig(),
		SSHRecorder: GetSSHRecorderConfig(),
	}
}

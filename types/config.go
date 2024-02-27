package types

import (
	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	Log         LogConfig
	SSHRecorder SSHRecorderConfig
	Loghead     LogheadConfig
}

type FileLoggerConfig struct {
	Enabled bool
	Dir     string
}

type ForwardingConfig struct {
	Enabled bool
	Addr    string
}

type LogConfig struct {
	Level  zerolog.Level
	Format string
}

type ProcessorConfig struct {
	FileLogger FileLoggerConfig
	Metrics    bool
	Hostinfo   bool
	Forward    ForwardingConfig
}

type ListenerConfig struct {
	Type string
	Addr string
	Port string
	TS   TSConfig
}

type TSConfig struct {
	AuthKey     string
	ControllURL string
	HostName    string
	Dir         string
}

type SSHRecorderConfig struct {
	Dir      string
	Listener ListenerConfig
}

type LogheadConfig struct {
	Processors ProcessorConfig
	Listener   ListenerConfig
}

const (
	JSONLogFormat = "json"
	TextLogFormat = "text"
)

func GetProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		FileLogger: GetFileLoggerConfig(),
		Metrics:    viper.GetBool("loghead.processors.metrics"),
		Hostinfo:   viper.GetBool("loghead.processors.hostinfo"),
		Forward:    GetForwardingConfig(),
	}
}

func GetForwardingConfig() ForwardingConfig {
	return ForwardingConfig{
		Enabled: viper.GetBool("loghead.processors.forward.enabled"),
		Addr:    viper.GetString("loghead.processors.forward.addr"),
	}
}

func GetListenerConfig(base string) ListenerConfig {
	return ListenerConfig{
		Type: viper.GetString(base + ".listener.type"),
		Addr: viper.GetString(base + ".listener.addr"),
		Port: viper.GetString(base + ".listener.port"),
		TS:   GetTSConfig(base + ".listener"),
	}
}

func GetTSConfig(base string) TSConfig {
	return TSConfig{
		AuthKey:     viper.GetString(base + ".tsnet.authKey"),
		ControllURL: viper.GetString(base + ".tsnet.controllURL"),
		HostName:    viper.GetString(base + ".tsnet.hostname"),
		Dir:         viper.GetString(base + ".tsnet.dir"),
	}
}

func GetSSHRecorderConfig() SSHRecorderConfig {
	return SSHRecorderConfig{
		Dir:      viper.GetString("ssh_recorder.dir"),
		Listener: GetListenerConfig("ssh_recorder"),
	}
}

func GetLogheadConfig() LogheadConfig {
	return LogheadConfig{
		Listener:   GetListenerConfig("loghead"),
		Processors: GetProcessorConfig(),
	}
}

func GetFileLoggerConfig() FileLoggerConfig {
	return FileLoggerConfig{
		Dir:     viper.GetString("loghead.processors.filelogger.dir"),
		Enabled: viper.GetBool("loghead.processors.filelogger.enabled"),
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

func LoadConfig() (*Config, error) {
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
		return nil, errors.New("Failed to read config")
	}

	var errorText string
	if (viper.GetString("ssh_recorder.listener.type") == "tsnet") && (!viper.IsSet("ssh_recorder.listener.tsnet.authKey")) {
		errorText += "Fatal config error: when using a tsnet listener, authkey must be provided\n"
	}
	if (viper.GetString("loghead.listener.type") == "tsnet") && (!viper.IsSet("loghead.listener.tsnet.authKey")) {
		errorText += "Fatal config error: when using a tsnet listener, authkey must be provided\n"
	}

	if errorText != "" {
		return nil, errors.New(strings.TrimSuffix(errorText, "\n"))
	}

	return &Config{
		Log:         GetLogConfig(),
		SSHRecorder: GetSSHRecorderConfig(),
		Loghead:     GetLogheadConfig(),
	}, nil
}

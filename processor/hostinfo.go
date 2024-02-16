package processor

import (
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
)

type HostInfo struct {
	AppConnector    bool   `mapstructure:"AppConnector"`
	BackendLogID    string `mapstructure:"BackendLogID"`
	Container       bool   `mapstructure:"Container"`
	Desktop         bool   `mapstructure:"Desktop"`
	Distro          string `mapstructure:"Distro"`
	DistroVersion   string `mapstructure:"DistroVersion"`
	GoArch          string `mapstructure:"GoArch"`
	GoArchVar       string `mapstructure:"GoArchVar"`
	GoVersion       string `mapstructure:"GoVersion"`
	Hostname        string `mapstructure:"Hostname"`
	IPNVersion      string `mapstructure:"IPNVersion"`
	Machine         string `mapstructure:"Machine"`
	OS              string `mapstructure:"OS"`
	OSVersion       string `mapstructure:"OSVersion"`
	Userspace       bool   `mapstructure:"Userspace"`
	UserspaceRouter bool   `mapstructure:"UserspaceRouter"`
}

func Process(msg LogtailMsg) {
	if h, ok := msg.Msg["Hostinfo"]; ok {
		var hi HostInfo
		err := mapstructure.Decode(h, &hi)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal HostInfo")
		}
		log.Debug().Msgf("HostInfo: %+v", hi)
	}
}

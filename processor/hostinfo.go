package processor

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
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

type HostInfoService struct {
}

func (hs *HostInfoService) Process(msg LogtailMsg) error {
	if h, ok := msg.Msg["Hostinfo"]; ok {
		var hi HostInfo
		err := mapstructure.Decode(h, &hi)
		if err != nil {
			return fmt.Errorf("unmarshaling HostInfo: %w", err)
		}
	}
	return nil
}

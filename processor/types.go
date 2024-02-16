package processor

type LogtailMsg struct {
	Msg       map[string]interface{}
	PrivateID string
}

type LogProcessor func(LogtailMsg)

const TailnodeCollection = "tailnode.log.tailscale.io"

package processor

type LogtailMsg struct {
	Msg       map[string]interface{}
	PrivateID string
}

type MsgProcessor func(LogtailMsg)
type LogProcessor func([]byte)

const TailnodeCollection = "tailnode.log.tailscale.io"

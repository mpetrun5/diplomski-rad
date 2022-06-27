package p2p

import "github.com/libp2p/go-libp2p-core/peer"

type MessageType uint8

const (
	TssKeyGenMsg MessageType = iota
	TssKeySignMsg
	TssReshareMsg
	TssInitiateMsg
	TssReadyMsg
	TssStartMsg
	Unknown
)

// String implement fmt.Stringer
func (msgType MessageType) String() string {
	switch msgType {
	case TssKeyGenMsg:
		return "TssKeyGenMsg"
	case TssKeySignMsg:
		return "TssKeySignMsg"
	case TssInitiateMsg:
		return "TssInitiateMsg"
	case TssStartMsg:
		return "TssStartMsg"
	case TssReadyMsg:
		return "TssReadyMsg"
	case TssReshareMsg:
		return "TssReshareMsg"
	default:
		return "Unknown"
	}
}

// WrappedMessage is a message sent to the stream iwth Broadcast.
type WrappedMessage struct {
	MessageType MessageType `json:"message_type"`
	SessionID   string      `json:"message_id"`
	Payload     []byte      `json:"payload"`
	From        peer.ID     `json:"from"`
}

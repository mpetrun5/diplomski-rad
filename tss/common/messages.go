package common

import (
	"encoding/json"
	"math/big"

	"github.com/libp2p/go-libp2p-core/peer"
)

type TssMessage struct {
	MsgBytes   []byte `json:"msgBytes"`
	From       string `json:"from"`
	IsBrodcast bool   `json:"isBrodcast"`
}

func MarshalTssMessage(msgBytes []byte, isBrodcast bool, from string) ([]byte, error) {
	tssMsg := &TssMessage{
		IsBrodcast: isBrodcast,
		From:       from,
		MsgBytes:   msgBytes,
	}

	msgBytes, err := json.Marshal(tssMsg)
	if err != nil {
		return []byte{}, err
	}

	return msgBytes, nil
}

func UnmarshalTssMessage(msgBytes []byte) (*TssMessage, error) {
	msg := &TssMessage{}
	err := json.Unmarshal(msgBytes, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

type InitiateMessage struct {
	SessionID string `json:"sessionID"`
	Process   string `json:"process"`
}

func MarhsalInitiateMessage(sessionID string, process string) ([]byte, error) {
	initiateMsg := &InitiateMessage{
		SessionID: sessionID,
		Process:   process,
	}

	msgBytes, err := json.Marshal(initiateMsg)
	if err != nil {
		return []byte{}, err
	}

	return msgBytes, nil
}

func UnmarshalInitiateMessage(msgBytes []byte) (*InitiateMessage, error) {
	msg := &InitiateMessage{}
	err := json.Unmarshal(msgBytes, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

type StartSignMessage struct {
	Peers []peer.ID `json:"peers"`
	Msg   *big.Int  `json:"msg"`
}

func MarshalStartSignMessage(peers []peer.ID, msg *big.Int) ([]byte, error) {
	startMsg := &StartSignMessage{
		Peers: peers,
		Msg:   msg,
	}

	msgBytes, err := json.Marshal(startMsg)
	if err != nil {
		return []byte{}, err
	}

	return msgBytes, nil
}

func UnmarshalStartSignMessage(msgBytes []byte) (*StartSignMessage, error) {
	msg := &StartSignMessage{}
	err := json.Unmarshal(msgBytes, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

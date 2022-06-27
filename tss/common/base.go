package common

import (
	p2p "diplomski/p2p"
	"fmt"

	"github.com/binance-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p-core/peer"
	host "github.com/libp2p/go-libp2p-host"
)

type BaseTss struct {
	Host          host.Host
	SessionID     string
	PartyID       *tss.PartyID
	Party         tss.Party
	PartyStore    map[string]*tss.PartyID
	Communication p2p.Communication
	Peers         []peer.ID
}

func (tss *BaseTss) Start() error {
	return nil
}

func (tss *BaseTss) PopulatePartyStore(parties tss.SortedPartyIDs) {
	for _, party := range parties {
		tss.PartyStore[party.Id] = party
	}
}

func (tss *BaseTss) IsParticipant(parties tss.SortedPartyIDs) bool {
	for _, party := range parties {
		if party.Id == tss.PartyID.Id {
			return true
		}
	}

	return false
}

// ProcessOutboundMessages listens to messages sent into channel by TSS process and sends them to corresponding peers.
func (tss *BaseTss) ProcessOutboundMessages(outChn chan tss.Message, messageType p2p.MessageType) {
	for {
		select {
		case msg := <-outChn:
			{
				msgBytes, routing, _ := msg.WireBytes()
				keygenMsgBytes, _ := MarshalTssMessage(msgBytes, routing.IsBroadcast, routing.From.Id)

				var peers peer.IDSlice
				if msg.IsBroadcast() {
					peers = tss.Peers
				} else {
					peers = GetPeersFromParties(msg.GetTo())
				}

				go tss.Communication.Broadcast(peers, keygenMsgBytes, messageType, tss.SessionID)
			}
		}
	}
}

// ProcessInboundMessages listens to incoming messages and updates TSS process state.
func (tss *BaseTss) ProcessInboundMessages(msgChan chan *p2p.WrappedMessage) {
	for {
		select {
		case wMsg := <-msgChan:
			{
				msg, _ := UnmarshalTssMessage(wMsg.Payload)
				go func() {
					ok, err := tss.Party.UpdateFromBytes(msg.MsgBytes, tss.PartyStore[msg.From], msg.IsBrodcast)
					if !ok && err != nil {
						fmt.Println(err)
					}
				}()
			}
		}
	}
}

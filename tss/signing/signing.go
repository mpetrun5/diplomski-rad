package signing

import (
	"fmt"
	"math/big"
	"time"

	p2p "github.com/mpetrun5/diplomski/p2p"
	"github.com/mpetrun5/diplomski/storage"
	"github.com/mpetrun5/diplomski/tss/common"

	"github.com/binance-chain/tss-lib/ecdsa/signing"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Signing struct {
	common.BaseTss
	endChn   chan *signing.SignatureData
	saveData storage.Keyshare
	parties  tss.SortedPartyIDs
}

func NewSigning(host host.Host, comm p2p.Communication, endChn chan *signing.SignatureData, keyshare storage.Keyshare) *Signing {
	partyID := common.CreatePartyID(host.ID().String())
	partyStore := make(map[string]*tss.PartyID)
	return &Signing{
		BaseTss: common.BaseTss{
			PartyStore:    partyStore,
			Host:          host,
			PartyID:       partyID,
			Communication: comm,
		},
		saveData: keyshare,
		endChn:   endChn,
	}
}

// Start starts a signing TSS process. Signature will be sent to end channel
// provided when creating the Signing struct.
func (s *Signing) Start(msg *big.Int, peers peer.IDSlice) error {
	s.parties = common.GetParties(peers)
	s.PopulatePartyStore(s.parties)

	outChn := make(chan tss.Message)
	msgChan := make(chan *p2p.WrappedMessage)

	s.Communication.Subscribe(p2p.TssKeySignMsg, s.SessionID, msgChan)
	go s.ProcessOutboundMessages(outChn, p2p.TssKeySignMsg)
	go s.ProcessInboundMessages(msgChan)

	ctx := tss.NewPeerContext(s.parties)
	params := tss.NewParameters(ctx, s.PartyStore[s.PartyID.Id], len(s.saveData.Peers), s.saveData.Threshold)
	s.Party = signing.NewLocalParty(msg, params, s.saveData.Key, outChn, s.endChn)

	go func() {
		err := s.Party.Start()
		if err != nil {
			fmt.Println(err)
		}
	}()
	return nil
}

// Initiate initiates a signing TSS process.
func (s *Signing) Initiate(msg *big.Int) error {
	s.SessionID = common.CalculateMsgID(msg)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTicker(15 * time.Second)
	defer timeout.Stop()

	readyChan := make(chan *p2p.WrappedMessage)
	readyMap := make(map[peer.ID]bool)
	readyMap[s.Host.ID()] = true

	s.Communication.Subscribe(p2p.TssReadyMsg, s.SessionID, readyChan)
	defer s.Communication.UnSubscribe(p2p.TssReadyMsg, s.SessionID)

	msgBytes, _ := common.MarhsalInitiateMessage(s.SessionID, "signing")
	go s.Communication.Broadcast(s.saveData.Peers, msgBytes, p2p.TssInitiateMsg, "initiate")
	for {
		select {
		case wMsg := <-readyChan:
			{
				readyMap[wMsg.From] = true
				if len(readyMap) == s.saveData.Threshold+1 {
					s.Peers = s.peerSubset(readyMap)
					startMsgBytes, err := common.MarshalStartSignMessage(s.Peers, msg)
					if err != nil {
						fmt.Println(err)
					}

					go s.Communication.Broadcast(s.Host.Peerstore().Peers(), startMsgBytes, p2p.TssStartMsg, s.SessionID)
					if s.IsParticipant(common.GetParties(s.Peers)) {
						go s.Start(msg, s.Peers)
					}

					return nil
				}
			}
		case <-ticker.C:
			{
				go s.Communication.Broadcast(s.Host.Peerstore().Peers(), []byte{}, p2p.TssInitiateMsg, s.SessionID)
			}
		case <-timeout.C:
			timeout.Stop()
			return nil
		}
	}
}

// WaitForStart waits for signing process TSS message.
func (s *Signing) WaitForStart() {
	timeout := time.NewTicker(15 * time.Second)
	defer timeout.Stop()

	msgChan := make(chan *p2p.WrappedMessage)
	startMsgChn := make(chan *p2p.WrappedMessage)

	s.Communication.Subscribe(p2p.TssInitiateMsg, s.SessionID, msgChan)
	defer s.Communication.UnSubscribe(p2p.TssInitiateMsg, s.SessionID)
	s.Communication.Subscribe(p2p.TssStartMsg, s.SessionID, startMsgChn)
	defer s.Communication.UnSubscribe(p2p.TssStartMsg, s.SessionID)

	for {
		select {
		case wMsg := <-startMsgChn:
			{
				startMsg, err := common.UnmarshalStartSignMessage(wMsg.Payload)
				if err != nil {
					fmt.Println(err)
				}

				s.Peers = startMsg.Peers
				if s.IsParticipant(common.GetParties(s.Peers)) {
					go s.Start(startMsg.Msg, startMsg.Peers)
				}

				return
			}
		case <-timeout.C:
			timeout.Stop()
			return
		}
	}
}

func (s *Signing) peerSubset(readyPeers map[peer.ID]bool) peer.IDSlice {
	peers := []peer.ID{}
	for peer := range readyPeers {
		peers = append(peers, peer)
	}

	return peers
}

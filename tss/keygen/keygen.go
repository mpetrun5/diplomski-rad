package keygen

import (
	p2p "diplomski/p2p"
	storage "diplomski/storage"
	"diplomski/tss/common"
	"diplomski/util"
	"fmt"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

type SaveDataStorer interface {
	SetSaveData(saveData storage.Keyshare) error
}

type Keygen struct {
	common.BaseTss
	storer    SaveDataStorer
	threshold int
}

func NewKeygen(host host.Host, comm p2p.Communication, storer SaveDataStorer, threshold int) *Keygen {
	partyID := common.CreatePartyID(host.ID().String())
	partyStore := make(map[string]*tss.PartyID)
	return &Keygen{
		BaseTss: common.BaseTss{
			PartyStore:    partyStore,
			Host:          host,
			PartyID:       partyID,
			Communication: comm,
			Peers:         host.Peerstore().Peers(),
		},
		storer:    storer,
		threshold: threshold,
	}
}

// Start stars a key generation TSS process.
func (k *Keygen) Start() error {
	parties := common.GetParties(k.Host.Peerstore().Peers())
	k.PopulatePartyStore(parties)
	ctx := tss.NewPeerContext(parties)
	params := tss.NewParameters(ctx, k.PartyStore[k.PartyID.Id], len(k.Host.Peerstore().Peers()), k.threshold)

	outChn := make(chan tss.Message)
	msgChn := make(chan *p2p.WrappedMessage)
	endChn := make(chan keygen.LocalPartySaveData)

	k.Communication.Subscribe(p2p.TssKeyGenMsg, k.SessionID, msgChn)

	go k.ProcessOutboundMessages(outChn, p2p.TssKeyGenMsg)
	go k.ProcessInboundMessages(msgChn)
	go k.processEndMessage(endChn)
	k.Party = keygen.NewLocalParty(params, outChn, endChn)

	go func() {
		err := k.Party.Start()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

// Initiate initiates a key generation TSS process.
func (k *Keygen) Initiate() error {
	k.SessionID = util.RandomString(32)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	readyChan := make(chan *p2p.WrappedMessage)
	readyMap := make(map[peer.ID]bool)
	readyMap[k.Host.ID()] = true

	k.Communication.Subscribe(p2p.TssReadyMsg, k.SessionID, readyChan)
	defer k.Communication.UnSubscribe(p2p.TssReadyMsg, k.SessionID)

	msgBytes, _ := common.MarhsalInitiateMessage(k.SessionID, "keygen")
	go k.Communication.Broadcast(k.Host.Peerstore().Peers(), msgBytes, p2p.TssInitiateMsg, "initiate")
	for {
		select {
		case wMsg := <-readyChan:
			{
				readyMap[wMsg.From] = true
				if len(readyMap) == len(k.Host.Peerstore().Peers()) {
					go k.Communication.Broadcast(k.Host.Peerstore().Peers(), []byte{}, p2p.TssStartMsg, k.SessionID)
					go k.Start()
					return nil
				}
			}
		case <-ticker.C:
			{
				go k.Communication.Broadcast(k.Host.Peerstore().Peers(), []byte{}, p2p.TssInitiateMsg, k.SessionID)
			}
		}
	}
}

// WaitForStart waits for key generation start message and starts TSS process.
func (s *Keygen) WaitForStart() {
	msgChan := make(chan *p2p.WrappedMessage)
	startMsgChn := make(chan *p2p.WrappedMessage)

	s.Communication.Subscribe(p2p.TssInitiateMsg, s.SessionID, msgChan)
	defer s.Communication.UnSubscribe(p2p.TssInitiateMsg, s.SessionID)
	s.Communication.Subscribe(p2p.TssStartMsg, s.SessionID, startMsgChn)
	defer s.Communication.UnSubscribe(p2p.TssStartMsg, s.SessionID)

	for {
		select {
		case <-startMsgChn:
			{
				go s.Start()
			}
		}
	}
}

func (k *Keygen) processEndMessage(endChn chan keygen.LocalPartySaveData) {
	for {
		select {
		case key := <-endChn:
			{
				saveData := storage.Keyshare{
					Peers:     k.Peers,
					Threshold: k.threshold,
					Key:       key,
				}
				err := k.storer.SetSaveData(saveData)
				if err != nil {
					fmt.Println(err)
				}

				k.Communication.UnSubscribe(p2p.TssReadyMsg, k.SessionID)

				fmt.Println("================================")
				fmt.Println("Successful keygen")
				fmt.Println("================================")
			}
		}
	}
}

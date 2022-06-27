package resharing

import (
	"fmt"
	"time"

	p2p "github.com/mpetrun5/diplomski/p2p"
	storage "github.com/mpetrun5/diplomski/storage"
	"github.com/mpetrun5/diplomski/tss/common"
	"github.com/mpetrun5/diplomski/util"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/ecdsa/resharing"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

type SaveDataStorer interface {
	GetSaveData() (storage.Keyshare, error)
	SetSaveData(saveData storage.Keyshare) error
}

type Resharing struct {
	common.BaseTss
	storer       SaveDataStorer
	newThreshold int
}

func NewResharing(host host.Host, comm p2p.Communication, storer SaveDataStorer, newThreshold int) *Resharing {
	partyID := common.CreatePartyID(host.ID().String())
	partyStore := make(map[string]*tss.PartyID)

	return &Resharing{
		BaseTss: common.BaseTss{
			PartyStore:    partyStore,
			Host:          host,
			PartyID:       partyID,
			Communication: comm,
			Peers:         host.Peerstore().Peers(),
		},
		storer: storer,
	}
}

// Start starts a resharing TSS process. Old parameters and read from keyshare and
// new resharing parameters and read from config.
func (rs *Resharing) Start() error {
	allParties := common.GetParties(rs.Host.Peerstore().Peers())

	saveData, err := rs.storer.GetSaveData()
	var key keygen.LocalPartySaveData
	var oldParties tss.SortedPartyIDs
	if err != nil {
		key = keygen.NewLocalPartySaveData(len(allParties))
		oldParties = common.GetParties(common.ExcludePeer(rs.Host.Peerstore().Peers(), rs.Host.ID()))
	} else {
		key = saveData.Key
		oldParties = common.GetParties(saveData.Peers)
	}
	newParties := rs.sortedNewParties(allParties, oldParties)
	fmt.Println(newParties)
	rs.PopulatePartyStore(newParties)

	oldCtx := tss.NewPeerContext(oldParties)
	newCtx := tss.NewPeerContext(newParties)
	params := tss.NewReSharingParameters(oldCtx, newCtx, rs.PartyStore[rs.PartyID.Id], len(oldParties), saveData.Threshold, len(newParties), rs.newThreshold)

	outChn := make(chan tss.Message)
	msgChn := make(chan *p2p.WrappedMessage)
	endChn := make(chan keygen.LocalPartySaveData)

	rs.Communication.Subscribe(p2p.TssReshareMsg, rs.SessionID, msgChn)
	go rs.ProcessOutboundMessages(outChn, p2p.TssReshareMsg)
	go rs.ProcessInboundMessages(msgChn)
	go rs.processEndMessage(endChn)

	rs.Party = resharing.NewLocalParty(params, key, outChn, endChn)

	go func() {
		err := rs.Party.Start()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

// Initiate initiates a resharing TSS process.
func (rs *Resharing) Initiate() error {
	rs.SessionID = util.RandomString(32)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	readyChan := make(chan *p2p.WrappedMessage)
	readyMap := make(map[peer.ID]bool)
	readyMap[rs.Host.ID()] = true

	rs.Communication.Subscribe(p2p.TssReadyMsg, rs.SessionID, readyChan)
	defer rs.Communication.UnSubscribe(p2p.TssReadyMsg, rs.SessionID)

	msgBytes, _ := common.MarhsalInitiateMessage(rs.SessionID, "resharing")
	go rs.Communication.Broadcast(rs.Host.Peerstore().Peers(), msgBytes, p2p.TssInitiateMsg, rs.SessionID)
	for {
		select {
		case wMsg := <-readyChan:
			{
				readyMap[wMsg.From] = true
				if len(readyMap) == len(rs.Host.Peerstore().Peers()) {
					go rs.Communication.Broadcast(rs.Host.Peerstore().Peers(), []byte{}, p2p.TssStartMsg, rs.SessionID)
					go rs.Start()
					return nil
				}
			}
		case <-ticker.C:
			{
				go rs.Communication.Broadcast(rs.Host.Peerstore().Peers(), []byte{}, p2p.TssInitiateMsg, rs.SessionID)
			}
		}
	}
}

// WaitForStart waits for resharing start message and starts TSS process.
func (rs *Resharing) WaitForStart() {
	msgChan := make(chan *p2p.WrappedMessage)
	startMsgChn := make(chan *p2p.WrappedMessage)

	rs.Communication.Subscribe(p2p.TssInitiateMsg, rs.SessionID, msgChan)
	defer rs.Communication.UnSubscribe(p2p.TssInitiateMsg, rs.SessionID)
	rs.Communication.Subscribe(p2p.TssStartMsg, rs.SessionID, startMsgChn)
	defer rs.Communication.UnSubscribe(p2p.TssStartMsg, rs.SessionID)

	for {
		select {
		case wMsg := <-msgChan:
			{
				go rs.Communication.Broadcast(peer.IDSlice{wMsg.From}, []byte{}, p2p.TssReadyMsg, rs.SessionID)
			}
		case <-startMsgChn:
			{
				go rs.Start()
				return
			}
		}
	}
}

func (rs *Resharing) processEndMessage(endChn chan keygen.LocalPartySaveData) {
	for {
		select {
		case key := <-endChn:
			{
				saveData := storage.Keyshare{
					Peers:     rs.Peers,
					Threshold: rs.newThreshold,
					Key:       key,
				}
				err := rs.storer.SetSaveData(saveData)
				if err != nil {
					fmt.Println(err)
				}

				rs.Communication.UnSubscribe(p2p.TssReshareMsg, rs.SessionID)
			}
		}
	}
}

func (rs *Resharing) sortedNewParties(allParties tss.SortedPartyIDs, oldParties tss.SortedPartyIDs) tss.SortedPartyIDs {
	newParties := make(tss.SortedPartyIDs, len(allParties))
	copy(newParties, oldParties)

	index := len(oldParties)
	for _, party := range allParties {
		if !common.IsParticipant(party, oldParties) {
			newParties[index] = party
			newParties[index].Index = index
			index++
		}
	}

	return newParties
}

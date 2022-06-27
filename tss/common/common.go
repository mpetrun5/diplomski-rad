package common

import (
	"fmt"
	"math/big"

	"github.com/binance-chain/tss-lib/tss"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

func CreatePartyID(peerID string) *tss.PartyID {
	key := big.NewInt(0).SetBytes([]byte(peerID))
	return tss.NewPartyID(peerID, peerID, key)

}

func GetParties(peers peer.IDSlice) tss.SortedPartyIDs {
	unsortedParties := make(tss.UnSortedPartyIDs, len(peers))

	for i, peer := range peers {
		unsortedParties[i] = CreatePartyID(peer.String())
	}

	return tss.SortPartyIDs(unsortedParties)
}

func GetPeersFromParties(parties []*tss.PartyID) []peer.ID {
	peers := make([]peer.ID, len(parties))
	for i, party := range parties {
		peerID, err := peer.Decode(party.Id)
		if err != nil {
			fmt.Println(err)
		}

		peers[i] = peerID
	}

	return peers
}

func CalculateMsgID(msg *big.Int) string {
	return common.Bytes2Hex(crypto.Keccak256(msg.Bytes()))
}

func ExcludePeer(peers peer.IDSlice, excluded peer.ID) peer.IDSlice {
	for i, peer := range peers {
		if peer.Pretty() == excluded.Pretty() {
			fmt.Println(peers)
			fmt.Println(i)
			return append(peers[:i], peers[i+1:]...)
		}
	}

	return peers
}

func IsParticipant(party *tss.PartyID, parties tss.SortedPartyIDs) bool {
	for _, existingParty := range parties {
		if party.Id == existingParty.Id {
			return true
		}
	}

	return false
}

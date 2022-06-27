package common

import (
	"reflect"
	"testing"

	"github.com/binance-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p-core/peer"
)

func TestGetPeersFromParties(t *testing.T) {
	type args struct {
		parties []*tss.PartyID
	}
	tests := []struct {
		name string
		args args
		want []peer.ID
	}{
		{
			name: "test",
			args: args{parties: []*tss.PartyID{CreatePartyID("QmcW3oMdSqoEcjbyd51auqC23vhKX6BqfcZcY2HJ3sKAZR")}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPeersFromParties(tt.args.parties); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPeersFromParties() = %v, want %v", got, tt.want)
			}
		})
	}
}

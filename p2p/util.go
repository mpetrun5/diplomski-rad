package p2p

import (
	"context"
	config "diplomski/config"
	"diplomski/p2p/discovery"
	"diplomski/util"

	"github.com/libp2p/go-libp2p-core/host"
)

// SetupCommunication is a helper method to configure host, communication and discovery.
func SetupCommunication(conf *config.Config) (host.Host, Communication, error) {
	privKey, err := util.LoadPrivateKey(conf.Key)
	if err != nil {
		return nil, nil, err
	}

	host, err := NewHost(privKey, conf)
	if err != nil {
		return nil, nil, err
	}
	kdht, err := discovery.NewKDHT(context.Background(), host, conf.BootstrapPeers)
	if err != nil {
		return nil, nil, err
	}
	go discovery.Discover(context.Background(), host, kdht, "")

	comm, err := NewCommunication(host, "/p2p/tss")
	if err != nil {
		return nil, nil, err
	}

	return host, comm, nil
}

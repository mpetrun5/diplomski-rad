package p2p

import (
	"fmt"

	config "github.com/mpetrun5/diplomski-rad/config"
	util "github.com/mpetrun5/diplomski-rad/util"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	noise "github.com/libp2p/go-libp2p-noise"
	"github.com/rs/zerolog/log"
)

// NewHost initiates libp2p host on given port and configures stream Noise protocol.
func NewHost(privKey crypto.PrivKey, rconf *config.Config) (host.Host, error) {
	logger := log.With().Logger()
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", rconf.Port)),
		libp2p.Identity(privKey),
		libp2p.DisableRelay(),
		libp2p.Security(noise.ID, noise.New),
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		logger.Error().Msg(
			"unable to create libp2p host",
		)
		return nil, err
	}

	fullAddr := util.GetHostAddress(h)
	logger.Info().Msg(
		fmt.Sprintf("host %s created", fullAddr),
	)

	return h, nil
}

package cli

import (
	"github.com/mpetrun5/diplomski/config"
	"github.com/mpetrun5/diplomski/p2p"
	"github.com/mpetrun5/diplomski/storage"
	"github.com/mpetrun5/diplomski/tss/common"
	"github.com/mpetrun5/diplomski/tss/keygen"
	"github.com/mpetrun5/diplomski/tss/resharing"
	"github.com/mpetrun5/diplomski/tss/signing"

	tssSigning "github.com/binance-chain/tss-lib/ecdsa/signing"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listenCMD = &cobra.Command{
		Use:   "listen",
		Short: "Run node in listener mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			listen()
			return nil
		},
	}
)

func init() {
}

func listen() {
	configPath := viper.GetString("config")
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}
	host, comm, err := p2p.SetupCommunication(conf)
	if err != nil {
		panic(err)
	}
	saveDataStorage := storage.NewSaveDataStorage(conf.KeysharePath)

	msgChan := make(chan *p2p.WrappedMessage)
	comm.Subscribe(p2p.TssInitiateMsg, "initiate", msgChan)
	for {
		select {
		case wMsg := <-msgChan:
			{
				initiateMsg, _ := common.UnmarshalInitiateMessage(wMsg.Payload)
				switch initiateMsg.Process {
				case "signing":
					{
						keyshare, _ := saveDataStorage.GetSaveData()
						s := signing.NewSigning(host, comm, make(chan *tssSigning.SignatureData, 1), keyshare)
						s.SessionID = initiateMsg.SessionID
						go s.WaitForStart()
					}
				case "keygen":
					{
						kg := keygen.NewKeygen(host, comm, saveDataStorage, conf.Threshold)
						kg.SessionID = initiateMsg.SessionID
						go kg.WaitForStart()
					}
				case "resharing":
					{
						rs := resharing.NewResharing(host, comm, saveDataStorage, conf.Threshold)
						rs.SessionID = initiateMsg.SessionID
						go rs.WaitForStart()
					}
				}
			}
		}
	}
}

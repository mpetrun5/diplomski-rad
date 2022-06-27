package cli

import (
	"math/big"

	"github.com/mpetrun5/diplomski-rad/config"
	"github.com/mpetrun5/diplomski-rad/executor"
	"github.com/mpetrun5/diplomski-rad/p2p"
	"github.com/mpetrun5/diplomski-rad/storage"
	"github.com/mpetrun5/diplomski-rad/transactor"
	"github.com/mpetrun5/diplomski-rad/tss/signing"

	tssSigning "github.com/binance-chain/tss-lib/ecdsa/signing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	network string
	to      string
	data    []byte
	value   int64

	signCMD = &cobra.Command{
		Use:   "send-transaction",
		Short: "Initiate a signature generation TSS process and send a transaction on the Ethereum network.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sendTransaction()
			return nil
		},
	}
)

func init() {
	signCMD.Flags().StringVar(&network, "network", "", "RPC endpoint of the blockchain network")
	signCMD.Flags().StringVar(&to, "to", "", "Address to which is transaction is to be sent")
	signCMD.Flags().BytesHexVar(&data, "data", []byte{}, "Transaction data")
	signCMD.Flags().Int64Var(&value, "value", 0, "Transaction value")
}

func sendTransaction() {
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
	keyshare, err := saveDataStorage.GetSaveData()
	if err != nil {
		panic(err)
	}

	t, err := transactor.NewTransactor(
		network,
		crypto.PubkeyToAddress(*keyshare.Key.ECDSAPub.ToECDSAPubKey()),
	)
	if err != nil {
		panic(err)
	}

	signatureChn := make(chan *tssSigning.SignatureData)
	signing := signing.NewSigning(host, comm, signatureChn, keyshare)
	exec := executor.NewExecutor(t, *signing, signatureChn)
	err = exec.Execute(common.HexToAddress(to), big.NewInt(value), data)
	if err != nil {
		panic(err)
	}
}

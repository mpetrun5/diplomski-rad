package cli

import (
	"github.com/mpetrun5/diplomski/config"
	"github.com/mpetrun5/diplomski/p2p"
	"github.com/mpetrun5/diplomski/storage"
	"github.com/mpetrun5/diplomski/tss/keygen"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	generateKeyCMD = &cobra.Command{
		Use:   "generate-key",
		Short: "Initiate a key generation TSS process.",
		RunE: func(cmd *cobra.Command, args []string) error {
			generateKey()
			return nil
		},
	}
)

func init() {
}

func generateKey() {
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

	kg := keygen.NewKeygen(host, comm, saveDataStorage, conf.Threshold)
	err = kg.Initiate()
	if err != nil {
		panic(err)
	}
}

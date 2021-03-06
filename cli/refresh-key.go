package cli

import (
	"github.com/mpetrun5/diplomski-rad/config"
	"github.com/mpetrun5/diplomski-rad/p2p"
	"github.com/mpetrun5/diplomski-rad/storage"
	"github.com/mpetrun5/diplomski-rad/tss/resharing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	refreshKeyCMD = &cobra.Command{
		Use:   "refresh-key",
		Short: "Initiate a key refresh TSS process.",
		RunE: func(cmd *cobra.Command, args []string) error {
			refreshKey()
			return nil
		},
	}
)

func init() {
}

func refreshKey() {
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

	kg := resharing.NewResharing(host, comm, saveDataStorage, conf.Threshold)
	err = kg.Initiate()
	if err != nil {
		panic(err)
	}
}

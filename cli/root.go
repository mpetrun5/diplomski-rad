package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCMD = &cobra.Command{
		Use: "",
	}
)

func init() {
	rootCMD.PersistentFlags().String("config", ".", "Path to JSON configuration file")
	_ = viper.BindPFlag("config", rootCMD.PersistentFlags().Lookup("config"))
}

// Execute register commands:
//  - listen
//  - generate-key
//  - refresh-key
//  - send-transaction
func Execute() {
	rootCMD.AddCommand(listenCMD, generateKeyCMD, refreshKeyCMD, signCMD)
	if err := rootCMD.Execute(); err != nil {
		panic(err)
	}
}

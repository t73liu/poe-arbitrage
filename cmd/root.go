package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var initialConfig = "https://raw.githubusercontent.com/t73liu/poe-arbitrage/master/default-config.json"
var defaultConfigFileName = "poe-arbitrage.json"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "poe-arbitrage",
	Short: "poe-arbitrage checks for bulk trading opportunities",
	Long: `
POE Arbitrage is a CLI that checks for inefficient bid-ask spreads for
bulk-item trades. It relies on the official POE Bulk Item Exchange
(https://www.pathofexile.com/trade/exchange) and is subject to rate-limits.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(
	//  &cfgFile,
	//  "config",
	//  "",
	//  fmt.Sprintf(
	//    "config file (default is $HOME%s)",
	//    string(filepath.Separator) + defaultConfigFile,
	//  ),
	//)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP(
	//  "toggle",
	//  "t",
	//  false,
	//  "Help message for toggle",
	//)
}

func initConfig() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.AutomaticEnv()
	viper.SetConfigType("json")

	customConfigFile := viper.GetString("POE_ARBITRAGE_CONFIG")
	isCustomConfig := customConfigFile != ""
	defaultConfigFilePath := filepath.Join(home, defaultConfigFileName)

	if isCustomConfig {
		viper.SetConfigFile(customConfigFile)
	} else {
		viper.SetConfigFile(defaultConfigFilePath)
	}

	err = viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else if isCustomConfig {
		fmt.Println("Could not find config file:", customConfigFile)
	} else {
		fmt.Println("Initializing default config file:", defaultConfigFilePath)
		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 5 * time.Second,
			},
		}

		resp, err := client.Get(initialConfig)
		defer resp.Body.Close()

		if err != nil {
			fmt.Println("Unable to download default config", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Unable to download default config", resp.Status)
			os.Exit(1)
		}

		if err := viper.ReadConfig(resp.Body); err != nil {
			fmt.Println("Unable to read the default config", err)
			os.Exit(1)
		}

		if err := viper.WriteConfig(); err != nil {
			fmt.Println("Unable to create default config", err)
			os.Exit(1)
		}
	}
}

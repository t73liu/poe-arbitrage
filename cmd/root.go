package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path/filepath"
	"poe-arbitrage/utils"
	"strings"
	"time"
)

const initialConfig = "https://raw.githubusercontent.com/t73liu/poe-arbitrage/master/default-config.json"
const defaultConfigFileName = "poe-arbitrage.json"

var customConfigFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "poe-arbitrage",
	Short: "poe-arbitrage checks for bulk trading opportunities",
	Long: `
poe-arbitrage is a CLI that checks for inefficient bid-ask
spreads for bulk-item trades. It relies on the official PoE
Bulk Item Exchange (https://www.pathofexile.com/trade/exchange)
and is subject to its rate-limits.
`,
	Version: "1.0.0",
}

type BulkItem struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	StackSize uint16 `json:"stackSize"`
}

type Config struct {
	League          string              `json:"league"`
	Hardcore        bool                `json:"hardcore"`
	ExcludeAFK      bool                `json:"excludeAFK"`
	IgnoredPlayers  []string            `json:"ignoredPlayers"`
	FavoritePlayers []string            `json:"favoritePlayers"`
	BulkItems       map[string]BulkItem `json:"bulkItems"`
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

	rootCmd.PersistentFlags().StringVar(
		&customConfigFile,
		"config",
		"",
		"config file (default is $HOME/poe-arbitrage.json)",
	)
}

// cobra.OnInitialize does not support functions that return errors
func initConfig() {
	if err := initConfigE(); err != nil {
		os.Exit(1)
	}
}

func initConfigE() error {
	viper.AutomaticEnv()
	viper.SetConfigType("json")

	customConfigFile = strings.TrimSpace(customConfigFile)

	if customConfigFile != "" {
		viper.SetConfigFile(customConfigFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Could not open config file:", err)
			return err
		}
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("Failed to detect home directory:", err)
			return err
		}

		defaultConfigFilePath := filepath.Join(home, defaultConfigFileName)
		viper.SetConfigFile(defaultConfigFilePath)

		if utils.FileExists(defaultConfigFilePath) {
			if err := viper.ReadInConfig(); err != nil {
				fmt.Println("Could not open config file:", err)
				return err
			}
		} else {
			fmt.Println("Initializing default config file:", defaultConfigFilePath)
			client := http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					TLSHandshakeTimeout: 5 * time.Second,
				},
			}

			resp, err := client.Get(initialConfig)
			if err != nil {
				fmt.Println("Unable to download default config:", err)
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println("Unable to download default config:", resp.Status)
				return err
			}

			if err := viper.ReadConfig(resp.Body); err != nil {
				fmt.Println("Unable to read the default config:", err)
				return err
			}

			if err := viper.WriteConfig(); err != nil {
				fmt.Println("Unable to create default config:", err)
				return err
			}
		}
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	return nil
}

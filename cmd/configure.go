package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"poe-arbitrage/utils"
	"strconv"
	"strings"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the CLI with various settings",
	Long: `
Update the CLI configuration with some common commands.
Update the config file directly for more custom operations.

The default config location is "$HOME/poe-arbitrage.json".
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var config Config
		if err := viper.Unmarshal(&config); err != nil {
			fmt.Println("Failed to parse config:", err)
			return err
		}

		leagueUpdated := cmd.Flags().Changed("league")
		if leagueUpdated {
			league, err := cmd.Flags().GetString("league")
			if err != nil {
				fmt.Println("Failed to parse --league:", err)
				return err
			}
			league = strings.TrimSpace(league)
			if league == "" {
				return errors.New("invalid league name")
			}
			config.League = league
		}

		hardcoreUpdated := cmd.Flags().Changed("hardcore")
		if hardcoreUpdated {
			hardcore, err := cmd.Flags().GetBool("hardcore")
			if err != nil {
				fmt.Println("Failed to parse --hardcore:", err)
				return err
			}
			config.Hardcore = hardcore
		}

		excludeAFKUpdated := cmd.Flags().Changed("exclude-afk")
		if excludeAFKUpdated {
			excludeAFK, err := cmd.Flags().GetBool("exclude-afk")
			if err != nil {
				fmt.Println("Failed to parse --exclude-afk:", err)
				return err
			}
			config.ExcludeAFK = excludeAFK
		}

		ignorePlayerUpdated := cmd.Flags().Changed("ignore-player")
		if ignorePlayerUpdated {
			ignoredPlayer, err := cmd.Flags().GetString("ignore-player")
			if err != nil {
				fmt.Println("Failed to parse --ignore-player:", err)
				return err
			}
			ignoredPlayer = strings.TrimSpace(ignoredPlayer)
			if ignoredPlayer == "" {
				return errors.New("invalid player name")
			}
			if utils.Contains(config.IgnoredPlayers, ignoredPlayer) {
				return errors.New(ignoredPlayer + " is already ignored")
			}
			config.IgnoredPlayers = append(config.IgnoredPlayers, ignoredPlayer)
		}

		favoritePlayerUpdated := cmd.Flags().Changed("favorite-player")
		if favoritePlayerUpdated {
			favoritePlayer, err := cmd.Flags().GetString("favorite-player")
			if err != nil {
				fmt.Println("Failed to parse --favorite-player:", err)
				return err
			}
			favoritePlayer = strings.TrimSpace(favoritePlayer)
			if favoritePlayer == "" {
				return errors.New("invalid player name")
			}
			if utils.Contains(config.FavoritePlayers, favoritePlayer) {
				return errors.New(favoritePlayer + " is already favorited")
			}
			config.FavoritePlayers = append(config.FavoritePlayers, favoritePlayer)
		}

		bulkItemUpdated := cmd.Flags().Changed("set-item")
		if bulkItemUpdated {
			itemSlice, err := cmd.Flags().GetStringSlice("set-item")
			if err != nil {
				fmt.Println("Failed to parse --set-item:", err)
				return err
			}

			if len(itemSlice) != 3 || utils.Contains(itemSlice, "") {
				return errors.New("invalid --item format, must provide id,name,stackSize")
			}

			stackSize, err := strconv.Atoi(itemSlice[2])
			if err != nil {
				fmt.Println("Unable to parse stack size from item:", err)
				return err
			}

			itemId := itemSlice[0]
			item := BulkItem{
				Id:        itemId,
				Name:      itemSlice[1],
				StackSize: uint(stackSize),
			}
			config.BulkItems[itemId] = item
		}

		configUpdated := leagueUpdated || hardcoreUpdated || excludeAFKUpdated ||
			ignorePlayerUpdated || favoritePlayerUpdated || bulkItemUpdated
		if configUpdated {
			jsonConfig, err := json.Marshal(config)
			if err != nil {
				fmt.Println("Unable to serialize the updated config:", err)
				return err
			}

			if err := viper.ReadConfig(bytes.NewBuffer(jsonConfig)); err != nil {
				fmt.Println("Unable to parse the updated config:", err)
				return err
			}

			if err := viper.WriteConfig(); err != nil {
				fmt.Println("Unable to write the updated config:", err)
				return err
			}

			fmt.Println("Config file updated.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.Flags().String(
		"league",
		"",
		"Set non-hardcore league name (capitalized)",
	)

	configureCmd.Flags().Bool(
		"hardcore",
		false,
		"Set league to hardcore version",
	)

	configureCmd.Flags().Bool(
		"exclude-afk",
		true,
		"Filter out AFK players from trades",
	)

	configureCmd.Flags().String(
		"ignore-player",
		"",
		"Add player to ignore list. Will be excluded from trades",
	)

	configureCmd.Flags().String(
		"favorite-player",
		"",
		"Add player to favorite list. Will be given preference in trades at same price",
	)

	configureCmd.Flags().StringSlice(
		"set-item",
		make([]string, 0),
		"Add/Update a bulk item in the CLI. Format is id,name,stackSize",
	)
}

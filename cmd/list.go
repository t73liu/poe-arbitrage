package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported bulk items",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Failed to retrieve --name value:", err)
			return err
		}

		var config Config
		if err := viper.Unmarshal(&config); err != nil {
			fmt.Println("Failed to parse config:", err)
			return err
		}

		trimmedSubstring := strings.ToLower(strings.TrimSpace(name))
		for _, item := range config.BulkItems {
			itemName := strings.ToLower(item.Name)
			if trimmedSubstring == "" || strings.Contains(itemName, trimmedSubstring) {
				fmt.Printf("%+v\n", item)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringP(
		"name",
		"n",
		"",
		"List items containing the provided string (case insensitive)",
	)
}

package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/Kschoffner13/GitSum/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage GitSum configuration",
}

var configSetKeyCmd = &cobra.Command{
	Use:   "set-key [key]",
	Short: "Store your Anthropic API key for future commands",
	Long: `Stores ANTHROPIC_API_KEY in GitSum's user-level config file so you
don't need to export it in every shell session.

An explicit "export ANTHROPIC_API_KEY=..." in your shell still takes
precedence over the stored value.

If you don't want the key to appear in your shell history, omit the
argument and you'll be prompted to enter it instead.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var key string
		if len(args) == 1 {
			key = args[0]
		} else {
			fmt.Fprint(cmd.OutOrStdout(), "Enter your Anthropic API key: ")
			line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
			if err != nil {
				return fmt.Errorf("reading key: %w", err)
			}
			key = strings.TrimSpace(line)
		}

		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		path, err := config.SetKey(key)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Saved ANTHROPIC_API_KEY to %s\n", path)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetKeyCmd)
	rootCmd.AddCommand(configCmd)
}

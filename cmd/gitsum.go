package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)



var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GitSum",
	Long:  `All software has versions. This is GitSum's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GitSum v0.1.0")
	},
}



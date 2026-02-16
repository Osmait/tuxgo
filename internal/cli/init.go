package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/josesaulburgos/tuxgo/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a local .tuxgo.yaml configuration file",
	Long: `Creates an example .tuxgo.yaml configuration file in the current
directory. Edit it to customize your tmux session layout.`,
	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	if err := config.InitLocalConfig(currentDir); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Created .tuxgo.yaml configuration file in current directory")
	fmt.Println("Edit it to customize your tmux session")
}

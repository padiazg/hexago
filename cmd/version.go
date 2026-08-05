/*
Copyright © 2026 Patricio Díaz <padiazg@gmail.com>
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/padiazg/hexago/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Shows HexaGo version",
		Run:   runVersion,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("simple", "s", false, "Prints only the version, useful for scripting")
}

func runVersion(cmd *cobra.Command, args []string) {
	simple, _ := cmd.Flags().GetBool("simple")
	if simple {
		fmt.Printf("%s", version.CurrentVersion().Version)
		return
	}

	var (
		stdWriter io.Writer = os.Stdout
		errWriter io.Writer = os.Stderr
	)

	version.Splash(stdWriter, errWriter)
}

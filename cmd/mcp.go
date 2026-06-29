/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"github.com/padiazg/hexago/internal/mcp"
	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the HexaGo MCP server (stdio)",
	Long: `Start HexaGo as a Model Context Protocol (MCP) server over stdio.

AI assistants (Claude Code, Claude Desktop, etc.) can use this server to scaffold
hexagonal architecture projects without leaving their conversation.

Each MCP tool delegates to the hexago CLI with --working-directory, so all
generation logic is shared with the regular CLI commands.

Register with Claude Code:
  claude mcp add hexago -- hexago mcp

Or scoped to a project:
  claude mcp add --scope project hexago -- hexago mcp`,
	RunE: runMCPServer,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	return mcp.Run()
}

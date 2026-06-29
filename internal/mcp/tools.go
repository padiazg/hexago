package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
)

// toolResult converts runSelf output into an MCP CallToolResult.
// Errors are reported as tool-level errors (IsError=true) so the LLM can see them.
func toolResult(out string, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("ERROR: %v\n\n%s", err, out)), nil
	}
	return mcp.NewToolResultText(out), nil
}

// runSelf executes the current hexago binary with the given arguments and returns
// the combined stdout+stderr output.
func runSelf(ctx context.Context, args ...string) (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot resolve hexago binary: %w", err)
	}
	c := exec.CommandContext(ctx, self, args...)
	out, err := c.CombinedOutput()
	return string(out), err
}

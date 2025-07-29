package tools

import (
	"context"
	"github.com/kagent-dev/tools/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"kyverno-agent/internal/commands"
)

func handleKyvernoApplyPolicy(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	policy := mcp.ParseString(request, "policy", "")

	if policy == "" {
		return mcp.NewToolResultError("Policy not specified"), nil
	}

	cmd := []string{"apply", "--cluster", "--policy-report", policy}

	output, err := runKyvernoCommand(ctx, cmd)
	if err != nil {
		return mcp.NewToolResultErrorf("Error applying kyverno policy: %v", err), nil
	}

	return mcp.NewToolResultText(output), nil
}

func runKyvernoCommand(ctx context.Context, cmd []string) (string, error) {
	kubeconfigPath := utils.GetKubeconfig()
	return commands.
		NewCommandBuilder("kyverno").
		WithArgs(cmd...).
		WithKubeconfig(kubeconfigPath).
		Execute(ctx)
}

func runKubectlCommand(ctx context.Context, cmd []string) (string, error) {
	kubeconfigPath := utils.GetKubeconfig()
	return commands.
		NewCommandBuilder("kubectl").
		WithArgs(cmd...).
		WithKubeconfig(kubeconfigPath).
		Execute(ctx)
}

func RegisterTools(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("kyverno_apply_policy"), handleKyvernoApplyPolicy)
}

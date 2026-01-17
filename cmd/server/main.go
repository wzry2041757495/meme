package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/shadow/meme/internal/core"
	"github.com/shadow/meme/internal/sources"
	"github.com/shadow/meme/internal/tools"
)

func main() {
	// 创建注册中心
	registry := core.NewRegistry()

	// 从环境变量读取配置
	config := &sources.Config{
		DouyinCookie:  os.Getenv("DOUYIN_COOKIE"),
		ImageProxyURL: os.Getenv("IMAGE_PROXY_URL"),
	}

	// 注册所有源
	sources.RegisterAllSources(registry, config)

	// 创建 MCP Server
	s := server.NewMCPServer(
		"meme-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// 注册 Tools
	s.AddTool(tools.NewSearchMemeTool(registry), tools.HandleSearchMeme(registry))
	s.AddTool(tools.NewListSourcesTool(), tools.HandleListSources(registry))

	// 启动 Stdio 服务
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/shadow/meme/internal/core"
	"github.com/shadow/meme/internal/sources"
)

// SearchMemeArgs search_meme 工具的参数
type SearchMemeArgs struct {
	Keyword string   `json:"keyword"`
	Sources []string `json:"sources,omitempty"` // 可选，指定搜索的源
	Page    int      `json:"page,omitempty"`
	Limit   int      `json:"limit,omitempty"`
}

// NewSearchMemeTool 创建 search_meme MCP Tool
func NewSearchMemeTool(registry *core.Registry) mcp.Tool {
	return mcp.NewTool(
		"search_meme",
		mcp.WithDescription("搜索表情包。支持从多个源并发搜索，返回去重后的结果列表。"),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("搜索关键词，如：猫、狗、开心、难过等"),
		),
		mcp.WithArray("sources",
			mcp.Description("可选，指定搜索的源ID列表。不指定则搜索所有源。可用源：qudoutu, doutula, pdan, sougou, douyin"),
		),
		mcp.WithNumber("page",
			mcp.Description("页码，默认为 1"),
		),
		mcp.WithNumber("limit",
			mcp.Description("每个源返回的最大数量，默认为 20"),
		),
	)
}

// HandleSearchMeme 处理 search_meme 请求
func HandleSearchMeme(registry *core.Registry) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 日志：接收到请求
		fmt.Fprintf(os.Stderr, "[SearchMeme] Request received. Params: %+v\n", request.Params.Arguments)

		// 解析参数
		var args SearchMemeArgs
		argsBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[SearchMeme] Failed to marshal params: %v\n", err)
			return mcp.NewToolResultError(fmt.Sprintf("参数解析失败: %v", err)), nil
		}
		if err := json.Unmarshal(argsBytes, &args); err != nil {
			fmt.Fprintf(os.Stderr, "[SearchMeme] Failed to unmarshal params: %v\n", err)
			return mcp.NewToolResultError(fmt.Sprintf("参数解析失败: %v", err)), nil
		}

		// 验证参数
		if args.Keyword == "" {
			fmt.Fprintf(os.Stderr, "[SearchMeme] Error: keyword is empty\n")
			return mcp.NewToolResultError("keyword 参数不能为空"), nil
		}

		// 设置默认值
		opts := core.DefaultSearchOptions()
		if args.Page > 0 {
			opts.Page = args.Page
		}
		if args.Limit > 0 {
			opts.Limit = args.Limit
		}

		fmt.Fprintf(os.Stderr, "[SearchMeme] Searching with args: keyword=%s, sources=%v, page=%d, limit=%d\n", args.Keyword, args.Sources, opts.Page, opts.Limit)

		// 执行搜索
		var result core.SearchResult
		if len(args.Sources) > 0 {
			result = registry.SearchSources(ctx, args.Keyword, args.Sources, opts)
		} else {
			result = registry.SearchAll(ctx, args.Keyword, opts)
		}

		fmt.Fprintf(os.Stderr, "[SearchMeme] Search completed. Found %d items. Duration: %dms\n", len(result.Memes), result.DurationMs)

		// 构造返回结果
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[SearchMeme] Failed to marshal result: %v\n", err)
			return mcp.NewToolResultError(fmt.Sprintf("结果序列化失败: %v", err)), nil
		}

		return mcp.NewToolResultText(string(resultJSON)), nil
	}
}

// NewListSourcesTool 创建 list_sources MCP Tool
func NewListSourcesTool() mcp.Tool {
	return mcp.NewTool(
		"list_sources",
		mcp.WithDescription("列出所有可用的表情包数据源及其信息"),
	)
}

// HandleListSources 处理 list_sources 请求
func HandleListSources(registry *core.Registry) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		fmt.Fprintf(os.Stderr, "[ListSources] Request received\n")

		infos := sources.GetAllSourceInfo(registry)

		fmt.Fprintf(os.Stderr, "[ListSources] Found %d sources\n", len(infos))

		resultJSON, err := json.MarshalIndent(infos, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ListSources] Failed to marshal result: %v\n", err)
			return mcp.NewToolResultError(fmt.Sprintf("结果序列化失败: %v", err)), nil
		}

		return mcp.NewToolResultText(string(resultJSON)), nil
	}
}

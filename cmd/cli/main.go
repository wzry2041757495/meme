package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shadow/meme/internal/core"
	"github.com/shadow/meme/internal/sources"
)

func main() {
	// 定义命令行参数
	keyword := flag.String("k", "", "搜索关键词 (必填)")
	sourceList := flag.String("s", "", "指定源，逗号分隔 (可选，如: pdan,qudoutu)")
	limit := flag.Int("l", 10, "每个源返回数量")
	page := flag.Int("p", 1, "页码")
	timeout := flag.Int("t", 15, "超时时间(秒)")
	listSources := flag.Bool("list", false, "列出所有可用源")
	outputJSON := flag.Bool("json", false, "输出 JSON 格式")
	verbose := flag.Bool("v", false, "显示详细信息")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Meme CLI - 表情包搜索命令行工具

用法:
  meme-cli -k <关键词> [选项]

示例:
  meme-cli -k 猫                    # 搜索 "猫" 相关表情包
  meme-cli -k 狗 -s pdan,qudoutu    # 只从指定源搜索
  meme-cli -k 开心 -l 5 -json       # 输出 JSON 格式
  meme-cli -list                    # 列出所有可用源

选项:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	// 创建注册中心并注册源
	if *verbose {
		fmt.Fprintln(os.Stderr, "🔧 初始化注册中心...")
	}
	registry := core.NewRegistry()
	config := &sources.Config{
		DouyinCookie:  os.Getenv("DOUYIN_COOKIE"),
		ImageProxyURL: os.Getenv("IMAGE_PROXY_URL"),
	}
	if *verbose {
		fmt.Fprintln(os.Stderr, "📦 正在注册数据源...")
	}
	sources.RegisterAllSources(registry, config)
	if *verbose {
		fmt.Fprintf(os.Stderr, "✅ 已加载 %d 个数据源\n", len(sources.GetAllSourceInfo(registry)))
	}

	// 列出所有源
	if *listSources {
		if *verbose {
			fmt.Fprintln(os.Stderr, "📋 执行 ListSources 操作...")
		}
		printSources(registry, *outputJSON)
		return
	}

	// 验证关键词
	if *keyword == "" {
		fmt.Fprintln(os.Stderr, "错误: 请使用 -k 指定搜索关键词")
		fmt.Fprintln(os.Stderr, "使用 -h 查看帮助")
		os.Exit(1)
	}

	// 解析指定的源
	var sourceIDs []string
	if *sourceList != "" {
		sourceIDs = strings.Split(*sourceList, ",")
		for i := range sourceIDs {
			sourceIDs[i] = strings.TrimSpace(sourceIDs[i])
		}
	}

	// 构造搜索选项
	opts := core.SearchOptions{
		Page:    *page,
		Limit:   *limit,
		Timeout: time.Duration(*timeout) * time.Second,
	}

	// 执行搜索
	ctx := context.Background()

	// 始终打印基本日志到 Stderr
	fmt.Fprintf(os.Stderr, "🚀 开始搜索: 关键词=%q, 页码=%d, 限制=%d\n", *keyword, *page, *limit)
	if len(sourceIDs) > 0 {
		fmt.Fprintf(os.Stderr, "🎯 指定源: %v\n", sourceIDs)
	} else {
		fmt.Fprintf(os.Stderr, "🌍 搜索范围: 所有源\n")
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "⏱️  超时设置: %d秒\n\n", *timeout)
	}

	startTime := time.Now()
	if *verbose {
		fmt.Fprintf(os.Stderr, "🚀 开始搜索: %s (页码:%d, 每页:%d)\n", *keyword, *page, *limit)
	}

	var result core.SearchResult
	start := time.Now()
	if len(sourceIDs) > 0 {
		result = registry.SearchSources(ctx, *keyword, sourceIDs, opts)
	} else {
		result = registry.SearchAll(ctx, *keyword, opts)
	}
	elapsed := time.Since(start)

	if *verbose {
		fmt.Fprintf(os.Stderr, "⏱️  底层逻辑执行耗时: %v\n", elapsed)
		fmt.Fprintf(os.Stderr, "📊 原始数据源状态: 成功 %d 个, 失败 %d 个\n", len(result.Sources), len(result.Errors))
		for id, err := range result.Errors {
			fmt.Fprintf(os.Stderr, "  - [%s]: %v\n", id, err)
		}
	}
	duration := time.Since(startTime)

	fmt.Fprintf(os.Stderr, "✅ 搜索结束: 耗时 %v, 找到 %d 个结果\n", duration, len(result.Memes))

	// 输出结果
	if *outputJSON {
		printJSON(result)
	} else {
		printPretty(result, *verbose)
	}
}

func printSources(registry *core.Registry, asJSON bool) {
	infos := sources.GetAllSourceInfo(registry)

	if asJSON {
		data, _ := json.MarshalIndent(infos, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Println("📦 可用的表情包源:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-12s %-10s %-30s %s\n", "ID", "名称", "描述", "认证")
	fmt.Println(strings.Repeat("-", 60))

	for _, info := range infos {
		auth := "❌"
		if info.RequiresAuth {
			auth = "✅ 需要"
		}
		fmt.Printf("%-12s %-10s %-30s %s\n", info.ID, info.Name, info.Description, auth)
	}
}

func printJSON(result core.SearchResult) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON 序列化失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printPretty(result core.SearchResult, verbose bool) {
	// 打印统计信息
	fmt.Printf("✅ 搜索完成! 耗时: %dms\n", result.DurationMs)
	fmt.Printf("📊 共找到 %d 个表情包\n", result.Total)

	if len(result.Sources) > 0 {
		fmt.Printf("🟢 成功的源: %s\n", strings.Join(result.Sources, ", "))
	}

	if len(result.Errors) > 0 {
		fmt.Printf("🔴 失败的源: ")
		errStrs := []string{}
		for id, err := range result.Errors {
			errStrs = append(errStrs, fmt.Sprintf("%s(%s)", id, err))
		}
		fmt.Println(strings.Join(errStrs, ", "))
	}

	fmt.Println()

	if len(result.Memes) == 0 {
		fmt.Println("😢 没有找到相关表情包")
		return
	}

	// 打印表情包列表
	fmt.Println("🎉 表情包列表:")
	fmt.Println(strings.Repeat("-", 80))

	for i, meme := range result.Memes {
		fmt.Printf("[%d] %s\n", i+1, meme.Title)
		fmt.Printf("    📦 来源: %s\n", meme.Platform)
		if verbose {
			fmt.Printf("    🔗 URL: %s\n", meme.URL)
			if meme.Format != "" {
				fmt.Printf("    📄 格式: %s\n", meme.Format)
			}
		} else {
			// 截断 URL 显示
			url := meme.URL
			if len(url) > 60 {
				url = url[:57] + "..."
			}
			fmt.Printf("    🔗 %s\n", url)
		}
		fmt.Println()
	}
}

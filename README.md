# Meme MCP Server

一个基于 Model Context Protocol (MCP) 的高性能表情包搜索服务。

## 特性

- 🚀 **高性能**: Go 实现，并发搜索多个数据源
- 🔌 **MCP 协议**: 可被 Claude Desktop、Cursor 等 AI 客户端直接调用
- 📦 **多源聚合**: 支持趣斗图、斗图啦、胖哒、搜狗、抖音等
- 🎯 **智能去重**: 自动识别并去除重复图片
- ⚡ **即插即用**: 单二进制文件，无需额外依赖

## 支持的数据源

| ID | 名称 | 说明 | 需要认证 |
|:---|:-----|:-----|:---------|
| `qudoutu` | 趣斗图 | qudoutu.cn | ❌ |
| `doutula` | 斗图啦 | doutupk.com | ❌ |
| `pdan` | 胖哒 | pdan.com.cn | ❌ |
| `sougou` | 搜狗表情 | pic.sogou.com | ❌ |
| `douyin` | 抖音 | douyin.com | ✅ Cookie |

## 快速开始

### 构建

```bash
# 安装依赖
make deps

# 构建
make build
```

### 运行

```bash
# 直接运行 (Stdio 模式)
./build/meme-server

# 带抖音 Cookie 运行
DOUYIN_COOKIE="your_cookie_here" ./build/meme-server
```

### 测试

```bash
# 测试搜索
make test-search

# 测试列出源
make test-list
```

## MCP Tools

### search_meme

搜索表情包。

**参数:**
- `keyword` (string, 必填): 搜索关键词
- `sources` (array, 可选): 指定搜索的源 ID 列表
- `page` (number, 可选): 页码，默认 1
- `limit` (number, 可选): 每个源返回的最大数量，默认 20

**示例:**
```json
{
  "name": "search_meme",
  "arguments": {
    "keyword": "猫",
    "sources": ["pdan", "qudoutu"],
    "limit": 10
  }
}
```

### list_sources

列出所有可用的数据源。

**参数:** 无

## 配置 Claude Desktop

在 `~/Library/Application Support/Claude/claude_desktop_config.json` 中添加:

```json
{
  "mcpServers": {
    "meme": {
      "command": "/path/to/meme-server",
      "env": {
        "DOUYIN_COOKIE": "your_cookie_here"
      }
    }
  }
}
```

## 配置 Cursor

在 Cursor 设置中添加 MCP Server 配置。

## 开发

### 添加新数据源

1. 在 `internal/sources/` 下创建新文件
2. 实现 `core.Source` 接口
3. 在 `internal/sources/register.go` 中注册

```go
type MySource struct {
    sources.BaseSource
}

func (s *MySource) Search(ctx context.Context, keyword string, opts core.SearchOptions) ([]core.Meme, error) {
    // 实现搜索逻辑
}
```

### 项目结构

```
.
├── cmd/server/          # MCP Server 入口
├── internal/
│   ├── core/            # 核心类型和注册中心
│   ├── sources/         # 数据源实现
│   └── tools/           # MCP Tools 定义
├── config/              # 配置文件
└── Makefile
```

## License

MIT

# Meme MCP Server

一个基于 Model Context Protocol (MCP) 的高性能表情包搜索服务。

它能够聚合多个表情包源的搜索结果，并通过 MCP 协议直接为 Claude Desktop、Cursor 等 AI 客户端提供服务。

## ✨ 特性

- 🚀 **高性能**: Go 语言实现，并发搜索多个数据源。
- 🔌 **MCP 协议**: 完美支持 Model Context Protocol，可无缝集成到 AI 工作流中。
- 📦 **多源聚合**: 支持 6 个主流表情包源。
- 🛡️ **防盗链支持**: 内置图片代理机制，解决部分源（如趣斗图）的图片 404/防盗链问题。
- 🎯 **智能去重**: 自动识别并去除重复图片。
- ⚡ **即插即用**: 单二进制文件，部署简单。

## 📚 支持的数据源

| ID | 名称 | 说明 | 配置要求 |
|:---|:-----|:-----|:---------|
| `doutula` | 斗图啦 | doutupk.com | 无 |
| `pdan` | 胖哒 | pdan.com.cn | 无 |
| `sougou` | 搜狗表情 | pic.sogou.com | 无 |
| `qudoutu` | 趣斗图 | qudoutu.cn | ⚠️ 需配置 `IMAGE_PROXY_URL` |
| `doutub` | 表情包API | api.doutub.com | ⚠️ 需配置 `IMAGE_PROXY_URL` |
| `douyin` | 抖音 | douyin.com | 🔐 需配置 `DOUYIN_COOKIE` |

> **注意**: 如果未配置相应的环境变量，对应的源将**不会被初始化**，也不会出现在搜索结果中。

## 🛠️ 配置说明

### 1. 图片代理配置 (`IMAGE_PROXY_URL`)

部分源（如`qudoutu`、`doutub`）开启了严格的防盗链保护，直接访问图片链接会返回 404。配置此环境变量后，返回的图片链接将被重写为代理地址。

**格式**:
您的代理服务地址，支持以下占位符：
- `{URL}` 或 `{SOURCE_URL}`: 原始图片链接 (会自动 URL 编码)
- `{REFERER}`: 该图片源对应的 Referer (会自动 URL 编码)

**示例**:
```bash
export IMAGE_PROXY_URL="https://my-proxy-worker.com/image?url={URL}&referer={REFERER}"
```

### 2. 抖音 Cookie (`DOUYIN_COOKIE`)

搜索抖音表情包需要有效的 Cookie。您可以在浏览器登录抖音网页版，按 F12 打开开发者工具，复制请求中的 Cookie 字符串。

```bash
export DOUYIN_COOKIE="your_cookie_string_here"
```

## 🚀 快速开始

### 构建

```bash
# 下载依赖
make deps

# 构建 Server 和 CLI
make build-all
```

### 命令行工具 (CLI) 测试

项目自带一个功能强大的 CLI 工具，方便测试和检索。

```bash
# 基础搜索
./build/meme-cli -k "猫"

# 指定源搜索 (需配置代理才会有 qudoutu)
export IMAGE_PROXY_URL="..."
./build/meme-cli -k "狗" -s qudoutu,doutub -l 5

# 列出当前可用的源 (检查配置是否生效)
./build/meme-cli -list
```

### 运行 MCP Server

```bash
# 设置环境变量并运行
export IMAGE_PROXY_URL="https://..."
export DOUYIN_COOKIE="..."
./build/meme-server
```

## 🤖 AI 客户端集成

### 配置 Claude Desktop

编辑配置文件:
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "meme": {
      "command": "/absolute/path/to/meme-server",
      "env": {
        "IMAGE_PROXY_URL": "https://your-proxy-service.com/?src={URL}&referer={REFERER}",
        "DOUYIN_COOKIE": "your_cookie_here"
      }
    }
  }
}
```

### 配置 Cursor

在 Cursor 的 MCP 设置中添加一个新的 Server：

- **Type**: `stdio`
- **Command**: `/absolute/path/to/meme-server`
- **Environment Variables**: 添加 `IMAGE_PROXY_URL` 和 `DOUYIN_COOKIE`

## 📦 MCP Tools

### `search_meme`
搜索表情包。

- `keyword` (string): 搜索关键词
- `sources` (array): 指定搜索源 ID (可选)
- `page` (number): 页码
- `limit` (number): 数量限制

### `list_sources`
列出当前已加载并可用的数据源。

## 📄 License

MIT

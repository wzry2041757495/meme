package sources

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/shadow/meme/internal/core"
)

// BaseSource 提供通用字段和方法，减少重复代码
type BaseSource struct {
	id          string
	name        string
	description string
	requireAuth bool
	client      *http.Client
}

func (b *BaseSource) ID() string          { return b.id }
func (b *BaseSource) Name() string        { return b.name }
func (b *BaseSource) Description() string { return b.description }
func (b *BaseSource) RequiresAuth() bool  { return b.requireAuth }

// newHTTPClient 创建带默认配置的 HTTP 客户端
// 增强 TLS 兼容性，解决某些网站的握手失败问题
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// 允许较旧的 TLS 版本以兼容更多网站
				MinVersion: tls.VersionTLS10,
				// 使用更宽松的加密套件
				CipherSuites: nil, // nil 表示使用 Go 默认的所有套件
				// 跳过证书验证 (某些网站证书配置有问题)
				InsecureSkipVerify: true,
			},
			// 连接配置
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			// 禁用 HTTP/2，某些网站对 HTTP/2 支持不好
			ForceAttemptHTTP2: false,
		},
	}
}

// fetchHTML 通用的 HTML 抓取方法
func fetchHTML(ctx context.Context, client *http.Client, targetURL string, headers map[string]string) (*goquery.Document, error) {
	fmt.Fprintf(os.Stderr, "🌐 [Request] GET %s\n", targetURL)
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置更完整的浏览器 Headers，模拟真实浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")

	// 设置自定义 Headers (会覆盖上面的默认值)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 处理 gzip 压缩
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("create gzip reader failed: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("parse HTML failed: %w", err)
	}

	return doc, nil
}

// ============ 趣斗图 (Qudoutu) ============

type QudoutuSource struct {
	BaseSource
}

func NewQudoutu() *QudoutuSource {
	return &QudoutuSource{
		BaseSource: BaseSource{
			id:          "qudoutu",
			name:        "趣斗图",
			description: "从 qudoutu.cn 搜索表情包",
			requireAuth: false,
			client:      newHTTPClient(),
		},
	}
}

func (s *QudoutuSource) Search(ctx context.Context, keyword string, opts core.SearchOptions) ([]core.Meme, error) {
	searchURL := fmt.Sprintf(
		"https://www.qudoutu.cn/search/?keyword=%s",
		url.QueryEscape(keyword),
	)

	doc, err := fetchHTML(ctx, s.client, searchURL, map[string]string{
		"Referer":        "https://www.qudoutu.cn/",
		"Sec-Fetch-Site": "same-origin",
	})
	if err != nil {
		return nil, err
	}

	var memes []core.Meme
	// 更精确地定位搜索结果区域：item-grid 下的 ul 中的 li
	// 如果精确选择器找不到，回退到通用选择器
	results := doc.Find("div.item-grid ul li")
	if results.Length() == 0 {
		// 回退到通用选择器，但只选择包含 a.Link 的 li
		results = doc.Find("li").Has("a.Link")
	}

	results.Each(func(i int, sel *goquery.Selection) {
		imgURL, exists := sel.Find("a.Link img").Attr("src")
		if !exists || imgURL == "" {
			return
		}

		// 处理相对路径
		if strings.HasPrefix(imgURL, "/") {
			imgURL = "https://www.qudoutu.cn" + imgURL
		}

		imgURL = core.NormalizeURL(imgURL)
		if !core.IsValidImageURL(imgURL) {
			return
		}

		title := strings.TrimSpace(sel.Find("p").Text())
		if title == "" {
			title = "趣斗图"
		}

		memes = append(memes, core.Meme{
			Title:    title,
			URL:      imgURL,
			Platform: s.id,
			Format:   core.DetectImageFormat(imgURL),
		})
	})

	if opts.Limit > 0 && len(memes) > opts.Limit {
		memes = memes[:opts.Limit]
	}

	return memes, nil
}

// ============ 斗图啦 (Doutula) ============

type DoutulaSource struct {
	BaseSource
}

func NewDoutula() *DoutulaSource {
	return &DoutulaSource{
		BaseSource: BaseSource{
			id:          "doutula",
			name:        "斗图啦",
			description: "从 doutupk.com 搜索表情包",
			requireAuth: false,
			client:      newHTTPClient(),
		},
	}
}

func (s *DoutulaSource) Search(ctx context.Context, keyword string, opts core.SearchOptions) ([]core.Meme, error) {
	searchURL := fmt.Sprintf(
		"https://www.doutupk.com/search?keyword=%s",
		url.QueryEscape(keyword),
	)

	doc, err := fetchHTML(ctx, s.client, searchURL, map[string]string{
		"Referer": "https://www.doutupk.com/",
	})
	if err != nil {
		return nil, err
	}

	var memes []core.Meme
	doc.Find("a.col-xs-6.col-md-2").Each(func(i int, sel *goquery.Selection) {
		title := strings.TrimSpace(sel.Find("p").Text())
		if title == "" {
			title = "斗图啦"
		}

		imgURL, exists := sel.Find("img.image_dtb").Attr("data-original")
		if !exists || imgURL == "" {
			return
		}

		// 强制 HTTPS
		imgURL = core.NormalizeURL(imgURL)
		if core.IsValidImageURL(imgURL) {
			memes = append(memes, core.Meme{
				Title:    title,
				URL:      imgURL,
				Platform: s.id,
				Format:   core.DetectImageFormat(imgURL),
			})
		}
	})

	if opts.Limit > 0 && len(memes) > opts.Limit {
		memes = memes[:opts.Limit]
	}

	return memes, nil
}

// ============ 胖哒 (Pdan) ============

type PdanSource struct {
	BaseSource
}

func NewPdan() *PdanSource {
	return &PdanSource{
		BaseSource: BaseSource{
			id:          "pdan",
			name:        "胖哒",
			description: "从 pdan.com.cn 搜索表情包",
			requireAuth: false,
			client:      newHTTPClient(),
		},
	}
}

func (s *PdanSource) Search(ctx context.Context, keyword string, opts core.SearchOptions) ([]core.Meme, error) {
	searchURL := fmt.Sprintf(
		"https://pdan.com.cn/?s=%s",
		url.QueryEscape(keyword),
	)

	doc, err := fetchHTML(ctx, s.client, searchURL, map[string]string{
		"Referer": "https://pdan.com.cn/",
	})
	if err != nil {
		return nil, err
	}

	var memes []core.Meme
	doc.Find("a.imageLink.image.loading").Each(func(i int, sel *goquery.Selection) {
		// 按优先级获取标题
		title := sel.AttrOr("title", "")
		if title == "" {
			title = sel.Find("img").AttrOr("alt", "")
		}
		if title == "" {
			title = strings.TrimSpace(sel.Find("span.bg").Text())
		}
		if title == "" {
			title = "胖哒"
		}

		// 优先 data-src，其次 src
		imgURL := sel.Find("img").AttrOr("data-src", "")
		if imgURL == "" {
			imgURL = sel.Find("img").AttrOr("src", "")
		}

		if imgURL == "" {
			return
		}

		imgURL = core.NormalizeURL(imgURL)
		if core.IsValidImageURL(imgURL) {
			memes = append(memes, core.Meme{
				Title:    title,
				URL:      imgURL,
				Platform: s.id,
				Format:   core.DetectImageFormat(imgURL),
			})
		}
	})

	if opts.Limit > 0 && len(memes) > opts.Limit {
		memes = memes[:opts.Limit]
	}

	return memes, nil
}

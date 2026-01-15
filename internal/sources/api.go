package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/shadow/meme/internal/core"
)

// ============ 搜狗表情 (Sougou) ============

type SougouSource struct {
	BaseSource
}

func NewSougou() *SougouSource {
	return &SougouSource{
		BaseSource: BaseSource{
			id:          "sougou",
			name:        "搜狗表情",
			description: "从搜狗图片搜索表情包 (JSON API)",
			requireAuth: false,
			client:      newHTTPClient(),
		},
	}
}

// sougouResponse 搜狗 PC 版 API 响应结构
// API: https://pic.sogou.com/napi/pc/searchList
type sougouResponse struct {
	Data struct {
		Items []struct {
			LocImageLink string `json:"locImageLink"` // CDN 链接
			ThumbUrl     string `json:"thumbUrl"`     // 缩略图链接
			OriPicUrl    string `json:"oriPicUrl"`    // 原始图片链接（优先使用）
			PicUrl       string `json:"picUrl"`       // 图片页面链接（备选）
			Title        string `json:"title"`        // 标题
			Width        int    `json:"width"`
			Height       int    `json:"height"`
		} `json:"items"`
	} `json:"data"`
	Status int `json:"status"`
}

func (s *SougouSource) Search(ctx context.Context, keyword string, opts core.SearchOptions) ([]core.Meme, error) {
	page := opts.Page
	if page < 1 {
		page = 1
	}

	// 计算分页参数
	pageSize := 48
	start := (page - 1) * pageSize

	// 构造新的 API URL
	// tagQSign 是固定的表情包标签签名
	params := url.Values{
		"mode":     {"1"},
		"tagQSign": {"表情包,5e604ff6"},
		"start":    {fmt.Sprintf("%d", start)},
		"xml_len":  {fmt.Sprintf("%d", pageSize)},
		"query":    {keyword},
		"channel":  {"pc_pic"},
		"scene":    {"pic_result"},
	}

	apiURL := "https://pic.sogou.com/napi/pc/searchList?" + params.Encode()
	fmt.Fprintf(os.Stderr, "🌐 [Request] GET %s\n", apiURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置完整的浏览器 Headers
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("X-Time4p", fmt.Sprintf("%d", time.Now().UnixMilli()))
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("Referer", "https://pic.sogou.com/pics")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data sougouResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode JSON failed: %w", err)
	}

	var memes []core.Meme
	for _, item := range data.Data.Items {
		// 优先使用原始图片链接，其次 CDN 链接，最后缩略图
		imgURL := item.OriPicUrl
		if imgURL == "" {
			imgURL = item.PicUrl
		}
		if imgURL == "" {
			imgURL = item.LocImageLink
		}
		if imgURL == "" {
			imgURL = item.ThumbUrl
		}
		if imgURL == "" {
			continue
		}

		imgURL = core.NormalizeURL(imgURL)
		if !core.IsValidImageURL(imgURL) {
			continue
		}

		title := item.Title
		if title == "" {
			title = "搜狗表情"
		}

		memes = append(memes, core.Meme{
			Title:    title,
			URL:      imgURL,
			Platform: s.id,
			Format:   core.DetectImageFormat(imgURL),
			Width:    item.Width,
			Height:   item.Height,
		})
	}

	if opts.Limit > 0 && len(memes) > opts.Limit {
		memes = memes[:opts.Limit]
	}

	return memes, nil
}

// ============ 抖音 (Douyin) ============

type DouyinSource struct {
	BaseSource
	cookie string
}

func NewDouyin(cookie string) *DouyinSource {
	return &DouyinSource{
		BaseSource: BaseSource{
			id:          "douyin",
			name:        "抖音",
			description: "从抖音搜索热门表情包 (需要 Cookie)",
			requireAuth: true,
			client:      newHTTPClient(),
		},
		cookie: cookie,
	}
}

// SetCookie 动态设置 Cookie
func (s *DouyinSource) SetCookie(cookie string) {
	s.cookie = cookie
}

// douyinResponse 抖音 API 响应结构
type douyinResponse struct {
	EmoticonData struct {
		StickerList []struct {
			Author struct {
				Name string `json:"name"`
			} `json:"author"`
			Origin struct {
				URLList []string `json:"url_list"`
			} `json:"origin"`
		} `json:"sticker_list"`
	} `json:"emoticon_data"`
}

func (s *DouyinSource) Search(ctx context.Context, keyword string, opts core.SearchOptions) ([]core.Meme, error) {
	if s.cookie == "" {
		return nil, fmt.Errorf("douyin source requires cookie configuration")
	}

	page := opts.Page
	if page < 1 {
		page = 1
	}
	cursor := (page - 1) * 10

	params := url.Values{
		"device_platform": {"webapp"},
		"aid":             {"1128"},
		"keyword":         {keyword},
		"cursor":          {fmt.Sprintf("%d", cursor)},
	}

	apiURL := "https://www.douyin.com/aweme/v1/web/im/resource/emoticon/search?" + params.Encode()
	fmt.Fprintf(os.Stderr, "🌐 [Request] GET %s\n", apiURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.douyin.com/")
	req.Header.Set("Cookie", s.cookie)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}
	fmt.Fprintf(os.Stderr, "🌐 [Response] %s\n", string(body))

	var data douyinResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode JSON failed: %w", err)
	}

	var memes []core.Meme
	for _, item := range data.EmoticonData.StickerList {
		if len(item.Origin.URLList) == 0 {
			continue
		}

		imgURL := core.NormalizeURL(item.Origin.URLList[0])

		// 过滤有效图片 URL
		if !core.IsValidImageURL(imgURL) {
			continue
		}

		title := item.Author.Name
		if title == "" {
			title = "抖音表情"
		}

		memes = append(memes, core.Meme{
			Title:    title,
			URL:      imgURL,
			Platform: s.id,
			Format:   core.DetectImageFormat(imgURL),
		})
	}

	if opts.Limit > 0 && len(memes) > opts.Limit {
		memes = memes[:opts.Limit]
	}

	return memes, nil
}

package sources

import "github.com/shadow/meme/internal/core"

// RegisterAllSources 注册所有内置源到注册中心
func RegisterAllSources(registry *core.Registry, config *Config) {
	// 注册无需认证的源
	registry.Register(NewQudoutu())
	registry.Register(NewDoutula())
	registry.Register(NewPdan())
	registry.Register(NewSougou())

	// 注册需要认证的源 (如果配置了 Cookie)
	if config != nil && config.DouyinCookie != "" {
		registry.Register(NewDouyin(config.DouyinCookie))
	}
}

// Config 源配置
type Config struct {
	DouyinCookie string `json:"douyin_cookie" yaml:"douyin_cookie"`
	// 可以扩展其他需要认证的源配置
}

// GetAllSourceInfo 获取所有源的信息 (用于 list_sources Tool)
func GetAllSourceInfo(registry *core.Registry) []SourceInfo {
	sources := registry.List()
	infos := make([]SourceInfo, 0, len(sources))

	for _, s := range sources {
		infos = append(infos, SourceInfo{
			ID:           s.ID(),
			Name:         s.Name(),
			Description:  s.Description(),
			RequiresAuth: s.RequiresAuth(),
		})
	}

	return infos
}

// SourceInfo 源信息结构
type SourceInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	RequiresAuth bool   `json:"requires_auth"`
}

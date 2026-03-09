// Package version 定义 qodercli 版本信息
package version

// Version qodercli 版本号
// 与官方版本保持同步
const Version = "0.1.30"

// BuildInfo 构建信息（可选，用于未来扩展）
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildTime string
	GoVersion string
}

// GetVersion 获取版本号
func GetVersion() string {
	return Version
}

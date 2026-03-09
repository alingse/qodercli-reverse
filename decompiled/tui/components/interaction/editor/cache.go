package editor

// InputCache 输入缓存
type InputCache struct {
	currentInput string
	savedInput   string
}

// NewInputCache 创建新的输入缓存
func NewInputCache() *InputCache {
	return &InputCache{}
}

// SetInput 设置当前输入
func (ic *InputCache) SetInput(input string) {
	ic.currentInput = input
}

// GetInput 获取当前输入
func (ic *InputCache) GetInput() string {
	return ic.currentInput
}

// Reset 重置缓存
func (ic *InputCache) Reset() {
	ic.currentInput = ""
	ic.savedInput = ""
}

// SaveInput 保存输入（用于临时切换）
func (ic *InputCache) SaveInput() {
	ic.savedInput = ic.currentInput
}

// RestoreInput 恢复保存的输入
func (ic *InputCache) RestoreInput() string {
	return ic.savedInput
}

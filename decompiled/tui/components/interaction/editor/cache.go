package editor

// InputCache input cache
type InputCache struct {
	currentInput string
	savedInput   string
}

// NewInputCache creates a new input cache
func NewInputCache() *InputCache {
	return &InputCache{}
}

// SetInput sets current input
func (ic *InputCache) SetInput(input string) {
	ic.currentInput = input
}

// GetInput gets current input
func (ic *InputCache) GetInput() string {
	return ic.currentInput
}

// Reset resets the cache
func (ic *InputCache) Reset() {
	ic.currentInput = ""
	ic.savedInput = ""
}

// SaveInput saves input (for temporary switching)
func (ic *InputCache) SaveInput() {
	ic.savedInput = ic.currentInput
}

// RestoreInput restores saved input
func (ic *InputCache) RestoreInput() string {
	return ic.savedInput
}

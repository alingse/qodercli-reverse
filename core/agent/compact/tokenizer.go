// Package compact 提供上下文压缩功能 - Tokenizer 实现
package compact

import (
	"strings"
	"unicode"
)

// SimpleTokenizer 简单的 Tokenizer 实现
type SimpleTokenizer struct {
	// 中文每个字符算 1.5 个 token
	chineseTokenFactor float64
	// 英文每个单词算 1.3 个 token
	englishTokenFactor float64
}

// NewSimpleTokenizer 创建简单的 Tokenizer
func NewSimpleTokenizer() *SimpleTokenizer {
	return &SimpleTokenizer{
		chineseTokenFactor: 1.5,
		englishTokenFactor: 1.3,
	}
}

// Count 计算文本的 Token 数
func (t *SimpleTokenizer) Count(text string) int {
	if text == "" {
		return 0
	}

	var chineseChars int
	var englishWords int
	var otherChars int

	currentWord := strings.Builder{}

	for _, r := range text {
		switch {
		case unicode.Is(unicode.Han, r):
			// 汉字
			chineseChars++
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			// 英文字母或数字
			currentWord.WriteRune(r)
		default:
			// 其他字符（空格、标点等）
			if currentWord.Len() > 0 {
				englishWords++
				currentWord.Reset()
			}
			otherChars++
		}
	}

	// 处理最后一个单词
	if currentWord.Len() > 0 {
		englishWords++
	}

	// 计算 Token 数
	tokens := 0
	tokens += int(float64(chineseChars) * t.chineseTokenFactor)
	tokens += int(float64(englishWords) * t.englishTokenFactor)
	tokens += otherChars/ 4 // 其他字符每 4 个算 1 个 token

	return tokens
}

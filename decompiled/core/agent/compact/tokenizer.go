// Package compact provides context compaction - Tokenizer implementations
package compact

import (
	"strings"
	"unicode"
)

// SimpleTokenizer simple tokenizer implementation
type SimpleTokenizer struct {
	// Chinese characters count as 1.5 tokens each
	chineseTokenFactor float64
	// English words count as 1.3 tokens each
	englishTokenFactor float64
}

// NewSimpleTokenizer creates a simple tokenizer
func NewSimpleTokenizer() *SimpleTokenizer {
	return &SimpleTokenizer{
		chineseTokenFactor: 1.5,
		englishTokenFactor: 1.3,
	}
}

// Count calculates the token count of text
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
			// Chinese character
			chineseChars++
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			// English letter or digit
			currentWord.WriteRune(r)
		default:
			// other characters (spaces, punctuation, etc.)
			if currentWord.Len() > 0 {
				englishWords++
				currentWord.Reset()
			}
			otherChars++
		}
	}

	// handle last word
	if currentWord.Len() > 0 {
		englishWords++
	}

	// calculate token count
	tokens := 0
	tokens += int(float64(chineseChars) * t.chineseTokenFactor)
	tokens += int(float64(englishWords) * t.englishTokenFactor)
	tokens += otherChars / 4 // other characters: 4 chars = 1 token

	return tokens
}

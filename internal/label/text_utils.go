package label

import (
	"regexp"
	"strings"
	"unicode"
)

// cleanText removes HTML tags, excessive whitespace, and normalizes the text
func cleanText(text string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, " ")

	// Remove URLs
	urlRe := regexp.MustCompile(`https?://[^\s]+`)
	text = urlRe.ReplaceAllString(text, " ")

	// Normalize whitespace
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

// isChineseText checks if the text is primarily Chinese
func isChineseText(text string) bool {
	chineseCount := 0
	totalCount := 0

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.Is(unicode.Han, r) {
			totalCount++
			if unicode.Is(unicode.Han, r) {
				chineseCount++
			}
		}
	}

	if totalCount == 0 {
		return false
	}

	// If more than 30% of characters are Chinese, consider it Chinese text
	return float64(chineseCount)/float64(totalCount) > 0.3
}

// splitWords splits text into words, handling both English and Chinese
func splitWords(text string, isChinese bool) []string {
	if isChinese {
		// For Chinese, we'll use character bigrams and trigrams
		return extractChineseNGrams(text)
	}

	// For English, split by word boundaries and filter
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	// Convert to lowercase and filter short words
	var filtered []string
	for _, word := range words {
		word = strings.ToLower(word)
		if len(word) >= 3 && !isStopWord(word) {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// extractChineseNGrams extracts Chinese character bigrams and trigrams
func extractChineseNGrams(text string) []string {
	runes := []rune(text)
	var ngrams []string

	// Extract bigrams (2-character sequences)
	for i := 0; i < len(runes)-1; i++ {
		if unicode.Is(unicode.Han, runes[i]) && unicode.Is(unicode.Han, runes[i+1]) {
			ngrams = append(ngrams, string(runes[i:i+2]))
		}
	}

	// Extract trigrams (3-character sequences) - these are often more meaningful
	for i := 0; i < len(runes)-2; i++ {
		if unicode.Is(unicode.Han, runes[i]) && unicode.Is(unicode.Han, runes[i+1]) && unicode.Is(unicode.Han, runes[i+2]) {
			ngrams = append(ngrams, string(runes[i:i+3]))
		}
	}

	return ngrams
}

// isStopWord checks if a word is a common stop word (English only)
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "be": true, "to": true, "of": true, "and": true,
		"a": true, "in": true, "that": true, "have": true, "i": true,
		"it": true, "for": true, "not": true, "on": true, "with": true,
		"he": true, "as": true, "you": true, "do": true, "at": true,
		"this": true, "but": true, "his": true, "by": true, "from": true,
		"they": true, "we": true, "say": true, "her": true, "she": true,
		"or": true, "an": true, "will": true, "my": true, "one": true,
		"all": true, "would": true, "there": true, "their": true, "what": true,
		"so": true, "up": true, "out": true, "if": true, "about": true,
		"who": true, "get": true, "which": true, "go": true, "me": true,
		"when": true, "make": true, "can": true, "like": true, "time": true,
		"no": true, "just": true, "him": true, "know": true, "take": true,
		"people": true, "into": true, "year": true, "your": true, "good": true,
		"some": true, "could": true, "them": true, "see": true, "other": true,
		"than": true, "then": true, "now": true, "look": true, "only": true,
		"come": true, "its": true, "over": true, "think": true, "also": true,
		"back": true, "after": true, "use": true, "two": true, "how": true,
		"our": true, "work": true, "first": true, "well": true, "way": true,
		"even": true, "new": true, "want": true, "because": true, "any": true,
		"these": true, "give": true, "day": true, "most": true, "us": true,
		"is": true, "was": true, "are": true, "been": true, "has": true,
		"had": true, "were": true, "said": true, "did": true, "having": true,
		"may": true, "should": true, "does": true, "being": true,
	}
	return stopWords[word]
}

// normalizeLabel normalizes a label string
func normalizeLabel(label string) string {
	// Trim whitespace
	label = strings.TrimSpace(label)

	// For English, capitalize first letter of each word
	if !isChineseText(label) {
		words := strings.Fields(label)
		for i, word := range words {
			runes := []rune(word)
			if len(runes) > 0 {
				words[i] = strings.ToUpper(string(runes[0])) + strings.ToLower(string(runes[1:]))
			}
		}
		label = strings.Join(words, " ")
	}

	return label
}

// validateLabel checks if a label meets the requirements
func validateLabel(label string) bool {
	if len(label) < MinLabelLength || len(label) > MaxLabelLength {
		return false
	}

	// Must contain at least one letter or Han character
	hasValidChar := false
	for _, r := range label {
		if unicode.IsLetter(r) || unicode.Is(unicode.Han, r) {
			hasValidChar = true
			break
		}
	}

	return hasValidChar
}

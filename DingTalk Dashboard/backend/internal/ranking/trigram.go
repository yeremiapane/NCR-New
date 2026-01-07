package ranking

import (
	"strings"
	"unicode"
)

// GenerateTrigrams creates a set of 3-character sequences from text
func GenerateTrigrams(text string) map[string]bool {
	// Normalize text: lowercase and remove extra spaces
	text = strings.ToLower(text)
	text = normalizeWhitespace(text)

	trigrams := make(map[string]bool)
	runes := []rune(text)

	if len(runes) < 3 {
		// For very short text, use the whole string as one "trigram"
		if len(runes) > 0 {
			trigrams[string(runes)] = true
		}
		return trigrams
	}

	for i := 0; i <= len(runes)-3; i++ {
		trigram := string(runes[i : i+3])
		trigrams[trigram] = true
	}

	return trigrams
}

// CalculateSimilarity computes Jaccard similarity between two trigram sets
// Returns a value between 0 (completely different) and 1 (identical)
func CalculateSimilarity(a, b map[string]bool) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0 // Both empty = identical
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0 // One empty = completely different
	}

	// Calculate intersection
	intersection := 0
	for trigram := range a {
		if b[trigram] {
			intersection++
		}
	}

	// Calculate union
	union := len(a) + len(b) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// LongestCommonSubstring finds the longest common substring between two strings
// Returns the LCS and its length
func LongestCommonSubstring(s1, s2 string) (string, int) {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	r1 := []rune(s1)
	r2 := []rune(s2)
	m := len(r1)
	n := len(r2)

	if m == 0 || n == 0 {
		return "", 0
	}

	// DP table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	maxLen := 0
	endIdx := 0

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if r1[i-1] == r2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				if dp[i][j] > maxLen {
					maxLen = dp[i][j]
					endIdx = i
				}
			}
		}
	}

	if maxLen == 0 {
		return "", 0
	}

	return string(r1[endIdx-maxLen : endIdx]), maxLen
}

// CalculateLCSSimilarity computes similarity based on LCS ratio
// Returns a value between 0 and 1
func CalculateLCSSimilarity(s1, s2 string) float64 {
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	_, lcsLen := LongestCommonSubstring(s1, s2)
	maxLen := len([]rune(s1))
	if len([]rune(s2)) > maxLen {
		maxLen = len([]rune(s2))
	}

	return float64(lcsLen) / float64(maxLen)
}

// CalculateCombinedSimilarity combines Trigram and LCS similarity
// weights: trigram=0.6, lcs=0.4 (configurable)
func CalculateCombinedSimilarity(s1, s2 string, trigramsA, trigramsB map[string]bool) float64 {
	trigramSim := CalculateSimilarity(trigramsA, trigramsB)
	lcsSim := CalculateLCSSimilarity(s1, s2)

	// Weighted combination
	return (trigramSim * 0.6) + (lcsSim * 0.4)
}

// normalizeWhitespace removes extra spaces and non-printable characters
func normalizeWhitespace(text string) string {
	var result strings.Builder
	lastWasSpace := true

	for _, r := range text {
		if unicode.IsSpace(r) {
			if !lastWasSpace {
				result.WriteRune(' ')
				lastWasSpace = true
			}
		} else if unicode.IsPrint(r) {
			result.WriteRune(r)
			lastWasSpace = false
		}
	}

	return strings.TrimSpace(result.String())
}

// ExtractKeywords extracts meaningful words from text (basic implementation)
func ExtractKeywords(text string) []string {
	// Common Indonesian stop words to filter out
	stopWords := map[string]bool{
		"yang": true, "dan": true, "di": true, "ke": true, "dari": true,
		"untuk": true, "dengan": true, "pada": true, "adalah": true, "ini": true,
		"itu": true, "atau": true, "tidak": true, "ada": true, "oleh": true,
		"akan": true, "sudah": true, "juga": true, "dapat": true, "bisa": true,
		"lebih": true, "sebagai": true, "dalam": true, "karena": true, "telah": true,
		"saat": true, "setelah": true, "harus": true, "menjadi": true, "seperti": true,
		"tersebut": true, "belum": true, "sehingga": true, "namun": true, "bila": true,
		"apabila": true, "bahwa": true, "yaitu": true, "antara": true, "tetapi": true,
		"tapi": true,
	}

	words := strings.Fields(strings.ToLower(text))
	var keywords []string

	for _, word := range words {
		// Clean word from punctuation
		word = strings.Trim(word, ".,;:!?\"'()[]{}/-")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// WordFrequency represents a word and its count
type WordFrequency struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// CountWordFrequencies counts word occurrences across multiple texts
func CountWordFrequencies(texts []string, limit int) []WordFrequency {
	wordCounts := make(map[string]int)

	for _, text := range texts {
		keywords := ExtractKeywords(text)
		for _, word := range keywords {
			wordCounts[word]++
		}
	}

	// Convert to slice and sort
	var freqs []WordFrequency
	for word, count := range wordCounts {
		if count >= 2 { // Only include words that appear at least twice
			freqs = append(freqs, WordFrequency{Word: word, Count: count})
		}
	}

	// Sort by count descending (bubble sort for simplicity)
	for i := 0; i < len(freqs)-1; i++ {
		for j := 0; j < len(freqs)-i-1; j++ {
			if freqs[j].Count < freqs[j+1].Count {
				freqs[j], freqs[j+1] = freqs[j+1], freqs[j]
			}
		}
	}

	// Limit results
	if len(freqs) > limit {
		freqs = freqs[:limit]
	}

	return freqs
}

// ExtractKeyPhrase extracts 2-5 most important words from a text to create a concise summary
func ExtractKeyPhrase(text string, maxWords int) string {
	if maxWords <= 0 {
		maxWords = 5
	}

	// Important domain-specific words that should be prioritized
	importantWords := map[string]int{
		// Problem types
		"rusak": 10, "pecah": 10, "retak": 10, "bocor": 10, "patah": 10,
		"bengkok": 10, "kotor": 10, "cacat": 10, "gagal": 10, "error": 10,
		"salah": 10, "keliru": 10, "tidak": 8, "kurang": 8, "lebih": 6,
		"beda": 8, "berbeda": 8, "miring": 10, "geser": 10, "longgar": 10,
		// Materials
		"material": 9, "bahan": 9, "kaca": 9, "aluminium": 9, "kayu": 9,
		"plat": 9, "besi": 9, "stainless": 9, "karet": 9, "plastik": 9,
		"cat": 9, "finishing": 9, "powder": 9, "coating": 9, "sealant": 9,
		// Components
		"pintu": 8, "jendela": 8, "frame": 8, "panel": 8, "handle": 8,
		"kunci": 8, "engsel": 8, "rel": 8, "roller": 8, "glass": 8,
		"profil": 8, "aksesoris": 8, "bracket": 8, "corner": 8, "gasket": 8,
		// Quality issues
		"dimensi": 9, "ukuran": 9, "warna": 9, "bentuk": 9, "kualitas": 9,
		"spec": 9, "spesifikasi": 9, "tolerance": 9, "standar": 9, "reject": 10,
		"defect": 10, "scratch": 10, "dent": 10, "baret": 10, "penyok": 10,
		// Process
		"welding": 8, "las": 8, "cutting": 8, "potong": 8, "drilling": 8,
		"bor": 8, "assembly": 8, "rakit": 8, "packing": 8, "kirim": 8,
		"produksi": 8, "proses": 7, "mesin": 8, "alat": 7, "setting": 8,
	}

	// Extract keywords
	keywords := ExtractKeywords(text)
	if len(keywords) == 0 {
		return ""
	}

	// Score each keyword
	type scoredWord struct {
		word  string
		score int
	}
	var scored []scoredWord

	for _, word := range keywords {
		score := 1
		if s, ok := importantWords[word]; ok {
			score = s
		} else if len(word) > 5 {
			score = 3 // Longer words are often more meaningful
		} else if len(word) > 3 {
			score = 2
		}
		scored = append(scored, scoredWord{word: word, score: score})
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := 0; j < len(scored)-i-1; j++ {
			if scored[j].score < scored[j+1].score {
				scored[j], scored[j+1] = scored[j+1], scored[j]
			}
		}
	}

	// Take top N unique words
	seen := make(map[string]bool)
	var result []string
	for _, sw := range scored {
		if !seen[sw.word] && len(result) < maxWords {
			result = append(result, sw.word)
			seen[sw.word] = true
		}
	}

	return strings.Join(result, " ")
}

// GetClusterSummary extracts a 2-5 word summary from multiple problem descriptions in a cluster
func GetClusterSummary(descriptions []string, maxWords int) string {
	if len(descriptions) == 0 {
		return ""
	}

	// Count word frequencies across all descriptions in the cluster
	wordCounts := make(map[string]int)
	for _, desc := range descriptions {
		keywords := ExtractKeywords(desc)
		seen := make(map[string]bool)
		for _, word := range keywords {
			if !seen[word] {
				wordCounts[word]++
				seen[word] = true
			}
		}
	}

	// Domain-specific important words (same as ExtractKeyPhrase)
	importantWords := map[string]int{
		"rusak": 10, "pecah": 10, "retak": 10, "bocor": 10, "patah": 10,
		"bengkok": 10, "kotor": 10, "cacat": 10, "gagal": 10, "error": 10,
		"salah": 10, "keliru": 10, "tidak": 8, "kurang": 8, "miring": 10,
		"geser": 10, "longgar": 10, "beda": 8, "berbeda": 8,
		"material": 9, "bahan": 9, "kaca": 9, "aluminium": 9, "kayu": 9,
		"plat": 9, "besi": 9, "stainless": 9, "karet": 9, "plastik": 9,
		"cat": 9, "finishing": 9, "powder": 9, "coating": 9, "sealant": 9,
		"pintu": 8, "jendela": 8, "frame": 8, "panel": 8, "handle": 8,
		"kunci": 8, "engsel": 8, "rel": 8, "roller": 8, "glass": 8,
		"profil": 8, "aksesoris": 8, "bracket": 8, "corner": 8, "gasket": 8,
		"dimensi": 9, "ukuran": 9, "warna": 9, "bentuk": 9, "kualitas": 9,
		"spec": 9, "spesifikasi": 9, "tolerance": 9, "standar": 9, "reject": 10,
		"defect": 10, "scratch": 10, "dent": 10, "baret": 10, "penyok": 10,
		"welding": 8, "las": 8, "cutting": 8, "potong": 8, "drilling": 8,
		"bor": 8, "assembly": 8, "rakit": 8, "packing": 8, "kirim": 8,
		"produksi": 8, "proses": 7, "mesin": 8, "alat": 7, "setting": 8,
	}

	// Score words by frequency * importance
	type scoredWord struct {
		word  string
		score float64
	}
	var scored []scoredWord

	for word, freq := range wordCounts {
		importance := 1.0
		if imp, ok := importantWords[word]; ok {
			importance = float64(imp)
		} else if len(word) > 5 {
			importance = 3.0
		}
		// Score = frequency * importance (words appearing in more descriptions are better)
		score := float64(freq) * importance
		scored = append(scored, scoredWord{word: word, score: score})
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := 0; j < len(scored)-i-1; j++ {
			if scored[j].score < scored[j+1].score {
				scored[j], scored[j+1] = scored[j+1], scored[j]
			}
		}
	}

	// Take top N unique words
	var result []string
	for i := 0; i < len(scored) && len(result) < maxWords; i++ {
		result = append(result, scored[i].word)
	}

	if len(result) == 0 {
		// Fallback: just take first few words from first description
		return ExtractKeyPhrase(descriptions[0], maxWords)
	}

	return strings.Join(result, " ")
}

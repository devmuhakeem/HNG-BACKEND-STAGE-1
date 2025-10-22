package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Properties struct {
	Length                int            `json:"length"`
	IsPalindrome          bool           `json:"is_palindrome"`
	UniqueCharacters      int            `json:"unique_characters"`
	WordCount             int            `json:"word_count"`
	SHA256Hash            string         `json:"sha256_hash"`
	CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
}

type StoredString struct {
	ID         string     `json:"id"`
	Value      string     `json:"value"`
	Properties Properties `json:"properties"`
	CreatedAt  string     `json:"created_at"`
}

type CreateReq struct {
	Value interface{} `json:"value"`
}

var (
	store = struct {
		sync.RWMutex
		m map[string]StoredString
	}{m: map[string]StoredString{}}
)

func computeHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func charFreqMap(s string) map[string]int {
	m := map[string]int{}
	for _, r := range s {
		m[string(r)]++
	}
	return m
}

func isPalindrome(s string) bool {
	rs := []rune(strings.ToLower(s))
	i, j := 0, len(rs)-1
	for i < j {
		if rs[i] != rs[j] {
			return false
		}
		i++
		j--
	}
	return true
}

func wordCount(s string) int {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0
	}
	parts := regexp.MustCompile(`\s+`).Split(trimmed, -1)
	return len(parts)
}

func analyzeString(s string) Properties {
	freq := charFreqMap(s)
	return Properties{
		Length:                len([]rune(s)),
		IsPalindrome:          isPalindrome(s),
		UniqueCharacters:      len(freq),
		WordCount:             wordCount(s),
		SHA256Hash:            computeHash(s),
		CharacterFrequencyMap: freq,
	}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func validateCreateBody(body CreateReq) (string, int, error) {
	if body.Value == nil {
		return "", http.StatusBadRequest, errors.New(`missing "value" field`)
	}
	switch v := body.Value.(type) {
	case string:
		return v, 0, nil
	default:
		return "", http.StatusUnprocessableEntity, errors.New(`"value" must be a string`)
	}
}

func postStringsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body CreateReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	val, code, err := validateCreateBody(body)
	if err != nil {
		writeJSON(w, code, map[string]string{"error": err.Error()})
		return
	}
	props := analyzeString(val)
	id := props.SHA256Hash
	store.RLock()
	_, exists := store.m[id]
	store.RUnlock()
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "string already exists in the system"})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item := StoredString{
		ID:         id,
		Value:      val,
		Properties: props,
		CreatedAt:  now,
	}
	store.Lock()
	store.m[id] = item
	store.Unlock()
	writeJSON(w, http.StatusCreated, item)
}

func getStringByValueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/strings/")
	decoded, err := url.PathUnescape(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid URL-encoded string"})
		return
	}
	if decoded == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing string value in path"})
		return
	}
	id := computeHash(decoded)
	store.RLock()
	item, exists := store.m[id]
	store.RUnlock()
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "string does not exist in the system"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func parseBoolParam(v string) (bool, error) {
	if v == "true" {
		return true, nil
	}
	if v == "false" {
		return false, nil
	}
	return false, errors.New("invalid boolean")
}

func getAllStringsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	var (
		filterIsPalindrome *bool
		minLength          *int
		maxLength          *int
		wordCountFilter    *int
		containsCharacter  *rune
	)
	if v := q.Get("is_palindrome"); v != "" {
		b, err := parseBoolParam(strings.ToLower(v))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid is_palindrome value"})
			return
		}
		filterIsPalindrome = &b
	}
	if v := q.Get("min_length"); v != "" {
		x, err := strconv.Atoi(v)
		if err != nil || x < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid min_length"})
			return
		}
		minLength = &x
	}
	if v := q.Get("max_length"); v != "" {
		x, err := strconv.Atoi(v)
		if err != nil || x < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid max_length"})
			return
		}
		maxLength = &x
	}
	if v := q.Get("word_count"); v != "" {
		x, err := strconv.Atoi(v)
		if err != nil || x < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid word_count"})
			return
		}
		wordCountFilter = &x
	}
	if v := q.Get("contains_character"); v != "" {
		rs := []rune(v)
		if len(rs) != 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "contains_character must be a single character"})
			return
		}
		containsCharacter = &rs[0]
	}
	store.RLock()
	results := make([]StoredString, 0, len(store.m))
	for _, item := range store.m {
		ok := true
		if filterIsPalindrome != nil && item.Properties.IsPalindrome != *filterIsPalindrome {
			ok = false
		}
		if minLength != nil && item.Properties.Length < *minLength {
			ok = false
		}
		if maxLength != nil && item.Properties.Length > *maxLength {
			ok = false
		}
		if wordCountFilter != nil && item.Properties.WordCount != *wordCountFilter {
			ok = false
		}
		if containsCharacter != nil {
			found := false
			for ch := range item.Properties.CharacterFrequencyMap {
				if []rune(ch)[0] == *containsCharacter {
					found = true
					break
				}
			}
			if !found {
				ok = false
			}
		}
		if ok {
			results = append(results, item)
		}
	}
	store.RUnlock()
	filtersApplied := map[string]interface{}{}
	if filterIsPalindrome != nil {
		filtersApplied["is_palindrome"] = *filterIsPalindrome
	}
	if minLength != nil {
		filtersApplied["min_length"] = *minLength
	}
	if maxLength != nil {
		filtersApplied["max_length"] = *maxLength
	}
	if wordCountFilter != nil {
		filtersApplied["word_count"] = *wordCountFilter
	}
	if containsCharacter != nil {
		filtersApplied["contains_character"] = string(*containsCharacter)
	}
	resp := map[string]interface{}{
		"data":            results,
		"count":           len(results),
		"filters_applied": filtersApplied,
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseNaturalLanguage(query string) (map[string]interface{}, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil, errors.New("empty query")
	}
	parsed := map[string]interface{}{}
	if strings.Contains(q, "single word") || strings.Contains(q, "single-word") || strings.Contains(q, "one word") {
		parsed["word_count"] = 1
	}
	if strings.Contains(q, "palindrom") {
		parsed["is_palindrome"] = true
	}
	reLonger := regexp.MustCompile(`longer than\s+(\d+)`)
	if m := reLonger.FindStringSubmatch(q); len(m) == 2 {
		n, err := strconv.Atoi(m[1])
		if err == nil {
			parsed["min_length"] = n + 1
		}
	}
	reLonger2 := regexp.MustCompile(`longer than\s+(\d+)\s+characters`)
	if m := reLonger2.FindStringSubmatch(q); len(m) == 2 {
		n, err := strconv.Atoi(m[1])
		if err == nil {
			parsed["min_length"] = n + 1
		}
	}
	reContains := regexp.MustCompile(`containing the letter\s+([a-zA-Z])|contain the letter\s+([a-zA-Z])|containing\s+([a-zA-Z])|contain\s+([a-zA-Z])`)
	if m := reContains.FindStringSubmatch(q); len(m) >= 5 {
		for i := 1; i <= 4; i++ {
			if m[i] != "" {
				parsed["contains_character"] = strings.ToLower(m[i])
				break
			}
		}
	}
	if strings.Contains(q, "first vowel") || strings.Contains(q, "first vowel a") {
		parsed["contains_character"] = "a"
	}
	if _, ok := parsed["word_count"]; !ok {
		reWords := regexp.MustCompile(`\b(\d+)\s+word`)
		if m := reWords.FindStringSubmatch(q); len(m) == 2 {
			n, err := strconv.Atoi(m[1])
			if err == nil {
				parsed["word_count"] = n
			}
		}
	}
	if len(parsed) == 0 {
		return nil, errors.New("unable to parse natural language query")
	}
	if min, ok1 := parsed["min_length"].(int); ok1 {
		if max, ok2 := parsed["max_length"].(int); ok2 && min > max {
			return nil, errors.New("conflicting filters")
		}
		if maxf, ok3 := parsed["max_length"].(float64); ok3 && min > int(maxf) {
			return nil, errors.New("conflicting filters")
		}
	}
	return parsed, nil
}

func applyParsedFilters(parsed map[string]interface{}) ([]StoredString, error) {
	store.RLock()
	defer store.RUnlock()
	results := []StoredString{}
	for _, item := range store.m {
		ok := true
		if v, okp := parsed["is_palindrome"]; okp {
			if b, ok2 := v.(bool); ok2 {
				if item.Properties.IsPalindrome != b {
					ok = false
				}
			}
		}
		if v, okp := parsed["word_count"]; okp {
			switch vv := v.(type) {
			case int:
				if item.Properties.WordCount != vv {
					ok = false
				}
			case float64:
				if item.Properties.WordCount != int(vv) {
					ok = false
				}
			}
		}
		if v, okp := parsed["min_length"]; okp {
			switch vv := v.(type) {
			case int:
				if item.Properties.Length < vv {
					ok = false
				}
			case float64:
				if item.Properties.Length < int(vv) {
					ok = false
				}
			}
		}
		if v, okp := parsed["max_length"]; okp {
			switch vv := v.(type) {
			case int:
				if item.Properties.Length > vv {
					ok = false
				}
			case float64:
				if item.Properties.Length > int(vv) {
					ok = false
				}
			}
		}
		if v, okp := parsed["contains_character"]; okp {
			chStr := fmt.Sprintf("%v", v)
			if chStr == "" {
				ok = false
			} else {
				found := false
				for ch := range item.Properties.CharacterFrequencyMap {
					if strings.EqualFold(ch, chStr) {
						found = true
						break
					}
				}
				if !found {
					ok = false
				}
			}
		}
		if ok {
			results = append(results, item)
		}
	}
	return results, nil
}

func naturalLanguageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("query")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter is required"})
		return
	}
	parsed, err := parseNaturalLanguage(q)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	results, err := applyParsedFilters(parsed)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	resp := map[string]interface{}{
		"data":  results,
		"count": len(results),
		"interpreted_query": map[string]interface{}{
			"original":       q,
			"parsed_filters": parsed,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

func deleteStringHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/strings/")
	decoded, err := url.PathUnescape(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid URL-encoded string"})
		return
	}
	if decoded == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing string value in path"})
		return
	}
	id := computeHash(decoded)
	store.Lock()
	_, exists := store.m[id]
	if !exists {
		store.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "string does not exist in the system"})
		return
	}
	delete(store.m, id)
	store.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	http.HandleFunc("/strings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postStringsHandler(w, r)
			return
		}
		if r.Method == http.MethodGet {
			getAllStringsHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/strings/filter-by-natural-language", naturalLanguageHandler)
	http.HandleFunc("/strings/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getStringByValueHandler(w, r)
		case http.MethodDelete:
			deleteStringHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	fmt.Println("Server running on :8080")
	_ = http.ListenAndServe(":8080", nil)
}

package embed

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxSeqLen = 256

// Tokenizer is a BERT WordPiece tokenizer loaded from Hugging Face tokenizer.json.
type Tokenizer struct {
	vocab  map[string]int
	unkID  int
	clsID  int
	sepID  int
	padID  int
	prefix string
	maxLen int
}

type hfTokenizerFile struct {
	AddedTokens []struct {
		ID      int    `json:"id"`
		Content string `json:"content"`
		Special bool   `json:"special"`
	} `json:"added_tokens"`
	Model struct {
		UnkToken                string         `json:"unk_token"`
		ContinuingSubwordPrefix string         `json:"continuing_subword_prefix"`
		Vocab                   map[string]int `json:"vocab"`
	} `json:"model"`
}

// LoadTokenizer reads a Hugging Face tokenizer.json (WordPiece).
func LoadTokenizer(path string) (*Tokenizer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tokenizer: %w", err)
	}
	var f hfTokenizerFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("tokenizer: %w", err)
	}
	if len(f.Model.Vocab) == 0 {
		return nil, fmt.Errorf("tokenizer: empty vocab")
	}
	t := &Tokenizer{
		vocab:  f.Model.Vocab,
		prefix: f.Model.ContinuingSubwordPrefix,
		maxLen: maxSeqLen,
	}
	if t.prefix == "" {
		t.prefix = "##"
	}
	unk := f.Model.UnkToken
	if unk == "" {
		unk = "[UNK]"
	}
	t.unkID = t.idOr(unk, 100)
	t.clsID = t.idOr("[CLS]", 101)
	t.sepID = t.idOr("[SEP]", 102)
	t.padID = t.idOr("[PAD]", 0)
	for _, a := range f.AddedTokens {
		if a.Content != "" {
			t.vocab[a.Content] = a.ID
		}
	}
	return t, nil
}

func (t *Tokenizer) idOr(tok string, fallback int) int {
	if id, ok := t.vocab[tok]; ok {
		return id
	}
	return fallback
}

// Encode returns input_ids and attention_mask (no padding; includes [CLS]/[SEP]).
func (t *Tokenizer) Encode(text string) (ids, mask []int64) {
	words := bertWords(bertNormalize(text))
	ids = []int64{int64(t.clsID)}
	for _, w := range words {
		pieces := t.wordPiece(w)
		if len(ids)+len(pieces)+1 > t.maxLen {
			break
		}
		for _, p := range pieces {
			ids = append(ids, int64(p))
		}
	}
	ids = append(ids, int64(t.sepID))
	if len(ids) > t.maxLen {
		ids = ids[:t.maxLen]
		ids[t.maxLen-1] = int64(t.sepID)
	}
	mask = make([]int64, len(ids))
	for i := range mask {
		mask[i] = 1
	}
	return ids, mask
}

func (t *Tokenizer) wordPiece(word string) []int {
	if word == "" {
		return nil
	}
	if id, ok := t.vocab[word]; ok {
		return []int{id}
	}
	var out []int
	start := 0
	for start < len(word) {
		end := len(word)
		var found int
		ok := false
		for end > start {
			substr := word[start:end]
			if start > 0 {
				substr = t.prefix + substr
			}
			if id, exists := t.vocab[substr]; exists {
				found = id
				ok = true
				break
			}
			_, size := utf8.DecodeLastRuneInString(word[:end])
			if size <= 0 {
				end--
			} else {
				end -= size
			}
		}
		if !ok {
			return []int{t.unkID}
		}
		out = append(out, found)
		start = end
	}
	return out
}

func bertNormalize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func bertWords(s string) []string {
	var words []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		words = append(words, cur.String())
		cur.Reset()
	}
	for _, r := range s {
		switch {
		case unicode.IsSpace(r):
			flush()
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			flush()
			words = append(words, string(r))
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return words
}

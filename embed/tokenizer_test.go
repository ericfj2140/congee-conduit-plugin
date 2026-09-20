package embed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokenizerWordPiece(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokenizer.json")
	body := `{
		"model": {
			"unk_token": "[UNK]",
			"continuing_subword_prefix": "##",
			"vocab": {
				"[PAD]": 0,
				"[UNK]": 1,
				"[CLS]": 2,
				"[SEP]": 3,
				"red": 4,
				"bicycle": 5,
				"bi": 6,
				"##cycle": 7,
				".": 8
			}
		}
	}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	tok, err := LoadTokenizer(path)
	if err != nil {
		t.Fatal(err)
	}
	ids, mask := tok.Encode("Red bicycle.")
	if len(ids) != len(mask) || ids[0] != 2 || ids[len(ids)-1] != 3 {
		t.Fatalf("ids=%v mask=%v", ids, mask)
	}
	if ids[1] != 4 || ids[2] != 5 {
		t.Fatalf("want red bicycle, got %v", ids)
	}
}

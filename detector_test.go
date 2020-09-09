package chardet

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

var (
	zhGB18030text = []byte{
		71, 111, 202, 199, 71, 111, 111, 103, 108, 101, 233, 95, 176, 108, 181, 196, 210, 187, 214, 214, 190, 142, 215, 103, 208, 205, 163, 172, 129, 75, 176, 108,
		208, 205, 163, 172, 178, 162, 190, 223, 211, 208, 192, 172, 187, 248, 187, 216, 202, 213, 185, 166, 196, 220, 181, 196, 177, 224, 179, 204, 211, 239, 209, 212,
		161, 163, 10,
	}
)

func TestDetector(t *testing.T) {
	type fileCharsetLanguagego struct {
		File     string
		IsHTML   bool
		Charset  string
		Language string
	}
	var data = []fileCharsetLanguagego{
		{"utf8.html", true, "UTF-8", ""},
		{"utf8_bom.html", true, "UTF-8", ""},
		{"8859_1_en.html", true, "ISO 8859-1", "en"},
		{"8859_1_da.html", true, "ISO 8859-1", "da"},
		{"8859_1_de.html", true, "ISO 8859-1", "de"},
		{"8859_1_es.html", true, "ISO 8859-1", "es"},
		{"8859_1_fr.html", true, "ISO 8859-1", "fr"},
		{"8859_1_pt.html", true, "ISO 8859-1", "pt"},
		{"shift_jis.html", true, "Shift_JIS", "ja"},
		{"gb18030.html", true, "GB18030", "zh"},
		{"euc_jp.html", true, "EUC-JP", "ja"},
		{"euc_kr.html", true, "EUC-KR", "ko"},
		{"big5.html", true, "Big5", "zh"},
	}

	textDetector := NewTextDetector()
	htmlDetector := NewHTMLDetector()
	buffer := make([]byte, 32<<10)
	for _, d := range data {
		f, err := os.Open(filepath.Join("testdata", d.File))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		size, _ := io.ReadFull(f, buffer)
		input := buffer[:size]
		var detector = textDetector
		if d.IsHTML {
			detector = htmlDetector
		}
		result, err := detector.DetectBest(input)
		if err != nil {
			t.Fatal(err)
		}
		if result.Charset != d.Charset {
			t.Errorf("%s: expected charset %s, actual %s", d.File, d.Charset, result.Charset)
		}
		if result.Language != d.Language {
			t.Errorf("%s: expected language %s, actual %s", d.File, d.Language, result.Language)
		}
	}
}

func TestNewTextDetector(t *testing.T) {
	detector := NewTextDetector()
	result, err := detector.DetectBest(zhGB18030text)
	if err == nil {
		if result.Charset != "GB18030" {
			t.Errorf("result.Charset = %s, want GB18030", result.Charset)
			return
		}
		if result.Language != "zh" {
			t.Errorf("result.Charset = %s, want zh", result.Language)
			return
		}
		return
	}
	t.Errorf(err.Error())
}

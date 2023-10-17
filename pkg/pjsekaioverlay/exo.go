package pjsekaioverlay

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func encodeString(str string) string {
	bytes := utf16.Encode([]rune(str))
	encoded := make([]string, 1024)
	if len(str) > 1024 {
		panic("too long string")
	}
	for i := range encoded {
		var hex string
		if i >= len(bytes) {
			hex = fmt.Sprintf("%04x", 0)
		} else {
			hex = fmt.Sprintf("%02x%02x", bytes[i]&0xff, bytes[i]>>8)
		}

		encoded[i] = hex
	}

	return strings.Join(encoded, "")
}

//go:embed main.exo
var rawBaseExo []byte

func WriteExoFiles(assets string, destDir string, title string, description string) error {
	baseExo := string(rawBaseExo)
	replacedExo := baseExo
	mapping := []string{
		"{assets}", strings.ReplaceAll(assets, "\\", "/"),
		"{dist}", strings.ReplaceAll(destDir, "\\", "/"),
		"{text:difficulty}", encodeString("MASTER"),
		"{text:title}", encodeString(title),
		"{text:description}", encodeString(description),
	}
	for i := range mapping {
		if i%2 == 0 {
			continue
		}
		if !strings.Contains(replacedExo, mapping[i-1]) {
      panic(fmt.Sprintf("exoファイルの生成に失敗しました（%sが見つかりません）", mapping[i-1]))
		}
		replacedExo = strings.ReplaceAll(replacedExo, mapping[i-1], mapping[i])
	}
	replacedExo = strings.ReplaceAll(replacedExo, "\n", "\r\n")
	encodedExo, err := io.ReadAll(transform.NewReader(
		strings.NewReader(replacedExo), japanese.ShiftJIS.NewEncoder()))
	if err != nil {
		return fmt.Errorf("エンコードに失敗しました（%w）", err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "main.exo"),
		encodedExo,
		0644); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました（%w）", err)
	}

	return nil
}

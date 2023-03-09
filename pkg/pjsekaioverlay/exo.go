package pjsekaioverlay

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

//go:embed exo/00_root.exo
var rootExo []byte

//go:embed exo/01_main.exo
var mainExo []byte

//go:embed exo/02_start.exo
var startExo []byte

//go:embed exo/03_bg.exo
var bgExo []byte

func WriteExoFiles(assets string, destDir string) error {
	filenames := []string{"00_root.exo", "01_main.exo", "02_start.exo", "03_bg.exo"}
	for i, baseExoRaw := range [][]byte{rootExo, mainExo, startExo, bgExo} {
		filename := filenames[i]

    baseExo := string(baseExoRaw)
    replacedExo := baseExo
		mapping := []string{
			"{assets}", assets,
			"{background}", filepath.Join(destDir, "background.png"),
			"{cover}", filepath.Join(destDir, "cover.png"),
			"{ped}", strings.ReplaceAll(filepath.Join(destDir, "data.ped"), "\\", "\\\\"),
		}
		for i := range mapping {
			if i%2 == 0 {
				continue
			}
			replacedExo = strings.ReplaceAll(replacedExo, mapping[i-1], mapping[i])
		}
    replacedExo = strings.ReplaceAll(replacedExo, "\n", "\r\n")
		encodedExo, err := io.ReadAll(transform.NewReader(
			strings.NewReader(replacedExo), japanese.ShiftJIS.NewEncoder()))
    if err != nil {
      return fmt.Errorf("エンコードに失敗しました（%w）", err)
    }
		if err := os.WriteFile(filepath.Join(destDir, filename),
			encodedExo,
			0644); err != nil {
			return fmt.Errorf("ファイルの書き込みに失敗しました（%w）", err)
		}

	}
	return nil
}

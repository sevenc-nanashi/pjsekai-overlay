package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/sevenc-nanashi/pjsekai-overlay/pkg/pjsekaioverlay"
  "github.com/srinathh/gokilo/rawmode"
  ansi "github.com/k0kubun/go-ansi"
)

func origMain(isOptionSpecified bool) {
	Title()

	var skipAviutlInstall bool
	flag.BoolVar(&skipAviutlInstall, "no-aviutl-install", false, "AviUtlオブジェクトのインストールをスキップします。")

	var outDir string
	flag.StringVar(&outDir, "out-dir", "./dist/_chartId_", "出力先ディレクトリを指定します。_chartId_ は譜面IDに置き換えられます。")

	var teamPower int
	flag.IntVar(&teamPower, "team-power", 250000, "総合力を指定します。")

	var apCombo bool
	flag.BoolVar(&apCombo, "ap-combo", true, "コンボのAP表示を有効にします。")

	flag.Usage = func() {
		ansi.Println("Usage: pjsekai-overlay [譜面ID] [オプション]")
		flag.PrintDefaults()
	}

	flag.Parse()

	if !skipAviutlInstall {
		success := pjsekaioverlay.TryInstallObject()
		if success {
			ansi.Println("AviUtlオブジェクトのインストールに成功しました。")
		}
	}

	var chartId string
	if flag.Arg(0) != "" {
		chartId = flag.Arg(0)
		ansi.Printf("譜面ID: %s\n", color.GreenString(chartId))
	} else {
		ansi.Print("譜面IDをプレフィックス込みで入力して下さい。\n> ")
		fmt.Scanln(&chartId)
		ansi.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartId))
	}

	chartSource, err := pjsekaioverlay.DetectChartSource(chartId)
	if err != nil {
		ansi.Println(color.RedString("譜面のサーバーを判別できませんでした。プレフィックスも込め、正しい譜面IDを入力して下さい。"))
		return
	}
	ansi.Printf("%s%s%s から譜面を取得中... ", RgbColorEscape(chartSource.Color), chartSource.Name, ResetEscape())
	chart, err := pjsekaioverlay.FetchChart(chartSource, chartId)

	if err != nil {
		ansi.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	ansi.Println(color.GreenString("成功"))
	ansi.Printf("  %s / %s - %s (Lv. %s)\n",
		color.CyanString(chart.Title),
		color.CyanString(chart.Artists),
		color.CyanString(chart.Author),
		color.MagentaString(strconv.Itoa(chart.Rating)),
	)

	ansi.Printf("exeのパスを取得中... ")
	executablePath, err := os.Executable()
	if err != nil {
		ansi.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	ansi.Println(color.GreenString("成功"))

	formattedOutDir := filepath.Join(filepath.Dir(executablePath), strings.Replace(outDir, "_chartId_", chartId, -1))
	ansi.Printf("出力先ディレクトリ: %s\n", color.CyanString(filepath.Dir(formattedOutDir)))

	ansi.Print("ジャケットをダウンロード中... ")
	err = pjsekaioverlay.DownloadCover(chartSource, chart, formattedOutDir)
	if err != nil {
		ansi.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	ansi.Println(color.GreenString("成功"))

	ansi.Print("背景をダウンロード中... ")
	err = pjsekaioverlay.DownloadBackground(chartSource, chart, formattedOutDir)
	if err != nil {
		ansi.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	ansi.Println(color.GreenString("成功"))

	ansi.Print("譜面を解析中... ")
	levelData, err := pjsekaioverlay.FetchLevelData(chartSource, chart)

	if err != nil {
		ansi.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	ansi.Println(color.GreenString("成功"))

	if !isOptionSpecified {
		ansi.Print("総合力を指定してください。\n> ")
		var tmpTeamPower string
		fmt.Scanln("%d", &tmpTeamPower)
		ansi.Printf("\033[A\033[2K\r> %s\n", color.GreenString(tmpTeamPower))
		teamPower, err = strconv.Atoi(tmpTeamPower)
	}

	ansi.Print("スコアを計算中... ")
	scoreData := pjsekaioverlay.CalculateScore(chart, levelData, teamPower)

	ansi.Println(color.GreenString("成功"))

	if !isOptionSpecified {
		ansi.Print("コンボのAP表示を有効にしますか？ (Y/n)\n> ")
		var tmpEnableComboAp string
		fmt.Scanln("%s", &tmpEnableComboAp)
		ansi.Printf("\033[A\033[2K\r> %s\n", color.GreenString(tmpEnableComboAp))
		if tmpEnableComboAp == "Y" || tmpEnableComboAp == "y" || tmpEnableComboAp == "" {
			apCombo = true
		}
	}
	executableDir := filepath.Dir(executablePath)
	assets := filepath.Join(executableDir, "assets")

	ansi.Print("pedファイルを生成中... ")

	err = pjsekaioverlay.WritePedFile(scoreData, assets, apCombo, filepath.Join(formattedOutDir, "data.ped"))

	if err != nil {
		ansi.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	ansi.Println(color.GreenString("成功"))

	ansi.Print("exoファイルを生成中... ")

	err = pjsekaioverlay.WriteExoFiles(assets, formattedOutDir)

	if err != nil {
		ansi.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	ansi.Println(color.GreenString("成功"))

	ansi.Println(color.GreenString("\n全ての処理が完了しました。https://github.com/sevenc-nanashi/pjsekai-overlay#readme を参考に、ファイルをAviUtlにインポートして下さい。"))
}

func main() {
	isOptionSpecified := len(os.Args) > 1

	origMain(isOptionSpecified)

	if !isOptionSpecified {
		ansi.Print(color.CyanString("\n何かキーを押すと終了します..."))

    before, _ := rawmode.Enable()
		bufio.NewReader(os.Stdin).ReadByte()
    rawmode.Restore(before)
	}
}

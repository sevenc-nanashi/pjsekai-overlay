package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/v57/github"
	"github.com/sevenc-nanashi/pjsekai-overlay/pkg/pjsekaioverlay"
	"github.com/sevenc-nanashi/pjsekai-overlay/pkg/sonolus"
	"github.com/srinathh/gokilo/rawmode"
	"golang.org/x/sys/windows"
)

func shouldCheckUpdate() bool {
	executablePath, err := os.Executable()
	if err != nil {
		return false
	}
	updateCheckFile, err := os.OpenFile(filepath.Join(filepath.Dir(executablePath), ".update-check"), os.O_RDONLY, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		return false
	}
	defer updateCheckFile.Close()

	scanner := bufio.NewScanner(updateCheckFile)
	scanner.Scan()
	lastCheckTime, err := strconv.ParseInt(scanner.Text(), 10, 64)
	if err != nil {
		return false
	}

	return time.Now().Unix()-lastCheckTime > 60*60*24
}

func checkUpdate() {
	githubClient := github.NewClient(nil)
	release, _, err := githubClient.Repositories.GetLatestRelease(context.Background(), "sevenc-nanashi", "pjsekai-overlay")
	if err != nil {
		return
	}

	executablePath, err := os.Executable()
	updateCheckFile, err := os.OpenFile(filepath.Join(filepath.Dir(executablePath), ".update-check"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer updateCheckFile.Close()
	updateCheckFile.WriteString(strconv.FormatInt(time.Now().Unix(), 10))

	latestVersion := strings.TrimPrefix(release.GetTagName(), "v")
	if latestVersion == pjsekaioverlay.Version {
		return
	}
	fmt.Printf("新しいバージョンがリリースされています：v%s -> v%s\n", pjsekaioverlay.Version, latestVersion)
	fmt.Printf("ダウンロード：%s\n", release.GetHTMLURL())
}

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
		fmt.Println("Usage: pjsekai-overlay [譜面ID] [オプション]")
		flag.PrintDefaults()
	}

	flag.Parse()

	if shouldCheckUpdate() {
		checkUpdate()
	}

	if !skipAviutlInstall {
		success := pjsekaioverlay.TryInstallObject()
		if success {
			fmt.Println("AviUtlオブジェクトのインストールに成功しました。")
		}
	}

	var chartId string
	if flag.Arg(0) != "" {
		chartId = flag.Arg(0)
		fmt.Printf("譜面ID: %s\n", color.GreenString(chartId))
	} else {
		fmt.Print("譜面IDをプレフィックス込みで入力して下さい。\n> ")
		fmt.Scanln(&chartId)
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartId))
	}

	chartSource, err := pjsekaioverlay.DetectChartSource(chartId)
	if err != nil {
		fmt.Println(color.RedString("譜面のサーバーを判別できませんでした。プレフィックスも込め、正しい譜面IDを入力して下さい。"))
		return
	}
	fmt.Printf("%s%s%s から譜面を取得中... ", RgbColorEscape(chartSource.Color), chartSource.Name, ResetEscape())
	chart, err := pjsekaioverlay.FetchChart(chartSource, chartId)

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}
	if chart.Engine.Version != 13 {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：このエンジンはサポートされていません。（バージョン%d）", chart.Engine.Version)))
		return
	}

	fmt.Println(color.GreenString("成功"))
	fmt.Printf("  %s / %s - %s (Lv. %s)\n",
		color.CyanString(chart.Title),
		color.CyanString(chart.Artists),
		color.CyanString(chart.Author),
		color.MagentaString(strconv.Itoa(chart.Rating)),
	)

	fmt.Printf("exeのパスを取得中... ")
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("成功"))

	cwd, err := os.Getwd()

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	formattedOutDir := filepath.Join(cwd, strings.Replace(outDir, "_chartId_", chartId, -1))
	fmt.Printf("出力先ディレクトリ: %s\n", color.CyanString(filepath.Dir(formattedOutDir)))

	fmt.Print("ジャケットをダウンロード中... ")
	err = pjsekaioverlay.DownloadCover(chartSource, chart, formattedOutDir)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("成功"))

	fmt.Print("背景をダウンロード中... ")
	err = pjsekaioverlay.DownloadBackground(chartSource, chart, formattedOutDir)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("成功"))

	fmt.Print("譜面を解析中... ")
	levelData, err := pjsekaioverlay.FetchLevelData(chartSource, chart)

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("成功"))

	if !isOptionSpecified {
		fmt.Print("総合力を指定してください。\n> ")
		var tmpTeamPower string
		fmt.Scanln(&tmpTeamPower)
		teamPower, err = strconv.Atoi(tmpTeamPower)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
			return
		}
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(tmpTeamPower))

	}

	fmt.Print("スコアを計算中... ")
	scoreData := pjsekaioverlay.CalculateScore(chart, levelData, teamPower)

	fmt.Println(color.GreenString("成功"))

	if !isOptionSpecified {
		fmt.Print("コンボのAP表示を有効にしますか？ (Y/n)\n> ")
		before, _ := rawmode.Enable()
		tmpEnableComboApByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpEnableComboAp := string(tmpEnableComboApByte)
		rawmode.Restore(before)
		fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpEnableComboAp))
		if tmpEnableComboAp == "Y" || tmpEnableComboAp == "y" || tmpEnableComboAp == "" {
			apCombo = true
		} else {
			apCombo = false
		}
	}
	executableDir := filepath.Dir(executablePath)
	assets := filepath.Join(executableDir, "assets")

	fmt.Print("pedファイルを生成中... ")

	err = pjsekaioverlay.WritePedFile(scoreData, assets, apCombo, filepath.Join(formattedOutDir, "data.ped"), sonolus.LevelInfo{Rating: chart.Rating})

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("成功"))

	fmt.Print("exoファイルを生成中... ")

	composerAndVocals := []string{chart.Artists, "？"}
	if separateAttempt := strings.Split(chart.Artists, " / "); chartSource.Id == "chart_cyanvas" && len(separateAttempt) <= 2 {
		composerAndVocals = separateAttempt
	}

	artists := fmt.Sprintf("作詞：？    作曲：%s    編曲：？\r\nVo：%s   譜面作成：%s", composerAndVocals[0], composerAndVocals[1], chart.Author)

	err = pjsekaioverlay.WriteExoFiles(assets, formattedOutDir, chart.Title, artists)

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("失敗：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("成功"))

	fmt.Println(color.GreenString("\n全ての処理が完了しました。READMEの規約を確認した上で、exoファイルをAviUtlにインポートして下さい。"))
}

func main() {
	isOptionSpecified := len(os.Args) > 1
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	origMain(isOptionSpecified)

	if !isOptionSpecified {
		fmt.Print(color.CyanString("\n何かキーを押すと終了します..."))

		before, _ := rawmode.Enable()
		bufio.NewReader(os.Stdin).ReadByte()
		rawmode.Restore(before)
	}
}

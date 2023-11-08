# pjsekai-overlay / プロセカ風動画作成補助ツール

pjsekai-overlay は、プロセカの創作譜面をプロセカ風の動画にするためのオープンソースのツールです。

> **Warning**
> **for English users:**\
> This tool is only for Japanese users (as AviUtl is only available in Japanese).\
> Please do not use this tool, **DO NOT open issues, or request help in Sonolus / Chart Cyanvas Discord servers**.

## 必須事項

- [AviUtl](http://spring-fragrance.mints.ne.jp/aviutl/) + [拡張編集プラグイン](http://spring-fragrance.mints.ne.jp/aviutl/) （[導入方法](https://aviutl.info/dl-innsuto-ru/)）  
  （強く推奨：[patch.aul](https://scrapbox.io/ePi5131/patch.aul)）
- AviUtlの基本的な知識

## 動画の作り方

1. [譜面を作る](https://wiki.purplepalette.net/create-charts)
2. [Sonolus](https://sonolus.com/)で譜面を撮影する
   - [Potato Leaves](https://github.com/sevenc-nanashi/potato_leaves)、または [Chart Cyanvas](https://cc.sevenc7c.com)で撮影してください。
   - 「Hide UI」をオンにしてください。
3. 撮影したプレイ動画のファイルをパソコンに転送する
   - Google Drive など
4. [ffmpeg](https://www.ffmpeg.org/)で再エンコードする
   - AviUtl で読み込むため
5. 下の利用方法に従って UI を後付けする

## 利用方法

0. 1280x720, 60fps で aviutl のプロジェクトを作成する
1. 右の Releases から最新のバージョンの zip をダウンロードする
2. zip を解凍する
3. AviUtl を起動する
   - pjsekai-overlay が起動する前に AviUtl を起動するとオブジェクトのインストールが行われます。
4. `pjsekai-overlay.exe` を起動する
5. 譜面 ID を入力する
   - Potato Leaves の場合は `ptlv-` を、Chart Cyanvas の場合は `chcy-` を先頭につけたまま入力してください。

## 注意

動画の概要欄などに、

- `名無し｡`という名前
- このリポジトリへのリンク
- `https://sevenc7c.com`へのリンク

が含まれている文章を載せて下さい。

#### 例

```
プロセカ風動画作成補助ツール：
  https://github.com/sevenc-nanashi/pjsekai-overlay
  作成：名無し｡ （ https://sevenc7c.com ）
```

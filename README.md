GM - Golang Minimum text editor (凸)/
===============

GM はGo言語で作成したミニマムなオールインワンバイナリなテキストエディターです。

- Emacs風キーバインド。Ctrl-F/B/N/P などで編集可能
- 「なんちゃってSKK」実装。SKK-JISYO.L も実行ファイル内に内蔵
- 文字コードはおとこらしく、常に UTF8, LF
- 設定ファイルなし。ユーザ辞書もメモリ上にしかない

使い方
------

```
gm ファイル名
```

ファイル名は省略できません。セーブは C-xC-s、終了は C-xC-c です。
「セーブしてよいですか?」「終了してよいですか」などという問い合わせをする親切さはまだありません。
コピーも行単位でしかサポートしていません。
(ペーストは対応してる)

ビルド方法
----------

まだ開発中なので、バイナリはリリースしていません。

make すると curl で [skk-dev/dict] から SKK-JISYO.L をダウンロードして、bzip2 で圧縮します。
作成された SKK-JISYO.L.bz2 は Go の embedパッケージで、実行可能ファイルの中に組み込まれます。

[skk-dev/dict]: https://github.com/skk-dev/dict

使っているパッケージ
--------------------

- [go-readline-ny] 一行入力用パッケージ
- [go-multiline-ny] Ctrl-P/N が押下されるたびに、上下の行に [go-readline-ny] の一行入力を移動させるという荒技で簡易テキストエディターを実現するパッケージ
- [go-readline-skk] なんちゃってSKK

本ツールは実は [go-readline-skk]、[go-multiline-ny] のテスト用ツールという位置付けだったりします。

[go-readline-ny]: https://github.com/nyaosorg/go-readline-ny
[go-readline-skk]: https://github.com/hymkor/go-readline-skk
[go-multiline-ny]: https://github.com/hymkor/go-multiline-ny

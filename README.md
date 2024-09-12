<h2 align="center">
    <p align="center">gh-migrate2</p>
</h2>

<h3 align="center">
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate2/issues">Report Bug</a> &nbsp; &nbsp;
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate2/issues">Request Feature</a>
</h3>

```bash
$ gh migrate -h
PRを作成します

Usage:
  gh-migrate2 [flags]

Flags:
  -a, --astgrep string   引数にあるymlファイルをast-grepとして実行します
  -c, --cmd string       引数にあるコマンドを実行します
  -f, --force            cacheを削除して再取得します
  -h, --help             help for gh-migrate2
  -r, --repo string      リポジトリ名
  -g, --semgrep string   引数にあるymlファイルをsemgrepとして実行します
  -s, --sh string        引数にあるシェルスクリプトファイルを実行します
```

## Usage

```bash
# Install
gh extension install HikaruEgashira/gh-migrate2
gh migrate --repo HikaruEgashira/gh-migrate2 -s "sed -cmd '' 's/gh-migrate2/gh-migrate22/g' README.md"

https://github.com/HikaruEgashira/gh-migrate2/pull/10
```

## Acknowledgements

- https://docs.github.com/ja/github-cli/github-cli/creating-github-cli-extensions

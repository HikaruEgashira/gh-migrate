<h2 align="center">
    <p align="center">gh-migrate</p>
</h2>

<h3 align="center">
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate/issues">Report Bug</a> &nbsp; &nbsp;
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate/issues">Request Feature</a>
</h3>

```bash
$ gh migrate -h
PRを作成します

Usage:
  gh-migrate [flags]

Flags:
  -c, --cmd string    引数にあるコマンドを実行します
  -f, --force         cacheを削除して再取得します
  -h, --help          help for gh-migrate
  -r, --repo string   リポジトリ名
  -s, --sh string     引数にあるシェルスクリプトファイルを実行します
```

## Usage

```bash
# Install
gh extension install HikaruEgashira/gh-migrate
gh migrate --repo HikaruEgashira/gh-migrate
<h2 align="center">
    <p align="center">gh-migrate</p>
</h2>
...
```
  
## Acknowledgements

- https://docs.github.com/ja/github-cli/github-cli/creating-github-cli-extensions

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
  -r, --repo string      リポジトリ名
  -f, --force            cacheを削除して再取得します

  -c, --cmd string       引数にあるコマンドを実行します
  -s, --sh string        引数にあるシェルスクリプトファイルを実行します
  -g, --semgrep string   引数にあるymlファイルをsemgrepとして実行します
  -a, --astgrep string   引数にあるymlファイルをast-grepとして実行します

      --open string       作成したPRをブラウザで開きます
      --with-dev string   作成したPRをgithub.devで開きます

  -h, --help             help for gh-migrate
```

## Usage

```bash
# Install
gh extension install HikaruEgashira/gh-migrate
```

### Example1

```bash
gh migrate --repo HikaruEgashira/gh-migrate -s "sed -cmd '' 's/gh-migrate/gh-migrate2/g' README.md"

https://github.com/HikaruEgashira/gh-migrate/pull/10
```

### Example2: GitHub Actionsのactions/checkoutをv4に変更する

```yml
# ./example/upgrade-checkout.yml
id: upgrade-checkout
language: yml
rule: {pattern: "uses: $NAME"}
constraints: {NAME: {regex: ^actions/checkout}}
fix: "uses: actions/checkout@v4"
```

```bash
gh api --paginate "/search/code?q=user:HikaruEgashira+actions/checkout" -q ".items.[].repository.name" | sort -u | xargs -I {} gh migrate --repo HikaruEgashira/{} --astgrep ./ast-grep/rules/upgrade-actions-checkout.yml
```

## Acknowledgements

- https://docs.github.com/ja/github-cli/github-cli/creating-github-cli-extensions

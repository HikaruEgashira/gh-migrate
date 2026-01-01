<h2 align="center">
    <p align="center">gh-migrate</p>
</h2>

<h3 align="center">
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate/issues">Report Bug</a> &nbsp; &nbsp;
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate/issues">Request Feature</a>
</h3>

```bash
$ gh migrate -h
Creates a PR

Usage:
  gh-migrate [flags]

Flags:
  -r, --repo string      Repository name
  -f, --force            Delete cache and re-fetch

  -c, --cmd string       Execute the command provided as an argument
  -s, --sh string        Execute the shell script file provided as an argument
  -g, --semgrep string   Execute the yml file provided as an argument as semgrep
  -a, --astgrep string   Execute the yml file provided as an argument as ast-grep

      --open string       Open the created PR in the browser
      --with-dev string   Open the created PR in github.dev

  -h, --help             help for gh-migrate
```

## Usage

```bash
# Install
gh extension install HikaruEgashira/gh-migrate
```

### Example1

```bash
gh migrate --repo HikaruEgashira/gh-migrate --cmd "sed -i '' 's/gh-migrate/gh-migrate2/g' README.md"

https://github.com/HikaruEgashira/gh-migrate/pull/10
```

### Example2: Upgrade GitHub Actions actions/checkout to v4

```yml
# ./example/upgrade-checkout.yml
id: upgrade-checkout
language: yml
rule: {pattern: "uses: $NAME"}
constraints: {NAME: {regex: ^actions/checkout}}
fix: "uses: actions/checkout@v4"
```

```bash
gh api --paginate "/search/code?q=user:HikaruEgashira+actions/checkout" -q ".items.[].repository.name" | sort -u | xargs -I {} gh migrate --repo HikaruEgashira/{} --astgrep rules/upgrade-actions-checkout.yml
```

### Example3: Claude Code Integration

Use Claude Code to make AI-powered changes across repositories.

```bash
# Single repository
gh migrate --repo owner/repo --prompt "Add node.js gitignore"

# Multiple repositories
gh migrate --repo owner/repo1,owner/repo2 --prompt "Fix typos in README"

# Auto-approve mode
gh migrate --repo owner/repo --prompt "Add comments to main.go" --auto-approve
```

## Acknowledgements

- https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions

<h2 align="center">
    <p align="center">gh-migrate</p>
</h2>

<h3 align="center">
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate/issues">Report Bug</a> &nbsp; &nbsp;
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate/issues">Request Feature</a>
</h3>

```bash
$ gh migrate -h
Usage:
  gh-migrate [flags]
  gh-migrate [command]

Available Commands:
  learn       Learn from a PR or commit and generate a reusable prompt

Flags:
      --auto-approve      Auto-approve permission requests from Claude Code
  -c, --cmd string        Execute command or script file (auto-detects if argument is a file path)
  -f, --force             Delete cache and re-fetch
  -h, --help              help for gh-migrate
      --open string       Open the created PR in the browser
  -P, --prompt string     Execute Claude Code with the prompt provided as an argument
  -r, --repo string       Specify repository name (multiple repositories can be specified with comma separation)
      --template string   Path to a local PR template file (overrides repository template)
  -t, --title string      Specify the title of the PR
      --with-dev string   Open the created PR in github.dev
  -w, --workpath string   Specify the path of the working directory
```

## Usage

```bash
# Install
gh extension install HikaruEgashira/gh-migrate
```

### Example1: Text Replacement with sed

```bash
gh migrate --repo HikaruEgashira/gh-migrate-demo --cmd "sed -i '' 's/Demo/Updated Demo/g' README.md"

https://github.com/HikaruEgashira/gh-migrate-demo/pull/19
```

### Example2: Add Security Policy with Claude Code

```bash
gh migrate --repo HikaruEgashira/gh-migrate-demo --prompt "Add SECURITY.md with vulnerability reporting guidelines"

https://github.com/HikaruEgashira/gh-migrate-demo/pull/22
```

### Example3: Learn from PR and Generate Reusable Prompt

```bash
# Learn from a PR and generate a Claude Code slash command
gh migrate learn https://github.com/owner/repo/pull/123 --name "add-gitignore"

# Learn from a commit
gh migrate learn https://github.com/owner/repo/commit/abc1234
```

### Example4: Use Custom PR Template

```bash
gh migrate --repo HikaruEgashira/gh-migrate-demo --prompt "Update dependencies" --template ./templates/pr-template.md
```

## Acknowledgements

- https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions

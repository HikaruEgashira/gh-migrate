<h2 align="center">
    <p align="center">gh-migrate</p>
</h2>

<h3 align="center">
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate/issues">Report Bug</a> &nbsp; &nbsp;
🔹<a  href="https://github.com/HikaruEgashira/gh-migrate/issues">Request Feature</a>
</h3>

```bash
$ gh migrate -h
gh-migrate is a tool that creates PRs for specified repositories.

Available Commands:
  exec        Execute a command or script and create a PR
  prompt      Execute Claude Code with a prompt and create a PR
  learn       Learn from a PR or commit and generate a reusable prompt
```

![demo](docs/examples/demo.gif)

## Usage

```bash
# Install
gh extension install HikaruEgashira/gh-migrate
```

### Example1: Text Replacement with sed

```bash
gh migrate exec "sed -i '' 's/Demo/Updated Demo/g' README.md" --repo HikaruEgashira/gh-migrate-demo

https://github.com/HikaruEgashira/gh-migrate-demo/pull/19
```

### Example2: Add Security Policy with Claude Code

```bash
gh migrate prompt "Add SECURITY.md with vulnerability reporting guidelines" --repo HikaruEgashira/gh-migrate-demo

https://github.com/HikaruEgashira/gh-migrate-demo/pull/22
```

### Example3: Learn from PR and Generate Reusable Prompt

```bash
# Learn from a PR and generate a Claude Code slash command
gh migrate learn https://github.com/owner/repo/pull/123 --name "add-license-file"

# Use slash command
gh migrate prompt --repo HikaruEgashira/gh-migrate-demo --prompt-file ./.claude/commands/add-license-file.md
```

### Example4: Use Custom PR Template

```bash
gh migrate prompt "Update dependencies" --repo HikaruEgashira/gh-migrate-demo --template ./templates/pr-template.md
```

## Acknowledgements

- https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions

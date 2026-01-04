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
  prompt    Execute Claude Code with a prompt and create a PR
  exec      Execute a command or script and create a PR
  learn     Learn from a PR or commit and generate a reusable prompt

Examples:
  gh migrate prompt "Add gitignore" --repo owner/repo
  gh migrate exec "sed -i 's/old/new/g' file.txt" --repo owner/repo
  gh migrate learn https://github.com/owner/repo/pull/123

For detailed usage examples and flag descriptions, please refer to the README.

Usage:
  gh-migrate [flags]
  gh-migrate [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  exec        Execute a command or script and create a PR
  help        Help about any command
  learn       Learn from a PR or commit and generate a reusable prompt
  prompt      Execute Claude Code with a prompt and create a PR

Flags:
  -h, --help   help for gh-migrate
```

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
gh migrate learn https://github.com/owner/repo/pull/123 --name "add-gitignore"

# Learn from a commit
gh migrate learn https://github.com/owner/repo/commit/abc1234
```

### Example4: Use Custom PR Template

```bash
gh migrate prompt "Update dependencies" --repo HikaruEgashira/gh-migrate-demo --template ./templates/pr-template.md
```

## Migration from v4.x to v5.0

v5.0 introduces a breaking change to the CLI structure. The `--prompt` and `--cmd` flags have been converted to subcommands.

### Before (v4.x)
```bash
gh migrate --repo owner/repo --prompt "Add gitignore"
gh migrate --repo owner/repo --cmd "sed -i 's/old/new/g' file.txt"
gh migrate --repo owner/repo --cmd "sed ..." --title "Custom Title"
```

### After (v5.0)
```bash
gh migrate prompt "Add gitignore" --repo owner/repo
gh migrate exec "sed -i 's/old/new/g' file.txt" --repo owner/repo
gh migrate exec "sed ..." --repo owner/repo --title "Custom Title"
```

### Key Changes
- `--prompt` → `prompt` subcommand with positional argument
- `--cmd` → `exec` subcommand with positional argument
- `--title` flag is now only available in `exec` subcommand
- All other flags remain the same but are now local to each subcommand

## Acknowledgements

- https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions

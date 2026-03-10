# grove 🌳

`grove` is a CLI for creating multi-repository development workspaces backed by `git worktree`.

A workspace is a directory containing one worktree per repository, while canonical clones are cached separately.

## Layout

By default, `grove` uses:

```text
~/.grove/repos/         # canonical cached clones
~/groves/<workspace>/   # generated workspaces
```

Example:

```text
~/groves/auth-feature/
  api/
  web/
  worker/
```

## Features

- initialize a workspace from multiple repositories
- add repositories to an existing workspace
- inspect branch and dirty state across all repos in a workspace
- remove workspace worktrees without deleting canonical clones
- navigate to a workspace with shell integration via `grove cd`

## Prerequisites

- Go 1.25+ to build the binary
- Git installed and available on `PATH`
- GitHub authentication via `gh auth login` or `GITHUB_TOKEN` for GitHub repo discovery

## Build

Standard build:

```bash
go build -o grove ./cmd/grove
```

If your environment restricts the default Go cache location:

```bash
GOCACHE=/tmp/grove-gocache CGO_ENABLED=0 go build -o grove ./cmd/grove
```

## Configuration

Config file:

```text
~/.config/grove/config.yaml
```

Example:

```yaml
workspace_root: ~/groves
repo_cache_root: ~/.grove/repos

github:
  enabled: true
  orgs:
    - acme
```

Notes:

- `github.orgs` is optional
- if `orgs` is empty, `grove` queries repositories visible to the authenticated GitHub user
- cached canonical clones under `repo_cache_root` are also used as a discovery source

## GitHub Auth

Auth priority:

1. `gh auth token`
2. `GITHUB_TOKEN`

If neither is available, GitHub discovery is skipped and `grove` falls back to cached repository metadata and cached canonical clones.

## Usage

Initialize a workspace:

```bash
./grove init auth-feature
```

Add repositories to the current workspace:

```bash
cd ~/groves/auth-feature
./grove add
```

Show workspace status:

```bash
./grove status auth-feature
```

Or from inside a workspace:

```bash
./grove status
```

Remove a workspace:

```bash
./grove remove auth-feature
```

Print a workspace path:

```bash
./grove path auth-feature
```

## Shell Integration

`grove cd` requires shell integration because a subprocess cannot change the parent shell's working directory.

Enable it for the current shell:

```bash
eval "$(./grove shell-init zsh)"
```

Then:

```bash
grove cd auth-feature
```

The generated shell integration also adds completion for:

```bash
grove cd <TAB>
```

To enable it permanently in `zsh`, add this to `~/.zshrc`:

```bash
eval "$(/absolute/path/to/grove shell-init zsh)"
```

## How It Works

When you initialize a workspace, `grove`:

1. discovers repositories from cached clones and GitHub
2. lets you choose repositories interactively
3. ensures a canonical clone exists under `repo_cache_root`
4. fetches latest refs
5. creates a worktree under the workspace directory

Workspace branches use:

```text
feature/<workspace>
```

If that branch does not exist, `grove` creates it from the repository default branch.

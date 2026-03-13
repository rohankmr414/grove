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
- remove repositories from an existing workspace
- inspect branch and dirty state across all repos in a workspace
- list known workspaces
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

## Install

Install the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/rohankmr414/grove/main/scripts/install.sh | sh
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/rohankmr414/grove/main/scripts/install.sh | sh -s -- v0.1.0
```

Upgrade an existing installation to the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/rohankmr414/grove/main/scripts/install.sh | sh
```

By default the script installs to `~/.local/bin`. Override that with `INSTALL_DIR`:

```bash
curl -fsSL https://raw.githubusercontent.com/rohankmr414/grove/main/scripts/install.sh | INSTALL_DIR=/usr/local/bin sh
```

Current install script targets:

- macOS: `amd64`, `arm64`
- Linux: `amd64`, `arm64`

Upgrade behavior:

- if the requested version is already installed at the target path, the script exits without replacing the binary
- if a different version is installed at the target path, the script replaces it in place
- if no compatible existing `grove version` output is available, the script proceeds with installation

After install or upgrade, the script prints shell integration commands so `grove cd` and auto-entering a workspace after `grove init` work immediately.
It also creates a default config at `$XDG_CONFIG_HOME/grove/config.yaml` when `XDG_CONFIG_HOME` is set, or `~/.config/grove/config.yaml` otherwise, if one does not already exist. The same default config directory also gets a `workspace-init/` folder for files you want copied into every new workspace.

## Configuration

Config file:

```text
$XDG_CONFIG_HOME/grove/config.yaml
# or, if XDG_CONFIG_HOME is unset:
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
- files placed in `$XDG_CONFIG_HOME/grove/workspace-init/`, or `~/.config/grove/workspace-init/` when `XDG_CONFIG_HOME` is unset, are copied into the root of each newly initialized workspace
- this is useful for files like `CLAUDE.md`, `.tool-versions`, or team-specific bootstrap docs

## GitHub Auth

Auth priority:

1. `gh auth token`
2. `GITHUB_TOKEN`

If neither is available, GitHub discovery is skipped and `grove` falls back to cached repository metadata and cached canonical clones.

GitHub repo metadata is cached at `~/.grove/cache/repos.json`. `grove` uses that cache immediately when it exists and refreshes it in the background during `init` and `repo add`, which keeps repeated commands fast without blocking on GitHub every time.

## Usage

Initialize a workspace:

```bash
./grove init auth-feature
```

Add repositories to the current workspace:

```bash
cd ~/groves/auth-feature
./grove repo add
```

Remove repositories from the current workspace:

```bash
cd ~/groves/auth-feature
./grove repo remove
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

List workspaces:

```bash
./grove list
```

Show command help:

```bash
./grove help init
./grove --help
```

## Shell Integration

`grove cd` and auto-jumping after `grove init` require shell integration because a subprocess cannot change the parent shell's working directory.

Enable it for the current shell:

```bash
eval "$(./grove shell-init zsh)"
```

Then:

```bash
grove cd auth-feature
```

With shell integration enabled, this will also drop you into the workspace after a successful init:

```bash
grove init auth-feature
```

The generated shell integration also loads Cobra-generated shell completion. You can also install completion directly:

```bash
./grove completion zsh
./grove completion bash
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
workspace/<workspace>
```

If that branch does not exist, `grove` creates it from the repository default branch.

Each workspace also includes VS Code metadata:

- `.vscode/settings.json` enables Git repository detection in subfolders and scans two levels deep, so opening the workspace root with `code .` finds each repo more reliably.
- top-level files in `~/.config/grove/workspace-init/` are copied into the workspace root before repositories are added.

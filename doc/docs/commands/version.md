# hexago version

Print HexaGo version information.

## Synopsis

```shell
hexago version [flags]
```

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--simple` | `-s` | bool | `false` | Print only the version number (useful for scripting) |

---

## Output

### Default (splash)

Running `hexago version` prints a banner with full build metadata:

```
┓┏      ┏┓    Version: 0.0.3
┣┫┏┓┓┏┏┓┃┓┏┓  Build: 2026-01-15T10:00:00Z
┛┗┗ ┛┗┗┻┗┛┗┛  Commit: abc1234
```

### Simple (`-s`)

Prints only the bare version string — no newline decoration:

```shell
hexago version --simple
# → v0.1.3
```

Useful in CI scripts or Makefiles:

```shell
VERSION=$(hexago version --simple)
docker build -t myapp:$VERSION .
```

---

## Examples

```shell
hexago version          # Full splash with build info
hexago version --simple # Version string only (e.g. v0.1.3)
hexago version -s       # Same, short flag
```

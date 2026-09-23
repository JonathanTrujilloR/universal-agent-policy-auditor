# Install the planned alpha

## Availability first

`v0.1.0-alpha.1` is a planned prerelease, not proof that any URL below is live. Use these links **only after** the [GitHub Release page](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/releases/tag/v0.1.0-alpha.1) shows the prerelease; GitHub Releases are the canonical source of actual availability.

The intended Linux amd64 assets are [auditor_v0.1.0-alpha.1_linux_amd64](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/releases/download/v0.1.0-alpha.1/auditor_v0.1.0-alpha.1_linux_amd64), [checksums.txt](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/releases/download/v0.1.0-alpha.1/checksums.txt), and [LICENSE](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/releases/download/v0.1.0-alpha.1/LICENSE). Linux amd64 only; no other platform or package manager is offered.

## Download, verify, and keep it isolated

After publication, download into a fresh directory so an existing `auditor` is not overwritten. No `sudo` is needed.

```bash
(
  set -eu
  mkdir -p "$HOME/.local/opt"
  INSTALL_DIR="$(mktemp -d "$HOME/.local/opt/auditor-alpha.XXXXXX")"
  cd "$INSTALL_DIR"
  curl --fail --location --remote-name https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/releases/download/v0.1.0-alpha.1/auditor_v0.1.0-alpha.1_linux_amd64
  curl --fail --location --remote-name https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/releases/download/v0.1.0-alpha.1/checksums.txt
  curl --fail --location --remote-name https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/releases/download/v0.1.0-alpha.1/LICENSE
  sha256sum --check checksums.txt
  install -m 755 auditor_v0.1.0-alpha.1_linux_amd64 "$INSTALL_DIR/auditor"
  "$INSTALL_DIR/auditor" version
  "$INSTALL_DIR/auditor" --help

  # Supported demo: OpenCode 1.18.27 legacy scalar permission only.
  ROOT="$(mktemp -d)"
  trap 'rm -rf -- "$ROOT"' EXIT
  CONFIG="$ROOT/opencode.json"
  printf '%s\n' '{"permission":{"read":"allow","edit":"deny","bash":"ask"}}' > "$CONFIG"
  "$INSTALL_DIR/auditor" audit opencode --root "$ROOT" --config "$CONFIG" --opencode-version 1.18.27 --format json
  "$INSTALL_DIR/auditor" compare opencode --root "$ROOT" --reference-config "$CONFIG" --target-config "$CONFIG" --opencode-version 1.18.27 --format text --no-color
  printf 'Installed executable: %s\n' "$INSTALL_DIR/auditor"
)
```

The one fail-closed subshell stops at the first failed setup, download, checksum, install, or demo command without changing the interactive shell. `INSTALL_DIR`, `ROOT`, and `CONFIG` are local to it; copy the printed absolute executable path for later manual use. The trap removes only the temporary demo root. Retain the included Apache-2.0 `LICENSE`. Checksums check file identity; they are not signatures, publisher-identity proof, or provenance records.

## Supported smoke example

The combined command above runs the only supported smoke: OpenCode `1.18.27` legacy scalar `permission` with explicit `read`, `edit`, and `bash`, using absolute paths and the documented flag order.

The auditor models static configuration semantics. It does not enforce runtime policy and is not a security or compliance guarantee. Claude Code is unsupported. See the checked-in [support evidence](../support/evidence.md), [safe-report limits](report-safety.md), and release smoke contract in [`scripts/test_release.py`](../scripts/test_release.py).

#!/usr/bin/env bash
# Local packaging only. No tag creation, GitHub calls or publication.
set -euo pipefail
export LC_ALL=C

if [[ $# != 2 || $1 != v0.1.0-alpha.1 ]]; then
  echo 'expected v0.1.0-alpha.1 and an output directory' >&2
  exit 1
fi
version=$1
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
mkdir -p -- "$2"
out=$(cd -- "$2" && pwd)
name="auditor_${version}_linux_amd64"
cd -- "$root"

# No timestamps, absolute build paths, VCS dirtiness or host CPU tuning.
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 GOFLAGS='' \
  go build -trimpath -buildvcs=false -ldflags="-s -w -buildid= -X main.version=$version" \
  -o "$out/$name" ./cmd/auditor
cp LICENSE "$out/LICENSE"
(cd -- "$out" && sha256sum "$name" > checksums.txt && sha256sum --check checksums.txt)

binary="$out/$name"
test "$("$binary" version)" = "auditor $version"
help=$("$binary" --help)
[[ $help == "auditor $version"$'\nUsage:'* ]]
fixture="$root/support/testdata/opencode-1.18.27-permission-legacy-scalar.json"
report=$("$binary" audit opencode --root "$root" --config "$fixture" \
  --opencode-version 1.18.27 --format json)
# Parse the actual packaged CLI output, rather than accepting a zero exit alone.
python3 -c 'import json, sys
report = json.load(sys.stdin)
assert report["tool"]["version"] == sys.argv[1]
assert report["result"] == "complete_no_findings"
assert report["schema"] == "auditor-report/v1alpha1"
' "$version" <<< "$report"
printf 'Smoke OK: version=%s help=OK audit=complete_no_findings checksum=OK\n' "$version"

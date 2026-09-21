"""Offline workflow contract and real Linux amd64 packaging checks (PyYAML)."""
import hashlib
import subprocess
import tempfile
import unittest
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
TAG = "v0.1.0-alpha.1"
NAME = f"auditor_{TAG}_linux_amd64"


class ReleaseContract(unittest.TestCase):
    def test_workflow(self):
        workflow = yaml.safe_load((ROOT / ".github/workflows/ci.yml").read_text())
        # PyYAML's YAML 1.1 loader interprets the GitHub key `on` as True.
        events = workflow.get("on", workflow.get(True))
        self.assertEqual(events, {"pull_request": None, "push": {
            "branches": ["main"], "tags": [TAG]}})
        self.assertEqual(workflow["permissions"], {"contents": "read"})
        jobs = workflow["jobs"]
        self.assertEqual(set(jobs), {"check", "artifacts"})
        check = jobs["check"]
        commands = "\n".join(s.get("run", "") for s in check["steps"])
        for command in ['test -z "$(gofmt -l .)"', "go test ./...",
                        "go vet ./...", "go build ./..."]:
            self.assertIn(command, commands)
        artifacts = jobs["artifacts"]
        self.assertEqual(artifacts["needs"], "check")
        self.assertEqual(artifacts["if"],
                         "github.event_name == 'push' && github.ref == 'refs/tags/" + TAG + "'")
        for job in jobs.values():
            self.assertEqual(job["runs-on"], "ubuntu-24.04")
            self.assertNotIn("permissions", job)
            steps = job["steps"]
            self.assertEqual(steps[0]["uses"], "actions/checkout@v4")
            self.assertEqual(steps[0]["with"], {"persist-credentials": False})
            self.assertEqual(steps[1]["uses"], "actions/setup-go@v5")
            self.assertEqual(steps[1]["with"], {"go-version-file": "go.mod", "cache": False})
        steps = artifacts["steps"]
        self.assertEqual(steps[2]["run"], 'bash scripts/release.sh "$RELEASE_TAG" "$RUNNER_TEMP/alpha"')
        self.assertEqual(steps[2]["env"], {"RELEASE_TAG": "${{ github.ref_name }}"})
        self.assertEqual(steps[3]["uses"], "actions/upload-artifact@v4")
        self.assertEqual(steps[3]["with"], {
            "name": f"auditor_{TAG}_linux_amd64",
            "path": "${{ runner.temp }}/alpha/*", "if-no-files-found": "error"})
        self.assertEqual(len(steps), 4)

    def test_package(self):
        with tempfile.TemporaryDirectory() as directory:
            out = Path(directory) / "artifacts"
            command = ["bash", str(ROOT / "scripts/release.sh"), TAG, str(out)]
            subprocess.run(command, cwd=ROOT, check=True)
            binary = out / NAME
            first = binary.read_bytes()
            self.assertEqual(first[:4], b"\x7fELF")
            self.assertEqual(first[4:6], b"\x02\x01")  # 64-bit little-endian
            self.assertEqual(first[18:20], b"\x3e\x00")  # AMD64 machine
            self.assertEqual(subprocess.check_output([binary, "version"]),
                             f"auditor {TAG}\n".encode())
            self.assertEqual((out / "checksums.txt").read_text(),
                             f"{hashlib.sha256(first).hexdigest()}  {NAME}\n")
            self.assertEqual({p.name for p in out.iterdir()}, {NAME, "checksums.txt", "LICENSE"})
            self.assertEqual((out / "LICENSE").read_bytes(), (ROOT / "LICENSE").read_bytes())
            subprocess.run(command, cwd=ROOT, check=True)
            self.assertEqual(first, binary.read_bytes(), "repeat build differs")
            binary.write_bytes(first + b"tampered")
            result = subprocess.run(["sha256sum", "--check", "checksums.txt"],
                                    cwd=out, capture_output=True)
            self.assertNotEqual(result.returncode, 0)
            for version in ["dev", "v0.1.0", "v0.1.0-alpha.2", ""]:
                result = subprocess.run(command[:2] + [version, str(out)],
                                        cwd=ROOT, capture_output=True)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(b"expected v0.1.0-alpha.1", result.stderr)


if __name__ == "__main__":
    unittest.main()

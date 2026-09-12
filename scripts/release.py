#!/usr/bin/env python3
"""Build versioned release archives without external packaging dependencies."""
import hashlib
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
import zipfile

tag = sys.argv[1] if len(sys.argv) == 2 else ""
if not re.fullmatch(r"v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)", tag):
    sys.exit("Usage: make release VERSION=v0.1.0 (stable vMAJOR.MINOR.PATCH tag)")
root = Path(__file__).resolve().parent.parent
destination = root / "dist" / "release" / tag
destination.mkdir(parents=True, exist_ok=True)
archives = []
targets = [("linux", "amd64"), ("linux", "arm64"), ("darwin", "amd64"), ("darwin", "arm64"), ("windows", "amd64")]
for system, arch in targets:
    name = f"hitmaker_{tag[1:]}_{system}_{arch}"
    with tempfile.TemporaryDirectory(prefix="hitmaker-release-") as temporary:
        folder = Path(temporary)
        binary = folder / ("hitmaker.exe" if system == "windows" else "hitmaker")
        env = dict(os.environ, CGO_ENABLED="0", GOOS=system, GOARCH=arch)
        subprocess.run([os.environ.get("GO", "go"), "build", "-trimpath", "-ldflags",
                        f"-s -w -X main.version={tag[1:]}", "-o", str(binary), "."],
                       cwd=root, env=env, check=True)
        shutil.copy2(root / "README.md", folder / "README.md")
        shutil.copytree(root / "examples" / "config", folder / "examples" / "config")
        if system == "windows":
            archive = destination / f"{name}.zip"
            with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as output:
                for item in sorted(folder.rglob("*")):
                    if item.is_file():
                        output.write(item, item.relative_to(folder))
        else:
            archive = destination / f"{name}.tar.gz"
            with tarfile.open(archive, "w:gz") as output:
                for item in sorted(folder.iterdir()):
                    output.add(item, arcname=item.name)
        archives.append(archive)
        print(archive.relative_to(root), flush=True)
checksums = "".join(f"{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n" for path in archives)
(destination / "SHA256SUMS").write_text(checksums)

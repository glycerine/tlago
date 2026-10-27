#!/usr/bin/env python3
"""Build a standalone alpha archive with examples and third-party notices."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--goos", help="Target Go OS (defaults to native)")
    parser.add_argument("--goarch", help="Target Go architecture (defaults to native)")
    parser.add_argument("--output", default="dist")
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    os.chdir(root)
    go = os.environ.get("GO", "go")
    go_command = [go]
    if sys.platform.startswith("linux") and shutil.which("taskset"):
        go_command = ["taskset", "-c", "0-7", go]
    env = dict(os.environ, GOMAXPROCS="8", CGO_ENABLED="0")
    if args.goos:
        env["GOOS"] = args.goos
    if args.goarch:
        env["GOARCH"] = args.goarch
    target = json.loads(subprocess.check_output(go_command + ["env", "-json", "GOOS", "GOARCH"], env=env))
    name = f"tlago-{target['GOOS']}-{target['GOARCH']}"
    output = Path(args.output).resolve()
    output.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix=".package-", dir=output) as temporary:
        stage = Path(temporary) / name
        stage.mkdir()
        binary = stage / ("tlago.exe" if target["GOOS"] == "windows" else "tlago")
        subprocess.run(go_command + ["build", "-p=8", "-trimpath", "-o", str(binary), "./cmd/tlago"],
                       env=env, check=True)
        for document in ("README.md", "ALPHA.md", "LICENSE", "THIRD_PARTY_NOTICES.md"):
            shutil.copy2(document, stage / document)
        shutil.copytree("examples/counter", stage / "examples" / "counter")
        notices = stage / "licenses"
        shutil.copytree("tlc/licenses", notices)
        shutil.copy2("tlc/JZLIB_LICENSE.txt", notices / "JZLIB_LICENSE.txt")
        shutil.copy2("test_vectors/CommunityModules/LICENSE", notices / "CommunityModules-LICENSE")
        # Preserve module-source notices alongside the binary's embedded modules.
        for source, label in (("stdlib/tla2tools", "StandardModules"),
                              ("stdlib/community", "CommunityModules"),
                              ("stdlib/tlaps", "TLAPS"), ("stdlib/apalache", "Apalache")):
            dest = notices / "embedded-modules" / label
            dest.mkdir(parents=True)
            pattern = "Apalache.tla" if label == "Apalache" else "*.tla"
            for module in Path(source).glob(pattern):
                shutil.copy2(module, dest / module.name)
        for module, label in (("github.com/glycerine/rbtree", "rbtree"), ("golang.org/x/text", "x-text")):
            directory = Path(subprocess.check_output(go_command + ["list", "-m", "-f", "{{.Dir}}", module],
                                                    env=env, text=True).strip())
            dest = notices / label
            dest.mkdir()
            shutil.copy2(directory / "LICENSE", dest / "LICENSE")
            if (directory / "PATENTS").exists():
                shutil.copy2(directory / "PATENTS", dest / "PATENTS")
        archive_format = "zip" if target["GOOS"] == "windows" else "gztar"
        archive = Path(shutil.make_archive(str(Path(temporary) / name), archive_format,
                                          root_dir=temporary, base_dir=name))
        final = output / archive.name
        os.replace(archive, final)
        digest = hashlib.sha256()
        with final.open("rb") as archive_file:
            for block in iter(lambda: archive_file.read(1024 * 1024), b""):
                digest.update(block)
        checksum = final.with_name(final.name + ".sha256")
        checksum.write_text(f"{digest.hexdigest()}  {final.name}\n", encoding="ascii")
    print(final)
    print(checksum)


if __name__ == "__main__":
    main()

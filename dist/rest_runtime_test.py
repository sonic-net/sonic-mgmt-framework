"""Exercise the packaged REST executables and generated UI with private fixtures."""

import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import posixpath
import socket
import ssl
import subprocess
import sys
import tarfile
import tempfile
import time
import unittest
import urllib.error
import urllib.request

RUNTIME, UI, COMMON, LIBRARIES, DATABASES = [Path(p).absolute() for p in sys.argv[1:6]]
del sys.argv[1:6]


def unpack(archive, root):
    def relocatable(member, destination):
        if member.issym() and member.linkname.startswith("/"):
            member = member.replace(linkname=posixpath.relpath(
                member.linkname.lstrip("/"), str(PurePosixPath(member.name).parent)))
        return tarfile.data_filter(member, destination)
    with tarfile.open(archive) as source:
        source.extractall(root, filter=relocatable)


def stop(process):
    if process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()


class InstalledRESTTest(unittest.TestCase):
    def test_packaged_rest_service(self):
        with tempfile.TemporaryDirectory(prefix="framework-rest-", dir="/tmp") as temporary:
            root = Path(temporary)
            installed = root / "installed"
            installed.mkdir()
            for package in (LIBRARIES, COMMON, RUNTIME, UI):
                unpack(package, installed)
            loader = installed / "usr/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2"
            if not loader.exists():
                loader = installed / "lib/x86_64-linux-gnu/ld-linux-x86-64.so.2"
            library_path = ":".join(sorted({str(p.parent) for p in installed.rglob("lib*.so*") if p.is_file()}))
            prefix = [str(loader), "--library-path", library_path]
            environment = dict(os.environ)
            environment.update({
                "DB_CONFIG_PATH": str(root / "database_config.json"),
                "CVL_SCHEMA_PATH": str(installed / "usr/sbin/schema"),
                "CVL_CFG_FILE": str(installed / "usr/sbin/cvl_cfg.json"),
                "YANG_MODELS_PATH": str(installed / "usr/models/yang"),
            })
            database = json.loads(DATABASES.read_text())
            redis_socket = str(root / "redis.sock")
            for instance in database["INSTANCES"].values():
                instance.update(hostname="127.0.0.1", port=0, unix_socket_path=redis_socket)
            for name, number in (("LOGLEVEL_DB", 3), ("ERROR_DB", 8), ("EVENT_DB", 9)):
                database["DATABASES"].setdefault(name, {"id": number, "separator": "|", "instance": "redis"})
            (root / "database_config.json").write_text(json.dumps(database))
            with (root / "redis.log").open("w+") as redis_log:
                redis = subprocess.Popen(prefix + [str(installed / "usr/bin/redis-server"),
                    "--port", "0", "--unixsocket", redis_socket, "--save", "", "--appendonly", "no",
                    "--databases", "32", "--dir", str(root)], stdout=redis_log, stderr=subprocess.STDOUT)
                self.addCleanup(stop, redis)
                deadline = time.monotonic() + 15
                while time.monotonic() < deadline and redis.poll() is None:
                    try:
                        with socket.socket(socket.AF_UNIX) as connection:
                            connection.settimeout(1)
                            connection.connect(redis_socket)
                            connection.sendall(b"*1\r\n$4\r\nPING\r\n")
                            self.assertEqual(connection.recv(64).strip(), b"+PONG")
                            break
                    except OSError:
                        time.sleep(0.05)
                else:
                    redis_log.seek(0)
                    self.fail("Private Redis failed: " + redis_log.read())
                try:
                    subprocess.run(prefix + [str(installed / "usr/sbin/generate_cert"),
                        "--host", "localhost,127.0.0.1", "--ecdsa-curve", "P256", "--duration", "1h"],
                        cwd=root, env=environment, check=True, timeout=20)
                    self.assertTrue((root / "cert.pem").is_file())
                    self.assertTrue((root / "key.pem").is_file())
                    context = ssl.create_default_context(cafile=str(root / "cert.pem"))
                    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}), urllib.request.HTTPSHandler(context=context))
                    for attempt in range(5):
                        with socket.socket() as reservation:
                            reservation.bind(("127.0.0.1", 0))
                            port = reservation.getsockname()[1]
                        with (root / "rest.log").open("w+") as log:
                            server = subprocess.Popen(prefix + [str(installed / "usr/sbin/rest_server"),
                                "-port", str(port), "-cert", str(root / "cert.pem"), "-key", str(root / "key.pem"),
                                "-rest_ui", str(installed / "etc/rest_server/ui"), "-client_auth", "none", "-logtostderr"],
                                cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT)
                            try:
                                base = f"https://127.0.0.1:{port}"
                                deadline = time.monotonic() + 30
                                while server.poll() is None and time.monotonic() < deadline:
                                    try:
                                        with opener.open(base + "/restconf/yang-library-version", timeout=1) as response:
                                            self.assertEqual(json.load(response)["ietf-restconf:yang-library-version"], "2016-06-21")
                                        break
                                    except (OSError, urllib.error.URLError):
                                        time.sleep(0.1)
                                else:
                                    log.seek(0)
                                    details = log.read()
                                    if "address already in use" in details and attempt < 4:
                                        continue
                                    self.fail("Packaged REST failed to start: " + details)
                                with opener.open(base + "/.well-known/host-meta", timeout=5) as response:
                                    self.assertIn(b"href='/restconf'", response.read())
                                with opener.open(base + "/restconf/operations", timeout=5) as response:
                                    operations = json.load(response)
                                    self.assertIn("operations", operations)
                                    self.assertIn("sonic-show-techsupport:sonic-show-techsupport-info", operations["operations"])
                                with opener.open(base + "/ui/index.html", timeout=5) as response:
                                    self.assertEqual(response.read(), (installed / "etc/rest_server/ui/index.html").read_bytes())
                                print(json.dumps({"runtime_sha256": hashlib.sha256(RUNTIME.read_bytes()).hexdigest(),
                                    "ui_sha256": hashlib.sha256(UI.read_bytes()).hexdigest(),
                                    "checks": ["installed certificate tool", "TLS trust", "RESTCONF version", "host-meta", "generated RPC registration", "installed UI bytes"],
                                    "auth": "none; host SSH/TACACS fixtures remain separately opt-in"}, sort_keys=True))
                                break
                            finally:
                                stop(server)
                finally:
                    stop(redis)


if __name__ == "__main__":
    unittest.main()

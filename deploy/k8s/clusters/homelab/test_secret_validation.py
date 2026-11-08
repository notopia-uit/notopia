import copy
import json
import pathlib
import subprocess
import tempfile
import unittest

import yaml

from test_infra import RECIPIENT, check_release_secret, find_secrets, validate_secret


class SecretValidationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        plaintext = "apiVersion: v1\nkind: Secret\nmetadata:\n  name: valkey-note\nstringData:\n  password: synthetic-only\n"
        encrypted = subprocess.check_output(
            ["sops", "--encrypt", "--age", RECIPIENT, "--encrypted-regex", "^(data|stringData)$", "--input-type", "yaml", "--output-type", "yaml", "--filename-override", "deploy/k8s/clusters/homelab/fixture.sops.yaml", "/dev/stdin"],
            input=plaintext,
            text=True,
        )
        cls.secret = yaml.safe_load(encrypted)

    def test_sops_generated_fixture(self):
        validate_secret(self.secret, "fixture")

    def test_plaintext_and_metadata_are_rejected(self):
        cases = []
        secret = copy.deepcopy(self.secret)
        secret["stringData"]["password"] = "synthetic-only"
        cases.append(secret)
        secret = copy.deepcopy(self.secret)
        secret["stringData"]["password"] = "ENC[not-encrypted]"
        cases.append(secret)
        for field in ("mac", "version", "lastmodified", "age", "encrypted_regex"):
            secret = copy.deepcopy(self.secret)
            secret["sops"].pop(field)
            cases.append(secret)
        secret = copy.deepcopy(self.secret)
        secret["sops"]["age"][0]["recipient"] = "age1wrong"
        cases.append(secret)
        secret = copy.deepcopy(self.secret)
        secret["sops"]["age"][0]["enc"] = "not an age envelope"
        cases.append(secret)
        for secret in cases:
            with self.subTest(secret=secret), self.assertRaises(ValueError):
                validate_secret(secret, "fixture")

    def test_scan_multidocument_and_both_extensions_outside_infra(self):
        base = pathlib.Path("tmp/homelab-wiring")
        base.mkdir(parents=True, exist_ok=True)
        with tempfile.TemporaryDirectory(dir=base) as directory:
            root = pathlib.Path(directory)
            for name in ("other.yml", "outside/unexpected.yaml"):
                path = root / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("kind: ConfigMap\n---\n" + yaml.safe_dump(self.secret))
            (root / "list.json").write_text(json.dumps({"apiVersion": "v1", "kind": "List", "items": [self.secret]}))
            self.assertEqual(len(list(find_secrets(root.rglob("*")))), 3)
            (root / "plaintext.yml").write_text("kind: Secret\nstringData:\n  password: synthetic-only\n")
            with self.assertRaises(ValueError):
                for path, secret in find_secrets(root.rglob("*")):
                    validate_secret(secret, path)

    def test_suspension_depends_on_rendered_secret_and_key(self):
        release = {"metadata": {"name": "valkey-note"}, "spec": {"suspend": True}}
        check_release_secret(release, [], "valkey-note", "password")
        release["spec"]["suspend"] = False
        with self.assertRaises(ValueError):
            check_release_secret(release, [], "valkey-note", "password")
        missing_key = copy.deepcopy(self.secret)
        missing_key["stringData"] = {"other": missing_key["stringData"]["password"]}
        with self.assertRaises(ValueError):
            check_release_secret(release, [missing_key], "valkey-note", "password")
        check_release_secret(release, [self.secret], "valkey-note", "password")


if __name__ == "__main__":
    unittest.main()

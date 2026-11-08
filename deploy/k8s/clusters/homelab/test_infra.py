import base64
import datetime
import pathlib
import re
import subprocess
import unittest

import yaml


ROOT = pathlib.Path(__file__).parent
INFRA = ROOT / "infra"
REPO = ROOT.parents[3]
RECIPIENT = "age1ytym5g6mxf0974syp2g2ecrh0y74tnl0u9ee4xm8caawhyv25pls3f2sy4"
ENC = re.compile(r"ENC\[AES256_GCM,data:([A-Za-z0-9+/=]*),iv:([A-Za-z0-9+/=]+),tag:([A-Za-z0-9+/=]+),type:str\]")


def encrypted_value(value):
    match = ENC.fullmatch(value) if isinstance(value, str) else None
    if not match:
        return False
    try:
        data, iv, tag = (base64.b64decode(part, validate=True) for part in match.groups())
    except ValueError:
        return False
    return bool(data) and len(iv) == 32 and len(tag) == 16


def validate_secret(secret, path):
    def require(condition, message):
        if not condition:
            raise ValueError(f"{path}: {message}")

    metadata = secret.get("sops")
    require(isinstance(metadata, dict), "missing SOPS metadata")
    require(metadata.get("encrypted_regex") == "^(data|stringData)$", "incorrect encrypted regex")
    require(bool(re.fullmatch(r"\d+\.\d+\.\d+", str(metadata.get("version", "")))), "invalid SOPS version")
    try:
        timestamp = datetime.datetime.fromisoformat(metadata.get("lastmodified", "").replace("Z", "+00:00"))
        require(timestamp.tzinfo is not None, "SOPS timestamp has no timezone")
    except (ValueError, TypeError, AttributeError):
        raise ValueError(f"{path}: invalid SOPS timestamp") from None
    require(encrypted_value(metadata.get("mac")), "invalid SOPS MAC")
    age = metadata.get("age")
    require(isinstance(age, list), "missing age recipients")
    recipients = [entry for entry in age if isinstance(entry, dict) and entry.get("recipient") == RECIPIENT]
    require(bool(recipients), "missing homelab age recipient")
    envelope = recipients[0].get("enc", "")
    require(isinstance(envelope, str), "invalid age envelope")
    match = re.fullmatch(r"-----BEGIN AGE ENCRYPTED FILE-----\s+([A-Za-z0-9+/=\s]+)-----END AGE ENCRYPTED FILE-----\s*", envelope)
    require(match is not None, "invalid age envelope")
    try:
        decoded = base64.b64decode("".join(match.group(1).split()), validate=True)
    except ValueError:
        raise ValueError(f"{path}: invalid age envelope encoding") from None
    require(decoded.startswith(b"age-encryption.org/v1\n"), "invalid age envelope payload")
    fields = [secret.get(field) for field in ("data", "stringData") if field in secret]
    require(bool(fields) and all(isinstance(field, dict) and field for field in fields), "missing secret values")
    require(all(encrypted_value(value) for field in fields for value in field.values()), "plaintext or malformed secret value")


def find_secrets(paths):
    def secrets_in(document):
        if not isinstance(document, dict):
            return
        if document.get("kind") == "Secret" and document.get("apiVersion", "v1") == "v1":
            yield document
        if document.get("kind") in ("List", "SecretList"):
            for item in document.get("items", []):
                yield from secrets_in(item)

    for path in paths:
        path = pathlib.Path(path)
        if not path.is_file() or path.suffix not in (".yaml", ".yml", ".json"):
            continue
        text = path.read_text()
        try:
            documents = list(yaml.safe_load_all(text))
        except yaml.YAMLError as error:
            if re.search(r"\bkind\s*:\s*['\"]?Secret\b", text):
                raise ValueError(f"{path}: cannot parse Secret document: {error}") from error
            continue
        for document in documents:
            for secret in secrets_in(document):
                yield path, secret


def check_release_secret(release, secrets, name, key):
    if release["spec"].get("suspend", False):
        return
    for secret in secrets:
        if secret.get("metadata", {}).get("name") == name and key in (secret.get("data", {}) | secret.get("stringData", {})):
            return
    raise ValueError(f"{release['metadata']['name']}: missing rendered Secret {name} key {key}; release must be suspended")


class InfrastructureTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        output = subprocess.check_output(["kubectl", "kustomize", str(INFRA)], text=True)
        cls.resources = list(yaml.safe_load_all(output))

    def test_clusters(self):
        clusters = {r["metadata"]["name"]: r["spec"] for r in self.resources if r["kind"] == "Cluster"}
        self.assertEqual(set(clusters), {f"cnpg-{s}" for s in ("note", "billing", "authorization", "document")})
        for name, spec in clusters.items():
            service = name.removeprefix("cnpg-")
            self.assertEqual(spec["instances"], 1)
            self.assertEqual(spec["storage"]["size"], "1Gi")
            self.assertEqual(spec["imageCatalogRef"]["major"], 18)
            self.assertEqual(spec["imageCatalogRef"]["name"], "postgresql-minimal-trixie")
            self.assertEqual(spec["bootstrap"]["initdb"], {"database": service, "owner": service})
            self.assertEqual(set(spec["managed"]["services"]["disabledDefaultServices"]), {"ro", "r"})

    def test_releases_and_values(self):
        releases = {r["metadata"]["name"]: r for r in self.resources if r["kind"] == "HelmRelease"}
        configs = {r["metadata"]["name"]: r for r in self.resources if r["kind"] == "ConfigMap"}
        secrets = [r for r in self.resources if r["kind"] == "Secret" and r.get("metadata", {}).get("namespace") == "notopia"]
        for service in ("note", "billing", "document"):
            release = releases[f"valkey-{service}"]
            check_release_secret(release, secrets, f"valkey-{service}", "password")
            values = yaml.safe_load(configs[release["spec"]["valuesFrom"][0]["name"]]["data"]["values.yaml"])
            self.assertTrue(values["dataStorage"]["enabled"])
            self.assertFalse(values["replica"]["enabled"])
            self.assertIn("maxmemory-policy noeviction", values["valkeyConfig"])
            self.assertIn("appendonly yes", values["valkeyConfig"])
            self.assertEqual(values["auth"]["usersExistingSecret"], f"valkey-{service}")
        release = releases["meilisearch"]
        check_release_secret(release, secrets, "meilisearch", "MEILI_MASTER_KEY")
        values = yaml.safe_load(configs[release["spec"]["valuesFrom"][0]["name"]]["data"]["values.yaml"])
        self.assertEqual(values["environment"]["MEILI_ENV"], "production")
        self.assertEqual(values["auth"]["existingMasterKeySecret"], "meilisearch")
        self.assertTrue(values["persistence"]["enabled"])

    def test_route(self):
        route = next(r for r in self.resources if r["kind"] == "HTTPRoute")
        self.assertEqual(route["spec"]["hostnames"], ["notopia-meilisearch.${PUBLIC_DOMAIN}"])
        rule = route["spec"]["rules"][0]
        self.assertEqual(rule["backendRefs"][0]["name"], "meilisearch")
        cors = next(f["cors"] for f in rule["filters"] if f["type"] == "CORS")
        self.assertEqual(cors["allowOrigins"], ["https://notopia.${PUBLIC_DOMAIN}", "https://notopia.${LOCAL_DOMAIN}"])

    def test_secrets_are_encrypted(self):
        output = subprocess.check_output(
            ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--", "*.yaml", "*.yml", "*.json"],
            cwd=REPO,
        )
        paths = (REPO / path.decode() for path in output.split(b"\0") if path)
        for path, secret in find_secrets(paths):
            validate_secret(secret, path)


if __name__ == "__main__":
    unittest.main()

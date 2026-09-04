import shutil
import tarfile
import tempfile
import unittest
import warnings
import zipfile
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
from aeib.generate import build, validate


class AEIBTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls._tmp = Path(tempfile.mkdtemp(prefix="aeib-tests-"))
        cls.corpus = cls._tmp / "corpus"
        cls.manifest = build(cls.corpus, 20260821)

    @classmethod
    def tearDownClass(cls):
        shutil.rmtree(cls._tmp, ignore_errors=True)

    def test_generated_corpus_validates(self):
        self.assertEqual(validate(self.corpus), [])

    def test_generation_is_deterministic_by_fixture_hash(self):
        with tempfile.TemporaryDirectory() as a, tempfile.TemporaryDirectory() as b:
            ma, mb = build(Path(a), 20260821), build(Path(b), 20260821)
            ha = [(x["fixture_id"], x["sha256"], x["bytes"]) for x in ma["fixtures"]]
            hb = [(x["fixture_id"], x["sha256"], x["bytes"]) for x in mb["fixtures"]]
            self.assertEqual(ha, hb)

    def test_required_families_present(self):
        families = {x["family"] for x in self.manifest["fixtures"]}
        self.assertTrue({"text", "path", "office", "pdf", "signature", "archive", "email", "binary", "mutation", "resource"} <= families)

    def test_duplicate_zip_names_are_real(self):
        with warnings.catch_warnings():
            warnings.simplefilter("ignore")
            with zipfile.ZipFile(self.corpus / "archives" / "duplicate_names.zip") as z:
                names = [x.filename for x in z.infolist()]
        self.assertGreaterEqual(names.count("dup.txt"), 2)

    def test_tar_contains_traversal_and_symlink_metadata(self):
        with tarfile.open(self.corpus / "archives" / "traversal_symlink.tar") as tf:
            members = tf.getmembers()
        self.assertTrue(any(x.name.startswith("../") for x in members))
        self.assertTrue(any(x.issym() for x in members))

    def test_invalid_utf8_fixture_contains_surrogate_encoding_bytes(self):
        self.assertIn(b"\xed\xa0\x80", (self.corpus / "email" / "invalid_utf8_surrogate.eml").read_bytes())

    def test_safety_cap(self):
        self.assertTrue(all(x["bytes"] <= self.manifest["max_fixture_bytes"] for x in self.manifest["fixtures"]))


if __name__ == "__main__":
    unittest.main()

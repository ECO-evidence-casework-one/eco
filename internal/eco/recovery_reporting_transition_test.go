package eco

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Fixture construction may use test conveniences; each acceptance action below
// deliberately uses the production existing-only OpenVault introduced by #133.
func TestRecoveryReportingLegacyStrictOpen(t *testing.T) {
	root, key, document := workspaceFormatFixture(t)
	delete(document, "revision")
	delete(document, "last_owner_txn")
	delete(document, "preservations")
	document["build_id"] = json.RawMessage(`"ECO-SYNTHETIC-OLDER-CANDIDATE"`)
	workspaceFormatWrite(t, root, key, workspaceFormatJSON(t, document))
	before := workspaceFormatTree(t, root)
	v, err := OpenVault(root)
	if err != nil { t.Fatal(err) }
	defer v.Close()
	if !reflect.DeepEqual(before, workspaceFormatTree(t, root)) { t.Fatal("strict open rewrote legacy state") }
	want := v.Snapshot()
	if want.Revision != 0 || len(want.Matters) != 1 || !want.Settings.LowSensory { t.Fatal("legacy data or zero defaults were lost") }
	if err := v.Save(); err != nil { t.Fatal(err) }
	if err := v.Close(); err != nil { t.Fatal(err) }
	reopened, err := OpenVault(root)
	if err != nil { t.Fatal(err) }
	defer reopened.Close()
	got := reopened.Snapshot()
	if got.Revision != 1 || got.LastOwnerTxn == "" || len(got.Matters) != 1 || got.Matters[0].Title != want.Matters[0].Title || !got.Settings.LowSensory || !got.CreatedAt.Equal(want.CreatedAt) { t.Fatal("strict save/reopen lost supported legacy records") }
}

func TestRecoveryReportingStrictUnknownFormat(t *testing.T) {
	root, key, document := workspaceFormatFixture(t)
	document["unrecognised_future_field"] = json.RawMessage(`true`)
	workspaceFormatWrite(t, root, key, workspaceFormatJSON(t, document))
	before := workspaceFormatTree(t, root)
	v, err := OpenVault(root)
	if v != nil { _ = v.Close() }
	if err == nil { t.Fatal("strict open accepted a format it cannot preserve") }
	if !reflect.DeepEqual(before, workspaceFormatTree(t, root)) { t.Fatal("format refusal changed existing state") }
}

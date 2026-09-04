package eco

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	tesseractRuntimeBundleChangeType = "tesseract-runtime-bundle-registered"
	wave4BuildManifestSHA256         = "19a04d1f229a56a49c720c4b0d1104768b58ca6e2d3dc7f50d95d0fd71f1b70e"
	wave4RuntimeInventorySHA256      = "1b5e1192e3ff209923a7643f16f6b2c7c464c293ff2389237e8cadc428512b0d"
	wave4OCRSmokeSHA256              = "4e046ad1d5b17e562ed5f24b18d1e74d40b6ad99271ace72124158d0da5b6535"
	wave4RuntimeInventoryFiles       = 151
	wave4TesseractSourceCommit       = "fb87a84b6e3384b424b2169c76de9031046861e9"
	wave4TessdataSourceCommit        = "87416418657359cb625c412a48b6e1d6d41c29bd"
)

var wave4RequiredRuntimePaths = []string{
	"bin/tesseract.exe",
	"bin/tesseract55.dll",
	"bin/leptonica-1.87.0.dll",
	"bin/libpng16.dll",
	"bin/jpeg62.dll",
	"bin/zlib1.dll",
	"share/tessdata/eng.traineddata",
	"share/tessdata/osd.traineddata",
}

type tesseractBundleSource struct {
	Name   string `json:"name"`
	Repo   string `json:"repo"`
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
}

type tesseractBundleManifest struct {
	Schema            int                     `json:"schema"`
	Platform          string                  `json:"platform"`
	Purpose           string                  `json:"purpose"`
	GitHubOnlySources bool                    `json:"github_only_sources"`
	LanguageModels    []string                `json:"language_models"`
	OCRSmokeTest      string                  `json:"ocr_smoke_test"`
	Sources           []tesseractBundleSource `json:"sources"`
}

type tesseractRuntimeHashRow struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type tesseractBundleVerification struct {
	Root                   string
	Executable             string
	TessdataDir            string
	BuildManifestSHA256    string
	RuntimeInventorySHA256 string
	OCRSmokeSHA256         string
}

type TesseractRuntimeRegistration struct {
	Root                   string    `json:"root"`
	Executable             string    `json:"executable"`
	TessdataDir            string    `json:"tessdata_dir"`
	Version                string    `json:"version"`
	BuildManifestSHA256    string    `json:"build_manifest_sha256"`
	RuntimeInventorySHA256 string    `json:"runtime_inventory_sha256"`
	OCRSmokeSHA256         string    `json:"ocr_smoke_sha256"`
	RegisteredAt           time.Time `json:"registered_at"`
	AuditChangeID          string    `json:"audit_change_id,omitempty"`
}

func requireAbsoluteTesseractDirectory(path, label string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("%s path is required", label)
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("%s must use an absolute path", label)
	}
	clean := filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", label, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", label, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", label)
	}
	return resolved, nil
}

func resolveTesseractBundleRoot(path string) (string, error) {
	root, err := requireAbsoluteTesseractDirectory(path, "Tesseract runtime bundle")
	if err != nil {
		return "", err
	}
	if strings.EqualFold(filepath.Base(root), "runtime") {
		parent := filepath.Dir(root)
		if info, statErr := os.Stat(filepath.Join(parent, "control")); statErr == nil && info.IsDir() {
			root = parent
		}
	}
	if _, err := requireAbsoluteTesseractDirectory(filepath.Join(root, "runtime"), "Tesseract runtime directory"); err != nil {
		return "", err
	}
	if _, err := requireAbsoluteTesseractDirectory(filepath.Join(root, "control"), "Tesseract control directory"); err != nil {
		return "", err
	}
	return root, nil
}

func safeTesseractRuntimeRelativePath(path string) (string, error) {
	path = strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	if path == "" || strings.HasPrefix(path, "/") {
		return "", errors.New("runtime inventory contains an empty or absolute path")
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || clean != path {
		return "", fmt.Errorf("runtime inventory contains an unsafe path %q", path)
	}
	return clean, nil
}

func hashExactControlFile(path, expected string, maxBytes int64) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() <= 0 || info.Size() > maxBytes {
		return fmt.Errorf("control file is missing, non-regular or unbounded: %s", filepath.Base(path))
	}
	hash, err := hashFile(path)
	if err != nil {
		return err
	}
	if hash != expected {
		return fmt.Errorf("control file SHA-256 does not match ECO's approved Wave 4 receipt: %s", filepath.Base(path))
	}
	return nil
}

func verifyTesseractRuntimeBundleFiles(path string) (tesseractBundleVerification, error) {
	root, err := resolveTesseractBundleRoot(path)
	if err != nil {
		return tesseractBundleVerification{}, err
	}
	control := filepath.Join(root, "control")
	runtimeRoot := filepath.Join(root, "runtime")
	manifestPath := filepath.Join(control, "BUILD_MANIFEST.json")
	inventoryPath := filepath.Join(control, "RUNTIME_FILE_HASHES.json")
	smokePath := filepath.Join(control, "OCR_SMOKE_RESULT.txt")
	if err := hashExactControlFile(manifestPath, wave4BuildManifestSHA256, 256*1024); err != nil {
		return tesseractBundleVerification{}, err
	}
	if err := hashExactControlFile(inventoryPath, wave4RuntimeInventorySHA256, 1024*1024); err != nil {
		return tesseractBundleVerification{}, err
	}
	if err := hashExactControlFile(smokePath, wave4OCRSmokeSHA256, 4096); err != nil {
		return tesseractBundleVerification{}, err
	}
	if smoke, readErr := os.ReadFile(smokePath); readErr != nil || strings.TrimSpace(string(smoke)) != "ECO OCR TEST 123" {
		return tesseractBundleVerification{}, errors.New("Wave 4 OCR smoke receipt is missing or inconsistent")
	}

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return tesseractBundleVerification{}, err
	}
	var manifest tesseractBundleManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return tesseractBundleVerification{}, fmt.Errorf("decode Tesseract build manifest: %w", err)
	}
	if manifest.Schema != 1 || manifest.Platform != "windows-x86_64" || !manifest.GitHubOnlySources || manifest.OCRSmokeTest != "PASS" {
		return tesseractBundleVerification{}, errors.New("Tesseract build manifest does not identify ECO's approved GitHub-only Windows build")
	}
	langs := map[string]bool{}
	for _, lang := range manifest.LanguageModels {
		langs[strings.ToLower(strings.TrimSpace(lang))] = true
	}
	if !langs["eng"] || !langs["osd"] {
		return tesseractBundleVerification{}, errors.New("Tesseract build manifest lacks the approved eng/osd language models")
	}
	sources := map[string]tesseractBundleSource{}
	for _, source := range manifest.Sources {
		sources[source.Name] = source
	}
	if source := sources["tesseract"]; source.Repo != "tesseract-ocr/tesseract" || source.Commit != wave4TesseractSourceCommit {
		return tesseractBundleVerification{}, errors.New("Tesseract build manifest source commit is not the qualified Wave 4 donor")
	}
	if source := sources["tessdata_fast"]; source.Repo != "tesseract-ocr/tessdata_fast" || source.Commit != wave4TessdataSourceCommit {
		return tesseractBundleVerification{}, errors.New("Tesseract language-data source commit is not the qualified Wave 4 donor")
	}

	inventoryBytes, err := os.ReadFile(inventoryPath)
	if err != nil {
		return tesseractBundleVerification{}, err
	}
	var rows []tesseractRuntimeHashRow
	if err := json.Unmarshal(inventoryBytes, &rows); err != nil {
		return tesseractBundleVerification{}, fmt.Errorf("decode Tesseract runtime inventory: %w", err)
	}
	if len(rows) != wave4RuntimeInventoryFiles {
		return tesseractBundleVerification{}, fmt.Errorf("Tesseract runtime inventory has %d files; expected %d", len(rows), wave4RuntimeInventoryFiles)
	}
	inventory := make(map[string]tesseractRuntimeHashRow, len(rows))
	resolvedRuntime, err := filepath.EvalSymlinks(runtimeRoot)
	if err != nil {
		return tesseractBundleVerification{}, err
	}
	runtimePrefix := resolvedRuntime + string(os.PathSeparator)
	for _, row := range rows {
		rel, err := safeTesseractRuntimeRelativePath(row.Path)
		if err != nil {
			return tesseractBundleVerification{}, err
		}
		if _, exists := inventory[rel]; exists {
			return tesseractBundleVerification{}, fmt.Errorf("duplicate Tesseract runtime inventory path %q", rel)
		}
		if row.Bytes <= 0 || !sha256TextPattern.MatchString(row.SHA256) {
			return tesseractBundleVerification{}, fmt.Errorf("invalid Tesseract runtime inventory record %q", rel)
		}
		full := filepath.Join(runtimeRoot, filepath.FromSlash(rel))
		info, err := os.Lstat(full)
		if err != nil {
			return tesseractBundleVerification{}, fmt.Errorf("Tesseract runtime file is missing: %s", rel)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() != row.Bytes {
			return tesseractBundleVerification{}, fmt.Errorf("Tesseract runtime file is non-regular or its size changed: %s", rel)
		}
		resolved, err := filepath.EvalSymlinks(full)
		if err != nil || !strings.HasPrefix(resolved, runtimePrefix) {
			return tesseractBundleVerification{}, fmt.Errorf("Tesseract runtime file escapes the verified bundle: %s", rel)
		}
		hash, err := hashFile(full)
		if err != nil || hash != strings.ToLower(row.SHA256) {
			return tesseractBundleVerification{}, fmt.Errorf("Tesseract runtime file SHA-256 changed: %s", rel)
		}
		inventory[rel] = row
	}
	for _, required := range wave4RequiredRuntimePaths {
		if _, ok := inventory[required]; !ok {
			return tesseractBundleVerification{}, fmt.Errorf("Tesseract runtime inventory is missing required file %s", required)
		}
	}

	walked := make([]string, 0, len(rows))
	err = filepath.WalkDir(runtimeRoot, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("Tesseract runtime contains symbolic link %s", p)
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("Tesseract runtime contains a non-regular entry %s", p)
		}
		rel, err := filepath.Rel(runtimeRoot, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if _, ok := inventory[rel]; !ok {
			return fmt.Errorf("Tesseract runtime contains an unrecorded file %s", rel)
		}
		walked = append(walked, rel)
		return nil
	})
	if err != nil {
		return tesseractBundleVerification{}, err
	}
	sort.Strings(walked)
	if len(walked) != len(rows) {
		return tesseractBundleVerification{}, errors.New("Tesseract runtime file count differs from the controlled inventory")
	}

	executable := filepath.Join(runtimeRoot, "bin", "tesseract.exe")
	tessdata := filepath.Join(runtimeRoot, "share", "tessdata")
	if _, err := requireAbsoluteRegularFile(executable, "Tesseract executable"); err != nil {
		return tesseractBundleVerification{}, err
	}
	if _, err := requireAbsoluteTesseractDirectory(tessdata, "Tesseract tessdata directory"); err != nil {
		return tesseractBundleVerification{}, err
	}
	return tesseractBundleVerification{
		Root:                   root,
		Executable:             executable,
		TessdataDir:            tessdata,
		BuildManifestSHA256:    wave4BuildManifestSHA256,
		RuntimeInventorySHA256: wave4RuntimeInventorySHA256,
		OCRSmokeSHA256:         wave4OCRSmokeSHA256,
	}, nil
}

func tesseractBundleLanguages(ctx context.Context, executable, tessdataDir string) (map[string]bool, error) {
	cmd := exec.CommandContext(ctx, executable, "--list-langs", "--tessdata-dir", tessdataDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Tesseract language-data check failed: %w", err)
	}
	if len(output) > 64*1024 {
		return nil, errors.New("Tesseract language-data output is unbounded")
	}
	langs := map[string]bool{}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if tesseractLanguagePattern.MatchString(line) {
			langs[line] = true
		}
	}
	if !langs["eng"] || !langs["osd"] {
		return nil, errors.New("verified Tesseract runtime cannot load its bundled eng/osd language data")
	}
	return langs, nil
}

func probeTesseractRuntimeBundle(ctx context.Context, verified tesseractBundleVerification) (string, error) {
	version, err := tesseractVersion(ctx, verified.Executable)
	if err != nil {
		return "", err
	}
	if _, err := tesseractBundleLanguages(ctx, verified.Executable, verified.TessdataDir); err != nil {
		return "", err
	}
	return strings.TrimSpace(version), nil
}

func (v *Vault) RegisterTesseractRuntimeBundle(path string) (TesseractRuntimeRegistration, error) {
	return v.RegisterTesseractRuntimeBundleContext(context.Background(), path)
}

func (v *Vault) RegisterTesseractRuntimeBundleContext(ctx context.Context, path string) (TesseractRuntimeRegistration, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return TesseractRuntimeRegistration{}, err
	}
	before, err := verifyTesseractRuntimeBundleFiles(path)
	if err != nil {
		return TesseractRuntimeRegistration{}, err
	}
	version, err := probeTesseractRuntimeBundle(ctx, before)
	if err != nil {
		return TesseractRuntimeRegistration{}, err
	}
	if version == "" || len([]rune(version)) > maxLocalToolIdentityRunes {
		return TesseractRuntimeRegistration{}, errors.New("Tesseract runtime version identity is missing or unbounded")
	}
	after, err := verifyTesseractRuntimeBundleFiles(before.Root)
	if err != nil {
		return TesseractRuntimeRegistration{}, errors.New("Tesseract runtime changed during registration: " + err.Error())
	}
	if before != after {
		return TesseractRuntimeRegistration{}, errors.New("Tesseract runtime identity changed during registration")
	}
	registeredAt := time.Now().UTC()
	registration := TesseractRuntimeRegistration{
		Root:                   before.Root,
		Executable:             before.Executable,
		TessdataDir:            before.TessdataDir,
		Version:                version,
		BuildManifestSHA256:    before.BuildManifestSHA256,
		RuntimeInventorySHA256: before.RuntimeInventorySHA256,
		OCRSmokeSHA256:         before.OCRSmokeSHA256,
		RegisteredAt:           registeredAt,
	}
	if err := validateTesseractRuntimeRegistration(registration); err != nil {
		return TesseractRuntimeRegistration{}, err
	}
	details := map[string]any{
		"root":                     registration.Root,
		"executable":               registration.Executable,
		"tessdata_dir":             registration.TessdataDir,
		"version":                  registration.Version,
		"build_manifest_sha256":    registration.BuildManifestSHA256,
		"runtime_inventory_sha256": registration.RuntimeInventorySHA256,
		"ocr_smoke_sha256":         registration.OCRSmokeSHA256,
		"registered_at":            registration.RegisteredAt.Format(time.RFC3339Nano),
		"verified":                 true,
		"verification_mode":        "complete-wave4-runtime-inventory",
		"upstream":                 "tesseract-ocr/tesseract",
		"license":                  "Apache-2.0",
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	v.addChangeUnlocked("user", tesseractRuntimeBundleChangeType, "Registered verified ECO Tesseract runtime bundle", details)
	registration.AuditChangeID = v.Workspace.Changes[0].ID
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return TesseractRuntimeRegistration{}, fmt.Errorf("persist Tesseract runtime registration: %w", err)
	}
	return registration, nil
}

func (v *Vault) RegisteredTesseractRuntimeBundle() (TesseractRuntimeRegistration, error) {
	ws := v.Snapshot()
	for _, change := range ws.Changes {
		if change.Type != tesseractRuntimeBundleChangeType {
			continue
		}
		return tesseractRuntimeRegistrationFromChange(change)
	}
	return TesseractRuntimeRegistration{}, os.ErrNotExist
}

func (v *Vault) VerifyRegisteredTesseractRuntimeBundle() (TesseractRuntimeRegistration, error) {
	return v.VerifyRegisteredTesseractRuntimeBundleContext(context.Background())
}

func (v *Vault) VerifyRegisteredTesseractRuntimeBundleContext(ctx context.Context) (TesseractRuntimeRegistration, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	registration, err := v.RegisteredTesseractRuntimeBundle()
	if err != nil {
		return TesseractRuntimeRegistration{}, err
	}
	if err := validateTesseractRuntimeRegistration(registration); err != nil {
		return TesseractRuntimeRegistration{}, err
	}
	verified, err := verifyTesseractRuntimeBundleFiles(registration.Root)
	if err != nil {
		return TesseractRuntimeRegistration{}, fmt.Errorf("registered Tesseract runtime is unavailable or changed; locate the verified runtime again: %w", err)
	}
	if verified.Executable != registration.Executable || verified.TessdataDir != registration.TessdataDir || verified.BuildManifestSHA256 != registration.BuildManifestSHA256 || verified.RuntimeInventorySHA256 != registration.RuntimeInventorySHA256 || verified.OCRSmokeSHA256 != registration.OCRSmokeSHA256 {
		return TesseractRuntimeRegistration{}, errors.New("registered Tesseract runtime identity changed; locate the verified runtime again")
	}
	version, err := probeTesseractRuntimeBundle(ctx, verified)
	if err != nil {
		return TesseractRuntimeRegistration{}, err
	}
	if version != registration.Version {
		return TesseractRuntimeRegistration{}, errors.New("registered Tesseract runtime version changed; locate the verified runtime again")
	}
	after, err := verifyTesseractRuntimeBundleFiles(registration.Root)
	if err != nil || after != verified {
		return TesseractRuntimeRegistration{}, errors.New("registered Tesseract runtime changed during verification")
	}
	return registration, nil
}

func tesseractRuntimeRegistrationFromChange(change ChangeRecord) (TesseractRuntimeRegistration, error) {
	if change.Type != tesseractRuntimeBundleChangeType {
		return TesseractRuntimeRegistration{}, errors.New("wrong change type")
	}
	if verified, ok := change.Details["verified"].(bool); !ok || !verified {
		return TesseractRuntimeRegistration{}, errors.New("verified Tesseract runtime marker is missing")
	}
	registeredAt, err := time.Parse(time.RFC3339Nano, detailString(change.Details, "registered_at"))
	if err != nil {
		return TesseractRuntimeRegistration{}, errors.New("Tesseract runtime registration timestamp is missing or invalid")
	}
	registration := TesseractRuntimeRegistration{
		Root:                   detailString(change.Details, "root"),
		Executable:             detailString(change.Details, "executable"),
		TessdataDir:            detailString(change.Details, "tessdata_dir"),
		Version:                detailString(change.Details, "version"),
		BuildManifestSHA256:    detailString(change.Details, "build_manifest_sha256"),
		RuntimeInventorySHA256: detailString(change.Details, "runtime_inventory_sha256"),
		OCRSmokeSHA256:         detailString(change.Details, "ocr_smoke_sha256"),
		RegisteredAt:           registeredAt.UTC(),
		AuditChangeID:          change.ID,
	}
	if err := validateTesseractRuntimeRegistration(registration); err != nil {
		return TesseractRuntimeRegistration{}, err
	}
	return registration, nil
}

func validateTesseractRuntimeRegistration(registration TesseractRuntimeRegistration) error {
	if !filepath.IsAbs(registration.Root) || !filepath.IsAbs(registration.Executable) || !filepath.IsAbs(registration.TessdataDir) {
		return errors.New("Tesseract runtime registration paths must be absolute")
	}
	if registration.Version == "" || len([]rune(registration.Version)) > maxLocalToolIdentityRunes {
		return errors.New("Tesseract runtime version is missing or unbounded")
	}
	if registration.BuildManifestSHA256 != wave4BuildManifestSHA256 || registration.RuntimeInventorySHA256 != wave4RuntimeInventorySHA256 || registration.OCRSmokeSHA256 != wave4OCRSmokeSHA256 {
		return errors.New("Tesseract runtime registration is not bound to ECO's qualified Wave 4 control receipts")
	}
	if registration.RegisteredAt.IsZero() {
		return errors.New("Tesseract runtime registration timestamp is missing")
	}
	return nil
}

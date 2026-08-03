package eco

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type filesystemOps struct {
	rename func(string, string) error
	remove func(string) error
}

var operatingFilesystem = filesystemOps{rename: os.Rename, remove: os.Remove}

func sameFilesystemPath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func normalDirectory(path, description string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(info) {
		return nil, fmt.Errorf("%s is a symbolic link, junction or reparse point", description)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a folder", description)
	}
	return info, nil
}

func canonicalNormalDirectory(path, description string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absolute = filepath.Clean(absolute)
	if _, err = normalDirectory(absolute, description); err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return "", err
	}
	return filepath.Clean(canonical), nil
}

func validateObjectsDirectory(root, objects string) (string, error) {
	absoluteRoot, err := normaliseWorkspaceRoot(root)
	if err != nil {
		return "", err
	}
	canonicalRoot, err := canonicalNormalDirectory(absoluteRoot, "the selected workspace folder")
	if err != nil {
		return "", err
	}
	expected := filepath.Join(absoluteRoot, "objects")
	if !sameFilesystemPath(objects, expected) {
		return "", errors.New("the encrypted object folder is not the selected workspace's direct objects folder")
	}
	canonicalObjects, err := canonicalNormalDirectory(expected, "the encrypted object folder")
	if err != nil {
		return "", err
	}
	if !sameFilesystemPath(canonicalObjects, filepath.Join(canonicalRoot, "objects")) {
		return "", errors.New("the encrypted object folder resolves outside the selected workspace")
	}
	return canonicalObjects, nil
}

func managedObjectTarget(objectsCanonical, name string) (string, error) {
	if !safeManagedObjectName(name) {
		return "", errors.New("the managed encrypted object name is invalid")
	}
	target := filepath.Join(objectsCanonical, name)
	if !sameFilesystemPath(filepath.Dir(target), objectsCanonical) || filepath.Base(target) != name {
		return "", errors.New("the managed encrypted object does not remain a direct child of the objects folder")
	}
	return target, nil
}

func canonicalWorkspaceParent(root string) (logicalRoot, canonicalParent string, err error) {
	logicalRoot, err = normaliseWorkspaceRoot(root)
	if err != nil {
		return "", "", err
	}
	parent := filepath.Dir(logicalRoot)
	canonicalParent, err = canonicalNormalDirectory(parent, "the workspace parent folder")
	if err != nil {
		return "", "", err
	}
	return logicalRoot, canonicalParent, nil
}

func canonicalWorkspacePath(root string) (string, error) {
	logicalRoot, canonicalParent, err := canonicalWorkspaceParent(root)
	if err != nil {
		return "", err
	}
	if _, statErr := os.Lstat(logicalRoot); statErr == nil {
		return canonicalNormalDirectory(logicalRoot, "the workspace folder")
	} else if !os.IsNotExist(statErr) {
		return "", statErr
	}
	return filepath.Join(canonicalParent, filepath.Base(logicalRoot)), nil
}

func validateDirectSibling(root, target, expectedBase string, mustExist bool) (string, error) {
	logicalRoot, canonicalParent, err := canonicalWorkspaceParent(root)
	if err != nil {
		return "", err
	}
	logicalTarget := filepath.Join(filepath.Dir(logicalRoot), expectedBase)
	if !sameFilesystemPath(filepath.Clean(target), logicalTarget) {
		return "", errors.New("the migration target is not the expected direct sibling")
	}
	info, statErr := os.Lstat(logicalTarget)
	if statErr != nil {
		if os.IsNotExist(statErr) && !mustExist {
			return filepath.Join(canonicalParent, expectedBase), nil
		}
		return "", statErr
	}
	if info.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(info) {
		return "", errors.New("the migration target is a symbolic link, junction or reparse point")
	}
	canonical, err := filepath.EvalSymlinks(logicalTarget)
	if err != nil {
		return "", err
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return "", err
	}
	canonical = filepath.Clean(canonical)
	if !sameFilesystemPath(filepath.Dir(canonical), canonicalParent) || !sameFilesystemPath(canonical, filepath.Join(canonicalParent, expectedBase)) {
		return "", errors.New("the migration target resolves outside the canonical workspace parent")
	}
	return canonical, nil
}

func removeNormalTree(root string, remove func(string) error) error {
	canonicalRoot, err := canonicalNormalDirectory(root, "the migration staging folder")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	paths := []string{}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if info.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(info) {
			return errors.New("migration cleanup found a symbolic link, junction or reparse point")
		}
		absolute, absErr := filepath.Abs(path)
		if absErr != nil {
			return absErr
		}
		relative, relErr := filepath.Rel(root, absolute)
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return errors.New("migration cleanup found a path outside its staging folder")
		}
		paths = append(paths, filepath.Clean(absolute))
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(paths, func(i, j int) bool { return len(paths[i]) > len(paths[j]) })
	for _, path := range paths {
		info, statErr := os.Lstat(path)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(info) {
			return errors.New("migration cleanup was blocked because the staging tree changed")
		}
		parentCanonical, evalErr := filepath.EvalSymlinks(filepath.Dir(path))
		if evalErr != nil {
			return evalErr
		}
		relative, relErr := filepath.Rel(canonicalRoot, filepath.Join(parentCanonical, filepath.Base(path)))
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return errors.New("migration cleanup target escaped the authenticated staging folder")
		}
		if err = remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

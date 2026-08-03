//go:build !windows

package eco

import "os"

func fileInfoHasReparsePoint(os.FileInfo) bool { return false }

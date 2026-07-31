package eco

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var unsafeName = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func NewID(prefix string) string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().UTC().Format("20060102T150405"), hex.EncodeToString(b))
}

func SafeDisplayName(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	name = unsafeName.ReplaceAllString(name, "_")
	name = strings.Trim(name, ". ")
	if name == "" {
		name = "unnamed-evidence"
	}
	if len([]rune(name)) > 180 {
		r := []rune(name)
		name = string(r[:160]) + "_" + string(r[len(r)-16:])
	}
	return name
}

func HumanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

//go:build !window && !linux

package servicemgmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func SystemOpen(target string) {
	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("xdg-open", target)
	case "windows":
		runDll32 := filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
		_ = exec.Command(runDll32, "url.dll,FileProtocolHandler", target)
	default:
		_ = exec.Command("open", target)
	}
}

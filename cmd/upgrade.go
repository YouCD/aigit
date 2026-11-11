package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// repoURL is the git repository cloned for self-upgrade. Release builds
// override it at compile time with -ldflags "-X github.com/YouCD/aigit/cmd.repoURL=..." so the
// upgrade always fetches from the repository the binary was released from.
var repoURL = "https://github.com/YouCD/aigit"

// selfUpgrade replaces the running binary with the given release version:
// it builds the release tag with the local Go toolchain and swaps the
// binary in place.
func selfUpgrade(version string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating current binary: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go toolchain not found; cannot build %s from source", version)
	}

	tmpDir, err := os.MkdirTemp("", "aigit-upgrade-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	if err := runQuiet(exec.Command("git", "clone", "--quiet", "--depth", "1", "--branch", version, repoURL, srcDir)); err != nil {
		return fmt.Errorf("downloading %s: %w", version, err)
	}

	newBin := filepath.Join(tmpDir, "aigit")
	build := exec.Command(goBin, "build", "-ldflags", "-s -w -X github.com/YouCD/aigit/cmd.Version="+version, "-o", newBin, ".")
	build.Dir = srcDir
	if err := runQuiet(build); err != nil {
		return fmt.Errorf("building %s: %w", version, err)
	}

	return replaceBinary(newBin, exe)
}

// replaceBinary swaps dst for newBin via a rename from a sibling temp file,
// so the running executable is replaced atomically rather than truncated.
func replaceBinary(newBin, dst string) error {
	data, err := os.ReadFile(newBin)
	if err != nil {
		return err
	}
	staged := dst + ".new"
	if err := os.WriteFile(staged, data, 0o755); err != nil {
		return fmt.Errorf("staging new binary (try upgrading with elevated permissions): %w", err)
	}
	if err := os.Rename(staged, dst); err != nil {
		os.Remove(staged)
		return fmt.Errorf("installing new binary: %w", err)
	}
	return nil
}

func runQuiet(cmd *exec.Cmd) error {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

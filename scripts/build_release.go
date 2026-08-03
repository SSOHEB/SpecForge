// Package main compiles and generates cross-platform releases for configforge.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type target struct {
	os   string
	arch string
	ext  string
}

func main() {
	version := getGitVersion()
	commit := getGitCommit()
	date := time.Now().Format(time.RFC3339)

	fmt.Printf("Building configforge %s (commit: %s, date: %s)...\n", version, commit, date)

	targets := []target{
		{"linux", "amd64", ""},
		{"linux", "arm64", ""},
		{"darwin", "amd64", ""},
		{"darwin", "arm64", ""},
		{"windows", "amd64", ".exe"},
	}

	distDir := "dist"
	if err := os.MkdirAll(distDir, 0755); err != nil {
		fmt.Printf("failed to create dist dir: %v\n", err)
		os.Exit(1)
	}

	var checksumLines []string

	for _, t := range targets {
		binaryName := fmt.Sprintf("configforge_%s_%s_%s%s", version, t.os, t.arch, t.ext)
		outPath := filepath.Join(distDir, binaryName)

		fmt.Printf("Building %s/%s -> %s...\n", t.os, t.arch, outPath)

		// Set environment variables for cross compilation
		cmd := exec.Command("go", "build",
			"-ldflags", fmt.Sprintf("-X configforge/cmd/configforge/cmd.version=%s -X configforge/cmd/configforge/cmd.commit=%s -X configforge/cmd/configforge/cmd.date=%s", version, commit, date),
			"-o", outPath,
			"./cmd/configforge",
		)
		cmd.Env = append(os.Environ(),
			"GOOS="+t.os,
			"GOARCH="+t.arch,
			"CGO_ENABLED=0",
		)

		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("failed to build target %s/%s: %v\nOutput: %s\n", t.os, t.arch, err, string(out))
			os.Exit(1)
		}

		// Calculate checksum
		checksum, err := calculateSHA256(outPath)
		if err != nil {
			fmt.Printf("failed to calculate checksum for %s: %v\n", outPath, err)
			os.Exit(1)
		}
		checksumLines = append(checksumLines, fmt.Sprintf("%s  %s", checksum, binaryName))
	}

	// Write checksums.txt
	checksumsContent := strings.Join(checksumLines, "\n") + "\n"
	checksumPath := filepath.Join(distDir, "checksums.txt")
	if err := os.WriteFile(checksumPath, []byte(checksumsContent), 0644); err != nil {
		fmt.Printf("failed to write checksums.txt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote checksums to %s\n", checksumPath)
	fmt.Println("Release build completed successfully.")
}

func getGitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--always")
	out, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(out))
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "none"
	}
	return strings.TrimSpace(string(out))
}

func calculateSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

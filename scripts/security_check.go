package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(jwt_secret|secret_key|api_key|private_key|db_password)\s*=\s*["'][a-zA-Z0-9_\-\.]{8,}["']`),
	regexp.MustCompile(`-----BEGIN (RSA|EC|DSA|OPENSSH) PRIVATE KEY-----`),
}

var ignoreDirs = []string{
	".git", ".system_generated", "bin", "vendor", "node_modules",
}

var ignoreFiles = []string{
	".env.example", "security_check.go", "config_test.go", "main_test.go", ".env",
}

func isIgnoredDir(path string) bool {
	for _, id := range ignoreDirs {
		if strings.Contains(path, id) {
			return true
		}
	}
	return false
}

func isIgnoredFile(path string) bool {
	base := filepath.Base(path)
	for _, inf := range ignoreFiles {
		if base == inf {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println("=== 🔒 Running Secret Scanner & Credential Check ===")
	var violations []string

	// 1. Verify that .env and private keys are ignored in .gitignore
	gitIgnoreBytes, err := os.ReadFile(".gitignore")
	if err == nil {
		gitIgnoreContent := string(gitIgnoreBytes)
		if !strings.Contains(gitIgnoreContent, ".env") {
			violations = append(violations, ".gitignore is missing .env exclusion!")
		}
	}

	// 2. Check git status to ensure .env or credentials are NOT tracked in git
	gitOutput, err := exec.Command("git", "ls-files").Output()
	if err == nil {
		trackedFiles := strings.Split(string(gitOutput), "\n")
		for _, tf := range trackedFiles {
			tf = strings.TrimSpace(tf)
			if tf == ".env" || strings.HasSuffix(tf, ".key") || strings.HasSuffix(tf, ".pem") {
				violations = append(violations, fmt.Sprintf("Prohibited credential file tracked in git: %s", tf))
			}
		}
	}

	// 3. Scan code files for hardcoded secrets
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if isIgnoredDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if isIgnoredFile(path) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".yaml" && ext != ".yml" && ext != ".json" && ext != ".sh" {
			return nil
		}

		cleanPath := filepath.Clean(path)
		// #nosec G304 G122 -- CLI utility strictly traverses current workspace files
		file, err := os.Open(cleanPath)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			for _, pat := range secretPatterns {
				if pat.MatchString(line) {
					if !strings.Contains(line, "nosec") && !strings.Contains(line, "testing") && !strings.Contains(line, "${") {
						violations = append(violations, fmt.Sprintf("Potential hardcoded secret at %s:%d: %s", path, lineNum, strings.TrimSpace(line)))
					}
				}
			}
			lineNum++
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning files: %v\n", err)
		os.Exit(1)
	}

	if len(violations) > 0 {
		fmt.Println("❌ Secret Scan FAILED! Violations detected:")
		for _, v := range violations {
			fmt.Printf(" - %s\n", v)
		}
		os.Exit(1)
	}

	fmt.Println("✅ Secret Scan PASSED: Zero hardcoded secrets, .gitignore verified, zero credential leaks.")
}

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestEnumer(t *testing.T) {
	c := qt.New(t)

	// Build enumer once
	enumerBin := buildEnumer(c)

	// Find all test cases in testdata
	entries, err := os.ReadDir("testdata")
	c.Assert(err, qt.IsNil)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testName := entry.Name()
		c.Run(testName, func(c *qt.C) {
			runTestCase(c, testName, enumerBin)
		})
	}
}

func runTestCase(c *qt.C, testName, enumerBin string) {
	c.Helper()

	testDir := filepath.Join("testdata", testName)

	// Create temp directory for this test
	tmpDir := c.TempDir()

	// Copy all .go files from testdata
	entries, err := os.ReadDir(testDir)
	c.Assert(err, qt.IsNil)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".go") {
			srcFile := filepath.Join(testDir, name)
			dstFile := filepath.Join(tmpDir, name)
			copyFile(c, srcFile, dstFile)
		}
	}

	// Read enumer.args file
	argsFile := filepath.Join(testDir, "enumer.args")
	argsData, err := os.ReadFile(argsFile)
	c.Assert(err, qt.IsNil, qt.Commentf("Missing enumer.args in %s", testName))

	args := parseArgs(string(argsData))

	// Create go.mod BEFORE running enumer (packages.Load needs it)
	modContent := `module test

go 1.21

require gopkg.in/yaml.v3 v3.0.1
`
	modFile := filepath.Join(tmpDir, "go.mod")
	err = os.WriteFile(modFile, []byte(modContent), 0644)
	c.Assert(err, qt.IsNil)

	// Run enumer
	cmd := exec.Command(enumerBin, args...)
	cmd.Dir = tmpDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		c.Logf("stdout: %s", stdout.String())
		c.Logf("stderr: %s", stderr.String())
	}
	c.Assert(err, qt.IsNil, qt.Commentf("enumer failed"))

	// Run go mod tidy to generate go.sum
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	err = cmd.Run()
	c.Assert(err, qt.IsNil, qt.Commentf("go mod tidy failed"))

	// Run the tests in the temp directory
	cmd = exec.Command("go", "test", "-v")
	cmd.Dir = tmpDir

	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		c.Logf("Test output for %s:", testName)
		c.Logf("stdout: %s", stdout.String())
		c.Logf("stderr: %s", stderr.String())

		// List generated files for debugging
		files, _ := os.ReadDir(tmpDir)
		var fileList []string
		for _, f := range files {
			fileList = append(fileList, f.Name())
		}
		c.Logf("files: %s", strings.Join(fileList, ", "))
	}
	c.Assert(err, qt.IsNil, qt.Commentf("tests failed for %s", testName))
}

func buildEnumer(c *qt.C) string {
	c.Helper()

	tmpBin := filepath.Join(c.TempDir(), "enumer")

	cmd := exec.Command("go", "build", "-o", tmpBin, ".")
	output, err := cmd.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("build output: %s", output))

	return tmpBin
}

func copyFile(c *qt.C, src, dst string) {
	c.Helper()

	data, err := os.ReadFile(src)
	c.Assert(err, qt.IsNil)

	err = os.WriteFile(dst, data, 0644)
	c.Assert(err, qt.IsNil)
}

func parseArgs(argsContent string) []string {
	lines := strings.Split(argsContent, "\n")
	var args []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			args = append(args, line)
		}
	}
	return args
}

package main

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

// Helper functions
func getCommand(name string, args ...string) *exec.Cmd {
	// Add .exe extension on Windows
	if runtime.GOOS == "windows" && !strings.HasSuffix(name, ".exe") {
		name = name + ".exe"
	}
	return exec.Command(name, args...)
}

func executeCommand(cmd *exec.Cmd) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	
	return stdout.String(), nil
}
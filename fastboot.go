package main

import (
    "fmt"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

func (t *FlashTool) getFastbootInfo() {
    if !t.isDeviceConnected() {
        t.appendLog("Error: No device connected!")
        return
    }

    startTime := time.Now()

    t.appendLog("Read Device Info Result:")
    t.appendLog("========= Device Information =========")

    // Get individual variables
    getVarInfo := func(varName string) string {
        cmd := exec.Command("fastboot", "getvar", varName)
        output, err := cmd.CombinedOutput()
        if err == nil {
            lines := strings.Split(string(output), "\n")
            for _, line := range lines {
                line = strings.TrimSpace(line)
                if strings.HasPrefix(line, varName) {
                    parts := strings.SplitN(line, ":", 2)
                    if len(parts) == 2 {
                        return strings.TrimSpace(parts[1])
                    }
                }
            }
        }
        return ""
    }

    // Define variables to check
    vars := []struct {
        label    string
        varNames []string
    }{
        
     
        {"Device Model", []string{"product", "device", "model", "product-name"}},
        {"Android Version", []string{"version", "build-version"}},
        {"Anti Number", []string{"anti"}},
        {"Serial NO:", []string{"serialno"}},
        {"Security Patch", []string{"current-slot"}},
        {"Software Version", []string{"version-baseband", "version-software"}},
        {"Root Access", []string{"secure"}},
        
    }

    // Get and display information
    for _, v := range vars {
        value := "Unknown"
        for _, varName := range v.varNames {
            if result := getVarInfo(varName); result != "" {
                value = result
                break
            }
        }
        t.appendLog(fmt.Sprintf("%-20s: %s", v.label, value))
    }

    // Additional device details
   

    // Calculate execution time
    executionTime := time.Since(startTime)

    t.appendLog("\n========= Operation Status =========")
    t.appendLog("✅ Completed successfully")
    t.appendLog(fmt.Sprintf("⏱️ Execution time: %.2fs", executionTime.Seconds()))
}

func (t *FlashTool) executeBatch() {
    t.logOutput.SetText("")
    t.appendLog(fmt.Sprintf("Starting execution of: %s", filepath.Base(t.filePath)))
    
    startTime := time.Now()
    
    if !t.isDeviceConnected() {
        t.appendLog("Error: No device connected!")
        return
    }
    
    cmd := exec.Command("cmd", "/C", t.filePath)
    output, err := cmd.CombinedOutput()
    
    executionTime := time.Since(startTime)
    
    if err != nil {
        t.appendLog(fmt.Sprintf("Error executing batch file:\n%s", string(output)))
        t.appendLog(fmt.Sprintf("Error details: %v", err))
        t.appendLog("\n=== Operation Status ===")
        t.appendLog("❌ Execution failed")
    } else {
        t.appendLog(string(output))
        t.appendLog("\n=== Operation Status ===")
        t.appendLog("✅ Completed successfully")
    }
    
    t.appendLog(fmt.Sprintf("⏱️ Execution time: %.2fs", executionTime.Seconds()))
}

func (t *FlashTool) checkFastbootDevice() {
    cmd := exec.Command("fastboot", "devices")
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        t.appendLog("=== Device Check ===")
        t.appendLog("❌ Error checking devices")
        t.appendLog(fmt.Sprintf("Error details: %v", err))
        return
    }
    
    if string(output) == "" {
        t.appendLog("=== Device Check ===")
        t.appendLog("❌ No devices found")
        return
    }
    
    t.appendLog("=== Device Check ===")
    t.appendLog("✅ Device connected")
    t.appendLog(string(output))
}

func (t *FlashTool) isDeviceConnected() bool {
    cmd := exec.Command("fastboot", "devices")
    output, err := cmd.CombinedOutput()
    return err == nil && len(output) > 0
}
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

// Fastboot reboot
func (t *FlashTool) fastbootReboot() {
    if !t.isDeviceConnected() {
        t.appendLog("Error: No device connected!")
        return
    }

    cmd := exec.Command("fastboot", "reboot")
    output, err := cmd.CombinedOutput()

    if err != nil {
        t.appendLog("Error rebooting device:")
        t.appendLog(fmt.Sprintf("%s\nError details: %v", string(output), err))
        return
    }

    t.appendLog("Device rebooted successfully.")
}

// Fastboot unlock bootloader
func (t *FlashTool) fastbootUnlock() {
    if !t.isDeviceConnected() {
        t.appendLog("❌ No device connected!")
        return
    }

    t.appendLog("🔓 Attempting to unlock bootloader...")
    t.appendLog("⚠️ WARNING: This will WIPE ALL DATA on your device!")
    t.appendLog("⚠️ Make sure you have a backup of important data!")
    
    t.appendLog("\n🔍 Checking unlock status...")
    
    // Check if already unlocked
    cmd := exec.Command("fastboot", "getvar", "unlocked")
    output, err := cmd.CombinedOutput()
    
    if err == nil {
        if strings.Contains(string(output), "unlocked: yes") {
            t.appendLog("✅ Bootloader is already unlocked!")
            return
        }
    }
    
    // Try standard unlock command
    t.appendLog("🚀 Executing unlock command...")
    cmd = exec.Command("fastboot", "oem", "unlock")
    output, err = cmd.CombinedOutput()
    
    if err != nil {
        t.appendLog("❌ Standard unlock failed, trying alternative...")
        // Try alternative unlock command
        cmd = exec.Command("fastboot", "flashing", "unlock")
        output, err = cmd.CombinedOutput()
        
        if err != nil {
            t.appendLog("❌ Unlock failed!")
            t.appendLog(fmt.Sprintf("Error: %v", err))
            t.appendLog("📌 Note: Make sure OEM unlocking is enabled in Developer Options")
            return
        }
    }
    
    t.appendLog("✅ Unlock command sent successfully!")
    t.appendLog("📱 Please check your device screen for confirmation")
    t.appendLog("🔽 Use Volume keys to navigate and Power to confirm")
    t.appendLog("\n" + string(output))
}
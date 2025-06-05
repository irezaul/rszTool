package main

import (
    "fmt"
    "os/exec"
    "strings"
    "time"
)

// Enhanced device connection check
func (t *FlashTool) isADBDeviceConnected() (bool, string) {
    cmd := exec.Command("adb", "devices")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return false, "Error executing ADB command"
    }

    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" && !strings.Contains(line, "List of devices") {
            parts := strings.Fields(line)
            if len(parts) >= 2 {
                switch parts[1] {
                case "device":
                    return true, parts[0] // Return true and device ID
                case "unauthorized":
                    return false, "Device unauthorized"
                case "offline":
                    return false, "Device offline"
                }
            }
        }
    }
    return false, "No device connected"
}

func (t *FlashTool) checkADBDevice() {
    t.logOutput.SetText("")
    t.appendLog("=== ADB Device Check ===")
    t.appendLog("Checking device connection...")

    connected, status := t.isADBDeviceConnected()
    if connected {
        t.appendLog("Status    : ✅ Device Connected")
        t.appendLog(fmt.Sprintf("Device ID : %s", status))
        
        // Get additional device info
        if model := t.getDeviceProp("ro.product.model"); model != "" {
            t.appendLog(fmt.Sprintf("Model     : %s", model))
        }
        if brand := t.getDeviceProp("ro.product.brand"); brand != "" {
            t.appendLog(fmt.Sprintf("Brand     : %s", brand))
        }
    } else {
        t.appendLog("Status    : ❌ Device Disconnected")
        t.appendLog(fmt.Sprintf("Reason    : %s", status))
        t.appendLog("\nTroubleshooting:")
        t.appendLog("1. Check USB connection")
        t.appendLog("2. Enable USB debugging")
        t.appendLog("3. Accept USB debugging prompt on device")
        t.appendLog("4. Try different USB port/cable")
    }
}

// Helper function to get device property
func (t *FlashTool) getDeviceProp(prop string) string {
    cmd := exec.Command("adb", "shell", "getprop", prop)
    output, err := cmd.CombinedOutput()
    if err == nil {
        return strings.TrimSpace(string(output))
    }
    return ""
}

func (t *FlashTool) getADBInfo() {
    t.logOutput.SetText("")
    connected, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("=== Device Information ===")
        t.appendLog("❌ Connection Error")
        t.appendLog(fmt.Sprintf("Details: %s", status))
        t.appendLog("\nPlease ensure:")
        t.appendLog("1. Device is connected via USB")
        t.appendLog("2. USB debugging is enabled")
        t.appendLog("3. Device is unlocked")
        return
    }

    startTime := time.Now()
    t.appendLog("Read Device Info Result:")
    t.appendLog("=== Device Information ===")

    // Enhanced device properties
    props := []struct {
        label string
        cmd   string
        type_ string // "prop" or "cmd"
    }{
        {"Brand", "ro.product.brand", "prop"},
        {"Model", "ro.product.model", "prop"},
        {"Device", "ro.product.device", "prop"},
        {"Hardware level", "ro.boot.hwlevel", "prop"},
        {"Android Version", "ro.build.version.release", "prop"},
        {"Security Patch", "ro.build.version.security_patch", "prop"},
        {"Build Number", "ro.build.display.id", "prop"},
        {"CPU Architecture", "ro.product.cpu.abi", "prop"},
        {"RAM", "cat /proc/meminfo | grep MemTotal", "cmd"},
        {"Storage", "df -h /data", "cmd"},
       
        {"Battery Level", "", "battery"},
        
    }

    for _, prop := range props {
        var value string
        switch prop.type_ {
        case "prop":
            value = t.getDeviceProp(prop.cmd)
        case "cmd":
            cmd := exec.Command("adb", "shell", prop.cmd)
            output, err := cmd.CombinedOutput()
            if err == nil {
                value = strings.TrimSpace(string(output))
                // Process specific outputs
                switch prop.label {
                case "RAM":
                    if parts := strings.Fields(value); len(parts) >= 2 {
                        value = parts[1] + " " + parts[2]
                    }
                case "Storage":
                    lines := strings.Split(value, "\n")
                    if len(lines) > 1 {
                        fields := strings.Fields(lines[1])
                        if len(fields) >= 4 {
                            value = fmt.Sprintf("Total: %s, Used: %s, Free: %s", fields[1], fields[2], fields[3])
                        }
                    }
                }
            }
        case "battery":
            cmd := exec.Command("adb", "shell", "dumpsys", "battery")
            output, err := cmd.CombinedOutput()
            if err == nil {
                lines := strings.Split(string(output), "\n")
                for _, line := range lines {
                    if strings.Contains(line, "level:") {
                        value = strings.TrimSpace(strings.Split(line, ":")[1]) + "%"
                        break
                    }
                }
            }
        }

        if value == "" {
            value = "Not available"
        }
        t.appendLog(fmt.Sprintf("%-20s: %s", prop.label, value))
    }

    // Additional system information

  

    executionTime := time.Since(startTime)
    t.appendLog("\n=== Operation Status ===")
    t.appendLog("✅ Information gathering completed")
    t.appendLog(fmt.Sprintf("⏱️ Execution time: %.2fs", executionTime.Seconds()))
}

func (t *FlashTool) adbReboot() {
    t.logOutput.SetText("")
    connected, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("=== Reboot Device ===")
        t.appendLog("❌ Operation Failed")
        t.appendLog(fmt.Sprintf("Reason: %s", status))
        return
    }

    t.appendLog("=== Rebooting Device ===")
    t.appendLog("⏳ Initiating reboot sequence...")
    cmd := exec.Command("adb", "reboot")
    if err := cmd.Run(); err != nil {
        t.appendLog("❌ Reboot failed")
        t.appendLog(fmt.Sprintf("Error: %v", err))
        return
    }
    t.appendLog("✅ Device is rebooting")
    t.appendLog("Please wait while the device restarts...")
}

func (t *FlashTool) adbRebootFastboot() {
    t.logOutput.SetText("")
    connected, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("=== Reboot to Fastboot ===")
        t.appendLog("❌ Operation Failed")
        t.appendLog(fmt.Sprintf("Reason: %s", status))
        return
    }

    t.appendLog("=== Rebooting to Fastboot ===")
    t.appendLog("⏳ Initiating fastboot reboot sequence...")
    cmd := exec.Command("adb", "reboot", "bootloader")
    if err := cmd.Run(); err != nil {
        t.appendLog("❌ Reboot to fastboot failed")
        t.appendLog(fmt.Sprintf("Error: %v", err))
        return
    }
    t.appendLog("✅ Device is rebooting to fastboot")
    t.appendLog("Please wait for fastboot mode...")
}

func (t *FlashTool) adbRebootRecovery() {
    t.logOutput.SetText("")
    connected, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("=== Reboot to Recovery ===")
        t.appendLog("❌ Operation Failed")
        t.appendLog(fmt.Sprintf("Reason: %s", status))
        return
    }

    t.appendLog("=== Rebooting to Recovery ===")
    t.appendLog("⏳ Initiating recovery reboot sequence...")
    cmd := exec.Command("adb", "reboot", "recovery")
    if err := cmd.Run(); err != nil {
        t.appendLog("❌ Reboot to recovery failed")
        t.appendLog(fmt.Sprintf("Error: %v", err))
        return
    }
    t.appendLog("✅ Device is rebooting to recovery")
    t.appendLog("Please wait for recovery mode...")
}
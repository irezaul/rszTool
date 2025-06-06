package main

import (
    "fmt"
    "os/exec"
    "strings"
    "time"
)

// Check if ADB device is connected
func (t *FlashTool) isADBDeviceConnected() (bool, string, string) {
    cmd := exec.Command("adb", "devices")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return false, "", "ADB not responding"
    }

    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" && !strings.Contains(line, "List of devices") {
            parts := strings.Fields(line)
            if len(parts) >= 2 {
                switch parts[1] {
                case "device":
                    return true, parts[0], "Connected"
                case "unauthorized":
                    return false, parts[0], "Unauthorized (Check USB debugging)"
                case "offline":
                    return false, parts[0], "Device offline (reconnect USB)"
                }
            }
        }
    }

    return false, "", "No device connected"
}

// ✅ Enable DIAG mode without root (if possible)
func (t *FlashTool) adbEnableDiag() {
    t.logOutput.SetText("")

    connected, deviceID, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("❌ No device connected")
        t.appendLog(fmt.Sprintf("Reason: %s", status))
        return
    }

    t.appendLog("🚀 Attempting to enable DIAG mode...")
    t.appendLog(fmt.Sprintf("Device ID: %s", deviceID))

    // Run commands to attempt diag activation
    cmds := [][]string{
        {"shell", "am", "start", "-n", "com.longcheertel.midtest/com.longcheertel.midtest.Diag"},
        
    }

    for _, args := range cmds {
        err := exec.Command("adb", args...).Run()
        if err != nil {
            t.appendLog(fmt.Sprintf("⚠️ Failed: adb %s", "Failed to connect"))
        } else {
            t.appendLog(fmt.Sprintf("✅ Success: adb %s",  "Enabled DIAG mode"))
        }
    }

    time.Sleep(1 * time.Second)

    // Verify that diag was enabled
    out, err := exec.Command("shell", "am", "start", "-n", "com.longcheertel.midtest/com.longcheertel.midtest.Diag").CombinedOutput()
    if err == nil {
        value := strings.TrimSpace(string(out))
        t.appendLog(fmt.Sprintf("🔍 Current USB Config: %s", value))
        if strings.Contains(value, "diag") {
            t.appendLog("✅ DIAG mode looks active!")
        } else {
            t.appendLog("⚠️ DIAG not confirmed. Device may need reboot.")
        }
    }

    t.appendLog("📌 Tip: Check Windows Device Manager > Ports (COM) for Qualcomm DIAG port.")
    
    t.appendLog("📌 Note: This method works only on some Qualcomm devices (no root needed).")
}

// ✅ Display basic device connection
func (t *FlashTool) checkADBDevice() {
    t.logOutput.SetText("")
    connected, deviceID, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("❌ No ADB device detected")
        t.appendLog(fmt.Sprintf("Status: %s", status))
        return
    }

    t.appendLog("✅ Device Connected")
    t.appendLog(fmt.Sprintf("Device ID : %s", deviceID))
    t.appendLog(fmt.Sprintf("Status    : %s", status))
    t.appendLog(fmt.Sprintf("Time      : %s", time.Now().Format("15:04:05")))
}

// ✅ Show detailed ADB info
func (t *FlashTool) getADBInfo() {
    t.logOutput.SetText("")
    connected, _, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("❌ Cannot read device info - not connected")
        t.appendLog(fmt.Sprintf("Status: %s", status))
        return
    }

    start := time.Now()
    t.appendLog("💡 Reading device information...")

    props := []struct {
        Label string
        Prop  string
    }{
        {"Brand", "ro.product.brand"},
        {"Model", "ro.product.model"},
        {"Device", "ro.product.device"},
        {"Android Version", "ro.build.version.release"},
        {"Build Number", "ro.build.display.id"},
        {"Security Patch", "ro.build.version.security_patch"},
        {"CPU", "ro.product.cpu.abi"},
        {"Bootloader", "ro.bootloader"},
        {"Battery Level", ""}, // special case
    }

    for _, prop := range props {
        var value string
        if prop.Label == "Battery Level" {
            cmd := exec.Command("adb", "shell", "dumpsys", "battery")
            out, err := cmd.CombinedOutput()
            if err == nil {
                lines := strings.Split(string(out), "\n")
                for _, line := range lines {
                    if strings.Contains(line, "level:") {
                        value = strings.Split(line, ":")[1]
                        value = strings.TrimSpace(value) + "%"
                        break
                    }
                }
            }
        } else {
            out, err := exec.Command("adb", "shell", "getprop", prop.Prop).CombinedOutput()
            if err == nil {
                value = strings.TrimSpace(string(out))
            }
        }

        if value == "" {
            value = "Unknown"
        }

        t.appendLog(fmt.Sprintf("%-18s: %s", prop.Label, value))
    }

    elapsed := time.Since(start)
    t.appendLog(fmt.Sprintf("\n✅ Info retrieved in %.2fs", elapsed.Seconds()))
}

// ✅ Reboot normally
func (t *FlashTool) adbReboot() {
    t.logOutput.SetText("")
    connected, _, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("❌ Cannot reboot - device not connected")
        t.appendLog(fmt.Sprintf("Status: %s", status))
        return
    }

    t.appendLog("🔁 Rebooting device...")
    err := exec.Command("adb", "reboot").Run()
    if err != nil {
        t.appendLog(fmt.Sprintf("❌ Reboot failed: %v", err))
        return
    }

    t.appendLog("✅ Reboot command sent")
}

// ✅ Reboot to bootloader / fastboot
func (t *FlashTool) adbRebootFastboot() {
    t.logOutput.SetText("")
    connected, _, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("❌ Cannot reboot to fastboot - no device detected")
        t.appendLog(fmt.Sprintf("Status: %s", status))
        return
    }

    t.appendLog("🚀 Rebooting to fastboot...")
    err := exec.Command("adb", "reboot", "bootloader").Run()
    if err != nil {
        t.appendLog(fmt.Sprintf("❌ Fastboot reboot failed: %v", err))
        return
    }

    t.appendLog("✅ Fastboot command sent")
}

// ✅ Reboot to recovery
func (t *FlashTool) adbRebootRecovery() {
    t.logOutput.SetText("")
    connected, _, status := t.isADBDeviceConnected()
    if !connected {
        t.appendLog("❌ Cannot reboot to recovery - no device detected")
        t.appendLog(fmt.Sprintf("Status: %s", status))
        return
    }

    t.appendLog("🛠 Rebooting to recovery...")
    err := exec.Command("adb", "reboot", "recovery").Run()
    if err != nil {
        t.appendLog(fmt.Sprintf("❌ Recovery reboot failed: %v", err))
        return
    }

    t.appendLog("✅ Recovery command sent")
}
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
        {"shell", "setprop", "sys.usb.config", "diag,adb"},
        {"shell", "setprop", "vendor.usb.config", "diag,adb"},
    }

    for i, args := range cmds {
        cmd := exec.Command("adb", args...)
        err := cmd.Run()
        if err != nil {
            t.appendLog(fmt.Sprintf("⚠️ Command %d failed: %v", i+1, err))
        } else {
            t.appendLog(fmt.Sprintf("✅ Command %d executed successfully", i+1))
        }
    }

    time.Sleep(1 * time.Second)

    // Try to get USB config
    cmd := exec.Command("adb", "shell", "getprop", "sys.usb.config")
    out, err := cmd.CombinedOutput()
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
        {"Phone Model", "ro.product.odm.marketname"},
        {"Device", "ro.product.device"},
        {"Region", "ro.miui.build.region"},
        {"Firmware State", "ro.product.mod_device"},
        {"CPU", "ro.boot.hardware"},
        {"Hardware Level", "ro.boot.hwlevel"},
        {"Manufacturer", "ro.product.system_ext.manufacturer"},
        {"Android Version", "ro.build.version.release"},
        {"Build Number", "ro.system_ext.build.version.incremental"},
        {"Security Patch", "ro.build.version.security_patch"},
        {"CPU-Product", "ro.product.cpu.abi"},
        {"Bootloader", "ro.secureboot.lockstate"},
        {"imei-1", "ro.ril.oem.imei"},
        {"imei-2", "ro.ril.oem.imei2"},
        {"IMEI", "ro.ril.miui.imei0"},
        {"IMEI2", "ro.ril.miui.imei1"},
        {"Battery Level", ""},
    }

    labelFormat := map[string]string{
        "Brand":           "%-21s: %s",
        "Model":           "%-21s: %s",
        "Phone Model":     "%-15s: %s",
        "Device":          "%-21s: %s",
        "Region":          "%-21s: %s",
        "Firmware State":  "%-15s: %s",
        "CPU":             "%-24s: %s",
        "Hardware Level":  "%-15s: %s",
        "Manufacturer":    "%-21s: %s",
        "Android Version": "%-20s: %s",
        "Build Number":    "%-20s: %s",
        "Security Patch":  "%-22s: %s",
        "CPU-Product":     "%-20s: %s",
        "Bootloader":      "%-21s: %s",
        "imei-1":          "%-26s: %s",
        "imei-2":          "%-26s: %s",
        "IMEI":            "%-26s: %s",
        "IMEI2":           "%-25s: %s",
        "Battery Level":   "%-20s: %s",
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
                        parts := strings.Split(line, ":")
                        if len(parts) >= 2 {
                            value = strings.TrimSpace(parts[1]) + "%"
                            break
                        }
                    }
                }
            }
        } else {
            out, err := exec.Command("adb", "shell", "getprop", prop.Prop).CombinedOutput()
            if err == nil {
                value = strings.TrimSpace(string(out))
            }
        }

        if value != "" && value != "Unknown" {
            format, exists := labelFormat[prop.Label]
            if !exists {
                format = "%-20s: %s"  // default format if not specified
            }
            t.appendLog(fmt.Sprintf(format, prop.Label, value))
        }
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
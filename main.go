package main

import (
    "fmt"
    "path/filepath"
    "syscall"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

type FlashTool struct {
    window    fyne.Window
    logOutput *widget.Entry
    filePath  string
}

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("RSZ Tool")
    
    tool := &FlashTool{
        window: myWindow,
    }
    
    tool.createUI()
    myWindow.Resize(fyne.NewSize(800, 600))
    myWindow.ShowAndRun()
}

func (t *FlashTool) createUI() {
    // Create log output
    t.logOutput = widget.NewMultiLineEntry()
    t.logOutput.SetPlaceHolder(`Logs will appear here...

Common codes to enable DM/DIAG mode (if needed):

*#*#717717#*#*  (Common for Xiaomi/Redmi)

*#*#83781#*#*   (Huawei)

*#*#4636#*#* Phone Info → Turn on "DM Mode"
`)
    t.logOutput.Wrapping = fyne.TextWrapWord

    // Create Fastboot tab content
    fastbootTab := t.createFastbootTab()

    // firmwareTab := t.createFirmwareTab()

    // Create ADB tab content
    adbTab := t.createADBTab()

    // Create tabs
    tabs := container.NewAppTabs(
        container.NewTabItem("Fastboot", fastbootTab),
        // container.NewTabItem("Firmware", firmwareTab),
        container.NewTabItem("ADB", adbTab),
    )
    tabs.SetTabLocation(container.TabLocationTop)

    // Bottom buttons
    clearButton := widget.NewButton("Clear Log", func() {
        t.logOutput.SetText("")
    })

    exitButton := widget.NewButton("Exit", func() {
        t.window.Close()
    })

    bottomButtons := container.NewHBox(
        clearButton,
        exitButton,
    )

    // Main content
    content := container.NewBorder(
        tabs,
        bottomButtons,
        nil,
        nil,
        t.logOutput,
    )
    
    t.window.SetContent(content)
}

func (t *FlashTool) createFastbootTab() fyne.CanvasObject {
    // Fastboot buttons
    fileButton := widget.NewButton("Select Batch File", func() {
        dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
            if err != nil {
                dialog.ShowError(err, t.window)
                return
            }
            if uri == nil {
                return
            }
            t.filePath = uri.URI().Path()
            t.logOutput.SetText("")
            t.appendLog("Selected file: " + filepath.Base(t.filePath))
        }, t.window)
    })
    
    executeButton := widget.NewButton("Execute Batch", func() {
        if t.filePath == "" {
            dialog.ShowError(fmt.Errorf("please select a batch file first"), t.window)
            return
        }
        go t.executeBatch()
    })
    
    deviceButton := widget.NewButton("Check Device", func() {
        t.logOutput.SetText("")
        go t.checkFastbootDevice()
    })

    infoButton := widget.NewButton("Device Info", func() {
        t.logOutput.SetText("")
        go t.getFastbootInfo()
    })

    // Create grid layout for fastboot buttons
    return container.NewGridWithColumns(4,
        fileButton,
        executeButton,
        deviceButton,
        infoButton,
    )
}

func (t *FlashTool) createADBTab() fyne.CanvasObject {
    // ADB buttons
    deviceButton := widget.NewButton("Check Device", func() {
        t.logOutput.SetText("")
        go t.checkADBDevice()
    })

    infoButton := widget.NewButton("Device Info", func() {
        t.logOutput.SetText("")
        go t.getADBInfo()
    })

    rebootButton := widget.NewButton("Reboot", func() {
        go t.adbReboot()
    })

    rebootFastbootButton := widget.NewButton("Reboot Fastboot", func() {
        go t.adbRebootFastboot()
    })

    rebootRecoveryButton := widget.NewButton("Reboot Recovery", func() {
        go t.adbRebootRecovery()
    })
        diagButton := widget.NewButton("Enable DIAG", func() {
        go t.adbEnableDiag()
    })

    // Create grid layout for ADB buttons
    return container.NewGridWithColumns(6,
        deviceButton,
        infoButton,
        rebootButton,
        rebootFastbootButton,
        rebootRecoveryButton,
        diagButton,
    )
}





func (t *FlashTool) appendLog(message string) {
    currentText := t.logOutput.Text
    if currentText == "" {
        t.logOutput.SetText(message)
    } else {
        t.logOutput.SetText(currentText + "\n" + message)
    }
}

func init() {
    // Hide console window
    console := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")
    if console.Find() == nil {
        showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
        if handle, _, _ := console.Call(); handle != 0 {
            showWindow.Call(handle, 0)
        }
    }
}
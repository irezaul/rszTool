package main

import (
    "syscall"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
)

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
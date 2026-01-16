package main

import (
	"fmt"
	"image/color"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const AppVersion = "1.2.0"

type DevPortsApp struct {
	myApp          fyne.App
	myWindow       fyne.Window
	table          *widget.Table
	refreshBtn     *widget.Button
	statusLbl      *widget.Label
	ports          []PortInfo
	portsMu        sync.RWMutex  // protects ports slice
	isScanning     atomic.Bool
	pendingRefresh atomic.Bool   // prevents multiple refresh goroutines
	quit           chan struct{}
}

func main() {
	// Validate configuration at startup
	if err := AppConfig.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	myApp := app.New()
	myApp.SetIcon(appIcon)

	myWindow := myApp.NewWindow("⚡ DevPorts Pro - Port Scanner")
	myWindow.Resize(fyne.NewSize(AppConfig.WindowWidth, AppConfig.WindowHeight))
	myWindow.CenterOnScreen()

	devApp := &DevPortsApp{
		myApp:    myApp,
		myWindow: myWindow,
		ports:    make([]PortInfo, 0),
		quit:     make(chan struct{}),
	}

	devApp.buildUI(myWindow)

	// Start initial scan
	go devApp.scanPorts()

	// Start auto-refresh timer (5 minutes)
	go devApp.startAutoRefresh()

	// Graceful shutdown handler
	myWindow.SetOnClosed(func() {
		close(devApp.quit)
	})

	myWindow.ShowAndRun()
}

func (da *DevPortsApp) buildUI(w fyne.Window) {
	// Set dark theme for terminal look
	da.myApp.Settings().SetTheme(theme.DarkTheme())

	// Modern terminal-style header with improved colors
	header := canvas.NewText("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", color.RGBA{0, 255, 150, 255})
	header.TextStyle.Monospace = true
	header.Alignment = fyne.TextAlignCenter

	title := canvas.NewText(fmt.Sprintf("⚡ DevPorts Pro v%s - Port Scanner", AppVersion), color.RGBA{0, 255, 200, 255})
	title.TextStyle.Monospace = true
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 18

	subtitle := canvas.NewText("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", color.RGBA{0, 255, 150, 255})
	subtitle.TextStyle.Monospace = true
	subtitle.Alignment = fyne.TextAlignCenter

	// Status label with improved styling
	da.statusLbl = widget.NewLabel("⚡ Ready to scan...")
	da.statusLbl.TextStyle.Monospace = true

	// Refresh button with improved styling
	da.refreshBtn = widget.NewButton("⟳ Refresh Scan", func() {
		if !da.isScanning.Load() {
			go da.scanPorts()
		}
	})
	da.refreshBtn.Importance = widget.HighImportance

	// Create table with compact rows
	da.table = widget.NewTable(
		func() (int, int) {
			da.portsMu.RLock()
			count := len(da.ports)
			da.portsMu.RUnlock()
			return count + 1, 4 // +1 for header
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("template")
			label.Resize(fyne.NewSize(100, 30))
			return container.NewHBox(label)
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			cell, ok := o.(*fyne.Container)
			if !ok {
				return
			}

			if i.Row == 0 {
				// Header row
				var text string
				switch i.Col {
				case 0:
					text = "Port"
				case 1:
					text = "PID"
				case 2:
					text = "Process"
				case 3:
					text = "Action"
				}
				label := widget.NewLabel(text)
				label.TextStyle.Bold = true
				cell.Objects = []fyne.CanvasObject{label}
			} else {
				da.portsMu.RLock()
				if i.Row-1 < len(da.ports) {
					port := da.ports[i.Row-1]
					da.portsMu.RUnlock()
					switch i.Col {
					case 0:
						label := widget.NewLabel(fmt.Sprintf("%d", port.Port))
						cell.Objects = []fyne.CanvasObject{label}
					case 1:
						label := widget.NewLabel(port.PID)
						cell.Objects = []fyne.CanvasObject{label}
					case 2:
						label := widget.NewLabel(port.Process)
						cell.Objects = []fyne.CanvasObject{label}
					case 3:
						if port.PID != "Unknown" && port.PID != "" && port.PID != "Timeout" {
							// Capture values in local variables BEFORE the closure
							pid := port.PID
							portNum := port.Port
							processName := port.Process

							killBtn := widget.NewButton("⨯ Kill", func() {
								da.showKillConfirmation(pid, portNum, processName)
							})
							killBtn.Importance = widget.DangerImportance
							killBtn.Resize(fyne.NewSize(90, 28))
							cell.Objects = []fyne.CanvasObject{killBtn}
						} else {
							label := widget.NewLabel("—")
							label.Alignment = fyne.TextAlignCenter
							cell.Objects = []fyne.CanvasObject{label}
						}
					}
				} else {
					da.portsMu.RUnlock()
					// Clear stale data when row is out of bounds
					cell.Objects = []fyne.CanvasObject{widget.NewLabel("")}
				}
			}
		})

	// Set optimized column widths for 1100px window
	da.table.SetColumnWidth(0, 120) // Port
	da.table.SetColumnWidth(1, 120) // PID
	da.table.SetColumnWidth(2, 550) // Process
	da.table.SetColumnWidth(3, 120) // Action

	// Info banner
	infoText := canvas.NewText(
		fmt.Sprintf("▸ Scanning ports %d-%d | Auto-refresh: %v | Click [Kill] to terminate process",
			AppConfig.PortRangeStart,
			AppConfig.PortRangeEnd,
			AppConfig.AutoRefreshInterval),
		color.RGBA{120, 120, 120, 255})
	infoText.TextStyle.Monospace = true
	infoText.Alignment = fyne.TextAlignCenter
	infoText.TextSize = 11

	// Top controls with terminal style
	topContainer := container.NewHBox(
		da.refreshBtn,
		widget.NewSeparator(),
		da.statusLbl,
	)

	// Header container
	headerContainer := container.NewVBox(
		header,
		title,
		subtitle,
		widget.NewSeparator(),
		infoText,
		widget.NewSeparator(),
	)

	// Footer
	footerText := canvas.NewText(
		fmt.Sprintf("━━━ DevPorts Pro © %d | Press [Refresh Scan] to update ━━━", time.Now().Year()),
		color.RGBA{0, 255, 150, 200})
	footerText.TextStyle.Monospace = true
	footerText.Alignment = fyne.TextAlignCenter
	footerText.TextSize = 10

	// Main content with terminal layout
	content := container.NewBorder(
		container.NewVBox(headerContainer, topContainer), // top
		footerText, // bottom
		nil,        // left
		nil,        // right
		da.table,   // center
	)

	w.SetContent(content)
}

func (da *DevPortsApp) scanPorts() {
	if !da.isScanning.CompareAndSwap(false, true) {
		return // Already scanning
	}
	defer da.isScanning.Store(false)

	da.statusLbl.SetText("⟳ Scanning ports...")
	da.refreshBtn.SetText("⏳ Scanning...")
	da.refreshBtn.Disable()

	// Use optimized concurrent port scanner
	startTime := time.Now()
	activePorts := ScanPorts()
	elapsed := time.Since(startTime)

	da.portsMu.Lock()
	da.ports = activePorts
	da.portsMu.Unlock()
	da.table.Refresh()

	da.statusLbl.SetText(fmt.Sprintf("✓ Scan complete: %d active ports found (%.2fs)", len(activePorts), elapsed.Seconds()))
	da.refreshBtn.SetText("⟳ Refresh Scan")
	da.refreshBtn.Enable()
}

func (da *DevPortsApp) showKillConfirmation(pid string, port int, process string) {
	if pid == "Unknown" || pid == "" {
		return
	}

	// Improved confirmation dialog
	title := "⚠️  Terminate Process"

	var message string
	if process != "Unknown" && process != "" {
		message = fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Process: %s\n"+
			"PID: %s\n"+
			"Port: %d\n\n"+
			"Are you sure you want to terminate this process?\n\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", process, pid, port)
	} else {
		message = fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"PID: %s\n"+
			"Port: %d\n\n"+
			"Are you sure you want to terminate this process?\n\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", pid, port)
	}

	dialog.ShowConfirm(
		title,
		message,
		func(confirmed bool) {
			if confirmed {
				da.executeKill(pid)
			}
		},
		da.myWindow,
	)
}

func (da *DevPortsApp) executeKill(pid string) {
	da.statusLbl.SetText(fmt.Sprintf("⏳ Terminating process PID %s...", pid))

	err := KillProcess(pid)
	if err != nil {
		da.statusLbl.SetText(fmt.Sprintf("✗ Failed to kill PID %s: %v", pid, err))
		// Show error dialog for better user feedback
		dialog.ShowError(fmt.Errorf("process termination failed: %v", err), da.myWindow)
	} else {
		da.statusLbl.SetText(fmt.Sprintf("✓ Process PID %s terminated successfully", pid))
	}

	// Always refresh after kill attempt to show current state
	if da.pendingRefresh.CompareAndSwap(false, true) {
		go func() {
			time.Sleep(AppConfig.PostKillRefreshDelay)
			da.pendingRefresh.Store(false)
			if !da.isScanning.Load() {
				da.scanPorts()
			}
		}()
	}
}

func (da *DevPortsApp) startAutoRefresh() {
	ticker := time.NewTicker(AppConfig.AutoRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !da.isScanning.Load() {
				go da.scanPorts()
			}
		case <-da.quit:
			return // Graceful exit
		}
	}
}

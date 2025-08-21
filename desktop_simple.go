package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type PortModel struct {
	walk.TableModelBase
	ports []PortInfo
}

func NewPortModel() *PortModel {
	m := &PortModel{}
	return m
}

func (m *PortModel) RowCount() int {
	return len(m.ports)
}

func (m *PortModel) Value(row, col int) interface{} {
	port := m.ports[row]
	switch col {
	case 0:
		return port.Port
	case 1:
		return port.PID
	case 2:
		return port.Process
	}
	return ""
}

func (m *PortModel) SetPorts(ports []PortInfo) {
	m.ports = ports
	m.PublishRowsReset()
}

func main() {
	var mainWindow *walk.MainWindow
	var table *walk.TableView
	var statusLabel *walk.Label
	var refreshBtn *walk.PushButton
	
	model := NewPortModel()
	
	// Initial scan
	fmt.Println("Starting DevPorts Pro Desktop...")
	go func() {
		ports := ScanPorts()
		model.SetPorts(ports)
		if statusLabel != nil {
			statusLabel.SetText(fmt.Sprintf("Found %d active ports", len(ports)))
		}
	}()

	err := MainWindow{
		AssignTo: &mainWindow,
		Title:    "DevPorts Pro - Desktop",
		Size:     Size{800, 600},
		Layout:   VBox{},
		Children: []Widget{
			// Header
			Label{
				Text: "🔍 DevPorts Pro - Port Scanner & Process Manager",
				Font: Font{PointSize: 14, Bold: true},
			},
			
			// Controls
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &refreshBtn,
						Text:     "🔄 Refresh Scan",
						OnClicked: func() {
							refreshBtn.SetText("Scanning...")
							refreshBtn.SetEnabled(false)
							go func() {
								ports := ScanPorts()
								model.SetPorts(ports)
								
								// Update UI in main thread
								mainWindow.Synchronize(func() {
									statusLabel.SetText(fmt.Sprintf("Found %d active ports", len(ports)))
									refreshBtn.SetText("🔄 Refresh Scan")
									refreshBtn.SetEnabled(true)
								})
							}()
						},
					},
					HSpacer{},
					Label{
						AssignTo: &statusLabel,
						Text:     "Ready to scan...",
					},
				},
			},
			
			// Table
			TableView{
				AssignTo:         &table,
				Model:           model,
				AlternatingRowBG: true,
				Columns: []TableViewColumn{
					{Title: "Port", Width: 80},
					{Title: "PID", Width: 80},
					{Title: "Process", Width: 300},
				},
				OnItemActivated: func() {
					if table.CurrentIndex() >= 0 && table.CurrentIndex() < len(model.ports) {
						port := model.ports[table.CurrentIndex()]
						if port.PID != "Unknown" && port.PID != "" {
							result := walk.MsgBox(mainWindow, "Kill Process", 
								fmt.Sprintf("Kill process PID %s on port %d?", port.PID, port.Port),
								walk.MsgBoxYesNo|walk.MsgBoxIconQuestion)
							if result == walk.DlgCmdYes {
								err := KillProcess(port.PID)
								if err != nil {
									walk.MsgBox(mainWindow, "Error", 
										fmt.Sprintf("Failed to kill PID %s: %v", port.PID, err),
										walk.MsgBoxOK|walk.MsgBoxIconError)
								} else {
									statusLabel.SetText(fmt.Sprintf("Killed PID %s", port.PID))
									// Refresh after kill
									go func() {
										time.Sleep(time.Second)
										ports := ScanPorts()
										model.SetPorts(ports)
									}()
								}
							}
						}
					}
				},
			},
			
			// Footer
			Label{
				Text: "Double-click a row to kill the process",
				Font: Font{PointSize: 9},
			},
		},
	}.Run()

	if err != nil {
		log.Fatal(err)
	}
}
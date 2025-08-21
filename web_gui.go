package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type WebApp struct {
	ports []PortInfo
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>DevPorts Pro - Web GUI</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .header { background: #2196F3; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .controls { background: white; padding: 15px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .button { background: #2196F3; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; margin-right: 10px; }
        .button:hover { background: #1976D2; }
        .kill-btn { background: #f44336; }
        .kill-btn:hover { background: #d32f2f; }
        .table-container { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); overflow: hidden; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f8f9fa; font-weight: bold; }
        .status { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .status.success { background: #d4edda; color: #155724; }
        .status.error { background: #f8d7da; color: #721c24; }
        .loading { display: none; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔍 DevPorts Pro - Web GUI</h1>
        <p>Port Scanner & Process Manager</p>
    </div>

    <div class="controls">
        <button class="button" onclick="scanPorts()">🔄 Refresh Scan</button>
        <button class="button" onclick="toggleAutoRefresh()">⏰ Auto-refresh (5min)</button>
        <span class="loading" id="loading">Scanning ports...</span>
    </div>

    <div id="status"></div>

    <div class="table-container">
        <table>
            <thead>
                <tr>
                    <th>Port</th>
                    <th>PID</th>
                    <th>Process</th>
                    <th>Action</th>
                </tr>
            </thead>
            <tbody id="ports-table">
                {{range .}}
                <tr>
                    <td>{{.Port}}</td>
                    <td>{{.PID}}</td>
                    <td>{{.Process}}</td>
                    <td>
                        {{if and (ne .PID "Unknown") (ne .PID "")}}
                        <button class="button kill-btn" onclick="killProcess('{{.PID}}', {{.Port}})">💀 Kill</button>
                        {{else}}
                        -
                        {{end}}
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <script>
        let autoRefresh = false;
        let refreshInterval;

        function showStatus(message, type = 'success') {
            const status = document.getElementById('status');
            status.innerHTML = '<div class="status ' + type + '">' + message + '</div>';
            setTimeout(() => status.innerHTML = '', 5000);
        }

        function showLoading(show) {
            document.getElementById('loading').style.display = show ? 'inline' : 'none';
        }

        async function scanPorts() {
            showLoading(true);
            try {
                const response = await fetch('/api/scan');
                const ports = await response.json();
                updateTable(ports);
                showStatus('Found ' + ports.length + ' active ports');
            } catch (error) {
                showStatus('Scan failed: ' + error.message, 'error');
            } finally {
                showLoading(false);
            }
        }

        async function killProcess(pid, port) {
            if (!confirm('Are you sure you want to kill process PID ' + pid + ' on port ' + port + '?')) {
                return;
            }

            try {
                const response = await fetch('/api/kill/' + pid, { method: 'POST' });
                const result = await response.text();
                showStatus('Successfully killed PID ' + pid);
                setTimeout(scanPorts, 1000); // Refresh after 1 second
            } catch (error) {
                showStatus('Failed to kill PID ' + pid + ': ' + error.message, 'error');
            }
        }

        function updateTable(ports) {
            const tbody = document.getElementById('ports-table');
            tbody.innerHTML = '';
            
            ports.forEach(port => {
                const row = tbody.insertRow();
                row.innerHTML = 
                    '<td>' + port.Port + '</td>' +
                    '<td>' + port.PID + '</td>' +
                    '<td>' + port.Process + '</td>' +
                    '<td>' + (port.PID !== 'Unknown' && port.PID !== '' ? 
                        '<button class="button kill-btn" onclick="killProcess(\'' + port.PID + '\', ' + port.Port + ')">💀 Kill</button>' : 
                        '-') + '</td>';
            });
        }

        function toggleAutoRefresh() {
            autoRefresh = !autoRefresh;
            if (autoRefresh) {
                refreshInterval = setInterval(scanPorts, 5 * 60 * 1000); // 5 minutes
                showStatus('Auto-refresh enabled (every 5 minutes)');
            } else {
                clearInterval(refreshInterval);
                showStatus('Auto-refresh disabled');
            }
        }

        // Initial scan
        scanPorts();
    </script>
</body>
</html>
`

func (wa *WebApp) handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("home").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, wa.ports)
}

func (wa *WebApp) handleScan(w http.ResponseWriter, r *http.Request) {
	wa.ports = ScanPorts()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wa.ports)
}

func (wa *WebApp) handleKill(w http.ResponseWriter, r *http.Request) {
	pid := r.URL.Path[len("/api/kill/"):]
	err := KillProcess(pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Process %s killed successfully", pid)
}

func main() {
	app := &WebApp{}
	
	// Initial scan
	fmt.Println("DevPorts Pro - Web GUI")
	fmt.Println("======================")
	fmt.Println("Starting initial port scan...")
	app.ports = ScanPorts()
	
	// Setup routes
	http.HandleFunc("/", app.handleHome)
	http.HandleFunc("/api/scan", app.handleScan)
	http.HandleFunc("/api/kill/", app.handleKill)
	
	// Start server
	fmt.Println("Starting web server on http://localhost:8070")
	fmt.Println("Open your browser and go to: http://localhost:8070")
	fmt.Println("Press Ctrl+C to stop")
	
	log.Fatal(http.ListenAndServe(":8070", nil))
}
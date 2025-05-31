// Dashboard JavaScript functionality
class WebhookBridgeDashboard {
    constructor() {
        this.eventSource = null;
        this.currentSection = 'dashboard';
        this.refreshInterval = null;
        this.init();
    }

    init() {
        this.setupNavigation();
        this.setupEventSource();
        this.loadInitialData();
        this.startAutoRefresh();
    }

    setupNavigation() {
        // Handle sidebar navigation
        document.querySelectorAll('.sidebar .nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = link.getAttribute('data-section');
                this.showSection(section);
                
                // Update active state
                document.querySelectorAll('.sidebar .nav-link').forEach(l => l.classList.remove('active'));
                link.classList.add('active');
            });
        });
    }

    showSection(section) {
        this.currentSection = section;
        
        // Hide all sections
        document.querySelectorAll('.content-section').forEach(s => s.classList.add('d-none'));
        
        // Show selected section
        const sectionElement = document.getElementById(`${section}-section`);
        if (sectionElement) {
            sectionElement.classList.remove('d-none');
        }
        
        // Update page title
        const titles = {
            'dashboard': 'Dashboard',
            'logs': 'Logs',
            'plugins': 'Plugins',
            'stats': 'Statistics',
            'system': 'System Information'
        };
        document.getElementById('page-title').textContent = titles[section] || section;
        
        // Load section-specific data
        this.loadSectionData(section);
    }

    setupEventSource() {
        // Setup Server-Sent Events for real-time logs
        if (this.eventSource) {
            this.eventSource.close();
        }
        
        this.eventSource = new EventSource('/dashboard/logs/stream');
        
        this.eventSource.onmessage = (event) => {
            const logEntry = JSON.parse(event.data);
            this.addLogEntry(logEntry);
        };
        
        this.eventSource.onerror = (error) => {
            console.error('EventSource failed:', error);
            // Attempt to reconnect after 5 seconds
            setTimeout(() => this.setupEventSource(), 5000);
        };
    }

    loadInitialData() {
        this.loadStats();
        this.loadRecentLogs();
        this.loadTopPlugins();
    }

    startAutoRefresh() {
        // Refresh data every 30 seconds
        this.refreshInterval = setInterval(() => {
            if (this.currentSection === 'dashboard') {
                this.loadStats();
                this.loadTopPlugins();
            }
        }, 30000);
    }

    async loadStats() {
        try {
            const response = await fetch('/api/dashboard/stats');
            const data = await response.json();
            
            this.updateStatsDisplay(data);
        } catch (error) {
            console.error('Failed to load stats:', error);
        }
    }

    updateStatsDisplay(stats) {
        const elements = {
            'total-requests': stats.system?.total_requests || 0,
            'total-executions': stats.system?.total_executions || 0,
            'total-errors': stats.system?.total_errors || 0,
            'memory-usage': Math.round(stats.system?.memory_usage_mb || 0)
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) {
                element.textContent = value.toLocaleString();
            }
        });
    }

    async loadRecentLogs() {
        try {
            const response = await fetch('/api/dashboard/logs?limit=10');
            const data = await response.json();
            
            const container = document.getElementById('recent-logs');
            if (data.logs && data.logs.length > 0) {
                container.innerHTML = data.logs.map(log => this.formatLogEntry(log)).join('');
            } else {
                container.innerHTML = '<div class="text-muted">No recent logs</div>';
            }
        } catch (error) {
            console.error('Failed to load recent logs:', error);
        }
    }

    async loadTopPlugins() {
        try {
            const response = await fetch('/api/dashboard/plugins');
            const data = await response.json();
            
            const container = document.getElementById('top-plugins');
            if (data.plugins && Object.keys(data.plugins).length > 0) {
                const sortedPlugins = Object.values(data.plugins)
                    .sort((a, b) => b.count - a.count)
                    .slice(0, 5);
                
                container.innerHTML = sortedPlugins.map(plugin => `
                    <div class="d-flex justify-content-between align-items-center mb-2">
                        <div>
                            <strong>${plugin.plugin}</strong>
                            <small class="text-muted d-block">${plugin.method}</small>
                        </div>
                        <span class="badge bg-primary">${plugin.count}</span>
                    </div>
                `).join('');
            } else {
                container.innerHTML = '<div class="text-muted">No plugin data</div>';
            }
        } catch (error) {
            console.error('Failed to load top plugins:', error);
        }
    }

    loadSectionData(section) {
        switch (section) {
            case 'logs':
                this.loadAllLogs();
                break;
            case 'plugins':
                this.loadPluginDetails();
                break;
            case 'stats':
                this.loadDetailedStats();
                break;
            case 'system':
                this.loadSystemInfo();
                break;
        }
    }

    async loadAllLogs() {
        try {
            const response = await fetch('/api/dashboard/logs?limit=100');
            const data = await response.json();
            
            const container = document.getElementById('logs-container');
            if (data.logs && data.logs.length > 0) {
                container.innerHTML = data.logs.map(log => this.formatLogEntry(log, true)).join('');
            } else {
                container.innerHTML = '<div class="text-muted">No logs available</div>';
            }
        } catch (error) {
            console.error('Failed to load logs:', error);
        }
    }

    formatLogEntry(log, detailed = false) {
        const timestamp = new Date(log.timestamp).toLocaleString();
        const levelClass = log.level.toLowerCase();
        
        let content = `
            <div class="log-entry ${levelClass}">
                <div class="d-flex justify-content-between align-items-start">
                    <div class="flex-grow-1">
                        <div class="fw-bold">${log.message}</div>
                        ${detailed ? `<small class="text-muted">Source: ${log.source}</small>` : ''}
                    </div>
                    <div class="text-end">
                        <span class="badge bg-${this.getLevelColor(log.level)}">${log.level}</span>
                        <small class="text-muted d-block">${timestamp}</small>
                    </div>
                </div>
        `;
        
        if (detailed && log.data) {
            content += `
                <div class="mt-2">
                    <small class="text-muted">Data:</small>
                    <pre class="small mt-1 mb-0">${JSON.stringify(log.data, null, 2)}</pre>
                </div>
            `;
        }
        
        content += '</div>';
        return content;
    }

    getLevelColor(level) {
        const colors = {
            'DEBUG': 'secondary',
            'INFO': 'info',
            'WARN': 'warning',
            'ERROR': 'danger'
        };
        return colors[level] || 'secondary';
    }

    addLogEntry(logEntry) {
        if (this.currentSection === 'logs') {
            const container = document.getElementById('logs-container');
            const newEntry = this.formatLogEntry(logEntry, true);
            container.insertAdjacentHTML('afterbegin', newEntry);
            
            // Remove old entries if too many
            const entries = container.querySelectorAll('.log-entry');
            if (entries.length > 100) {
                entries[entries.length - 1].remove();
            }
        }
        
        // Update recent logs on dashboard
        if (this.currentSection === 'dashboard') {
            const recentContainer = document.getElementById('recent-logs');
            const newEntry = this.formatLogEntry(logEntry);
            recentContainer.insertAdjacentHTML('afterbegin', newEntry);
            
            // Keep only 10 recent entries
            const entries = recentContainer.querySelectorAll('.log-entry');
            if (entries.length > 10) {
                entries[entries.length - 1].remove();
            }
        }
    }

    async clearLogs() {
        if (confirm('Are you sure you want to clear all logs?')) {
            try {
                await fetch('/dashboard/logs', { method: 'DELETE' });
                document.getElementById('logs-container').innerHTML = '<div class="text-muted">Logs cleared</div>';
                document.getElementById('recent-logs').innerHTML = '<div class="text-muted">No recent logs</div>';
            } catch (error) {
                console.error('Failed to clear logs:', error);
                alert('Failed to clear logs');
            }
        }
    }

    async loadPluginDetails() {
        // Implementation for detailed plugin view
        const container = document.getElementById('plugins-container');
        container.innerHTML = '<div class="text-center"><i class="fas fa-spinner fa-spin"></i> Loading plugins...</div>';
        
        try {
            const response = await fetch('/api/v1/plugins');
            const data = await response.json();
            
            if (data.plugins && data.plugins.length > 0) {
                container.innerHTML = data.plugins.map(plugin => `
                    <div class="card mb-3">
                        <div class="card-body">
                            <h6 class="card-title">${plugin.name}</h6>
                            <p class="card-text">${plugin.description}</p>
                            <div class="d-flex justify-content-between align-items-center">
                                <small class="text-muted">Methods: ${plugin.supported_methods.join(', ')}</small>
                                <button class="btn btn-sm btn-primary" onclick="testPlugin('${plugin.name}')">
                                    <i class="fas fa-play"></i> Test
                                </button>
                            </div>
                        </div>
                    </div>
                `).join('');
            } else {
                container.innerHTML = '<div class="text-muted">No plugins available</div>';
            }
        } catch (error) {
            console.error('Failed to load plugins:', error);
            container.innerHTML = '<div class="text-danger">Failed to load plugins</div>';
        }
    }

    async loadDetailedStats() {
        // Implementation for detailed statistics view
        const container = document.getElementById('stats-container');
        container.innerHTML = '<div class="text-center"><i class="fas fa-spinner fa-spin"></i> Loading statistics...</div>';
        
        try {
            const response = await fetch('/api/dashboard/stats');
            const data = await response.json();
            
            container.innerHTML = `
                <div class="row">
                    <div class="col-md-6">
                        <h6>System Metrics</h6>
                        <table class="table table-sm">
                            <tr><td>Uptime</td><td>${data.system?.uptime || 'N/A'}</td></tr>
                            <tr><td>Total Requests</td><td>${data.system?.total_requests || 0}</td></tr>
                            <tr><td>Total Executions</td><td>${data.system?.total_executions || 0}</td></tr>
                            <tr><td>Error Rate</td><td>${(data.system?.error_rate || 0).toFixed(2)}%</td></tr>
                            <tr><td>Requests/sec</td><td>${(data.system?.requests_per_sec || 0).toFixed(2)}</td></tr>
                            <tr><td>Memory Usage</td><td>${Math.round(data.system?.memory_usage_mb || 0)} MB</td></tr>
                            <tr><td>Goroutines</td><td>${data.system?.goroutines || 0}</td></tr>
                        </table>
                    </div>
                    <div class="col-md-6">
                        <h6>Plugin Statistics</h6>
                        <div id="plugin-stats-list">
                            ${Object.values(data.plugins || {}).map(plugin => `
                                <div class="mb-2">
                                    <strong>${plugin.plugin}</strong> (${plugin.method})
                                    <div class="progress" style="height: 20px;">
                                        <div class="progress-bar" style="width: ${Math.min(plugin.count / 10 * 100, 100)}%">
                                            ${plugin.count} executions
                                        </div>
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                </div>
            `;
        } catch (error) {
            console.error('Failed to load detailed stats:', error);
            container.innerHTML = '<div class="text-danger">Failed to load statistics</div>';
        }
    }

    async loadSystemInfo() {
        // Implementation for system information view
        const container = document.getElementById('system-container');
        container.innerHTML = '<div class="text-center"><i class="fas fa-spinner fa-spin"></i> Loading system information...</div>';
        
        try {
            const response = await fetch('/api/dashboard/system');
            const data = await response.json();
            
            container.innerHTML = `
                <div class="row">
                    <div class="col-md-4">
                        <h6>Server Configuration</h6>
                        <table class="table table-sm">
                            <tr><td>Address</td><td>${data.server?.address || 'N/A'}</td></tr>
                            <tr><td>Mode</td><td>${data.server?.mode || 'N/A'}</td></tr>
                            <tr><td>Uptime</td><td>${data.server?.uptime || 'N/A'}</td></tr>
                        </table>
                    </div>
                    <div class="col-md-4">
                        <h6>Executor Configuration</h6>
                        <table class="table table-sm">
                            <tr><td>Address</td><td>${data.executor?.address || 'N/A'}</td></tr>
                            <tr><td>Timeout</td><td>${data.executor?.timeout || 'N/A'}s</td></tr>
                        </table>
                    </div>
                    <div class="col-md-4">
                        <h6>Logging Configuration</h6>
                        <table class="table table-sm">
                            <tr><td>Level</td><td>${data.logging?.level || 'N/A'}</td></tr>
                            <tr><td>Format</td><td>${data.logging?.format || 'N/A'}</td></tr>
                        </table>
                    </div>
                </div>
            `;
        } catch (error) {
            console.error('Failed to load system info:', error);
            container.innerHTML = '<div class="text-danger">Failed to load system information</div>';
        }
    }
}

// Global functions
function refreshData() {
    if (window.dashboard) {
        window.dashboard.loadSectionData(window.dashboard.currentSection);
        if (window.dashboard.currentSection === 'dashboard') {
            window.dashboard.loadInitialData();
        }
    }
}

function testPlugin(pluginName) {
    const payload = prompt(`Enter JSON payload for ${pluginName}:`, '{"test": "data"}');
    if (payload) {
        try {
            const data = JSON.parse(payload);
            fetch(`/dashboard/plugins/${pluginName}/execute`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            })
            .then(response => response.json())
            .then(result => {
                alert(`Plugin executed successfully!\nResult: ${JSON.stringify(result, null, 2)}`);
            })
            .catch(error => {
                alert(`Plugin execution failed: ${error.message}`);
            });
        } catch (error) {
            alert('Invalid JSON payload');
        }
    }
}

function clearLogs() {
    if (window.dashboard) {
        window.dashboard.clearLogs();
    }
}

// Initialize dashboard when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new WebhookBridgeDashboard();
});

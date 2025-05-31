// Modern Dashboard JavaScript - Simplified and clean

class ModernDashboard {
    constructor() {
        this.currentSection = 'dashboard';
        this.refreshInterval = null;
        this.init();
    }

    init() {
        // Initialize Lucide icons
        if (typeof lucide !== 'undefined') {
            lucide.createIcons();
        }

        // Set up auto-refresh
        this.setupAutoRefresh();
        
        // Load initial data
        this.loadDashboardData();
    }

    setupAutoRefresh() {
        // Auto-refresh every 30 seconds for dashboard
        this.refreshInterval = setInterval(() => {
            if (this.currentSection === 'dashboard') {
                this.loadDashboardData();
            }
        }, 30000);
    }

    showSection(sectionName) {
        // Hide all sections
        document.querySelectorAll('.section').forEach(section => {
            section.classList.add('hidden');
        });

        // Show selected section
        const targetSection = document.getElementById(sectionName + '-section');
        if (targetSection) {
            targetSection.classList.remove('hidden');
        }

        // Update navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active', 'bg-accent', 'text-accent-foreground');
        });
        
        // Find and activate the clicked nav item
        const activeNavItem = document.querySelector(`[onclick="showSection('${sectionName}')"]`);
        if (activeNavItem) {
            activeNavItem.classList.add('active', 'bg-accent', 'text-accent-foreground');
        }

        // Update page title
        const titles = {
            'dashboard': 'Dashboard',
            'plugins': 'Plugins',
            'workers': 'Workers',
            'logs': 'Logs',
            'config': 'Configuration'
        };
        
        const titleElement = document.getElementById('page-title');
        if (titleElement && titles[sectionName]) {
            titleElement.textContent = titles[sectionName];
        }

        // Update current section
        this.currentSection = sectionName;

        // Load section data
        this.loadSectionData(sectionName);
    }

    async loadSectionData(section) {
        try {
            switch(section) {
                case 'dashboard':
                    await this.loadDashboardData();
                    break;
                case 'plugins':
                    await this.loadPluginsData();
                    break;
                case 'workers':
                    await this.loadWorkersData();
                    break;
                case 'logs':
                    await this.loadLogsData();
                    break;
                case 'config':
                    await this.loadConfigData();
                    break;
            }
        } catch (error) {
            console.error('Failed to load section data:', error);
            this.showError(`Failed to load ${section} data`);
        }
    }

    async loadDashboardData() {
        try {
            const [statusResponse, metricsResponse] = await Promise.all([
                fetch('/api/dashboard/status'),
                fetch('/api/dashboard/metrics')
            ]);

            if (!statusResponse.ok || !metricsResponse.ok) {
                throw new Error('Failed to fetch dashboard data');
            }

            const status = await statusResponse.json();
            const metrics = await metricsResponse.json();

            // Update stats cards
            this.updateElement('total-requests', metrics.requests?.total?.toLocaleString() || '0');
            this.updateElement('success-rate', this.calculateSuccessRate(metrics.requests) + '%');
            this.updateElement('avg-response', metrics.performance?.avg_response_time || '0ms');
            this.updateElement('active-workers', metrics.workers?.active || '0');

            // Load recent activity
            await this.loadRecentActivity();
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            this.showError('Failed to load dashboard data');
        }
    }

    calculateSuccessRate(requests) {
        if (!requests || !requests.total || requests.total === 0) {
            return '0.0';
        }
        return ((requests.success / requests.total) * 100).toFixed(1);
    }

    async loadRecentActivity() {
        try {
            const response = await fetch('/api/dashboard/logs');
            if (!response.ok) {
                throw new Error('Failed to fetch logs');
            }
            
            const data = await response.json();
            const activityContainer = document.getElementById('recent-activity');
            
            if (!activityContainer) return;

            const activityHtml = data.logs?.slice(0, 5).map(log => {
                const levelColors = {
                    'info': 'text-blue-500',
                    'warn': 'text-yellow-500',
                    'error': 'text-red-500',
                    'debug': 'text-gray-500'
                };

                return `
                    <div class="flex items-start space-x-3 p-3 rounded-lg bg-muted/50">
                        <div class="w-2 h-2 rounded-full ${levelColors[log.level] || 'text-gray-500'} mt-2"></div>
                        <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium">${this.escapeHtml(log.message)}</p>
                            <p class="text-xs text-muted-foreground">${new Date(log.timestamp).toLocaleString()}</p>
                        </div>
                    </div>
                `;
            }).join('') || '<p class="text-muted-foreground">No recent activity</p>';

            activityContainer.innerHTML = activityHtml;
        } catch (error) {
            console.error('Failed to load recent activity:', error);
        }
    }

    async loadPluginsData() {
        try {
            const response = await fetch('/api/dashboard/plugins');
            if (!response.ok) {
                throw new Error('Failed to fetch plugins');
            }
            
            const data = await response.json();
            const pluginsContainer = document.getElementById('plugins-list');
            
            if (!pluginsContainer) return;

            const pluginsHtml = data.plugins?.map(plugin => {
                const statusColor = plugin.status === 'active' ? 'bg-green-500' : 'bg-gray-500';

                return `
                    <div class="border border-border rounded-lg p-4 mb-4">
                        <div class="flex items-center justify-between mb-2">
                            <h4 class="font-semibold">${this.escapeHtml(plugin.name)}</h4>
                            <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium">
                                <div class="w-2 h-2 ${statusColor} rounded-full mr-1"></div>
                                ${this.escapeHtml(plugin.status)}
                            </span>
                        </div>
                        <p class="text-sm text-muted-foreground mb-2">${this.escapeHtml(plugin.description)}</p>
                        <div class="flex items-center justify-between text-xs text-muted-foreground">
                            <span>Version: ${this.escapeHtml(plugin.version)}</span>
                            <span>Last used: ${new Date(plugin.last_used).toLocaleString()}</span>
                        </div>
                    </div>
                `;
            }).join('') || '<p class="text-muted-foreground">No plugins available</p>';

            pluginsContainer.innerHTML = pluginsHtml;
        } catch (error) {
            console.error('Failed to load plugins data:', error);
            this.showError('Failed to load plugins data');
        }
    }

    async loadWorkersData() {
        try {
            const response = await fetch('/api/dashboard/workers');
            if (!response.ok) {
                throw new Error('Failed to fetch workers');
            }
            
            const data = await response.json();
            const workersContainer = document.getElementById('workers-list');
            
            if (!workersContainer) return;

            const workersHtml = `
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                    <div class="bg-muted/50 rounded-lg p-4">
                        <p class="text-sm text-muted-foreground">Total Workers</p>
                        <p class="text-2xl font-bold">${data.pool?.workers || 0}</p>
                    </div>
                    <div class="bg-muted/50 rounded-lg p-4">
                        <p class="text-sm text-muted-foreground">Queue Size</p>
                        <p class="text-2xl font-bold">${data.pool?.queue_size || 0}</p>
                    </div>
                    <div class="bg-muted/50 rounded-lg p-4">
                        <p class="text-sm text-muted-foreground">Completed Jobs</p>
                        <p class="text-2xl font-bold">${data.pool?.completed_jobs || 0}</p>
                    </div>
                    <div class="bg-muted/50 rounded-lg p-4">
                        <p class="text-sm text-muted-foreground">Failed Jobs</p>
                        <p class="text-2xl font-bold">${data.pool?.failed_jobs || 0}</p>
                    </div>
                </div>
                <div class="space-y-4">
                    ${data.workers?.map(worker => `
                        <div class="border border-border rounded-lg p-4">
                            <div class="flex items-center justify-between">
                                <h4 class="font-semibold">Worker ${worker.id}</h4>
                                <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium">
                                    <div class="w-2 h-2 bg-green-500 rounded-full mr-1"></div>
                                    ${this.escapeHtml(worker.status)}
                                </span>
                            </div>
                            <div class="mt-2 text-sm text-muted-foreground">
                                <p>Jobs processed: ${worker.jobs_processed || 0}</p>
                                <p>Last activity: ${new Date(worker.last_activity).toLocaleString()}</p>
                            </div>
                        </div>
                    `).join('') || '<p class="text-muted-foreground">No workers available</p>'}
                </div>
            `;

            workersContainer.innerHTML = workersHtml;
        } catch (error) {
            console.error('Failed to load workers data:', error);
            this.showError('Failed to load workers data');
        }
    }

    async loadLogsData() {
        try {
            const response = await fetch('/api/dashboard/logs');
            if (!response.ok) {
                throw new Error('Failed to fetch logs');
            }
            
            const data = await response.json();
            const logsContainer = document.getElementById('logs-list');
            
            if (!logsContainer) return;

            const logsHtml = data.logs?.map(log => {
                const levelColors = {
                    'info': 'text-blue-500',
                    'warn': 'text-yellow-500',
                    'error': 'text-red-500',
                    'debug': 'text-gray-500'
                };

                return `
                    <div class="border border-border rounded-lg p-4 mb-2">
                        <div class="flex items-center justify-between mb-1">
                            <span class="text-xs font-medium ${levelColors[log.level] || 'text-gray-500'} uppercase">${this.escapeHtml(log.level)}</span>
                            <span class="text-xs text-muted-foreground">${new Date(log.timestamp).toLocaleString()}</span>
                        </div>
                        <p class="text-sm">${this.escapeHtml(log.message)}</p>
                        ${log.plugin ? `<p class="text-xs text-muted-foreground mt-1">Plugin: ${this.escapeHtml(log.plugin)}</p>` : ''}
                        ${log.component ? `<p class="text-xs text-muted-foreground mt-1">Component: ${this.escapeHtml(log.component)}</p>` : ''}
                    </div>
                `;
            }).join('') || '<p class="text-muted-foreground">No logs available</p>';

            logsContainer.innerHTML = logsHtml;
        } catch (error) {
            console.error('Failed to load logs data:', error);
            this.showError('Failed to load logs data');
        }
    }

    async loadConfigData() {
        try {
            const response = await fetch('/api/dashboard/config');
            if (!response.ok) {
                throw new Error('Failed to fetch config');
            }
            
            const config = await response.json();
            const configContainer = document.getElementById('config-display');
            
            if (!configContainer) return;

            const configHtml = `
                <div class="space-y-6">
                    <div>
                        <h4 class="font-semibold mb-3">Server Configuration</h4>
                        <div class="bg-muted/50 rounded-lg p-4">
                            <pre class="text-sm">${JSON.stringify(config.server, null, 2)}</pre>
                        </div>
                    </div>
                    <div>
                        <h4 class="font-semibold mb-3">Python Configuration</h4>
                        <div class="bg-muted/50 rounded-lg p-4">
                            <pre class="text-sm">${JSON.stringify(config.python, null, 2)}</pre>
                        </div>
                    </div>
                    <div>
                        <h4 class="font-semibold mb-3">Logging Configuration</h4>
                        <div class="bg-muted/50 rounded-lg p-4">
                            <pre class="text-sm">${JSON.stringify(config.logging, null, 2)}</pre>
                        </div>
                    </div>
                    <div>
                        <h4 class="font-semibold mb-3">Directories Configuration</h4>
                        <div class="bg-muted/50 rounded-lg p-4">
                            <pre class="text-sm">${JSON.stringify(config.directories, null, 2)}</pre>
                        </div>
                    </div>
                </div>
            `;

            configContainer.innerHTML = configHtml;
        } catch (error) {
            console.error('Failed to load config data:', error);
            this.showError('Failed to load config data');
        }
    }

    updateElement(id, value) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = value;
        }
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    showError(message) {
        console.error(message);
        // You could implement a toast notification system here
    }

    refreshData() {
        this.loadSectionData(this.currentSection);
    }
}

// Global functions for HTML onclick handlers
function showSection(sectionName) {
    if (window.dashboard) {
        window.dashboard.showSection(sectionName);
    }
}

function refreshData() {
    if (window.dashboard) {
        window.dashboard.refreshData();
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new ModernDashboard();
});

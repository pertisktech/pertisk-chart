// Pertisk Chart Repository - Frontend Application
const API_BASE = '/api';

// State management
const state = {
    charts: [],
    filteredCharts: [],
    currentView: 'home',
    currentChart: null,
    searchQuery: '',
    sortBy: 'name',
    token: localStorage.getItem('auth_token'),
    user: null
};

// Initialize application
document.addEventListener('DOMContentLoaded', () => {
    initializeApp();
});

async function initializeApp() {
    initializeTheme();
    setupEventListeners();
    await checkAuth();
    
    // Load domain and repository configuration first (non-blocking)
    const configPromise = (async () => {
        try {
            const domainResponse = await fetch('/api/config/domain');
            if (domainResponse.ok) {
                const data = await domainResponse.json();
                if (data.domain) {
                    repositoryDomain = data.domain;
                }
            }
            
            // Try to load repository name from admin config (only if user is admin)
            if (state.token && state.user && state.user.is_admin) {
                try {
                    const configResponse = await fetch('/api/admin/config', {
                        headers: {
                            'Authorization': `Bearer ${state.token}`
                        }
                    });
                    if (configResponse.ok) {
                        const config = await configResponse.json();
                        if (config.repository_name) {
                            repositoryName = config.repository_name;
                        }
                    } else if (configResponse.status === 403) {
                        // User is not admin, silently ignore
                    }
                } catch (err) {
                    // Silently fail if not admin or config not available
                }
            }
        } catch (error) {
            console.error('Failed to load configuration:', error);
        }
    })();
    
    // Load charts (critical for routing)
    await loadCharts();
    updateStats();
    
    // Wait for config to complete (non-blocking, but good to have)
    await configPromise;
    
    // Initialize routing - read current URL and show appropriate view
    // Charts are already loaded above, so routing should work
    const currentPath = window.location.pathname;
    console.log('Initializing app, current path:', currentPath);
    const route = getRouteFromPath(currentPath);
    console.log('Parsed route:', route);
    handleRoute(route, false); // Don't update history on initial load
    
    updateAuthUI();
    
    // Listen for browser back/forward buttons
    window.addEventListener('popstate', (event) => {
        if (event.state) {
            handleRoute({ view: event.state.view, params: event.state.params || {} }, false);
        } else {
            // Fallback: parse current URL
            const path = window.location.pathname;
            const route = getRouteFromPath(path);
            handleRoute(route, false);
        }
    });
}

// Event Listeners
function setupEventListeners() {
    // Search
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('input', handleSearch);
    }

    // Initialize modern dropdowns
    initializeModernDropdowns();

    // Sort (hidden select for compatibility)
    const sortSelect = document.getElementById('sortSelect');
    if (sortSelect) {
        sortSelect.addEventListener('change', (e) => {
            state.sortBy = e.target.value;
            applyFilters();
        });
    }

    // Management search and sort
    const managementSearch = document.getElementById('managementSearch');
    const managementSort = document.getElementById('managementSort');
    
    if (managementSearch) {
        managementSearch.addEventListener('input', () => {
            if (state.currentView === 'myCharts') {
                renderMyCharts();
            }
        });
    }
    
    if (managementSort) {
        managementSort.addEventListener('change', () => {
            if (state.currentView === 'myCharts') {
                renderMyCharts();
            }
        });
    }

    // Theme toggle
    const themeToggle = document.getElementById('themeToggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }

    // Upload button
    const uploadBtn = document.getElementById('uploadBtn');
    if (uploadBtn) {
        uploadBtn.addEventListener('click', () => {
            if (!state.token) {
                showModal('loginModal');
                return;
            }
            showModal('uploadModal');
        });
    }

    // Modal controls
    const closeModal = document.getElementById('closeModal');
    const cancelUpload = document.getElementById('cancelUpload');
    if (closeModal) {
        closeModal.addEventListener('click', () => hideModal('uploadModal'));
    }
    if (cancelUpload) {
        cancelUpload.addEventListener('click', () => hideModal('uploadModal'));
    }

    // Upload form
    const uploadForm = document.getElementById('uploadForm');
    if (uploadForm) {
        uploadForm.addEventListener('submit', handleUpload);
    }

    // Auth buttons
    const loginBtn = document.getElementById('loginBtn');
    const registerBtn = document.getElementById('registerBtn');
    const logoutBtn = document.getElementById('logoutBtn');
    
    if (loginBtn) {
        loginBtn.addEventListener('click', () => showModal('loginModal'));
    }
    if (registerBtn) {
        registerBtn.addEventListener('click', () => showModal('registerModal'));
    }
    if (logoutBtn) {
        logoutBtn.addEventListener('click', handleLogout);
    }

    // Login form
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    // Register form
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', handleRegister);
    }

    // Modal close handlers
    setupModalClose('loginModal', 'closeLoginModal', 'cancelLogin');
    setupModalClose('registerModal', 'closeRegisterModal', 'cancelRegister');
    setupModalClose('uploadModal', 'closeModal', 'cancelUpload');
}

function setupModalClose(modalId, closeBtnId, cancelBtnId) {
    const modal = document.getElementById(modalId);
    const closeBtn = document.getElementById(closeBtnId);
    const cancelBtn = document.getElementById(cancelBtnId);
    
    if (modal) {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                hideModal(modalId);
            }
        });
    }
    if (closeBtn) {
        closeBtn.addEventListener('click', () => hideModal(modalId));
    }
    if (cancelBtn) {
        cancelBtn.addEventListener('click', () => hideModal(modalId));
    }
}

// Load charts from API
async function loadCharts() {
    try {
        const response = await fetch(`${API_BASE}/charts`);
        if (!response.ok) throw new Error('Failed to load charts');
        
        const data = await response.json();
        state.charts = data;
        state.filteredCharts = data;
        renderCharts();
    } catch (error) {
        console.error('Error loading charts:', error);
        showError('Failed to load charts. Please refresh the page.');
    }
}

// Render charts grid
function renderCharts() {
    const grid = document.getElementById('chartsGrid');
    const emptyState = document.getElementById('emptyState');
    
    if (!grid) return;

    if (state.filteredCharts.length === 0) {
        grid.style.display = 'none';
        if (emptyState) emptyState.style.display = 'block';
        return;
    }

    if (emptyState) emptyState.style.display = 'none';
    grid.style.display = 'grid';

    grid.innerHTML = state.filteredCharts.map(chart => createChartCard(chart)).join('');
    
    // Add click listeners to cards
    grid.querySelectorAll('.chart-card').forEach(card => {
        card.addEventListener('click', (e) => {
            // Don't navigate if clicking on a link or button
            if (e.target.tagName === 'A' || e.target.closest('a') || e.target.closest('.chart-actions')) {
                return;
            }
            const chartName = card.dataset.chartName;
            if (chartName) {
                showChartDetail(chartName);
            }
        });
    });
}

// Create chart card HTML
function createChartCard(chart) {
    const latestVersion = chart.versions && chart.versions.length > 0 ? chart.versions[0] : null;
    const versionCount = chart.versions ? chart.versions.length : 0;
    
    return `
        <div class="chart-card" data-chart-name="${chart.name}">
            <div class="chart-header">
                ${chart.icon ? `<img src="${chart.icon}" alt="${chart.name}" class="chart-icon" onerror="this.style.display='none'">` : ''}
                <div style="flex: 1;">
                    <h3 class="chart-title">
                        <a href="/charts/${encodeURIComponent(chart.name)}" onclick="event.preventDefault(); showChartDetail('${chart.name}'); return false;">${escapeHtml(chart.name)}</a>
                    </h3>
                    <p class="chart-description">${escapeHtml(chart.description || 'No description available')}</p>
                </div>
            </div>
            <div class="chart-meta">
                <div class="chart-meta-item">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
                    </svg>
                    <span>${versionCount} version${versionCount !== 1 ? 's' : ''}</span>
                </div>
                ${latestVersion ? `
                <div class="chart-meta-item">
                    <span class="version-badge">${escapeHtml(latestVersion.version)}</span>
                </div>
                ` : ''}
            </div>
        </div>
    `;
}

// Show chart detail view
async function showChartDetail(chartName, shouldUpdateURL = true) {
    const chart = state.charts.find(c => c.name === chartName);
    if (!chart) {
        showError('Chart not found');
        // Redirect to charts view if chart not found
        if (shouldUpdateURL) {
            showView('charts');
        }
        return;
    }

    // Load domain configuration
    try {
        const response = await fetch('/api/config/domain');
        if (response.ok) {
            const data = await response.json();
            if (data.domain) {
                repositoryDomain = data.domain;
            }
        }
    } catch (error) {
        console.error('Failed to load domain:', error);
    }

    state.currentChart = chart;
    state.currentView = 'chartDetail';
    
    // Update URL if needed
    if (shouldUpdateURL) {
        updateURL('chartDetail', { chartName });
    }
    
    // Show the detail view container
    const detailView = document.getElementById('chartDetailView');
    if (!detailView) return;
    
    // Hide other views
    const homeView = document.getElementById('homeView');
    const chartsView = document.getElementById('chartsView');
    const myChartsView = document.getElementById('myChartsView');
    const adminView = document.getElementById('adminView');
    
    if (homeView) homeView.style.display = 'none';
    if (chartsView) chartsView.style.display = 'none';
    if (myChartsView) myChartsView.style.display = 'none';
    if (adminView) adminView.style.display = 'none';
    if (detailView) detailView.style.display = 'block';

    // Render the chart detail content
    renderChartDetailContent(chart);
}

// Load chart detail asynchronously (waits for charts to load if needed)
async function loadChartDetailAsync(chartName) {
    const detailView = document.getElementById('chartDetailView');
    if (!detailView) {
        console.error('Chart detail view element not found');
        return;
    }
    
    console.log('Loading chart detail for:', chartName);
    console.log('Current charts:', state.charts);
    
    // If charts are not loaded yet, wait for them
    if (!state.charts || state.charts.length === 0) {
        try {
            // Show loading state
            detailView.innerHTML = '<div class="loading-state"><p>Loading charts...</p></div>';
            await loadCharts();
            console.log('Charts loaded:', state.charts);
        } catch (error) {
            console.error('Failed to load charts:', error);
            detailView.innerHTML = `
                <div class="error-state">
                    <h3>Failed to load charts</h3>
                    <p>Please refresh the page or try again later.</p>
                    <button class="btn-primary" onclick="window.location.reload()" style="margin-top: var(--spacing-md);">Refresh Page</button>
                </div>
            `;
            return;
        }
    }
    
    // Now find and render the chart (case-insensitive search)
    const chart = state.charts.find(c => {
        const match = c.name.toLowerCase() === chartName.toLowerCase() || c.name === chartName;
        if (match) console.log('Found chart:', c.name);
        return match;
    });
    
    if (chart) {
        console.log('Rendering chart:', chart.name);
        state.currentChart = chart;
        renderChartDetailContent(chart);
    } else {
        console.warn('Chart not found:', chartName, 'Available charts:', state.charts.map(c => c.name));
        // Chart not found - show helpful error
        detailView.innerHTML = `
            <div class="error-state">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 64px; height: 64px; margin-bottom: var(--spacing-md); opacity: 0.5;">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                </svg>
                <h3>Chart Not Found</h3>
                <p>The chart "${escapeHtml(chartName)}" does not exist in this repository.</p>
                <p style="color: var(--text-muted); font-size: var(--font-size-sm); margin-top: var(--spacing-sm);">
                    Available charts: ${state.charts.length > 0 ? state.charts.map(c => c.name).join(', ') : 'None'}
                </p>
                <button class="btn-primary" onclick="showView('charts')" style="margin-top: var(--spacing-md);">Browse All Charts</button>
            </div>
        `;
    }
}

// Render chart detail content (separated to avoid circular calls)
function renderChartDetailContent(chart) {
    const detailView = document.getElementById('chartDetailView');
    if (!detailView) {
        console.error('Detail view element not found in renderChartDetailContent');
        return;
    }
    
    // Ensure the view is visible
    detailView.style.display = 'block';
    detailView.style.visibility = 'visible';
    
    console.log('Rendering chart detail content for:', chart.name);

    const latestVersion = chart.versions && chart.versions.length > 0 ? chart.versions[0] : null;
    const versionCount = chart.versions ? chart.versions.length : 0;
    
    // Format default values
    const defaultIcon = '<svg class="detail-icon-placeholder" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>';
    const displayIcon = chart.icon ? 
        `<img src="${chart.icon}" alt="${chart.name}" class="detail-icon" onerror="this.outerHTML='${defaultIcon.replace(/'/g, "\\'")}'">` : 
        defaultIcon;
    
    const displayDescription = chart.description || 'No description available for this chart.';
    const displayHome = chart.home || 'Not specified';
    const displayLatestVersion = latestVersion ? latestVersion.version : 'N/A';
    const displayAppVersion = latestVersion && latestVersion.appVersion ? latestVersion.appVersion : 'N/A';
    const displayCreated = latestVersion && latestVersion.created ? 
        new Date(latestVersion.created).toLocaleDateString() : 'N/A';

    detailView.innerHTML = `
        <div class="detail-header">
            <div class="detail-icon-wrapper">
                ${displayIcon}
            </div>
            <div class="detail-info">
                <div class="detail-title-section">
                    <h1 class="detail-title">${escapeHtml(chart.name)}</h1>
                    ${latestVersion ? `<span class="detail-latest-badge">v${escapeHtml(latestVersion.version)}</span>` : ''}
                </div>
                <p class="detail-description">${escapeHtml(displayDescription)}</p>
                
                <div class="detail-meta-grid">
                    <div class="detail-meta-item">
                        <span class="detail-meta-label">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="meta-icon">
                                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
                            </svg>
                            Versions
                        </span>
                        <span class="detail-meta-value">${versionCount}</span>
                    </div>
                    
                    <div class="detail-meta-item">
                        <span class="detail-meta-label">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="meta-icon">
                                <path d="M20 6L9 17l-5-5"></path>
                            </svg>
                            Latest Version
                        </span>
                        <span class="detail-meta-value">${escapeHtml(displayLatestVersion)}</span>
                    </div>
                    
                    <div class="detail-meta-item">
                        <span class="detail-meta-label">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="meta-icon">
                                <rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
                                <line x1="8" y1="21" x2="16" y2="21"></line>
                                <line x1="12" y1="17" x2="12" y2="21"></line>
                            </svg>
                            App Version
                        </span>
                        <span class="detail-meta-value">${escapeHtml(displayAppVersion)}</span>
                    </div>
                    
                    <div class="detail-meta-item">
                        <span class="detail-meta-label">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="meta-icon">
                                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
                                <polyline points="9 22 9 12 15 12 15 22"></polyline>
                            </svg>
                            Home
                        </span>
                        <span class="detail-meta-value">
                            ${chart.home ? 
                                `<a href="${chart.home}" target="_blank" rel="noopener noreferrer">${escapeHtml(displayHome)}</a>` : 
                                `<span class="text-muted">${escapeHtml(displayHome)}</span>`
                            }
                        </span>
                    </div>
                    
                    <div class="detail-meta-item">
                        <span class="detail-meta-label">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="meta-icon">
                                <circle cx="12" cy="12" r="10"></circle>
                                <polyline points="12 6 12 12 16 14"></polyline>
                            </svg>
                            Created
                        </span>
                        <span class="detail-meta-value">${escapeHtml(displayCreated)}</span>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="versions-section">
            <div class="section-header-inline">
                <h2 class="section-title">Available Versions</h2>
                <span class="version-count-badge">${versionCount} version${versionCount !== 1 ? 's' : ''}</span>
            </div>
            <ul class="versions-list">
                ${chart.versions && chart.versions.length > 0 ? 
                    chart.versions.map((v, index) => `
                        <li class="version-item">
                            <div class="version-info">
                                <div class="version-badge-large">${escapeHtml(v.version)}</div>
                                <div class="version-details">
                                    ${v.appVersion ? `
                                        <div class="version-detail-row">
                                            <span class="version-detail-label">App Version:</span>
                                            <span class="version-detail-value">${escapeHtml(v.appVersion)}</span>
                                        </div>
                                    ` : ''}
                                    ${v.created ? `
                                        <div class="version-detail-row">
                                            <span class="version-detail-label">Created:</span>
                                            <span class="version-detail-value">${new Date(v.created).toLocaleDateString()}</span>
                                        </div>
                                    ` : ''}
                                    ${v.description ? `
                                        <div class="version-detail-row">
                                            <span class="version-detail-label">Description:</span>
                                            <span class="version-detail-value">${escapeHtml(v.description)}</span>
                                        </div>
                                    ` : ''}
                                </div>
                            </div>
                            <div class="version-actions">
                                ${v.urls && v.urls.length > 0 ? `
                                    <a href="${v.urls[0]}" class="btn-download" download title="Download ${escapeHtml(chart.name)} ${escapeHtml(v.version)}">
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                                            <polyline points="7 10 12 15 17 10"></polyline>
                                            <line x1="12" y1="15" x2="12" y2="3"></line>
                                        </svg>
                                        Download
                                    </a>
                                ` : ''}
                                ${state.token ? `
                                    <button class="btn-delete" onclick="deleteChart('${escapeHtml(chart.name)}', '${escapeHtml(v.version)}')" title="Delete version ${escapeHtml(v.version)}">
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                                            <polyline points="3 6 5 6 21 6"></polyline>
                                            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                                        </svg>
                                        Delete
                                    </button>
                                ` : ''}
                            </div>
                        </li>
                    `).join('') 
                    : `
                    <li class="version-item-empty">
                        <div class="empty-version-message">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 48px; height: 48px; margin-bottom: var(--spacing-md); opacity: 0.5;">
                                <circle cx="12" cy="12" r="10"></circle>
                                <line x1="12" y1="8" x2="12" y2="12"></line>
                                <line x1="12" y1="16" x2="12.01" y2="16"></line>
                            </svg>
                            <p>No versions available for this chart</p>
                        </div>
                    </li>
                    `
                }
            </ul>
        </div>
        
        <div class="installation-section">
            <div class="installation-header">
                <h3 class="section-subtitle">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="section-icon">
                        <path d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path>
                    </svg>
                    Installation Instructions
                </h3>
            </div>
            
            <div class="installation-tabs">
                <button class="install-tab active" data-tab="quick" onclick="switchInstallTab('quick')">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
                    </svg>
                    Quick Install
                </button>
                <button class="install-tab" data-tab="version" onclick="switchInstallTab('version')">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <circle cx="12" cy="12" r="10"></circle>
                        <polyline points="12 6 12 12 16 14"></polyline>
                    </svg>
                    Specific Version
                </button>
                <button class="install-tab" data-tab="values" onclick="switchInstallTab('values')">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                        <polyline points="14 2 14 8 20 8"></polyline>
                        <line x1="16" y1="13" x2="8" y2="13"></line>
                        <line x1="16" y1="17" x2="8" y2="17"></line>
                        <polyline points="10 9 9 9 8 9"></polyline>
                    </svg>
                    With Values
                </button>
            </div>

            <div class="install-tab-content active" id="install-quick">
                <div class="install-step">
                    <div class="step-number">1</div>
                    <div class="step-content">
                        <h4 class="step-title">Add the repository</h4>
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('repo-add-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="repo-add-${chart.name}">helm repo add ${repositoryName} ${repositoryDomain}</code>
                        </div>
                    </div>
                </div>

                <div class="install-step">
                    <div class="step-number">2</div>
                    <div class="step-content">
                        <h4 class="step-title">Update repository index</h4>
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('repo-update-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="repo-update-${chart.name}">helm repo update</code>
                        </div>
                    </div>
                </div>

                <div class="install-step">
                    <div class="step-number">3</div>
                    <div class="step-content">
                        <h4 class="step-title">Install the chart</h4>
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('install-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="install-${chart.name}">helm install ${escapeHtml(chart.name)} ${repositoryName}/${escapeHtml(chart.name)}</code>
                        </div>
                    </div>
                </div>
            </div>

            <div class="install-tab-content" id="install-version">
                <div class="install-step">
                    <div class="step-number">1</div>
                    <div class="step-content">
                        <h4 class="step-title">Add the repository</h4>
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('repo-add-ver-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="repo-add-ver-${chart.name}">helm repo add ${repositoryName} ${repositoryDomain}</code>
                        </div>
                    </div>
                </div>

                <div class="install-step">
                    <div class="step-number">2</div>
                    <div class="step-content">
                        <h4 class="step-title">Update repository index</h4>
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('repo-update-ver-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="repo-update-ver-${chart.name}">helm repo update</code>
                        </div>
                    </div>
                </div>

                <div class="install-step">
                    <div class="step-number">3</div>
                    <div class="step-content">
                        <h4 class="step-title">Install specific version</h4>
                        ${latestVersion ? `
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('install-ver-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="install-ver-${chart.name}">helm install ${escapeHtml(chart.name)} ${repositoryName}/${escapeHtml(chart.name)} --version ${escapeHtml(latestVersion.version)}</code>
                        </div>
                        ` : '<p class="text-muted">No versions available</p>'}
                    </div>
                </div>
            </div>

            <div class="install-tab-content" id="install-values">
                <div class="install-step">
                    <div class="step-number">1</div>
                    <div class="step-content">
                        <h4 class="step-title">Add the repository</h4>
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('repo-add-val-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="repo-add-val-${chart.name}">helm repo add ${repositoryName} ${repositoryDomain}</code>
                        </div>
                    </div>
                </div>

                <div class="install-step">
                    <div class="step-number">2</div>
                    <div class="step-content">
                        <h4 class="step-title">Update repository index</h4>
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('repo-update-val-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="repo-update-val-${chart.name}">helm repo update</code>
                        </div>
                    </div>
                </div>

                <div class="install-step">
                    <div class="step-number">3</div>
                    <div class="step-content">
                        <h4 class="step-title">Install with custom values</h4>
                        <div class="code-block-modern">
                            <div class="code-header">
                                <span class="code-lang">bash</span>
                                <button class="copy-btn" onclick="copyToClipboard('install-val-${chart.name}', this)" title="Copy command">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                            <code class="code-content" id="install-val-${chart.name}">helm install ${escapeHtml(chart.name)} ${repositoryName}/${escapeHtml(chart.name)} -f values.yaml</code>
                        </div>
                        <p class="step-hint">Create a <code>values.yaml</code> file with your custom configuration</p>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="default-values-section" id="defaultValuesSection-${chart.name}">
            <div class="section-header-inline">
                <h2 class="section-title">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="section-icon">
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                        <polyline points="14 2 14 8 20 8"></polyline>
                        <line x1="16" y1="13" x2="8" y2="13"></line>
                        <line x1="16" y1="17" x2="8" y2="17"></line>
                        <polyline points="10 9 9 9 8 9"></polyline>
                    </svg>
                    Default Values
                </h2>
                <div class="values-actions">
                    <button class="btn-secondary btn-sm" onclick="loadDefaultValues('${chart.name}', '${latestVersion ? latestVersion.version : ''}')" id="loadValuesBtn-${chart.name}">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                            <polyline points="7 10 12 15 17 10"></polyline>
                            <line x1="12" y1="15" x2="12" y2="3"></line>
                        </svg>
                        Show Values
                    </button>
                    ${latestVersion ? `
                        <a href="/api/charts/${encodeURIComponent(chart.name)}/${encodeURIComponent(latestVersion.version)}/values.yaml" 
                           class="btn-download btn-sm" 
                           download="${chart.name}-${latestVersion.version}-values.yaml"
                           title="Download values.yaml">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                                <polyline points="7 10 12 15 17 10"></polyline>
                                <line x1="12" y1="15" x2="12" y2="3"></line>
                            </svg>
                            Download
                        </a>
                    ` : ''}
                </div>
            </div>
            <div class="values-content" id="valuesContent-${chart.name}" style="display: none;">
                <div class="values-loading" id="valuesLoading-${chart.name}">
                    <p>Loading default values...</p>
                </div>
                <div class="values-error" id="valuesError-${chart.name}" style="display: none;">
                    <p>Failed to load default values. This chart may not have a values.yaml file.</p>
                </div>
                <div class="values-display" id="valuesDisplay-${chart.name}" style="display: none;">
                    <div class="yaml-editor-container">
                        <div class="code-header">
                            <span class="code-lang">yaml</span>
                            <div class="editor-actions">
                                <button class="copy-btn" onclick="copyYamlContent('values-editor-${chart.name}', this)" title="Copy values.yaml">
                                    <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                    </svg>
                                    <span class="copy-text">Copy</span>
                                </button>
                            </div>
                        </div>
                        <textarea id="values-editor-${chart.name}" class="yaml-editor-textarea"></textarea>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    // Load default values if available
    if (latestVersion) {
        // Small delay to ensure DOM is ready
        setTimeout(() => {
            loadDefaultValues(chart.name, latestVersion.version);
        }, 100);
    }
}

// Load default values for a chart
async function loadDefaultValues(chartName, version) {
    const valuesSection = document.getElementById(`defaultValuesSection-${chartName}`);
    const valuesContent = document.getElementById(`valuesContent-${chartName}`);
    const valuesLoading = document.getElementById(`valuesLoading-${chartName}`);
    const valuesError = document.getElementById(`valuesError-${chartName}`);
    const valuesDisplay = document.getElementById(`valuesDisplay-${chartName}`);
    const loadBtn = document.getElementById(`loadValuesBtn-${chartName}`);
    
    if (!valuesContent || !valuesLoading || !valuesError || !valuesDisplay) {
        console.error('Values section elements not found');
        return;
    }
    
    // Show content area
    valuesContent.style.display = 'block';
    valuesLoading.style.display = 'block';
    valuesError.style.display = 'none';
    valuesDisplay.style.display = 'none';
    
    // Update button text
    if (loadBtn) {
        loadBtn.innerHTML = `
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                <polyline points="23 4 23 10 17 10"></polyline>
                <polyline points="1 20 1 14 7 14"></polyline>
                <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
            Loading...
        `;
        loadBtn.disabled = true;
    }
    
    try {
        const response = await fetch(`/api/charts/${encodeURIComponent(chartName)}/${encodeURIComponent(version)}/values`);
        
        if (!response.ok) {
            if (response.status === 404) {
                // Chart doesn't have values.yaml
                valuesLoading.style.display = 'none';
                valuesError.style.display = 'block';
                if (loadBtn) {
                    loadBtn.innerHTML = `
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                            <polyline points="7 10 12 15 17 10"></polyline>
                            <line x1="12" y1="15" x2="12" y2="3"></line>
                        </svg>
                        Show Values
                    `;
                    loadBtn.disabled = false;
                }
                return;
            }
            throw new Error(`Failed to load values: ${response.statusText}`);
        }
        
        const data = await response.json();
        const yamlContent = data.yaml || '';
        
        // Hide loading, show display
        valuesLoading.style.display = 'none';
        valuesError.style.display = 'none';
        valuesDisplay.style.display = 'block';
        
        // Initialize CodeMirror editor
        const textarea = document.getElementById(`values-editor-${chartName}`);
        if (textarea && typeof CodeMirror !== 'undefined') {
            // Destroy existing editor if any
            if (textarea.cmEditor) {
                textarea.cmEditor.toTextArea();
            }
            
            // Get current theme
            const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
            const cmTheme = currentTheme === 'dark' ? 'monokai' : 'default';
            
            // Create new CodeMirror instance
            const editor = CodeMirror.fromTextArea(textarea, {
                mode: 'yaml',
                theme: cmTheme,
                lineNumbers: true,
                lineWrapping: true,
                readOnly: true,
                indentUnit: 2,
                tabSize: 2,
                viewportMargin: Infinity,
                autoRefresh: true
            });
            
            // Set the content
            editor.setValue(yamlContent);
            
            // Store reference for later use
            textarea.cmEditor = editor;
            
            // Refresh editor after a short delay to ensure proper rendering
            setTimeout(() => {
                editor.refresh();
            }, 100);
        } else if (textarea) {
            // Fallback if CodeMirror is not loaded
            textarea.value = yamlContent;
        }
        
        // Update button to show "Hide Values"
        if (loadBtn) {
            loadBtn.innerHTML = `
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                    <polyline points="18 6 6 18"></polyline>
                    <polyline points="6 6 18 18"></polyline>
                </svg>
                Hide Values
            `;
            loadBtn.onclick = () => {
                valuesContent.style.display = 'none';
                loadBtn.innerHTML = `
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                        <polyline points="7 10 12 15 17 10"></polyline>
                        <line x1="12" y1="15" x2="12" y2="3"></line>
                    </svg>
                    Show Values
                `;
                loadBtn.onclick = () => loadDefaultValues(chartName, version);
            };
            loadBtn.disabled = false;
        }
        
    } catch (error) {
        console.error('Error loading default values:', error);
        valuesLoading.style.display = 'none';
        valuesError.style.display = 'block';
        if (loadBtn) {
            loadBtn.innerHTML = `
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 16px; height: 16px; margin-right: 4px;">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                    <polyline points="7 10 12 15 17 10"></polyline>
                    <line x1="12" y1="15" x2="12" y2="3"></line>
                </svg>
                Show Values
            `;
            loadBtn.disabled = false;
        }
    }
}

// Handle search
function handleSearch(e) {
    state.searchQuery = e.target.value.toLowerCase();
    applyFilters();
}

// Apply filters and sorting
function applyFilters() {
    let filtered = [...state.charts];

    // Apply search filter
    if (state.searchQuery) {
        filtered = filtered.filter(chart => 
            chart.name.toLowerCase().includes(state.searchQuery) ||
            (chart.description && chart.description.toLowerCase().includes(state.searchQuery))
        );
    }

    // Apply sorting
    switch (state.sortBy) {
        case 'name':
            filtered.sort((a, b) => a.name.localeCompare(b.name));
            break;
        case 'recent':
            filtered.sort((a, b) => {
                const aDate = a.versions && a.versions.length > 0 ? new Date(a.versions[0].created) : 0;
                const bDate = b.versions && b.versions.length > 0 ? new Date(b.versions[0].created) : 0;
                return bDate - aDate;
            });
            break;
        case 'versions':
            filtered.sort((a, b) => {
                const aCount = a.versions ? a.versions.length : 0;
                const bCount = b.versions ? b.versions.length : 0;
                return bCount - aCount;
            });
            break;
    }

    state.filteredCharts = filtered;
    renderCharts();
}

// Handle upload
async function handleUpload(e) {
    e.preventDefault();
    const fileInput = document.getElementById('chartFile');
    const statusDiv = document.getElementById('uploadStatus');
    
    if (!fileInput || !fileInput.files[0]) {
        showStatus(statusDiv, 'Please select a chart file', 'error');
        return;
    }

    const formData = new FormData();
    formData.append('chart', fileInput.files[0]);

    showStatus(statusDiv, 'Uploading...', 'info');

    try {
        const headers = {};
        if (state.token) {
            headers['Authorization'] = `Bearer ${state.token}`;
        }
        
        const response = await fetch(`${API_BASE}/charts`, {
            method: 'POST',
            headers: headers,
            body: formData
        });

        const data = await response.json();

        if (response.ok) {
            showStatus(statusDiv, `Chart ${data.name} v${data.version} uploaded successfully!`, 'success');
            fileInput.value = '';
            setTimeout(async () => {
                hideModal('uploadModal');
                await loadCharts();
                updateStats();
                if (state.currentView === 'myCharts') {
                    renderMyCharts();
                }
            }, 1500);
        } else {
            showStatus(statusDiv, `Error: ${data.error}`, 'error');
        }
    } catch (error) {
        showStatus(statusDiv, `Error: ${error.message}`, 'error');
    }
}

// Delete chart
async function deleteChart(name, version) {
    if (!confirm(`Are you sure you want to delete ${name} version ${version}?`)) {
        return;
    }

    try {
        const headers = {};
        if (state.token) {
            headers['Authorization'] = `Bearer ${state.token}`;
        }
        
        const response = await fetch(`${API_BASE}/charts/${encodeURIComponent(name)}/${encodeURIComponent(version)}`, {
            method: 'DELETE',
            headers: headers
        });

        if (response.ok) {
            await loadCharts();
            updateStats();
            
            // Update views if needed
            if (state.currentView === 'myCharts') {
                renderMyCharts();
            }
            
            // If we're viewing this chart's detail, go back to charts view
            if (state.currentView === 'chartDetail' && state.currentChart && state.currentChart.name === name) {
                if (state.token) {
                    showView('myCharts');
                } else {
                    showView('charts');
                }
            }
        } else {
            const data = await response.json();
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        alert(`Error: ${error.message}`);
    }
}

// Update statistics
function updateStats() {
    const totalCharts = state.charts.length;
    const totalVersions = state.charts.reduce((sum, chart) => 
        sum + (chart.versions ? chart.versions.length : 0), 0
    );

    const chartsEl = document.getElementById('totalCharts');
    const versionsEl = document.getElementById('totalVersions');
    
    if (chartsEl) chartsEl.textContent = totalCharts;
    if (versionsEl) versionsEl.textContent = totalVersions;
}

// URL Routing
function getRouteFromPath(path) {
    // Remove leading slash and split
    const parts = path.split('/').filter(p => p);
    
    if (parts.length === 0) {
        return { view: 'home', params: {} };
    }
    
    const view = parts[0];
    const params = {};
    
    switch (view) {
        case 'charts':
            if (parts.length === 1) {
                return { view: 'charts', params: {} };
            } else if (parts.length === 2) {
                // /charts/:name
                params.chartName = decodeURIComponent(parts[1]);
                return { view: 'chartDetail', params };
            }
            break;
        case 'my-charts':
            return { view: 'myCharts', params: {} };
        case 'admin':
            return { view: 'admin', params: {} };
        default:
            return { view: 'home', params: {} };
    }
    
    return { view: 'home', params: {} };
}

function updateURL(viewName, params = {}) {
    let path = '/';
    
    switch (viewName) {
        case 'home':
            path = '/';
            break;
        case 'charts':
            path = '/charts';
            break;
        case 'myCharts':
            path = '/my-charts';
            break;
        case 'admin':
            path = '/admin';
            break;
        case 'chartDetail':
            if (params.chartName) {
                path = `/charts/${encodeURIComponent(params.chartName)}`;
            } else {
                path = '/charts';
            }
            break;
        default:
            path = '/';
    }
    
    // Update URL without reloading page
    if (window.location.pathname !== path) {
        window.history.pushState({ view: viewName, params }, '', path);
    }
}

function handleRoute(route, updateHistory = true) {
    const { view, params } = route;
    
    // Update URL if needed
    if (updateHistory) {
        updateURL(view, params);
    }
    
    // Show the view
    showViewInternal(view, params);
}

function showViewInternal(viewName, params = {}) {
    state.currentView = viewName;

    // Hide all views - use both ID and class selectors to ensure all are hidden
    const homeView = document.getElementById('homeView');
    const chartsView = document.getElementById('chartsView');
    const myChartsView = document.getElementById('myChartsView');
    const detailView = document.getElementById('chartDetailView');
    const adminView = document.getElementById('adminView');

    // Hide all sections
    if (homeView) {
        homeView.style.display = 'none';
        homeView.style.visibility = 'hidden';
    }
    if (chartsView) {
        chartsView.style.display = 'none';
        chartsView.style.visibility = 'hidden';
    }
    if (myChartsView) {
        myChartsView.style.display = 'none';
        myChartsView.style.visibility = 'hidden';
    }
    if (detailView) {
        detailView.style.display = 'none';
        detailView.style.visibility = 'hidden';
    }
    if (adminView) {
        adminView.style.display = 'none';
        adminView.style.visibility = 'hidden';
    }

    // Show selected view
    switch (viewName) {
        case 'home':
            if (homeView) {
                homeView.style.display = 'block';
                homeView.style.visibility = 'visible';
            }
            break;
        case 'charts':
            if (chartsView) {
                chartsView.style.display = 'block';
                chartsView.style.visibility = 'visible';
                applyFilters();
            }
            break;
        case 'myCharts':
            if (myChartsView) {
                myChartsView.style.display = 'block';
                myChartsView.style.visibility = 'visible';
                renderMyCharts();
            }
            break;
        case 'admin':
            if (adminView) {
                adminView.style.display = 'block';
                adminView.style.visibility = 'visible';
                loadAdminData();
            }
            break;
        case 'chartDetail':
            if (params.chartName) {
                console.log('Showing chart detail for:', params.chartName);
                // Show the detail view container
                if (detailView) {
                    detailView.style.display = 'block';
                    detailView.style.visibility = 'visible';
                    // Show loading state immediately
                    detailView.innerHTML = '<div class="loading-state"><p>Loading chart...</p></div>';
                } else {
                    console.error('Detail view element not found!');
                }
                // Wait for charts to load if not loaded yet, then render
                loadChartDetailAsync(params.chartName).catch(error => {
                    console.error('Error loading chart detail:', error);
                    if (detailView) {
                        detailView.innerHTML = `
                            <div class="error-state">
                                <h3>Error Loading Chart</h3>
                                <p>${escapeHtml(error.message || 'An error occurred')}</p>
                                <button class="btn-primary" onclick="showView('charts')" style="margin-top: var(--spacing-md);">Back to Charts</button>
                            </div>
                        `;
                    }
                });
            } else if (state.currentChart) {
                if (detailView) {
                    detailView.style.display = 'block';
                    detailView.style.visibility = 'visible';
                }
                renderChartDetailContent(state.currentChart);
            } else {
                // Fallback to charts view if no chart specified
                if (chartsView) {
                    chartsView.style.display = 'block';
                    chartsView.style.visibility = 'visible';
                }
                applyFilters();
            }
            break;
        default:
            if (homeView) {
                homeView.style.display = 'block';
                homeView.style.visibility = 'visible';
            }
    }
}

// View management
function showView(viewName, params = {}) {
    handleRoute({ view: viewName, params }, true);
}

// Modal management
function showModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = 'flex';
        document.body.style.overflow = 'hidden';
    }
}

function hideModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = 'none';
        document.body.style.overflow = '';
    }
}

// Status messages
function showStatus(message, type) {
    const statusDiv = document.getElementById('uploadStatus');
    if (statusDiv) {
        statusDiv.textContent = message;
        statusDiv.className = `upload-status ${type}`;
    }
}

function showError(message) {
    // Could implement a toast notification system here
    console.error(message);
}

// Utility functions
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Authentication functions
async function checkAuth() {
    if (state.token) {
        try {
            const response = await fetch(`${API_BASE}/auth/me`, {
                headers: {
                    'Authorization': `Bearer ${state.token}`
                }
            });
            if (response.ok) {
                state.user = await response.json();
                return true;
            } else {
                // Token invalid, clear it
                state.token = null;
                localStorage.removeItem('auth_token');
            }
        } catch (error) {
            console.error('Auth check failed:', error);
            state.token = null;
            localStorage.removeItem('auth_token');
        }
    }
    return false;
}

async function handleLogin(e) {
    e.preventDefault();
    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;
    const statusDiv = document.getElementById('loginStatus');

    showStatus(statusDiv, 'Logging in...', 'info');

    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();

        if (response.ok) {
            state.token = data.token;
            state.user = data.user;
            localStorage.setItem('auth_token', data.token);
            showStatus(statusDiv, 'Login successful!', 'success');
            setTimeout(() => {
                hideModal('loginModal');
                updateAuthUI();
                document.getElementById('loginForm').reset();
            }, 1000);
        } else {
            showStatus(statusDiv, `Error: ${data.error}`, 'error');
        }
    } catch (error) {
        showStatus(statusDiv, `Error: ${error.message}`, 'error');
    }
}

async function handleRegister(e) {
    e.preventDefault();
    const username = document.getElementById('registerUsername').value;
    const email = document.getElementById('registerEmail').value;
    const password = document.getElementById('registerPassword').value;
    const statusDiv = document.getElementById('registerStatus');

    showStatus(statusDiv, 'Registering...', 'info');

    try {
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, email, password })
        });

        const data = await response.json();

        if (response.ok) {
            state.token = data.token;
            state.user = data.user;
            localStorage.setItem('auth_token', data.token);
            showStatus(statusDiv, 'Registration successful!', 'success');
            setTimeout(() => {
                hideModal('registerModal');
                updateAuthUI();
                document.getElementById('registerForm').reset();
            }, 1000);
        } else {
            showStatus(statusDiv, `Error: ${data.error}`, 'error');
        }
    } catch (error) {
        showStatus(statusDiv, `Error: ${error.message}`, 'error');
    }
}

function handleLogout() {
    state.token = null;
    state.user = null;
    localStorage.removeItem('auth_token');
    updateAuthUI();
}

function updateAuthUI() {
    const authButtons = document.getElementById('authButtons');
    const userMenu = document.getElementById('userMenu');
    const usernameDisplay = document.getElementById('usernameDisplay');
    const uploadBtn = document.getElementById('uploadBtn');
    const myChartsLink = document.getElementById('myChartsLink');
    const adminLink = document.getElementById('adminLink');

    if (state.token && state.user) {
        // User is logged in
        if (authButtons) authButtons.style.display = 'none';
        if (userMenu) userMenu.style.display = 'flex';
        if (usernameDisplay) usernameDisplay.textContent = state.user.username;
        if (uploadBtn) uploadBtn.style.display = 'flex';
        if (myChartsLink) myChartsLink.style.display = 'block';
        // Show admin link only if user is admin
        if (adminLink) adminLink.style.display = state.user.is_admin ? 'block' : 'none';
    } else {
        // User is not logged in
        if (authButtons) authButtons.style.display = 'flex';
        if (userMenu) userMenu.style.display = 'none';
        if (uploadBtn) uploadBtn.style.display = 'none';
        if (myChartsLink) myChartsLink.style.display = 'none';
    }
}

// Render My Charts management view
function renderMyCharts() {
    const grid = document.getElementById('myChartsGrid');
    const emptyState = document.getElementById('myChartsEmpty');
    
    if (!grid) return;

    // Filter charts (for now, show all - in future can filter by uploader)
    const myCharts = state.charts;

    if (myCharts.length === 0) {
        grid.style.display = 'none';
        if (emptyState) emptyState.style.display = 'block';
        updateMyChartsStats(0, 0, null);
        return;
    }

    if (emptyState) emptyState.style.display = 'none';
    grid.style.display = 'grid';

    // Apply management filters
    let filtered = [...myCharts];
    const searchInput = document.getElementById('managementSearch');
    const sortSelect = document.getElementById('managementSort');
    
    if (searchInput && searchInput.value) {
        const query = searchInput.value.toLowerCase();
        filtered = filtered.filter(chart => 
            chart.name.toLowerCase().includes(query) ||
            (chart.description && chart.description.toLowerCase().includes(query))
        );
    }

    // Sort
    if (sortSelect) {
        switch (sortSelect.value) {
            case 'name':
                filtered.sort((a, b) => a.name.localeCompare(b.name));
                break;
            case 'recent':
                filtered.sort((a, b) => {
                    const aDate = a.versions && a.versions.length > 0 ? new Date(a.versions[0].created) : 0;
                    const bDate = b.versions && b.versions.length > 0 ? new Date(b.versions[0].created) : 0;
                    return bDate - aDate;
                });
                break;
            case 'versions':
                filtered.sort((a, b) => {
                    const aCount = a.versions ? a.versions.length : 0;
                    const bCount = b.versions ? b.versions.length : 0;
                    return bCount - aCount;
                });
                break;
        }
    }

    grid.innerHTML = filtered.map(chart => createManagementChartCard(chart)).join('');
    
    // Add event listeners
    grid.querySelectorAll('.chart-card-management').forEach(card => {
        card.addEventListener('click', (e) => {
            if (!e.target.closest('.chart-actions-menu')) {
                const chartName = card.dataset.chartName;
                showChartDetail(chartName);
            }
        });
    });

    // Calculate stats
    const totalVersions = myCharts.reduce((sum, chart) => sum + (chart.versions ? chart.versions.length : 0), 0);
    const latestUpload = myCharts.length > 0 && myCharts[0].versions && myCharts[0].versions.length > 0 
        ? myCharts[0].versions[0].created 
        : null;
    
    updateMyChartsStats(myCharts.length, totalVersions, latestUpload);
}

// Create management chart card
function createManagementChartCard(chart) {
    const latestVersion = chart.versions && chart.versions.length > 0 ? chart.versions[0] : null;
    const versionCount = chart.versions ? chart.versions.length : 0;
    const totalSize = chart.versions ? chart.versions.reduce((sum, v) => sum + (v.size || 0), 0) : 0;
    
    return `
        <div class="chart-card-management" data-chart-name="${chart.name}">
            <div class="chart-actions-menu" onclick="event.stopPropagation();">
                <button class="chart-action-btn" onclick="showChartDetail('${chart.name}')" title="View Details">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                        <circle cx="12" cy="12" r="3"></circle>
                    </svg>
                </button>
                ${latestVersion ? `
                <button class="chart-action-btn" onclick="deleteChart('${chart.name}', '${latestVersion.version}')" title="Delete Latest Version">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="3 6 5 6 21 6"></polyline>
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                    </svg>
                </button>
                ` : ''}
            </div>
            <div class="chart-header">
                ${chart.icon ? `<img src="${chart.icon}" alt="${chart.name}" class="chart-icon" onerror="this.style.display='none'">` : ''}
                <div style="flex: 1;">
                    <h3 class="chart-title">
                        <a href="/charts/${encodeURIComponent(chart.name)}" onclick="event.preventDefault(); showChartDetail('${chart.name}'); return false;">${escapeHtml(chart.name)}</a>
                    </h3>
                    <p class="chart-description">${escapeHtml(chart.description || 'No description available')}</p>
                </div>
            </div>
            <div class="chart-management-info">
                <div class="chart-info-row">
                    <span class="chart-info-label">Versions:</span>
                    <span class="chart-info-value">${versionCount}</span>
                </div>
                ${latestVersion ? `
                <div class="chart-info-row">
                    <span class="chart-info-label">Latest:</span>
                    <span class="chart-info-value">${escapeHtml(latestVersion.version)}</span>
                </div>
                ` : ''}
                ${chart.home ? `
                <div class="chart-info-row">
                    <span class="chart-info-label">Home:</span>
                    <a href="${chart.home}" target="_blank" class="chart-info-value" style="color: var(--primary); text-decoration: none;">${escapeHtml(chart.home)}</a>
                </div>
                ` : ''}
            </div>
        </div>
    `;
}

// Update My Charts statistics
function updateMyChartsStats(totalCharts, totalVersions, latestUpload) {
    const totalEl = document.getElementById('myTotalCharts');
    const versionsEl = document.getElementById('myTotalVersions');
    const latestEl = document.getElementById('myLatestUpload');
    
    if (totalEl) totalEl.textContent = totalCharts;
    if (versionsEl) versionsEl.textContent = totalVersions;
    
    if (latestEl) {
        if (latestUpload) {
            const date = new Date(latestUpload);
            const now = new Date();
            const diffMs = now - date;
            const diffMins = Math.floor(diffMs / 60000);
            const diffHours = Math.floor(diffMs / 3600000);
            const diffDays = Math.floor(diffMs / 86400000);
            
            if (diffMins < 60) {
                latestEl.textContent = `${diffMins}m ago`;
            } else if (diffHours < 24) {
                latestEl.textContent = `${diffHours}h ago`;
            } else if (diffDays < 7) {
                latestEl.textContent = `${diffDays}d ago`;
            } else {
                latestEl.textContent = date.toLocaleDateString();
            }
        } else {
            latestEl.textContent = '-';
        }
    }
}

// Helper function for status messages
function showStatus(element, message, type) {
    if (element) {
        element.textContent = message;
        element.className = `upload-status ${type}`;
    }
}


// Initialize modern dropdowns
function initializeModernDropdowns() {
    // Initialize sort dropdown
    const sortDropdown = document.getElementById('sortSelectDropdown');
    if (sortDropdown) {
        setupModernDropdown(sortDropdown, 'sortSelect', 'name');
    }

    // Initialize management sort dropdown
    const managementDropdown = document.getElementById('managementSortDropdown');
    if (managementDropdown) {
        setupModernDropdown(managementDropdown, 'managementSort', 'name');
    }
}

// Setup a modern dropdown
function setupModernDropdown(dropdownElement, selectId, defaultValue) {
    const toggle = dropdownElement.querySelector('.dropdown-toggle');
    const menu = dropdownElement.querySelector('.dropdown-menu');
    const items = dropdownElement.querySelectorAll('.dropdown-item');
    const hiddenSelect = document.getElementById(selectId);
    
    if (!toggle || !menu || !hiddenSelect) return;
    
    // Set initial value
    const initialValue = hiddenSelect.value || defaultValue;
    setDropdownValue(dropdownElement, initialValue, hiddenSelect);
    
    // Toggle dropdown
    toggle.addEventListener('click', (e) => {
        e.stopPropagation();
        const isOpen = dropdownElement.classList.contains('open');
        
        // Close all other dropdowns
        document.querySelectorAll('.modern-dropdown').forEach(dd => {
            if (dd !== dropdownElement) {
                dd.classList.remove('open');
                dd.querySelector('.dropdown-toggle')?.setAttribute('aria-expanded', 'false');
            }
        });
        
        if (isOpen) {
            dropdownElement.classList.remove('open');
            toggle.setAttribute('aria-expanded', 'false');
        } else {
            dropdownElement.classList.add('open');
            toggle.setAttribute('aria-expanded', 'true');
        }
    });
    
    // Handle item selection
    items.forEach(item => {
        item.addEventListener('click', (e) => {
            e.stopPropagation();
            const value = item.dataset.value;
            setDropdownValue(dropdownElement, value, hiddenSelect);
            dropdownElement.classList.remove('open');
            toggle.setAttribute('aria-expanded', 'false');
            
            // Trigger change event on hidden select
            const event = new Event('change', { bubbles: true });
            hiddenSelect.dispatchEvent(event);
        });
    });
    
    // Close on outside click
    document.addEventListener('click', (e) => {
        if (!dropdownElement.contains(e.target)) {
            dropdownElement.classList.remove('open');
            toggle.setAttribute('aria-expanded', 'false');
        }
    });
    
    // Close on Escape key
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && dropdownElement.classList.contains('open')) {
            dropdownElement.classList.remove('open');
            toggle.setAttribute('aria-expanded', 'false');
        }
    });
}

// Set dropdown value
function setDropdownValue(dropdownElement, value, hiddenSelect) {
    const selectedText = dropdownElement.querySelector(`.dropdown-item[data-value="${value}"] span`)?.textContent || '';
    const selectedSpan = dropdownElement.querySelector('.dropdown-selected');
    const items = dropdownElement.querySelectorAll('.dropdown-item');
    
    // Update selected text
    if (selectedSpan) {
        selectedSpan.textContent = selectedText;
    }
    
    // Update active state
    items.forEach(item => {
        const itemValue = item.dataset.value;
        const check = item.querySelector('.dropdown-check');
        
        if (itemValue === value) {
            item.classList.add('active');
            if (check) check.style.display = 'block';
        } else {
            item.classList.remove('active');
            if (check) check.style.display = 'none';
        }
    });
    
    // Update hidden select
    if (hiddenSelect) {
        hiddenSelect.value = value;
    }
}

// Make functions available globally for onclick handlers
// Copy to clipboard function
async function copyToClipboard(elementId, button) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    const text = element.textContent || element.innerText;
    
    try {
        await navigator.clipboard.writeText(text);
        
        // Update button to show success
        const copyText = button.querySelector('.copy-text');
        const originalText = copyText.textContent;
        copyText.textContent = 'Copied!';
        button.classList.add('copied');
        
        setTimeout(() => {
            copyText.textContent = originalText;
            button.classList.remove('copied');
        }, 2000);
    } catch (err) {
        console.error('Failed to copy:', err);
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = text;
        textArea.style.position = 'fixed';
        textArea.style.opacity = '0';
        document.body.appendChild(textArea);
        textArea.select();
        try {
            document.execCommand('copy');
            const copyText = button.querySelector('.copy-text');
            const originalText = copyText.textContent;
            copyText.textContent = 'Copied!';
            button.classList.add('copied');
            setTimeout(() => {
                copyText.textContent = originalText;
                button.classList.remove('copied');
            }, 2000);
        } catch (fallbackErr) {
            console.error('Fallback copy failed:', fallbackErr);
        }
        document.body.removeChild(textArea);
    }
}

// Switch installation tab
function switchInstallTab(tabName) {
    // Hide all tab contents
    document.querySelectorAll('.install-tab-content').forEach(content => {
        content.classList.remove('active');
    });
    
    // Remove active class from all tabs
    document.querySelectorAll('.install-tab').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Show selected tab content
    const targetContent = document.getElementById(`install-${tabName}`);
    if (targetContent) {
        targetContent.classList.add('active');
    }
    
    // Add active class to clicked tab
    const clickedTab = document.querySelector(`[data-tab="${tabName}"]`);
    if (clickedTab) {
        clickedTab.classList.add('active');
    }
}

// Admin functions
let repositoryDomain = 'http://localhost:7080';
let repositoryName = 'pertisk';

async function loadAdminData() {
    // Check if user is admin before loading admin config
    if (!state.token || !state.user || !state.user.is_admin) {
        console.warn('User is not an admin, cannot load admin configuration');
        return;
    }

    // Load all configuration
    try {
        const response = await fetch('/api/admin/config', {
            headers: {
                'Authorization': `Bearer ${state.token}`
            }
        });
        
        if (response.ok) {
            const config = await response.json();
            
            // Load domain configuration
            if (config.domain) {
                repositoryDomain = config.domain;
                const domainInput = document.getElementById('domainInput');
                if (domainInput) {
                    domainInput.value = config.domain;
                    updateDomainPreview(config.domain);
                }
            }
            
            // Load system settings
            const siteNameInput = document.getElementById('siteNameInput');
            const siteDescriptionInput = document.getElementById('siteDescriptionInput');
            const repositoryNameInput = document.getElementById('repositoryNameInput');
            
            if (siteNameInput && config.site_name) {
                siteNameInput.value = config.site_name;
            }
            if (siteDescriptionInput && config.site_description) {
                siteDescriptionInput.value = config.site_description;
            }
            if (repositoryNameInput && config.repository_name) {
                repositoryNameInput.value = config.repository_name;
                repositoryName = config.repository_name;
            }
        } else if (response.status === 403) {
            console.warn('Access denied: User is not an admin');
        } else {
            console.error('Failed to load admin configuration:', response.status);
        }
    } catch (error) {
        console.error('Failed to load configuration:', error);
    }

    // Load users if on users tab
    const usersTab = document.getElementById('admin-users');
    if (usersTab && usersTab.classList.contains('active')) {
        loadUsers();
    }
}

function updateDomainPreview(domain) {
    const repoUrlPreview = document.getElementById('repoUrlPreview');
    const helmCommandPreview = document.getElementById('helmCommandPreview');
    const repositoryName = document.getElementById('repositoryNameInput')?.value || 'pertisk';
    
    if (repoUrlPreview) {
        repoUrlPreview.textContent = domain || 'http://localhost:7080';
    }
    if (helmCommandPreview) {
        const domainValue = domain || 'http://localhost:7080';
        helmCommandPreview.textContent = `helm repo add ${repositoryName} ${domainValue}`;
    }
}

async function loadUsers() {
    const usersList = document.getElementById('usersList');
    if (!usersList) return;

    try {
        const response = await fetch('/api/admin/users', {
            headers: {
                'Authorization': `Bearer ${state.token}`
            }
        });

        if (response.ok) {
            const users = await response.json();
            renderUsers(users);
        } else {
            usersList.innerHTML = '<div class="error-state">Failed to load users</div>';
        }
    } catch (error) {
        usersList.innerHTML = '<div class="error-state">Error loading users</div>';
    }
}

function renderUsers(users) {
    const usersList = document.getElementById('usersList');
    if (!usersList) return;

    if (users.length === 0) {
        usersList.innerHTML = '<div class="empty-state">No users found</div>';
        return;
    }

    usersList.innerHTML = users.map(user => `
        <div class="user-item">
            <div class="user-info">
                <div class="user-name">${escapeHtml(user.username)}</div>
                <div class="user-email">${escapeHtml(user.email)}</div>
                <div class="user-meta">
                    <span class="user-date">Joined: ${new Date(user.created_at).toLocaleDateString()}</span>
                </div>
            </div>
            <div class="user-actions">
                <label class="admin-toggle">
                    <input type="checkbox" ${user.is_admin ? 'checked' : ''} 
                           onchange="toggleAdmin('${user.id}', this.checked)">
                    <span>Admin</span>
                </label>
            </div>
        </div>
    `).join('');
}

async function toggleAdmin(userId, isAdmin) {
    try {
        const response = await fetch(`/api/admin/users/${userId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${state.token}`
            },
            body: JSON.stringify({ is_admin: isAdmin })
        });

        if (!response.ok) {
            const data = await response.json();
            alert(`Failed to update user: ${data.error}`);
            // Reload users to revert UI
            loadUsers();
        }
    } catch (error) {
        alert(`Error updating user: ${error.message}`);
        loadUsers();
    }
}

function switchAdminTab(tabName) {
    // Hide all tab contents
    document.querySelectorAll('.admin-tab-content').forEach(content => {
        content.classList.remove('active');
    });
    
    // Remove active class from all tabs
    document.querySelectorAll('.admin-tab').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Show selected tab content
    const targetContent = document.getElementById(`admin-${tabName}`);
    if (targetContent) {
        targetContent.classList.add('active');
    }
    
    // Add active class to clicked tab
    const clickedTab = document.querySelector(`.admin-tab[data-tab="${tabName}"]`);
    if (clickedTab) {
        clickedTab.classList.add('active');
    }

    // Load data for the tab
    if (tabName === 'users') {
        loadUsers();
    } else if (tabName === 'domain') {
        // Update domain preview when switching to domain tab
        const domainInput = document.getElementById('domainInput');
        if (domainInput) {
            updateDomainPreview(domainInput.value.trim() || repositoryDomain);
        }
    }
}

// Setup admin forms
document.addEventListener('DOMContentLoaded', () => {
    // Domain form
    const domainForm = document.getElementById('domainForm');
    if (domainForm) {
        const domainInput = document.getElementById('domainInput');
        const repositoryNameInput = document.getElementById('repositoryNameInput');
        
        // Update preview on input change
        if (domainInput) {
            domainInput.addEventListener('input', () => {
                updateDomainPreview(domainInput.value.trim() || repositoryDomain);
            });
        }
        if (repositoryNameInput) {
            repositoryNameInput.addEventListener('input', () => {
                updateDomainPreview(domainInput?.value.trim() || repositoryDomain);
            });
        }
        
        domainForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const statusDiv = document.getElementById('domainStatus');
            
            if (!domainInput || !statusDiv) return;

            const domain = domainInput.value.trim();
            if (!domain) {
                showStatus(statusDiv, 'Domain is required', 'error');
                return;
            }

            // Validate URL format
            try {
                new URL(domain);
            } catch (err) {
                showStatus(statusDiv, 'Invalid URL format. Please include protocol (http:// or https://)', 'error');
                return;
            }

            try {
                const response = await fetch('/api/admin/config', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${state.token}`
                    },
                    body: JSON.stringify({ key: 'domain', value: domain })
                });

                if (response.ok) {
                    const data = await response.json();
                    repositoryDomain = data.value;
                    updateDomainPreview(data.value);
                    showStatus(statusDiv, 'Domain updated successfully!', 'success');
                } else {
                    const data = await response.json();
                    showStatus(statusDiv, `Error: ${data.error}`, 'error');
                }
            } catch (error) {
                showStatus(statusDiv, `Error: ${error.message}`, 'error');
            }
        });
    }

    // System settings form
    const systemForm = document.getElementById('systemForm');
    if (systemForm) {
        systemForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const statusDiv = document.getElementById('systemStatus');
            const siteNameInput = document.getElementById('siteNameInput');
            const siteDescriptionInput = document.getElementById('siteDescriptionInput');
            const repositoryNameInput = document.getElementById('repositoryNameInput');
            
            if (!statusDiv) return;

            const updates = [];
            
            if (siteNameInput && siteNameInput.value.trim()) {
                updates.push({ key: 'site_name', value: siteNameInput.value.trim() });
            }
            if (siteDescriptionInput && siteDescriptionInput.value.trim()) {
                updates.push({ key: 'site_description', value: siteDescriptionInput.value.trim() });
            }
                    if (repositoryNameInput && repositoryNameInput.value.trim()) {
                        const repoName = repositoryNameInput.value.trim();
                        updates.push({ key: 'repository_name', value: repoName });
                        repositoryName = repoName; // Update global variable
                    }

            if (updates.length === 0) {
                showStatus(statusDiv, 'No changes to save', 'info');
                return;
            }

            try {
                // Save all settings
                const promises = updates.map(update => 
                    fetch('/api/admin/config', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'Authorization': `Bearer ${state.token}`
                        },
                        body: JSON.stringify(update)
                    })
                );

                const results = await Promise.all(promises);
                const allOk = results.every(r => r.ok);

                if (allOk) {
                    showStatus(statusDiv, 'Settings saved successfully!', 'success');
                    // Update domain preview if repository name changed
                    if (repositoryNameInput && repositoryNameInput.value.trim()) {
                        updateDomainPreview(repositoryDomain);
                    }
                } else {
                    showStatus(statusDiv, 'Some settings failed to save', 'error');
                }
            } catch (error) {
                showStatus(statusDiv, `Error: ${error.message}`, 'error');
            }
        });
    }

    // Admin data will be loaded when admin view is shown or during initialization
    // No need to load here as it requires admin authentication
});

window.showChartDetail = showChartDetail;
window.deleteChart = deleteChart;
window.showView = showView;
// Copy YAML content from CodeMirror editor
async function copyYamlContent(editorId, button) {
    const textarea = document.getElementById(editorId);
    if (!textarea) return;
    
    let text = '';
    if (textarea.cmEditor) {
        // CodeMirror editor
        text = textarea.cmEditor.getValue();
    } else {
        // Fallback to textarea value
        text = textarea.value;
    }
    
    try {
        await navigator.clipboard.writeText(text);
        
        // Update button to show success
        const copyText = button.querySelector('.copy-text');
        if (copyText) {
            const originalText = copyText.textContent;
            copyText.textContent = 'Copied!';
            setTimeout(() => {
                copyText.textContent = originalText;
            }, 2000);
        }
    } catch (err) {
        console.error('Failed to copy:', err);
        // Fallback for older browsers
        const fallbackTextarea = document.createElement('textarea');
        fallbackTextarea.value = text;
        fallbackTextarea.style.position = 'fixed';
        fallbackTextarea.style.opacity = '0';
        document.body.appendChild(fallbackTextarea);
        fallbackTextarea.select();
        try {
            document.execCommand('copy');
            const copyText = button.querySelector('.copy-text');
            if (copyText) {
                const originalText = copyText.textContent;
                copyText.textContent = 'Copied!';
                setTimeout(() => {
                    copyText.textContent = originalText;
                }, 2000);
            }
        } catch (e) {
            console.error('Fallback copy failed:', e);
        }
        document.body.removeChild(fallbackTextarea);
    }
}

window.copyToClipboard = copyToClipboard;
window.copyYamlContent = copyYamlContent;
window.switchInstallTab = switchInstallTab;
window.switchAdminTab = switchAdminTab;
window.toggleAdmin = toggleAdmin;

// Auto-show charts view when searching
document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            if (e.target.value && state.currentView === 'home') {
                showView('charts');
            }
        });
    }
});

// Theme Management
function initializeTheme() {
    const savedTheme = localStorage.getItem('theme') || 'light';
    setTheme(savedTheme);
}

function setTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
    updateThemeIcon(theme);
    // Update CodeMirror theme for all editors
    updateCodeMirrorThemes(theme);
}

// Update CodeMirror themes when theme changes
function updateCodeMirrorThemes(theme) {
    // Find all CodeMirror editors
    const textareas = document.querySelectorAll('.yaml-editor-textarea');
    textareas.forEach(textarea => {
        if (textarea.cmEditor) {
            const editor = textarea.cmEditor;
            const newTheme = theme === 'dark' ? 'monokai' : 'default';
            editor.setOption('theme', newTheme);
            editor.refresh();
        }
    });
}

function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    setTheme(newTheme);
}

function updateThemeIcon(theme) {
    const lightIcon = document.getElementById('themeIconLight');
    const darkIcon = document.getElementById('themeIconDark');
    
    if (lightIcon && darkIcon) {
        if (theme === 'dark') {
            lightIcon.style.display = 'block';
            darkIcon.style.display = 'none';
        } else {
            lightIcon.style.display = 'none';
            darkIcon.style.display = 'block';
        }
    }
}

// Make toggleTheme available globally
window.toggleTheme = toggleTheme;


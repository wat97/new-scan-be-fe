// ===== Configuration =====
// Use relative URL when served via nginx (Docker), otherwise use localhost for dev
const API_BASE = window.location.hostname === 'localhost' && window.location.port === '3000'
    ? 'http://localhost:8080/api'
    : '/api';
let currentUser = null;
let token = null;
let qrScanner = null;
let currentScannedBarcode = null;
let weeklyChart = null;

// ===== Initialization =====
document.addEventListener('DOMContentLoaded', () => {
    checkAuth();
    setupEventListeners();
});

function checkAuth() {
    token = localStorage.getItem('token');
    const userData = localStorage.getItem('user');

    if (token && userData) {
        currentUser = JSON.parse(userData);
        showMainApp();
    } else {
        showLoginPage();
    }
}

function showLoginPage() {
    document.getElementById('loginPage').classList.add('active');
    document.getElementById('mainApp').classList.remove('active');
}

function showMainApp() {
    document.getElementById('loginPage').classList.remove('active');
    document.getElementById('mainApp').classList.add('active');

    // Set user info
    document.getElementById('userName').textContent = currentUser.name;
    document.getElementById('userRole').textContent = currentUser.role === 'admin' ? 'Administrator' : 'User';
    document.getElementById('userAvatar').textContent = currentUser.name.charAt(0).toUpperCase();

    // Set mobile user button
    const mobileUserBtn = document.getElementById('mobileUserBtn');
    if (mobileUserBtn) {
        mobileUserBtn.textContent = currentUser.name.charAt(0).toUpperCase();
    }

    // Show/hide admin menus
    if (currentUser.role === 'admin') {
        document.body.classList.add('is-admin');
    } else {
        document.body.classList.remove('is-admin');
    }

    // Load scanner as main page (instead of dashboard)
    navigateTo('scanner');
}

// ===== Event Listeners =====
function setupEventListeners() {
    // Login form
    document.getElementById('loginForm').addEventListener('submit', handleLogin);

    // Logout
    document.getElementById('logoutBtn').addEventListener('click', handleLogout);

    // Navigation
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', () => {
            const page = item.dataset.page;
            navigateTo(page);
        });
    });

    // Scanner buttons
    document.getElementById('btnSesuai')?.addEventListener('click', () => submitScan(true));
    document.getElementById('btnTidakSesuai')?.addEventListener('click', () => submitScan(false));

    // History filter
    document.getElementById('btnFilterHistory')?.addEventListener('click', loadHistory);

    // Add unit button
    document.getElementById('btnAddUnit')?.addEventListener('click', showAddUnitModal);

    // Add user button
    document.getElementById('btnAddUser')?.addEventListener('click', showAddUserModal);

    // Export button
    document.getElementById('btnExport')?.addEventListener('click', exportToExcel);

    // Modal
    document.getElementById('modalClose')?.addEventListener('click', closeModal);
    document.getElementById('modal')?.addEventListener('click', (e) => {
        if (e.target.id === 'modal') closeModal();
    });

    // Mobile Menu
    document.getElementById('mobileMenuBtn')?.addEventListener('click', toggleMobileMenu);
    document.getElementById('sidebarOverlay')?.addEventListener('click', closeMobileMenu);

    // Bottom Navigation
    document.querySelectorAll('.bottom-nav-item').forEach(item => {
        item.addEventListener('click', () => {
            const page = item.dataset.page;
            navigateTo(page);
            updateBottomNav(page);
        });
    });

    // FAB buttons for mobile
    document.getElementById('fabAddUnit')?.addEventListener('click', showAddUnitModal);
    document.getElementById('fabAddUser')?.addEventListener('click', showAddUserModal);
}

// ===== Mobile Menu =====
function toggleMobileMenu() {
    const sidebar = document.querySelector('.sidebar');
    const overlay = document.getElementById('sidebarOverlay');
    sidebar.classList.toggle('open');
    overlay.classList.toggle('active');
}

function closeMobileMenu() {
    const sidebar = document.querySelector('.sidebar');
    const overlay = document.getElementById('sidebarOverlay');
    sidebar.classList.remove('open');
    overlay.classList.remove('active');
}

function updateBottomNav(page) {
    document.querySelectorAll('.bottom-nav-item').forEach(item => {
        item.classList.toggle('active', item.dataset.page === page);
    });
}

// ===== Navigation =====
let currentPage = null;

function navigateTo(page) {
    // Stop scanner if leaving scanner page
    if (currentPage === 'scanner' && page !== 'scanner') {
        stopScanner();
    }

    currentPage = page;

    // Update sidebar nav
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.toggle('active', item.dataset.page === page);
    });

    // Update bottom nav
    updateBottomNav(page);

    // Close mobile menu if open
    closeMobileMenu();

    // Update content
    document.querySelectorAll('.content-page').forEach(p => {
        p.classList.remove('active');
    });
    document.getElementById(`${page}Page`)?.classList.add('active');

    // Load page data
    switch (page) {
        case 'dashboard':
            loadDashboard();
            break;
        case 'scanner':
            initScanner();
            break;
        case 'history':
            loadHistory();
            break;
        case 'users':
            loadUsers();
            break;
        case 'reports':
            loadReports();
            break;
    }
}

// ===== Auth =====
async function handleLogin(e) {
    e.preventDefault();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const errorDiv = document.getElementById('loginError');

    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();

        if (!response.ok) {
            errorDiv.textContent = data.error || 'Login gagal';
            return;
        }

        token = data.token;
        currentUser = data.user;

        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify(currentUser));

        errorDiv.textContent = '';
        showMainApp();

    } catch (error) {
        errorDiv.textContent = 'Tidak dapat terhubung ke server';
    }
}

function handleLogout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    token = null;
    currentUser = null;
    currentPage = null;

    // Stop camera if active
    stopScanner();

    showLoginPage();
}

// ===== API Helper =====
async function apiFetch(endpoint, options = {}) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
            ...options.headers
        }
    });

    if (response.status === 401) {
        handleLogout();
        throw new Error('Unauthorized');
    }

    return response;
}

// ===== Dashboard =====
async function loadDashboard() {
    try {
        const response = await apiFetch('/reports/summary');
        const data = await response.json();

        document.getElementById('statTotalToday').textContent = data.today?.total || 0;
        document.getElementById('statMatchToday').textContent = data.today?.match || 0;
        document.getElementById('statNotMatchToday').textContent = data.today?.not_match || 0;

        loadWeeklyChart();

    } catch (error) {
        console.error('Error loading dashboard:', error);
    }
}

async function loadWeeklyChart() {
    try {
        const response = await apiFetch('/reports/daily?days=7');
        const data = await response.json();

        const labels = data.map(d => d.date).reverse();
        const matchData = data.map(d => d.match).reverse();
        const notMatchData = data.map(d => d.not_match).reverse();

        const ctx = document.getElementById('weeklyChart').getContext('2d');

        if (weeklyChart) {
            weeklyChart.destroy();
        }

        weeklyChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'Sesuai',
                        data: matchData,
                        backgroundColor: 'rgba(16, 185, 129, 0.8)',
                        borderRadius: 6
                    },
                    {
                        label: 'Tidak Sesuai',
                        data: notMatchData,
                        backgroundColor: 'rgba(239, 68, 68, 0.8)',
                        borderRadius: 6
                    }
                ]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: {
                        labels: { color: '#94a3b8' }
                    }
                },
                scales: {
                    x: {
                        ticks: { color: '#94a3b8' },
                        grid: { color: 'rgba(148, 163, 184, 0.1)' }
                    },
                    y: {
                        ticks: { color: '#94a3b8' },
                        grid: { color: 'rgba(148, 163, 184, 0.1)' }
                    }
                }
            }
        });

    } catch (error) {
        console.error('Error loading chart:', error);
    }
}

// ===== Scanner =====
let scannerActive = false;
let scannerInitialized = false;

async function initScanner() {
    const scanResult = document.getElementById('scanResult');
    scanResult.classList.add('hidden');

    // If scanner is already active, don't reinitialize
    if (scannerActive && qrScanner) {
        console.log('Scanner already active, skipping init');
        return;
    }

    // Create scanner instance only once
    if (!qrScanner) {
        qrScanner = new Html5Qrcode("qrReader");
        scannerInitialized = true;
        console.log('Created new Html5Qrcode instance');
    }

    // Check if already scanning
    try {
        const state = qrScanner.getState();
        if (state === 2) { // Already scanning
            console.log('Scanner already running');
            scannerActive = true;
            return;
        }
    } catch (e) {
        // getState might fail if not initialized properly
    }

    try {
        await qrScanner.start(
            { facingMode: "environment" },
            {
                fps: 10,
                aspectRatio: 1.333333, // 4:3 aspect ratio
                disableFlip: false
            },
            onScanSuccess,
            () => { }
        );
        scannerActive = true;
        console.log('Scanner started');
    } catch (err) {
        console.error('Error starting scanner:', err);
        showToast('Tidak dapat mengakses kamera', 'error');
        scannerActive = false;
    }
}

async function stopScanner() {
    console.log('stopScanner called, scannerActive:', scannerActive);

    if (qrScanner && scannerActive) {
        try {
            const state = qrScanner.getState();
            console.log('Scanner state:', state);

            // Only stop if scanning (state 2) or paused (state 3)
            if (state === 2 || state === 3) {
                await qrScanner.stop();
                console.log('Scanner stopped successfully');
            }
            scannerActive = false;
        } catch (e) {
            console.log('Error stopping scanner:', e);
            scannerActive = false;
        }
    }
    // Don't set qrScanner to null - keep the instance for reuse
    // This prevents repeated permission requests
}

async function onScanSuccess(barcode) {
    try {
        await qrScanner.stop();
    } catch (e) { }

    // Simpan barcode yang di-scan untuk digunakan saat submit
    currentScannedBarcode = barcode;

    // Tampilkan hasil scan langsung tanpa lookup unit
    document.getElementById('scannedUnitName').textContent = barcode;
    document.getElementById('scannedUnitQR').textContent = 'Barcode berhasil di-scan';
    document.getElementById('scannedUnitLocation').textContent = '';
    document.getElementById('scannedUnitGrade').textContent = 'Apakah grade sesuai?';

    document.getElementById('scanResult').classList.remove('hidden');
    document.getElementById('scanNotes').value = '';
}

async function submitScan(isMatch) {
    if (!currentScannedBarcode) return;

    const notes = document.getElementById('scanNotes').value;

    try {
        const response = await apiFetch('/scans', {
            method: 'POST',
            body: JSON.stringify({
                barcode: currentScannedBarcode,
                is_match: isMatch,
                notes: notes
            })
        });

        if (response.ok) {
            showToast(isMatch ? 'Ditandai SESUAI' : 'Ditandai TIDAK SESUAI', 'success');
            currentScannedBarcode = null;
            document.getElementById('scanResult').classList.add('hidden');
            setTimeout(initScanner, 1500);
        } else {
            const data = await response.json();
            showToast(data.error || 'Error menyimpan scan', 'error');
        }

    } catch (error) {
        showToast('Error menyimpan scan', 'error');
    }
}

// ===== History =====
async function loadHistory() {
    const date = document.getElementById('filterDate').value;
    const isMatch = document.getElementById('filterMatch').value;

    let params = new URLSearchParams();
    if (date) params.append('date', date);
    if (isMatch) params.append('is_match', isMatch);

    try {
        const response = await apiFetch(`/scans?${params.toString()}`);
        const scans = await response.json();

        const tbody = document.getElementById('historyTableBody');
        tbody.innerHTML = scans.map(scan => `
            <tr>
                <td>${formatDateTime(scan.scanned_at)}</td>
                <td>${scan.barcode || '-'}</td>
                <td>
                    <span class="badge ${scan.is_match ? 'badge-success' : 'badge-danger'}">
                        ${scan.is_match ? 'Sesuai' : 'Tidak Sesuai'}
                    </span>
                </td>
                <td>${scan.user?.name || '-'}</td>
                <td>${scan.notes || '-'}</td>
            </tr>
        `).join('');

    } catch (error) {
        console.error('Error loading history:', error);
    }
}


// ===== Users =====
async function loadUsers() {
    try {
        const response = await apiFetch('/users');
        const users = await response.json();

        // Desktop table view
        const tbody = document.getElementById('usersTableBody');
        tbody.innerHTML = users.map(user => `
            <tr>
                <td>${user.username}</td>
                <td>${user.name}</td>
                <td>
                    <span class="badge ${user.role === 'admin' ? 'badge-primary' : 'badge-success'}">
                        ${user.role === 'admin' ? 'Admin' : 'User'}
                    </span>
                </td>
                <td class="action-buttons">
                    <button class="btn btn-sm btn-secondary" onclick="editUser(${user.id})">Edit</button>
                    <button class="btn btn-sm btn-danger" onclick="deleteUser(${user.id})">Hapus</button>
                </td>
            </tr>
        `).join('');

        // Mobile card view
        const mobileCards = document.getElementById('usersMobileCards');
        if (mobileCards) {
            if (users.length === 0) {
                mobileCards.innerHTML = `
                    <div class="empty-state">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                            <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" />
                            <circle cx="9" cy="7" r="4" />
                            <path d="M23 21v-2a4 4 0 00-3-3.87" />
                            <path d="M16 3.13a4 4 0 010 7.75" />
                        </svg>
                        <h3>Belum ada user</h3>
                        <p>Tap tombol + untuk menambahkan user baru</p>
                    </div>
                `;
            } else {
                mobileCards.innerHTML = users.map(user => `
                    <div class="mobile-data-card" onclick="editUser(${user.id})">
                        <div class="mobile-card-header">
                            <div>
                                <div class="mobile-card-title">${user.name}</div>
                                <div class="mobile-card-subtitle">@${user.username}</div>
                            </div>
                            <span class="badge ${user.role === 'admin' ? 'badge-primary' : 'badge-success'}">
                                ${user.role === 'admin' ? 'Admin' : 'User'}
                            </span>
                        </div>
                        <div class="mobile-card-actions">
                            <button class="btn btn-secondary btn-sm" onclick="event.stopPropagation(); editUser(${user.id})">
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                                    <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" />
                                    <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
                                </svg>
                                Edit
                            </button>
                            <button class="btn btn-danger btn-sm" onclick="event.stopPropagation(); deleteUser(${user.id})">
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                                    <polyline points="3,6 5,6 21,6" />
                                    <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2" />
                                </svg>
                                Hapus
                            </button>
                        </div>
                    </div>
                `).join('');
            }
        }

    } catch (error) {
        console.error('Error loading users:', error);
    }
}

function showAddUserModal() {
    document.getElementById('modalTitle').textContent = 'Tambah User';
    document.getElementById('modalBody').innerHTML = `
        <form id="userForm">
            <div class="form-group">
                <label>Username</label>
                <input type="text" name="username" required placeholder="Username">
            </div>
            <div class="form-group">
                <label>Password</label>
                <input type="password" name="password" required placeholder="Password min 6 karakter" minlength="6">
            </div>
            <div class="form-group">
                <label>Nama</label>
                <input type="text" name="name" required placeholder="Nama lengkap">
            </div>
            <div class="form-group">
                <label>Role</label>
                <select name="role" required>
                    <option value="user">User</option>
                    <option value="admin">Admin</option>
                </select>
            </div>
            <button type="submit" class="btn btn-primary btn-block">Simpan</button>
        </form>
    `;

    document.getElementById('userForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        try {
            const response = await apiFetch('/users', {
                method: 'POST',
                body: JSON.stringify(Object.fromEntries(formData))
            });

            if (response.ok) {
                showToast('User berhasil ditambahkan', 'success');
                closeModal();
                loadUsers();
            } else {
                const data = await response.json();
                showToast(data.error || 'Error', 'error');
            }
        } catch (error) {
            showToast('Error', 'error');
        }
    });

    openModal();
}

async function editUser(id) {
    try {
        const response = await apiFetch(`/users/${id}`);
        const user = await response.json();

        document.getElementById('modalTitle').textContent = 'Edit User';
        document.getElementById('modalBody').innerHTML = `
            <form id="userForm">
                <div class="form-group">
                    <label>Nama</label>
                    <input type="text" name="name" value="${user.name}" required placeholder="Nama lengkap">
                </div>
                <div class="form-group">
                    <label>Password Baru (kosongkan jika tidak diubah)</label>
                    <input type="password" name="password" placeholder="Password min 6 karakter" minlength="6">
                </div>
                <div class="form-group">
                    <label>Role</label>
                    <select name="role" required>
                        <option value="user" ${user.role === 'user' ? 'selected' : ''}>User</option>
                        <option value="admin" ${user.role === 'admin' ? 'selected' : ''}>Admin</option>
                    </select>
                </div>
                <button type="submit" class="btn btn-primary btn-block">Update</button>
            </form>
        `;

        document.getElementById('userForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);

            // Remove empty password
            if (!data.password) delete data.password;

            try {
                const response = await apiFetch(`/users/${id}`, {
                    method: 'PUT',
                    body: JSON.stringify(data)
                });

                if (response.ok) {
                    showToast('User berhasil diupdate', 'success');
                    closeModal();
                    loadUsers();
                } else {
                    const data = await response.json();
                    showToast(data.error || 'Error', 'error');
                }
            } catch (error) {
                showToast('Error', 'error');
            }
        });

        openModal();

    } catch (error) {
        showToast('Error', 'error');
    }
}

async function deleteUser(id) {
    if (!confirm('Yakin ingin menghapus user ini?')) return;

    try {
        const response = await apiFetch(`/users/${id}`, { method: 'DELETE' });

        if (response.ok) {
            showToast('User berhasil dihapus', 'success');
            loadUsers();
        } else {
            const data = await response.json();
            showToast(data.error || 'Error', 'error');
        }
    } catch (error) {
        showToast('Error', 'error');
    }
}

// ===== Reports =====
async function loadReports() {
    try {
        const response = await apiFetch('/reports/summary');
        const data = await response.json();

        document.getElementById('weekTotal').textContent = data.week?.total || 0;
        document.getElementById('weekMatch').textContent = data.week?.match || 0;
        document.getElementById('weekNotMatch').textContent = data.week?.not_match || 0;

        document.getElementById('monthTotal').textContent = data.month?.total || 0;
        document.getElementById('monthMatch').textContent = data.month?.match || 0;
        document.getElementById('monthNotMatch').textContent = data.month?.not_match || 0;

        if (currentUser?.role === 'admin') {
            loadUserPerformance();
        }

    } catch (error) {
        console.error('Error loading reports:', error);
    }
}

async function loadUserPerformance() {
    try {
        const response = await apiFetch('/reports/users');
        const users = await response.json();

        const tbody = document.getElementById('userPerformanceBody');
        tbody.innerHTML = users.map(user => `
            <tr>
                <td>${user.user_name}</td>
                <td>${user.total}</td>
                <td class="match">${user.match}</td>
                <td class="notmatch">${user.not_match}</td>
            </tr>
        `).join('');

    } catch (error) {
        console.error('Error loading user performance:', error);
    }
}

async function exportToExcel() {
    try {
        const response = await fetch(`${API_BASE}/reports/export`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `scan_report_${new Date().toISOString().split('T')[0]}.xlsx`;
            a.click();
            window.URL.revokeObjectURL(url);
            showToast('Export berhasil', 'success');
        } else {
            showToast('Export gagal', 'error');
        }
    } catch (error) {
        showToast('Export gagal', 'error');
    }
}

// ===== Modal =====
function openModal() {
    document.getElementById('modal').classList.add('active');
}

function closeModal() {
    document.getElementById('modal').classList.remove('active');
}

// ===== Toast =====
function showToast(message, type = 'info') {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = `toast show ${type}`;

    setTimeout(() => {
        toast.classList.remove('show');
    }, 3000);
}

// ===== Utilities =====
function formatDateTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('id-ID', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

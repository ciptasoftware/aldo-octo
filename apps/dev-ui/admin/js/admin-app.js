// admin-app.js
// Router utama untuk mengontrol SPA (Single Page Application) tanpa memuat ulang layar

import { LoginView } from './views/login.js';
import { DashboardView } from './views/dashboard.js';
import { ProductsView } from './views/products.js';
import { CategoriesView } from './views/categories.js';
import { BrandsView } from './views/brands.js';

// DOM Elements
const appRoot = document.getElementById('app-root');

// Mendefinisikan Rute URL Hash ke Komponen View
const routes = {
    '': DashboardView,
    '#dashboard': DashboardView,
    '#products': ProductsView,
    '#categories': CategoriesView,
    '#brands': BrandsView,
    '#login': LoginView
};

/**
 * Merender Skeleton Tata Letak Dasar (Sidebar + Header)
 * Ini hanya untuk menu yang terkunci auth. Login view menimpa ini.
 */
export function renderAdminLayout(contentHTML) {
    const user = JSON.parse(localStorage.getItem('admin_user') || '{}');
    
    return `
        <div class="admin-layout sidebar-collapsed" id="adminMainLayout">
            <!-- Sidebar -->
            <aside class="sidebar">
                <div class="sidebar-brand">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 16V7a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v9m16 0H4m16 0 1.28 2.55a1 1 0 0 1-.9 1.45H3.62a1 1 0 0 1-.9-1.45L4 16"/></svg>
                    <span>OBS Admin</span>
                </div>
                <nav class="sidebar-nav">
                    <a href="#dashboard" class="nav-link ${location.hash==='' || location.hash==='#dashboard' ? 'active' : ''}">
                        <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
                        <span>Dashboard</span>
                    </a>
                    <a href="#products" class="nav-link ${location.hash==='#products' ? 'active' : ''}">
                        <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path><line x1="7" y1="7" x2="7.01" y2="7"></line></svg>
                        <span>Products</span>
                    </a>
                    <a href="#categories" class="nav-link ${location.hash==='#categories' ? 'active' : ''}">
                        <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
                        <span>Categories</span>
                    </a>
                    <a href="#brands" class="nav-link ${location.hash==='#brands' ? 'active' : ''}">
                        <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>
                        <span>Brands</span>
                    </a>
                    <a href="#orders" class="nav-link ${location.hash==='#orders' ? 'active' : ''}">
                        <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="9" cy="21" r="1"></circle><circle cx="20" cy="21" r="1"></circle><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path></svg>
                        <span>Orders</span>
                    </a>
                </nav>
                <div class="sidebar-footer">
                    <button id="btnLogout" class="btn-logout-sidebar" title="Keluar">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                            <polyline points="16 17 21 12 16 7"></polyline>
                            <line x1="21" y1="12" x2="9" y2="12"></line>
                        </svg>
                        <span>Logout</span>
                    </button>
                </div>
            </aside>

            <!-- Main Content Area -->
            <main class="main-content">
                <header class="topbar">
                    <div class="search-wrap" style="display:flex; align-items:center; gap: 1rem;">
                        <button id="btnToggleSidebar" class="btn btn-icon btn-outline" title="Toggle Menu">
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="6" x2="21" y2="6"></line><line x1="3" y1="18" x2="21" y2="18"></line></svg>
                        </button>
                        <!-- TBD ruang search global -->
                    </div>
                    <div class="user-profile">
                        <span style="font-size: 0.875rem; font-weight:500;">${user.name || 'Admin Pengelola'}</span>
                        <div class="avatar-circle">${(user.name || 'A').charAt(0)}</div>
                    </div>
                </header>

                <div class="content-wrapper" id="page-content">
                    ${contentHTML}
                </div>
            </main>
        </div>
    `;
}

/**
 * Menangani Inisiasi dan Peralihan Layar
 */
async function router() {
    let hash = window.location.hash;
    
    // Auth Guard (Cegah masuk Dasbor tanpa token)
    const token = localStorage.getItem('admin_token');
    if (!token && hash !== '#login') {
        window.location.hash = '#login';
        return;
    }

    // Hindari redirect manual ke `#login` jika user sudah login
    if (token && hash === '#login') {
        window.location.hash = '#dashboard';
        return;
    }

    // Resolusi Komponen View
    const ViewComponent = routes[hash] || DashboardView;

    // Bersihkan listener / interval dari view sebelumnya
    appRoot.innerHTML = ''; 

    try {
        // Init dan merender interface layar aktif
        await ViewComponent.render(appRoot);
        
        // Pasang Event Listener spesifik untuk Layout Admin (bila bukan di menu Login)
        if (hash !== '#login') {
            document.getElementById('btnToggleSidebar').addEventListener('click', () => {
                document.getElementById('adminMainLayout').classList.toggle('sidebar-collapsed');
            });
            
            document.getElementById('btnLogout').addEventListener('click', () => {
                localStorage.removeItem('admin_token');
                localStorage.removeItem('admin_user');
                window.location.hash = '#login';
            });
        }

    } catch (e) {
        console.error("Routing error: ", e);
        appRoot.innerHTML = `<h3 style="padding:2rem;">Fatal Error loading page. Check console.</h3>`;
    }
}

// Intersep perubahan URL
window.addEventListener('hashchange', router);

// Perintah Muat Awal saat file diload
window.addEventListener('DOMContentLoaded', router);

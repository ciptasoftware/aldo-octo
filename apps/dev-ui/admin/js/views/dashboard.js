// views/dashboard.js
import { renderAdminLayout } from '../admin-app.js';
import { apiClient } from '../api-client.js';

export const DashboardView = {
    async render(rootElement) {
        // Build kerangka dalam admin layout HTML
        const html = renderAdminLayout(`
            <div class="page-header">
                <div>
                    <h1 class="page-title">Dashboard Overview</h1>
                    <p style="color:var(--text-muted); margin-top:0.25rem;">Rangkuman eksekutif toko harian.</p>
                </div>
            </div>

            <div class="overview-grid">
                <div class="stat-card">
                    <div class="stat-icon">💰</div>
                    <div class="stat-details">
                        <h3>Total Pendapatan</h3>
                        <p id="statIncome">Rp 0</p>
                    </div>
                </div>
                <div class="stat-card">
                    <div class="stat-icon">📦</div>
                    <div class="stat-details">
                        <h3>Semua Pesanan</h3>
                        <p id="statOrders">0</p>
                    </div>
                </div>
                <div class="stat-card">
                    <div class="stat-icon">🏷️</div>
                    <div class="stat-details">
                        <h3>Total Produk Aktif</h3>
                        <p id="statProducts">0</p>
                    </div>
                </div>
                <div class="stat-card">
                    <div class="stat-icon">👥</div>
                    <div class="stat-details">
                        <h3>Jumlah Kategori</h3>
                        <p id="statCat">0</p>
                    </div>
                </div>
            </div>
            
            <div class="table-container" style="padding: 1.5rem; margin-top: 2rem;">
                <h3 style="margin-bottom:1rem; font-size:1.125rem;">Pesanan Menunggu Dikirim</h3>
                <p style="color:var(--text-muted); font-size:0.875rem;">(Implementasi Tabel Pesanan Ringkas menyusul)</p>
            </div>
        `);

        rootElement.innerHTML = html;

        // Fetch data stastik asli 
        try {
            // Paralel Fetch agar kencang
            const [productsRes, ordersRes, catRes] = await Promise.all([
                apiClient('/products?limit=1'), // Asumsi endpoint ini punya format standard paginasi dgn total records
                apiClient('/orders'), // All orders 
                apiClient('/categories')
            ]);
            
            // Render Products Count dari paginated response
            if (productsRes.total !== undefined) {
                document.getElementById('statProducts').textContent = productsRes.total;
            }

            // Render Orders Count dan Hitung Profit Total dari backend
            if (Array.isArray(ordersRes)) {
                document.getElementById('statOrders').textContent = ordersRes.length;
                let totalIncome = 0;
                ordersRes.forEach(o => {
                    // Hanya menghitung order sukses/selesai
                    if (o.status !== 'cancelled') {
                        totalIncome += o.total_amount;
                    }
                });
                
                document.getElementById('statIncome').textContent = new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(totalIncome);
            }

            if (Array.isArray(catRes)) {
                document.getElementById('statCat').textContent = catRes.length;
            }

        } catch (error) {
            console.error("Gagal menarik data untuk dasbor:", error);
        }
    }
};

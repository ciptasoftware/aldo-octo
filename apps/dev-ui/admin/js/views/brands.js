// views/brands.js
import { renderAdminLayout } from '../admin-app.js';
import { apiClient, showToast } from '../api-client.js';

export const BrandsView = {
    async render(rootElement) {
        const html = renderAdminLayout(`
            <div class="page-header">
                <div>
                    <h1 class="page-title">Manajemen Brand</h1>
                    <p style="color:var(--text-muted); margin-top:0.25rem;">Kelola daftar merk/brand yang tersedia di toko.</p>
                </div>
                <button class="btn btn-primary" id="btnShowAddBrand">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                    Tambah Brand
                </button>
            </div>

            <div class="table-container">
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th width="80">ID</th>
                            <th>Nama Brand</th>
                            <th>Slug</th>
                            <th>Logo URL</th>
                            <th width="100">Aksi</th>
                        </tr>
                    </thead>
                    <tbody id="tblBodyBrands">
                        <tr><td colspan="5" style="text-align:center; padding:2rem;">Loading data...</td></tr>
                    </tbody>
                </table>
            </div>

            <!-- Modal Form -->
            <div class="modal-overlay" id="modalBrandForm">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3 id="modalTitle">Buat Brand Baru</h3>
                        <button class="btn-close" id="btnCloseModal">&times;</button>
                    </div>
                    <form id="frmSaveBrand">
                        <div class="modal-body">
                            <div class="form-group">
                                <label>Nama Brand</label>
                                <input type="text" class="form-control" id="fName" required>
                            </div>
                            <div class="form-group">
                                <label>Slug (Identitas URL)</label>
                                <input type="text" class="form-control" id="fSlug" required>
                            </div>
                            <div class="form-group">
                                <label>Logo URL</label>
                                <input type="text" class="form-control" id="fLogo" placeholder="https://...">
                            </div>
                            <div class="form-group">
                                <label>Deskripsi</label>
                                <textarea class="form-control" id="fDesc" rows="3"></textarea>
                            </div>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-outline" id="btnCancelModal">Batal</button>
                            <button type="submit" class="btn btn-primary" id="btnSubmitForm">Simpan Brand</button>
                        </div>
                    </form>
                </div>
            </div>
        `);

        rootElement.innerHTML = html;

        const tbody = document.getElementById('tblBodyBrands');
        const modal = document.getElementById('modalBrandForm');
        let currentEditId = null;

        async function fetchBrands() {
            try {
                const brands = await apiClient('/brands');
                if (!brands || brands.length === 0) {
                    tbody.innerHTML = `<tr><td colspan="5" style="text-align:center; padding:2rem; color:var(--text-muted)">Tidak ada data brand.</td></tr>`;
                    return;
                }

                tbody.innerHTML = brands.map(b => `
                    <tr>
                        <td>#${b.id}</td>
                        <td><strong>${b.name}</strong></td>
                        <td><code>${b.slug}</code></td>
                        <td>${b.logo_url ? `<img src="${b.logo_url}" height="30" alt="">` : '-'}</td>
                        <td>
                            <div style="display:flex; gap:0.5rem">
                                <button class="btn btn-icon btn-outline btn-edit" data-id="${b.id}" title="Edit">✏️</button>
                                <button class="btn btn-icon btn-outline btn-delete" data-id="${b.id}" data-name="${b.name}" title="Hapus">🗑️</button>
                            </div>
                        </td>
                    </tr>
                `).join('');

                // Bind events
                document.querySelectorAll('.btn-edit').forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        const id = parseInt(e.currentTarget.getAttribute('data-id'));
                        const brand = brands.find(x => x.id === id);
                        if (!brand) return;

                        currentEditId = id;
                        document.getElementById('modalTitle').innerText = 'Edit Brand: ' + brand.name;
                        document.getElementById('fName').value = brand.name;
                        document.getElementById('fSlug').value = brand.slug;
                        document.getElementById('fLogo').value = brand.logo_url || '';
                        document.getElementById('fDesc').value = brand.description || '';
                        modal.classList.add('active');
                    });
                });

                document.querySelectorAll('.btn-delete').forEach(btn => {
                    btn.addEventListener('click', async (e) => {
                        const id = e.currentTarget.getAttribute('data-id');
                        const name = e.currentTarget.getAttribute('data-name');
                        if (confirm(`Hapus brand "${name}"? Produk terkait mungkin akan terpengaruh.`)) {
                            try {
                                await apiClient(`/brands/${id}`, { method: 'DELETE' });
                                showToast("Brand berhasil dihapus", "success");
                                fetchBrands();
                            } catch (err) {
                                showToast(err.message, "error");
                            }
                        }
                    });
                });

            } catch (err) {
                tbody.innerHTML = `<tr><td colspan="5" style="text-align:center; color:red">Error: ${err.message}</td></tr>`;
            }
        }

        fetchBrands();

        document.getElementById('btnShowAddBrand').addEventListener('click', () => {
            currentEditId = null;
            document.getElementById('modalTitle').innerText = 'Buat Brand Baru';
            document.getElementById('frmSaveBrand').reset();
            modal.classList.add('active');
        });

        const closeModals = () => modal.classList.remove('active');
        document.getElementById('btnCloseModal').addEventListener('click', closeModals);
        document.getElementById('btnCancelModal').addEventListener('click', closeModals);

        document.getElementById('fName').addEventListener('input', (e) => {
            if (!currentEditId) {
                document.getElementById('fSlug').value = e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)+/g, '');
            }
        });

        document.getElementById('frmSaveBrand').addEventListener('submit', async (e) => {
            e.preventDefault();
            const payload = {
                name: document.getElementById('fName').value.trim(),
                slug: document.getElementById('fSlug').value.trim(),
                logo_url: document.getElementById('fLogo').value.trim(),
                description: document.getElementById('fDesc').value.trim()
            };

            try {
                if (currentEditId) {
                    await apiClient(`/brands/${currentEditId}`, { method: 'PUT', body: payload });
                    showToast("Brand diperbarui!", "success");
                } else {
                    await apiClient('/brands', { method: 'POST', body: payload });
                    showToast("Brand ditambahkan!", "success");
                }
                closeModals();
                fetchBrands();
            } catch (err) {
                showToast(err.message, "error");
            }
        });
    }
};

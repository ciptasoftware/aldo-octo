// views/categories.js
import { renderAdminLayout } from '../admin-app.js';
import { apiClient, showToast } from '../api-client.js';

export const CategoriesView = {
    async render(rootElement) {
        const html = renderAdminLayout(`
            <div class="page-header">
                <div>
                    <h1 class="page-title">Manajemen Kategori</h1>
                    <p style="color:var(--text-muted); margin-top:0.25rem;">Kelola pengelompokan produk.</p>
                </div>
                <button class="btn btn-primary" id="btnShowAddCat">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                    Tambah Kategori
                </button>
            </div>

            <div class="table-container">
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th width="80">ID</th>
                            <th>Nama Kategori</th>
                            <th>Slug</th>
                            <th>Deskripsi</th>
                            <th width="100">Aksi</th>
                        </tr>
                    </thead>
                    <tbody id="tblBodyCats">
                        <tr><td colspan="5" style="text-align:center; padding:2rem;">Loading data...</td></tr>
                    </tbody>
                </table>
            </div>

            <!-- Modal Form -->
            <div class="modal-overlay" id="modalCatForm">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3 id="modalTitle">Buat Kategori Baru</h3>
                        <button class="btn-close" id="btnCloseModal">&times;</button>
                    </div>
                    <form id="frmSaveCat">
                        <div class="modal-body">
                            <div class="form-group">
                                <label>Nama Kategori</label>
                                <input type="text" class="form-control" id="fName" required>
                            </div>
                            <div class="form-group">
                                <label>Slug (Identitas URL)</label>
                                <input type="text" class="form-control" id="fSlug" required>
                            </div>
                            <div class="form-group">
                                <label>Deskripsi</label>
                                <textarea class="form-control" id="fDesc" rows="3"></textarea>
                            </div>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-outline" id="btnCancelModal">Batal</button>
                            <button type="submit" class="btn btn-primary" id="btnSubmitForm">Simpan Kategori</button>
                        </div>
                    </form>
                </div>
            </div>
        `);

        rootElement.innerHTML = html;

        const tbody = document.getElementById('tblBodyCats');
        const modal = document.getElementById('modalCatForm');
        let currentEditId = null;

        async function fetchCats() {
            try {
                const cats = await apiClient('/categories');
                if (!cats || cats.length === 0) {
                    tbody.innerHTML = `<tr><td colspan="5" style="text-align:center; padding:2rem; color:var(--text-muted)">Tidak ada data kategori.</td></tr>`;
                    return;
                }

                tbody.innerHTML = cats.map(c => `
                    <tr>
                        <td>#${c.id}</td>
                        <td><strong>${c.name}</strong></td>
                        <td><code>${c.slug}</code></td>
                        <td>${c.description || '-'}</td>
                        <td>
                            <div style="display:flex; gap:0.5rem">
                                <button class="btn btn-icon btn-outline btn-edit" data-id="${c.id}" title="Edit">✏️</button>
                                <button class="btn btn-icon btn-outline btn-delete" data-id="${c.id}" data-name="${c.name}" title="Hapus">🗑️</button>
                            </div>
                        </td>
                    </tr>
                `).join('');

                // Bind events
                document.querySelectorAll('.btn-edit').forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        const id = parseInt(e.currentTarget.getAttribute('data-id'));
                        const cat = cats.find(x => x.id === id);
                        if (!cat) return;

                        currentEditId = id;
                        document.getElementById('modalTitle').innerText = 'Edit Kategori: ' + cat.name;
                        document.getElementById('fName').value = cat.name;
                        document.getElementById('fSlug').value = cat.slug;
                        document.getElementById('fDesc').value = cat.description || '';
                        modal.classList.add('active');
                    });
                });

                document.querySelectorAll('.btn-delete').forEach(btn => {
                    btn.addEventListener('click', async (e) => {
                        const id = e.currentTarget.getAttribute('data-id');
                        const name = e.currentTarget.getAttribute('data-name');
                        if (confirm(`Hapus kategori "${name}"? Produk terkait mungkin akan terpengaruh.`)) {
                            try {
                                await apiClient(`/categories/${id}`, { method: 'DELETE' });
                                showToast("Kategori berhasil dihapus", "success");
                                fetchCats();
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

        fetchCats();

        document.getElementById('btnShowAddCat').addEventListener('click', () => {
            currentEditId = null;
            document.getElementById('modalTitle').innerText = 'Buat Kategori Baru';
            document.getElementById('frmSaveCat').reset();
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

        document.getElementById('frmSaveCat').addEventListener('submit', async (e) => {
            e.preventDefault();
            const payload = {
                name: document.getElementById('fName').value.trim(),
                slug: document.getElementById('fSlug').value.trim(),
                description: document.getElementById('fDesc').value.trim()
            };

            try {
                if (currentEditId) {
                    await apiClient(`/categories/${currentEditId}`, { method: 'PUT', body: payload });
                    showToast("Kategori diperbarui!", "success");
                } else {
                    await apiClient('/categories', { method: 'POST', body: payload });
                    showToast("Kategori ditambahkan!", "success");
                }
                closeModals();
                fetchCats();
            } catch (err) {
                showToast(err.message, "error");
            }
        });
    }
};

// views/products.js
import { renderAdminLayout } from '../admin-app.js';
import { apiClient, showToast } from '../api-client.js';

export const ProductsView = {
    async render(rootElement) {
        
        let brandsData = [];
        let catsData = [];

        // Tarik data referensi brand & category untuk dropdown create product
        try {
            const [bRes, cRes] = await Promise.all([
                apiClient('/brands'),
                apiClient('/categories')
            ]);
            brandsData = bRes || [];
            catsData = cRes || [];
        } catch (e) {
            console.error(e);
        }

        const html = renderAdminLayout(`
            <div class="page-header">
                <div>
                    <h1 class="page-title">Master Data Produk</h1>
                    <p style="color:var(--text-muted); margin-top:0.25rem;">Kelola seluruh inventaris yang dijual ke publik.</p>
                </div>
                <button class="btn btn-primary" id="btnShowAddProduct">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                    Tambah Produk
                </button>
            </div>

            <!-- Tabel Produk -->
            <div class="table-container">
                <div class="table-header">
                    <input type="text" class="search-box" id="inpSearch" placeholder="Cari ketikan nama produk...">
                    <button class="btn btn-outline" id="btnRefresh">Refresh Data</button>
                </div>
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th width="60">Foto</th>
                            <th>Nama Produk</th>
                            <th>Kategori</th>
                            <th>Brand</th>
                            <th>Harga (IDR)</th>
                            <th>Stok</th>
                            <th width="100">Aksi</th>
                        </tr>
                    </thead>
                    <tbody id="tblBodyProducts">
                        <!-- Disuntikkan lewat JS -->
                    </tbody>
                </table>
            </div>

            <!-- Overlay Modal Form -->
            <div class="modal-overlay" id="modalProductForm">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3 id="modalTitle">Buat Produk Baru</h3>
                        <button class="btn-close" id="btnCloseModal">&times;</button>
                    </div>
                    <form id="frmSaveProduct">
                    <div class="modal-body">
                        <div class="form-group">
                            <label>Nama Produk Lengkap</label>
                            <input type="text" class="form-control" id="fName" required>
                        </div>
                        <div class="form-group">
                            <label>Pengenal Unik (Slug)</label>
                            <input type="text" class="form-control" id="fSlug" placeholder="contoh: tas-ransel-korean" required>
                        </div>
                        <div style="display:flex; gap:1rem;">
                            <div class="form-group" style="flex:1">
                                <label>Target Brand</label>
                                <div class="autocomplete-wrapper">
                                    <input type="text" class="form-control" id="inpDisplayBrand" placeholder="Ketik untuk mencari brand..." autocomplete="off" required>
                                    <input type="hidden" id="fBrandId">
                                    <div class="autocomplete-list" id="listBrand"></div>
                                </div>
                            </div>
                            <div class="form-group" style="flex:1">
                                <label>Target Kategori</label>
                                <div class="autocomplete-wrapper">
                                    <input type="text" class="form-control" id="inpDisplayCat" placeholder="Ketik untuk mencari kategori..." autocomplete="off" required>
                                    <input type="hidden" id="fCatId">
                                    <div class="autocomplete-list" id="listCat"></div>
                                </div>
                            </div>
                        </div>
                        <div style="display:flex; gap:1rem;">
                            <div class="form-group" style="flex:1">
                                <label>Harga Resmi (IDR)</label>
                                <input type="number" class="form-control" id="fPrice" min="0" step="0.01" required>
                            </div>
                            <div class="form-group" style="flex:1">
                                <label>Stok Gudang</label>
                                <input type="number" class="form-control" id="fStock" min="0" required>
                            </div>
                        </div>
                        <div class="form-group">
                            <label>Thumbnail Path (Link/URL Image)</label>
                            <input type="text" class="form-control" id="fThumbnail" placeholder="https://..." value="https://dummyjson.com/image/i/products/1/thumbnail.jpg">
                        </div>
                        <div class="quill-editor-container">
                            <label style="display:block; margin-bottom:0.5rem; font-size:0.875rem; font-weight:500; color:#475569;">Deskripsi Produk (Rich Text)</label>
                            <div id="fDescription"></div>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-outline" id="btnCancelModal">Batal</button>
                        <button type="submit" class="btn btn-primary" id="btnSubmitForm">Simpan Produk</button>
                    </div>
                    </form>
                </div>
            </div>
        `);

        rootElement.innerHTML = html;

        // Reference Variables
        const tbody = document.getElementById('tblBodyProducts');
        const modal = document.getElementById('modalProductForm');
        let currentProducts = [];
        let currentEditId = null;

        // Initialize Quill
        const quill = new Quill('#fDescription', {
            theme: 'snow',
            placeholder: 'Tulis deskripsi detail produk di sini...'
        });

        // ==== FUNCTION: Load Data ===
        async function fetchProducts(searchQuery = '') {
            tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding: 2rem;">Loading data dari server...</td></tr>`;
            try {
                const queryParam = searchQuery ? `&q=${encodeURIComponent(searchQuery)}` : '';
                // Sesuai dokumentasi di docs/api.md
                const resp = await apiClient(`/products?limit=50${queryParam}`);
                currentProducts = resp.products || [];
                let items = currentProducts;
                
                if (items.length === 0) {
                    tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:2rem; color:var(--text-muted)">Tidak ada data produk ditemukan.</td></tr>`;
                    return;
                }

                tbody.innerHTML = items.map(p => `
                    <tr>
                        <td><img src="${p.thumbnail || '#'}" alt="" class="thumb-cell" loading="lazy"></td>
                        <td>
                            <strong style="display:block; color:#0f172a">${p.title}</strong>
                            <small style="color:var(--text-muted); font-size:0.75rem;">SKU: ${p.sku||'-'}</small>
                        </td>
                        <td><span class="badge" style="background:#e0f2fe;color:#0369a1">${p.category || 'Uncategorized'}</span></td>
                        <td>${p.brand || 'Vendor ?'}</td>
                        <td style="font-weight:600; font-family:monospace;">Rp${p.price.toLocaleString('id-ID')}
                             ${p.discountPercentage > 0 ? `<br><small style="color:#ef4444; font-weight:bold;">-${p.discountPercentage}%</small>` : ''}
                        </td>
                        <td>
                            <span class="badge ${p.stock > 10 ? 'badge-success' : (p.stock > 0 ? 'badge-warning' : 'badge-danger')}">
                                ${p.stock} Unit
                            </span>
                        </td>
                        <td>
                            <div style="display:flex; gap:0.5rem">
                                <button class="btn btn-icon btn-outline btn-edit" data-id="${p.id}" title="Edit Produk">
                                    ✏️
                                </button>
                                <button class="btn btn-icon btn-outline btn-delete" data-id="${p.id}" data-name="${p.title}" title="Hapus Produk">
                                    🗑️
                                </button>
                            </div>
                        </td>
                    </tr>
                `).join('');

                // Mendaftarkan event binding untuk tombol hapus
                document.querySelectorAll('.btn-delete').forEach(btn => {
                    btn.addEventListener('click', async (e) => {
                        const id = e.currentTarget.getAttribute('data-id');
                        const name = e.currentTarget.getAttribute('data-name');
                        if (confirm(`Yakin ingin MENGHAPUS produk ini permanen?\n\n"${name}"`)) {
                            try {
                                await apiClient(`/products/${id}`, { method: 'DELETE' });
                                showToast("Produk berhasil dihapus", "success");
                                fetchProducts(); // Reload table
                            } catch (err) {
                                showToast(err.message, "error");
                            }
                        }
                    });
                });

                // Mendaftarkan event binding untuk tombol edit
                document.querySelectorAll('.btn-edit').forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        const id = parseInt(e.currentTarget.getAttribute('data-id'), 10);
                        const product = currentProducts.find(p => p.id === id);
                        if (!product) return;
                        
                        currentEditId = id;
                        document.getElementById('modalTitle').innerText = 'Edit Produk: ' + product.title;
                        document.getElementById('btnSubmitForm').innerText = 'Simpan Perubahan';
                        
                        document.getElementById('fName').value = product.title;
                        document.getElementById('fSlug').value = product.sku;
                        
                        // Autofill comboboxes
                        if (product.brand_id) {
                            document.getElementById('fBrandId').value = product.brand_id;
                            const bx = brandsData.find(x => x.id == product.brand_id);
                            document.getElementById('inpDisplayBrand').value = bx ? bx.name : product.brand;
                        }
                        if (product.category_id) {
                            document.getElementById('fCatId').value = product.category_id;
                            const cx = catsData.find(x => x.id == product.category_id);
                            document.getElementById('inpDisplayCat').value = cx ? cx.name : product.category;
                        }
                        
                        document.getElementById('fPrice').value = product.price;
                        document.getElementById('fStock').value = product.stock;
                        document.getElementById('fThumbnail').value = product.thumbnail;
                        
                        // Set Quill Content
                        quill.root.innerHTML = product.description || '';

                        modal.classList.add('active');
                    });
                });

            } catch (err) {
                tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:2rem; color:var(--color-danger)">Gagal memuat produk: ${err.message}</td></tr>`;
            }
        }

        // Jalankan Fetch awal
        fetchProducts();

        // Bind Search Events
        const searchBox = document.getElementById('inpSearch');
        let searchTimeoutId = null;
        searchBox.addEventListener('keyup', (e) => {
            clearTimeout(searchTimeoutId);
            searchTimeoutId = setTimeout(() => {
                fetchProducts(e.target.value);
            }, 400); // 400ms debounce
        });
        document.getElementById('btnRefresh').addEventListener('click', () => fetchProducts());

        // ==== BINDING UI MODAL FORM ====
        document.getElementById('btnShowAddProduct').addEventListener('click', () => {
            currentEditId = null;
            document.getElementById('modalTitle').innerText = 'Buat Produk Baru';
            document.getElementById('btnSubmitForm').innerText = 'Tambahkan Produk';
            document.getElementById('frmSaveProduct').reset();
            quill.root.innerHTML = ''; // Reset Quill
            modal.classList.add('active');
        });

        // Hide overlay rules
        const closeModals = () => modal.classList.remove('active');
        document.getElementById('btnCloseModal').addEventListener('click', closeModals);
        document.getElementById('btnCancelModal').addEventListener('click', closeModals);

        // Auto-Generate Slug Helper
        document.getElementById('fName').addEventListener('input', (e) => {
            const val = e.target.value;
            // Lowercase and replace non alphabet with hyphens
            document.getElementById('fSlug').value = val.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)+/g, '');
        });

        // Action POST / Simpan ke Server
        document.getElementById('frmSaveProduct').addEventListener('submit', async (e) => {
            e.preventDefault();
            const btnSubmit = document.getElementById('btnSubmitForm');
            const originalText = btnSubmit.innerHTML;
            btnSubmit.innerHTML = 'Menyimpan...';
            btnSubmit.disabled = true;

            const payload = {
                title: document.getElementById('fName').value.trim(),
                sku: document.getElementById('fSlug').value.trim(),
                description: quill.root.innerHTML, // Get Quill content
                brand_id: parseInt(document.getElementById('fBrandId').value, 10),
                category_id: parseInt(document.getElementById('fCatId').value, 10),
                price: parseFloat(document.getElementById('fPrice').value),
                discountPercentage: 0,
                stock: parseInt(document.getElementById('fStock').value, 10),
                thumbnail: document.getElementById('fThumbnail').value.trim()
            };

            try {
                if (currentEditId) {
                    await apiClient(`/products/${currentEditId}`, {
                        method: 'PUT',
                        body: payload
                    });
                    showToast("Produk berhasil diperbarui!", "success");
                } else {
                    await apiClient('/products', {
                        method: 'POST',
                        body: payload
                    });
                    showToast("Produk berhasil ditambahkan!", "success");
                }
                closeModals();
                fetchProducts();
            } catch (err) {
                showToast("Gagal Disimpan: " + err.message, "error");
            } finally {
                btnSubmit.innerHTML = originalText;
                btnSubmit.disabled = false;
            }
        });

        // ==== AUTOCOMPLETE LOGIC ====
        function setupAutocomplete(inputId, hiddenId, listId, dataArray, endpointPath, noun) {
            const inp = document.getElementById(inputId);
            const hidden = document.getElementById(hiddenId);
            const list = document.getElementById(listId);

            inp.addEventListener('keydown', function(e) {
                if (e.key === 'Enter') {
                    e.preventDefault(); // prevent form submission on enter
                }
            });

            inp.addEventListener('input', function() {
                const val = this.value.trim().toLowerCase();
                list.innerHTML = '';
                hidden.value = ''; // Invalidate selection
                if (!val) {
                    list.classList.remove('active');
                    return;
                }

                const matches = dataArray.filter(x => x.name.toLowerCase().includes(val));
                if (matches.length > 0) {
                    matches.forEach(m => {
                        const div = document.createElement('div');
                        div.className = 'autocomplete-item';
                        div.innerHTML = m.name;
                        div.addEventListener('click', () => {
                            inp.value = m.name;
                            hidden.value = m.id;
                            list.innerHTML = '';
                            list.classList.remove('active');
                        });
                        list.appendChild(div);
                    });
                    list.classList.add('active');
                } else {
                    list.innerHTML = `
                        <div class="autocomplete-item not-found">
                            <i style="color:var(--text-muted)">Tidak ditemukan.</i><br>
                            <button type="button" class="btn btn-primary" id="btnQuickAdd_${listId}" style="margin-top:0.5rem; width:100%; font-size:0.75rem;">+ Tambah "${this.value}"</button>
                        </div>
                    `;
                    document.getElementById(`btnQuickAdd_${listId}`).addEventListener('click', async () => {
                        const newName = inp.value.trim();
                        try {
                            const slug = newName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)+/g, '');
                            const resp = await apiClient(endpointPath, {
                                method: 'POST',
                                body: { name: newName, slug: slug }
                            });
                            showToast(`${noun} berhasil ditambahkan!`, "success");
                            // Push to local array
                            dataArray.push(resp);
                            // Auto select
                            inp.value = resp.name;
                            hidden.value = resp.id;
                            list.innerHTML = '';
                            list.classList.remove('active');
                        } catch (err) {
                            showToast(`Gagal tambah ${noun}: ${err.message}`, "error");
                        }
                    });
                    list.classList.add('active');
                }
            });

            document.addEventListener('click', function(e) {
                if (e.target !== inp && !list.contains(e.target)) {
                    list.classList.remove('active');
                }
            });
        }

        setupAutocomplete('inpDisplayBrand', 'fBrandId', 'listBrand', brandsData, '/brands', 'Brand');
        setupAutocomplete('inpDisplayCat', 'fCatId', 'listCat', catsData, '/categories', 'Kategori');

    }
};

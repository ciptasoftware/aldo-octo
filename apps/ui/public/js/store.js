/* ============================================================
   Go-AI Store — Pure JS Storefront Interactions
   ============================================================ */

document.addEventListener('DOMContentLoaded', () => {
    // 1. Hero Carousel Controller
    initHeroCarousel();

    // 2. Client-Side Search, Category Filter, and Sorting
    initStoreFilters();

    // 3. Product Detail Modal
    initProductModal();

    // 4. Cart Counter & Toast Notification
    initCartActions();
});

// --- Hero Carousel Controller ---
function initHeroCarousel() {
    const slides = document.querySelectorAll('.hero-slide');
    const dots = document.querySelectorAll('.slider-dot');
    if (!slides.length) return;

    let currentIndex = 0;
    let autoInterval = null;

    function showSlide(index) {
        slides.forEach((s, i) => s.classList.toggle('active', i === index));
        dots.forEach((d, i) => d.classList.toggle('active', i === index));
        currentIndex = index;
    }

    dots.forEach((dot, index) => {
        dot.addEventListener('click', () => {
            showSlide(index);
            resetAutoPlay();
        });
    });

    function nextSlide() {
        showSlide((currentIndex + 1) % slides.length);
    }

    function resetAutoPlay() {
        if (autoInterval) clearInterval(autoInterval);
        autoInterval = setInterval(nextSlide, 5000);
    }

    resetAutoPlay();
}

// --- Store Search, Category & Brand Filters, and Sorting ---
function initStoreFilters() {
    const searchInput = document.getElementById('storeSearchInput');
    const searchCategorySelect = document.getElementById('searchCategorySelect');
    const filterPills = document.querySelectorAll('.filter-pill');
    const catNavItems = document.querySelectorAll('.cat-nav-item');
    const sortSelect = document.getElementById('storeSortSelect');
    const productGrid = document.getElementById('storeProductGrid');
    const emptyState = document.getElementById('storeEmptyState');
    const activeFilterBadge = document.getElementById('activeFilterBadge');
    const activeFilterText = document.getElementById('activeFilterText');

    if (!productGrid) return;

    let selectedCategory = 'all';
    let selectedBrand = 'all';
    let selectedCategoryName = '';
    let selectedBrandName = '';
    let searchQuery = '';

    function updateFilterUI() {
        // Update Category Pills UI
        filterPills.forEach(p => {
            p.classList.toggle('active', (selectedCategory !== 'all' && p.getAttribute('data-category') === selectedCategory) || (selectedCategory === 'all' && selectedBrand === 'all' && p.getAttribute('data-category') === 'all'));
        });

        // Update Nav Strip UI
        catNavItems.forEach(n => {
            n.classList.toggle('active', (selectedCategory !== 'all' && n.getAttribute('data-category') === selectedCategory) || (selectedCategory === 'all' && selectedBrand === 'all' && n.getAttribute('data-category') === 'all'));
        });

        if (searchCategorySelect) {
            searchCategorySelect.value = selectedCategory === 'all' ? '' : selectedCategory;
        }

        // Active Filter Notification Badge
        if (activeFilterBadge && activeFilterText) {
            if (selectedBrand !== 'all') {
                activeFilterText.innerHTML = `Filter Aktif: Brand <strong>"${selectedBrandName || selectedBrand}"</strong>`;
                activeFilterBadge.style.display = 'flex';
            } else if (selectedCategory !== 'all') {
                activeFilterText.innerHTML = `Filter Aktif: Kategori <strong>"${selectedCategoryName || selectedCategory}"</strong>`;
                activeFilterBadge.style.display = 'flex';
            } else if (searchQuery) {
                activeFilterText.innerHTML = `Pencarian: <strong>"${searchQuery}"</strong>`;
                activeFilterBadge.style.display = 'flex';
            } else {
                activeFilterBadge.style.display = 'none';
            }
        }
    }

    function filterAndSortProducts() {
        const cards = Array.from(productGrid.querySelectorAll('.product-card'));
        let visibleCount = 0;

        cards.forEach(card => {
            const cardCatSlug = (card.getAttribute('data-category-slug') || card.getAttribute('data-category') || '').toLowerCase();
            const cardCatName = (card.getAttribute('data-category-name') || '').toLowerCase();
            const cardBrandSlug = (card.getAttribute('data-brand-slug') || card.getAttribute('data-brand') || '').toLowerCase();
            const cardBrandName = (card.getAttribute('data-brand-name') || '').toLowerCase();
            const cardName = (card.getAttribute('data-name') || '').toLowerCase();

            const matchCat = (selectedCategory === 'all' || 
                cardCatSlug === selectedCategory.toLowerCase() || 
                cardCatName === selectedCategory.toLowerCase());

            const matchBrand = (selectedBrand === 'all' || 
                cardBrandSlug === selectedBrand.toLowerCase() || 
                cardBrandName === selectedBrand.toLowerCase());

            const matchSearch = !searchQuery || 
                cardName.includes(searchQuery) || 
                cardCatName.includes(searchQuery) || 
                cardBrandName.includes(searchQuery);

            if (matchCat && matchBrand && matchSearch) {
                card.style.display = 'flex';
                visibleCount++;
            } else {
                card.style.display = 'none';
            }
        });

        // Show/Hide Empty State
        if (emptyState) {
            emptyState.style.display = visibleCount === 0 ? 'block' : 'none';
        }

        // Apply Sorting
        sortProducts();
    }

    function sortProducts() {
        if (!sortSelect) return;
        const sortVal = sortSelect.value;
        const cards = Array.from(productGrid.querySelectorAll('.product-card'));

        cards.sort((a, b) => {
            const priceA = parseFloat(a.getAttribute('data-price') || '0');
            const priceB = parseFloat(b.getAttribute('data-price') || '0');
            const nameA = (a.getAttribute('data-name') || '').toLowerCase();
            const nameB = (b.getAttribute('data-name') || '').toLowerCase();

            if (sortVal === 'price-asc') return priceA - priceB;
            if (sortVal === 'price-desc') return priceB - priceA;
            if (sortVal === 'name-asc') return nameA.localeCompare(nameB);
            return 0; // Default order
        });

        cards.forEach(card => productGrid.appendChild(card));
    }

    // Global Filter Handlers
    window.filterByCategory = function(catSlug, displayName = '', scrollToSection = true) {
        selectedCategory = catSlug;
        selectedCategoryName = displayName;
        selectedBrand = 'all';
        updateFilterUI();
        filterAndSortProducts();
        if (scrollToSection) {
            const sec = document.getElementById('product-section');
            if (sec) sec.scrollIntoView({ behavior: 'smooth' });
        }
    };

    window.filterByBrand = function(brandSlug, displayName = '', scrollToSection = true) {
        selectedBrand = brandSlug;
        selectedBrandName = displayName;
        selectedCategory = 'all';
        updateFilterUI();
        filterAndSortProducts();
        if (scrollToSection) {
            const sec = document.getElementById('product-section');
            if (sec) sec.scrollIntoView({ behavior: 'smooth' });
        }
    };

    window.resetAllFilters = function() {
        selectedCategory = 'all';
        selectedBrand = 'all';
        selectedCategoryName = '';
        selectedBrandName = '';
        searchQuery = '';
        if (searchInput) searchInput.value = '';
        updateFilterUI();
        filterAndSortProducts();
    };

    // Category Filter Pills & Nav Items Event
    filterPills.forEach(pill => {
        pill.addEventListener('click', () => {
            const slug = pill.getAttribute('data-category');
            window.filterByCategory(slug, pill.innerText.trim());
        });
    });

    catNavItems.forEach(item => {
        item.addEventListener('click', () => {
            const slug = item.getAttribute('data-category');
            window.filterByCategory(slug, item.innerText.trim());
        });
    });

    if (searchCategorySelect) {
        searchCategorySelect.addEventListener('change', (e) => {
            const val = e.target.value || 'all';
            window.filterByCategory(val);
        });
    }

    // Live Search Input Event
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            searchQuery = e.target.value.trim().toLowerCase();
            updateFilterUI();
            filterAndSortProducts();
        });
    }

    // Sort Dropdown Event
    if (sortSelect) {
        sortSelect.addEventListener('change', () => {
            sortProducts();
        });
    }

    // Check URL parameters (e.g. ?cat=slug or ?brand=slug)
    const urlParams = new URLSearchParams(window.location.search);
    const catParam = urlParams.get('cat');
    const brandParam = urlParams.get('brand');
    if (catParam) {
        window.filterByCategory(catParam, '', false);
    } else if (brandParam) {
        window.filterByBrand(brandParam, '', false);
    }
}

// --- Product Modal ---
function initProductModal() {
    const modal = document.getElementById('storeProductModal');
    if (!modal) return;

    window.openProductModal = function(btn) {
        const card = btn.closest('.product-card');
        if (!card) return;

        const name = card.getAttribute('data-name');
        const price = card.getAttribute('data-price-formatted');
        const cat = card.getAttribute('data-category-name');
        const brand = card.getAttribute('data-brand-name');
        const desc = card.getAttribute('data-desc');
        const thumb = card.getAttribute('data-thumb');
        const stock = card.getAttribute('data-stock');

        document.getElementById('modalProdTitle').innerText = name;
        document.getElementById('modalProdPrice').innerText = price;
        document.getElementById('modalProdCat').innerText = cat;
        document.getElementById('modalProdBrand').innerText = brand;
        document.getElementById('modalProdDesc').innerText = desc || 'Tidak ada deskripsi rinci.';
        document.getElementById('modalProdStock').innerText = stock + ' unit tersedia';

        const imgEl = document.getElementById('modalProdImg');
        if (thumb) {
            imgEl.src = thumb;
            imgEl.style.display = 'block';
        } else {
            imgEl.style.display = 'none';
        }

        modal.style.display = 'flex';
    };

    window.closeProductModal = function() {
        modal.style.display = 'none';
    };

    modal.addEventListener('click', (e) => {
        if (e.target === modal) closeProductModal();
    });
}

// --- Cart Actions & Toast ---
function initCartActions() {
    updateCartHeaderBadges();

    window.addToCart = function(id, name, price, thumbnail, slug) {
        let cart = [];
        try {
            cart = JSON.parse(localStorage.getItem('aldo_store_cart')) || [];
        } catch(e) {
            cart = [];
        }

        const existingIndex = cart.findIndex(item => item.id === id);
        if (existingIndex >= 0) {
            cart[existingIndex].qty = (cart[existingIndex].qty || 1) + 1;
        } else {
            cart.push({
                id: id,
                name: name,
                price: parseFloat(price) || 0,
                thumbnail: thumbnail || '',
                slug: slug || '',
                qty: 1
            });
        }

        localStorage.setItem('aldo_store_cart', JSON.stringify(cart));
        updateCartHeaderBadges();
        showStoreToast(`"${name}" ditambahkan ke keranjang!`, 'success');
    };

    window.handleHomeBuyBtn = function(btn) {
        if (!btn) return;
        const id = parseInt(btn.getAttribute('data-id')) || 0;
        const name = btn.getAttribute('data-name') || '';
        const price = parseFloat(btn.getAttribute('data-price')) || 0;
        const thumb = btn.getAttribute('data-thumb') || '';
        const slug = btn.getAttribute('data-slug') || '';
        window.addToCart(id, name, price, thumb, slug);
    };
}

function updateCartHeaderBadges() {
    let cart = [];
    try {
        cart = JSON.parse(localStorage.getItem('aldo_store_cart')) || [];
    } catch(e) {
        cart = [];
    }

    let totalUnits = 0;
    let grandTotal = 0;
    cart.forEach(item => {
        const qty = item.qty || 1;
        totalUnits += qty;
        grandTotal += (item.price || 0) * qty;
    });

    const badge = document.getElementById('storeCartBadge');
    if (badge) {
        badge.innerText = totalUnits;
        badge.style.display = totalUnits > 0 ? 'flex' : 'none';
    }

    const totalEl = document.getElementById('storeCartTotal');
    if (totalEl) {
        const intPart = Math.round(grandTotal);
        totalEl.innerText = 'Rp ' + intPart.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ".");
    }
}

function showStoreToast(message, type = 'success') {
    let container = document.getElementById('storeToastContainer');
    if (!container) {
        container = document.createElement('div');
        container.id = 'storeToastContainer';
        container.style.cssText = 'position:fixed; bottom:1.5rem; right:1.5rem; z-index:9999; display:flex; flex-direction:column; gap:0.5rem;';
        document.body.appendChild(container);
    }

    const toast = document.createElement('div');
    toast.style.cssText = `
        background: ${type === 'success' ? '#10b981' : '#ef4444'};
        color: white;
        padding: 0.85rem 1.25rem;
        border-radius: 10px;
        font-size: 0.875rem;
        font-weight: 600;
        box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1);
        display: flex;
        align-items: center;
        gap: 0.5rem;
        animation: toastIn 0.2s ease forwards;
    `;
    toast.innerHTML = `<span>✓</span> <span>${message}</span>`;

    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateY(10px)';
        toast.style.transition = 'all 0.3s ease';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

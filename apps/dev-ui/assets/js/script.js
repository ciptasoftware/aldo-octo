/**
 * OBSShop - Main JavaScript
 * Fetches products from https://dummyjson.com/products
 * Renders category-based product grid, handles search, filter, sort, modal, cart, etc.
 */

/* ============================================================
   CONSTANTS & STATE
   ============================================================ */
const API_BASE = 'https://dummyjson.com/products';
const PRODUCTS_PER_PAGE = 194; // fetch all at once, display grouped by category

const CATEGORY_ICONS = {
  'beauty': '💄',
  'fragrances': '🌸',
  'furniture': '🛋️',
  'groceries': '🛍️',
  'home-decoration': '🏠',
  'kitchen-accessories': '🍳',
  'laptops': '💻',
  'mens-shirts': '👔',
  'mens-shoes': '👟',
  'mens-watches': '⌚',
  'mobile-accessories': '📱',
  'motorcycle': '🏍️',
  'skin-care': '🧴',
  'smartphones': '📱',
  'sports-accessories': '⚽',
  'sunglasses': '🕶️',
  'tablets': '📲',
  'tops': '👕',
  'vehicle': '🚗',
  'womens-bags': '👜',
  'womens-dresses': '👗',
  'womens-jewellery': '💍',
  'womens-shoes': '👠',
  'womens-watches': '⌚',
};

const CATEGORY_CLASS_MAP = {
  'beauty': 'cat-beauty',
  'fragrances': 'cat-fragrances',
  'furniture': 'cat-furniture',
  'groceries': 'cat-groceries',
  'home-decoration': 'cat-home-decoration',
  'kitchen-accessories': 'cat-electronics',
  'laptops': 'cat-electronics',
  'smartphones': 'cat-electronics',
  'tablets': 'cat-electronics',
  'mobile-accessories': 'cat-electronics',
  'mens-shirts': 'cat-default',
  'mens-shoes': 'cat-default',
  'mens-watches': 'cat-default',
  'skin-care': 'cat-beauty',
  'sports-accessories': 'cat-sports-accessories',
  'sunglasses': 'cat-default',
  'tops': 'cat-default',
  'vehicle': 'cat-automotive',
  'motorcycle': 'cat-automotive',
  'womens-bags': 'cat-beauty',
  'womens-dresses': 'cat-beauty',
  'womens-jewellery': 'cat-beauty',
  'womens-shoes': 'cat-beauty',
  'womens-watches': 'cat-default',
};

let state = {
  allProducts: [],
  filteredProducts: [],
  categories: [],
  activeCategory: 'all',
  searchQuery: '',
  sortMode: 'default',
  cartItems: [],
  wishlist: new Set(),
  currentPage: 1,
  productsPerBatch: 10,
  loading: false,
};

/* ============================================================
   DOM REFS
   ============================================================ */
const dom = {
  searchInput: document.getElementById('search-input'),
  searchBtn: document.getElementById('search-btn'),
  searchCategorySelect: document.getElementById('search-category-select'),
  navList: document.getElementById('nav-list'),
  categoryGrid: document.getElementById('category-grid'),
  filterTags: document.getElementById('filter-tags'),
  sortSelect: document.getElementById('sort-select'),
  productsContainer: document.getElementById('products-container'),
  loadingState: document.getElementById('loading-state'),
  noResults: document.getElementById('no-results'),
  loadMoreWrap: document.getElementById('load-more-wrap'),
  loadMoreBtn: document.getElementById('load-more-btn'),
  loadMoreText: document.getElementById('load-more-text'),
  loadMoreSpinner: document.getElementById('load-more-spinner'),
  cartCount: document.getElementById('cart-count'),
  cartBtn: document.getElementById('cart-btn'),
  productModal: document.getElementById('product-modal'),
  modalBox: document.getElementById('modal-box'),
  modalClose: document.getElementById('modal-close'),
  modalBody: document.getElementById('modal-body'),
  toastContainer: document.getElementById('toast-container'),
  backToTop: document.getElementById('back-to-top'),
  resetFilterBtn: document.getElementById('reset-filter-btn'),
  heroSlider: document.getElementById('hero-slider'),
  sliderPrev: document.getElementById('slider-prev'),
  sliderNext: document.getElementById('slider-next'),
  sliderDots: document.getElementById('slider-dots'),
  footerTopStrip: document.getElementById('footer-top-strip'),
};

/* ============================================================
   HERO SLIDER
   ============================================================ */
let currentSlide = 0;
let sliderInterval = null;

function initSlider() {
  const slides = document.querySelectorAll('.hero-slide');
  const dots = document.querySelectorAll('.dot');

  function goToSlide(idx) {
    slides.forEach((s, i) => s.classList.toggle('active', i === idx));
    dots.forEach((d, i) => d.classList.toggle('active', i === idx));
    currentSlide = idx;
  }

  dom.sliderPrev.addEventListener('click', () => {
    goToSlide((currentSlide - 1 + slides.length) % slides.length);
    resetInterval();
  });

  dom.sliderNext.addEventListener('click', () => {
    goToSlide((currentSlide + 1) % slides.length);
    resetInterval();
  });

  dots.forEach((dot, i) => {
    dot.addEventListener('click', () => { goToSlide(i); resetInterval(); });
  });

  function resetInterval() {
    clearInterval(sliderInterval);
    sliderInterval = setInterval(() => goToSlide((currentSlide + 1) % slides.length), 5000);
  }
  sliderInterval = setInterval(() => goToSlide((currentSlide + 1) % slides.length), 5000);
}

/* ============================================================
   HERO SVG PLACEHOLDERS (generated inline)
   ============================================================ */
function createHeroSVGs() {
  // Create simple inline SVG for hero images
  const svgData = [
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 250">
      <defs><radialGradient id="g1" cx="50%" cy="50%" r="50%"><stop offset="0%" stop-color="rgba(255,255,255,0.4)"/><stop offset="100%" stop-color="rgba(255,255,255,0)"/></radialGradient></defs>
      <ellipse cx="150" cy="125" rx="140" ry="110" fill="url(#g1)"/>
      <text x="150" y="80" text-anchor="middle" font-size="80" font-family="serif">🏷️</text>
      <text x="150" y="160" text-anchor="middle" font-size="22" font-weight="900" fill="white" font-family="sans-serif">FLASH SALE</text>
      <text x="150" y="190" text-anchor="middle" font-size="36" font-weight="900" fill="#FFD600" font-family="sans-serif">70% OFF</text>
    </svg>`,
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 250">
      <defs><radialGradient id="g2" cx="50%" cy="50%" r="50%"><stop offset="0%" stop-color="rgba(255,255,255,0.4)"/><stop offset="100%" stop-color="rgba(255,255,255,0)"/></radialGradient></defs>
      <ellipse cx="150" cy="125" rx="140" ry="110" fill="url(#g2)"/>
      <text x="150" y="80" text-anchor="middle" font-size="80">✨</text>
      <text x="150" y="155" text-anchor="middle" font-size="20" font-weight="900" fill="white" font-family="sans-serif">NEW ARRIVAL</text>
      <text x="150" y="190" text-anchor="middle" font-size="16" fill="rgba(255,255,255,0.85)" font-family="sans-serif">Premium Quality</text>
    </svg>`,
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 250">
      <defs><radialGradient id="g3" cx="50%" cy="50%" r="50%"><stop offset="0%" stop-color="rgba(255,255,255,0.4)"/><stop offset="100%" stop-color="rgba(255,255,255,0)"/></radialGradient></defs>
      <ellipse cx="150" cy="125" rx="140" ry="110" fill="url(#g3)"/>
      <text x="150" y="80" text-anchor="middle" font-size="80">🚚</text>
      <text x="150" y="155" text-anchor="middle" font-size="20" font-weight="900" fill="white" font-family="sans-serif">FREE SHIPPING</text>
      <text x="150" y="190" text-anchor="middle" font-size="15" fill="rgba(255,255,255,0.85)" font-family="sans-serif">No Minimum Order</text>
    </svg>`
  ];

  const imgIds = ['hero-img-1', 'hero-img-2', 'hero-img-3'];
  imgIds.forEach((id, i) => {
    const img = document.getElementById(id);
    if (img) {
      const blob = new Blob([svgData[i]], { type: 'image/svg+xml' });
      img.src = URL.createObjectURL(blob);
    }
  });
}

/* ============================================================
   FETCH PRODUCTS
   ============================================================ */
async function fetchAllProducts() {
  try {
    showLoading(true);
    const res = await fetch(`${API_BASE}?limit=194&select=id,title,price,thumbnail,category,rating,brand,discountPercentage,stock,description`);
    if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
    const data = await res.json();

    state.allProducts = data.products || [];
    state.filteredProducts = [...state.allProducts];

    // Extract unique categories
    state.categories = [...new Set(state.allProducts.map(p => p.category))].sort();

    buildNav();
    buildCategoryGrid();
    buildFilterTags();
    renderProducts();
    showLoading(false);
  } catch (err) {
    console.error('Failed to fetch products:', err);
    showLoading(false);
    showToast('❌ Gagal memuat produk. Silakan refresh halaman.', 'error');
    dom.productsContainer.innerHTML = `
      <div class="no-results" style="display:block;">
        <div class="no-results-icon">⚠️</div>
        <h3>Gagal memuat data</h3>
        <p>Periksa koneksi internet Anda dan coba lagi.</p>
        <button onclick="location.reload()" class="reset-btn">Refresh Halaman</button>
      </div>
    `;
  }
}

/* ============================================================
   BUILD NAV / CATEGORIES
   ============================================================ */
function buildNav() {
  // Clear existing category links (keep first: Beranda)
  const existing = dom.navList.querySelectorAll('li:nth-child(n+2)');
  existing.forEach(li => li.remove());

  state.categories.forEach(cat => {
    const li = document.createElement('li');
    const a = document.createElement('a');
    a.href = '#';
    a.className = 'nav-link';
    a.dataset.category = cat;
    a.textContent = formatCategoryName(cat);
    a.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveCategory(cat);
    });
    li.appendChild(a);
    dom.navList.appendChild(li);
  });

  const berandaLink = dom.navList.querySelector('[data-category="all"]');
  if (berandaLink) {
    berandaLink.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveCategory('all');
    });
  }

  // Populate search category select
  if (dom.searchCategorySelect) {
    state.categories.forEach(cat => {
      const opt = document.createElement('option');
      opt.value = cat;
      opt.textContent = formatCategoryName(cat);
      dom.searchCategorySelect.appendChild(opt);
    });
    dom.searchCategorySelect.addEventListener('change', () => {
      if (dom.searchCategorySelect.value !== 'all') {
        setActiveCategory(dom.searchCategorySelect.value);
      } else {
        setActiveCategory('all');
      }
    });
  }
}

function buildCategoryGrid() {
  dom.categoryGrid.innerHTML = '';

  // All Products card
  const allCard = createCategoryCard('all', '🛒', 'Semua Produk');
  allCard.classList.add('active');
  dom.categoryGrid.appendChild(allCard);

  state.categories.forEach(cat => {
    const icon = CATEGORY_ICONS[cat] || '📦';
    const card = createCategoryCard(cat, icon, formatCategoryName(cat));
    dom.categoryGrid.appendChild(card);
  });
}

function createCategoryCard(cat, icon, name) {
  const card = document.createElement('div');
  card.className = 'category-card';
  card.dataset.category = cat;

  const gradient = getCategoryGradient(cat);
  card.innerHTML = `
    <div class="cat-icon-wrap" style="background:${gradient}">
      <span>${icon}</span>
    </div>
    <span class="cat-name">${name}</span>
  `;
  card.addEventListener('click', () => setActiveCategory(cat));
  return card;
}

function buildFilterTags() {
  dom.filterTags.innerHTML = `<button class="filter-tag active" data-category="all">Semua</button>`;

  state.categories.forEach(cat => {
    const btn = document.createElement('button');
    btn.className = 'filter-tag';
    btn.dataset.category = cat;
    btn.textContent = formatCategoryName(cat);
    btn.addEventListener('click', () => setActiveCategory(cat));
    dom.filterTags.appendChild(btn);
  });
}

/* ============================================================
   FILTERING & SORTING
   ============================================================ */
function setActiveCategory(cat) {
  state.activeCategory = cat;
  state.currentPage = 1;
  applyFilters();
  updateActiveStates();
  window.scrollTo({ top: document.getElementById('product-section').offsetTop - 70, behavior: 'smooth' });
}

function applyFilters() {
  let products = [...state.allProducts];

  // Category filter
  if (state.activeCategory !== 'all') {
    products = products.filter(p => p.category === state.activeCategory);
  }

  // Search filter
  if (state.searchQuery.trim()) {
    const q = state.searchQuery.toLowerCase().trim();
    products = products.filter(p =>
      p.title.toLowerCase().includes(q) ||
      (p.brand && p.brand.toLowerCase().includes(q)) ||
      p.category.toLowerCase().includes(q)
    );
  }

  // Sort
  products = sortProducts(products, state.sortMode);

  state.filteredProducts = products;
  renderProducts();
}

function sortProducts(products, mode) {
  const arr = [...products];
  switch (mode) {
    case 'price-asc':   return arr.sort((a, b) => a.price - b.price);
    case 'price-desc':  return arr.sort((a, b) => b.price - a.price);
    case 'rating-desc': return arr.sort((a, b) => b.rating - a.rating);
    case 'name-asc':    return arr.sort((a, b) => a.title.localeCompare(b.title));
    default:            return arr;
  }
}

function updateActiveStates() {
  // Nav links
  document.querySelectorAll('.nav-link').forEach(link => {
    link.classList.toggle('active', link.dataset.category === state.activeCategory);
  });
  // Category cards
  document.querySelectorAll('.category-card').forEach(card => {
    card.classList.toggle('active', card.dataset.category === state.activeCategory);
  });
  // Filter tags
  document.querySelectorAll('.filter-tag').forEach(tag => {
    tag.classList.toggle('active', tag.dataset.category === state.activeCategory);
  });
}

/* ============================================================
   RENDER PRODUCTS
   ============================================================ */
function renderProducts() {
  dom.productsContainer.innerHTML = '';

  if (state.filteredProducts.length === 0) {
    dom.noResults.style.display = 'block';
    dom.loadMoreWrap.style.display = 'none';
    return;
  }

  dom.noResults.style.display = 'none';

  if (state.activeCategory === 'all' && state.searchQuery === '' && state.sortMode === 'default') {
    // Group by category
    renderByCategory();
  } else {
    // Flat grid
    renderFlatGrid(state.filteredProducts);
  }
}

function renderByCategory() {
  const grouped = {};
  state.filteredProducts.forEach(p => {
    if (!grouped[p.category]) grouped[p.category] = [];
    grouped[p.category].push(p);
  });

  Object.entries(grouped).forEach(([cat, products]) => {
    const section = createCategorySection(cat, products);
    dom.productsContainer.appendChild(section);
  });

  dom.loadMoreWrap.style.display = 'none';
}

function createCategorySection(cat, products) {
  const section = document.createElement('section');
  const catClass = CATEGORY_CLASS_MAP[cat] || 'cat-default';
  section.className = `category-section ${catClass}`;
  section.id = `cat-${cat}`;

  const icon = CATEGORY_ICONS[cat] || '📦';
  const accentColor = getCategoryAccentColor(cat);

  section.innerHTML = `
    <div class="category-header" style="border-left: 4px solid ${accentColor};">
      <div class="cat-header-left">
        <span class="cat-header-icon">${icon}</span>
        <span class="cat-header-name">${formatCategoryName(cat)}</span>
        <span class="cat-header-count">${products.length} produk</span>
      </div>
      <div class="cat-header-right">
        <button class="cat-view-all" data-category="${cat}">Lihat Semua ›</button>
      </div>
    </div>
    <div class="products-grid" id="grid-${cat}"></div>
  `;

  const grid = section.querySelector(`#grid-${cat}`);
  products.forEach(product => {
    grid.appendChild(createProductCard(product));
  });

  // View all button
  section.querySelector('.cat-view-all').addEventListener('click', () => {
    setActiveCategory(cat);
  });

  return section;
}

function renderFlatGrid(products) {
  const wrapper = document.createElement('div');
  wrapper.className = 'category-section cat-default';

  const header = document.createElement('div');
  header.className = 'category-header';
  const accentColor = getCategoryAccentColor(state.activeCategory);
  header.style.borderLeft = `4px solid ${accentColor}`;

  const title = state.activeCategory !== 'all'
    ? `${CATEGORY_ICONS[state.activeCategory] || '📦'} ${formatCategoryName(state.activeCategory)}`
    : '🔍 Hasil Pencarian';

  header.innerHTML = `
    <div class="cat-header-left">
      <span class="cat-header-name">${title}</span>
      <span class="cat-header-count">${products.length} produk</span>
    </div>
  `;

  const grid = document.createElement('div');
  grid.className = 'products-grid';

  products.forEach(p => grid.appendChild(createProductCard(p)));

  wrapper.appendChild(header);
  wrapper.appendChild(grid);
  dom.productsContainer.appendChild(wrapper);
}

/* ============================================================
   PRODUCT CARD
   ============================================================ */
function createProductCard(product) {
  const card = document.createElement('article');
  card.className = 'product-card';
  card.dataset.id = product.id;

  const isWishlisted = state.wishlist.has(product.id);
  const discount = product.discountPercentage ? Math.round(product.discountPercentage) : null;
  const originalPrice = discount ? (product.price / (1 - discount / 100)).toFixed(2) : null;

  card.innerHTML = `
    <div class="product-img-wrap">
      <img
        class="product-img"
        src="${product.thumbnail}"
        alt="${product.title}"
        loading="lazy"
        onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22200%22><rect width=%22200%22 height=%22200%22 fill=%22%23f0f0f0%22/><text x=%22100%22 y=%22110%22 text-anchor=%22middle%22 font-size=%2248%22>📦</text></svg>'"
      />
      <div class="badge-wrap">
        ${discount ? `<span class="badge-item badge-sale">-${discount}%</span>` : ''}
        ${product.stock < 10 ? `<span class="badge-item badge-hot">Hot</span>` : ''}
      </div>
      <button class="card-wishlist${isWishlisted ? ' active' : ''}" data-id="${product.id}" aria-label="Wishlist">
        ${isWishlisted ? '❤️' : '🤍'}
      </button>
    </div>
    <div class="product-info">
      ${product.brand ? `<span class="product-brand">${product.brand}</span>` : ''}
      <h3 class="product-name">${product.title}</h3>
      <div class="product-rating">
        <div class="stars">${generateStars(product.rating)}</div>
        <span class="rating-val">${product.rating.toFixed(1)}</span>
      </div>
      <div class="product-price-row">
        <div>
          <span class="product-price">$${product.price.toFixed(2)}</span>
          ${originalPrice ? `<br><span class="product-original-price">$${originalPrice}</span>` : ''}
        </div>
        <button class="card-add-btn" data-id="${product.id}" aria-label="Tambah ke keranjang">+</button>
      </div>
    </div>
  `;

  // Click card → open modal
  card.addEventListener('click', (e) => {
    if (e.target.closest('.card-wishlist') || e.target.closest('.card-add-btn')) return;
    openModal(product);
  });

  // Wishlist toggle
  card.querySelector('.card-wishlist').addEventListener('click', (e) => {
    e.stopPropagation();
    toggleWishlist(product.id, e.currentTarget);
  });

  // Add to cart
  card.querySelector('.card-add-btn').addEventListener('click', (e) => {
    e.stopPropagation();
    addToCart(product);
  });

  return card;
}

function generateStars(rating) {
  const full = Math.floor(rating);
  const half = rating - full >= 0.4 ? 1 : 0;
  const empty = 5 - full - half;
  return (
    '★'.repeat(full).split('').map(() => `<span class="star filled">★</span>`).join('') +
    (half ? `<span class="star half">★</span>` : '') +
    '★'.repeat(empty).split('').map(() => `<span class="star">★</span>`).join('')
  );
}

/* ============================================================
   MODAL
   ============================================================ */
function openModal(product) {
  const discount = product.discountPercentage ? Math.round(product.discountPercentage) : null;
  const originalPrice = discount ? (product.price / (1 - discount / 100)).toFixed(2) : null;
  const isWished = state.wishlist.has(product.id);
  const inStock = !product.stock || product.stock > 0;

  dom.modalBody.innerHTML = `
    <div class="modal-img-col">
      <div class="modal-img-wrap">
        <img class="modal-img" src="${product.thumbnail}" alt="${product.title}"
          onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22200%22><rect width=%22200%22 height=%22200%22 fill=%22%23f5f5f5%22/><text x=%22100%22 y=%22110%22 text-anchor=%22middle%22 font-size=%2248%22>📦</text></svg>'"
        />
      </div>
    </div>
    <div class="modal-info-col">
      <span class="modal-category-badge">${CATEGORY_ICONS[product.category] || '📦'} ${formatCategoryName(product.category)}</span>
      <h2 class="modal-title">${product.title}</h2>
      ${product.brand ? `<p class="modal-brand">Dijual oleh <strong>${product.brand}</strong></p>` : ''}
      <div class="modal-rating-row">
        <div class="stars">${generateStars(product.rating)}</div>
        <span class="rating-val">${product.rating.toFixed(1)}</span>
        ${product.stock !== undefined ? `<span class="stock-info">| Stok: ${product.stock}</span>` : ''}
      </div>
      <div class="modal-price-row">
        <span class="modal-price">$${product.price.toFixed(2)}</span>
        ${originalPrice ? `<span class="modal-orig-price">$${originalPrice}</span>` : ''}
        ${discount ? `<span class="modal-discount-badge">Hemat ${discount}%</span>` : ''}
      </div>
      ${product.description ? `<p class="modal-desc">${product.description}</p>` : ''}
      <p class="modal-stock-row">${inStock ? '✓ Tersedia &amp; Siap Kirim' : '✗ Stok Habis'}</p>
      <div class="modal-actions">
        <button class="btn-add-cart" id="modal-add-cart" data-id="${product.id}">
          🛒 Tambah ke Keranjang
        </button>
        <button class="btn-buy-now" id="modal-buy-now" data-id="${product.id}">
          ⚡ Beli Sekarang
        </button>
        <button class="btn-wishlist-modal" id="modal-wishlist" data-id="${product.id}">
          ${isWished ? '❤️ Hapus dari Wishlist' : '🤍 Tambah ke Wishlist'}
        </button>
      </div>
      <p class="modal-secure-note">🔒 Transaksi aman &amp; terenkripsi | Garansi Uang Kembali 30 Hari</p>
    </div>
  `;

  dom.productModal.style.display = 'flex';
  document.body.style.overflow = 'hidden';

  document.getElementById('modal-add-cart').addEventListener('click', () => addToCart(product));
  document.getElementById('modal-buy-now').addEventListener('click', () => {
    addToCart(product);
    showToast('⚡ Lanjut ke checkout...', 'success');
    closeModal();
  });
  document.getElementById('modal-wishlist').addEventListener('click', (e) => {
    toggleWishlist(product.id, null);
    e.currentTarget.textContent = state.wishlist.has(product.id)
      ? '❤️ Hapus dari Wishlist'
      : '🤍 Tambah ke Wishlist';
  });
}

function closeModal() {
  dom.productModal.style.display = 'none';
  document.body.style.overflow = '';
}

dom.modalClose.addEventListener('click', closeModal);
dom.productModal.addEventListener('click', (e) => {
  if (e.target === dom.productModal) closeModal();
});
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') closeModal();
});

/* ============================================================
   CART
   ============================================================ */
function addToCart(product) {
  const existing = state.cartItems.find(i => i.id === product.id);
  if (existing) {
    existing.qty++;
  } else {
    state.cartItems.push({ ...product, qty: 1 });
  }
  updateCartBadge();
  showToast(`🛒 "${product.title}" ditambahkan ke keranjang!`, 'success');
}

function updateCartBadge() {
  const total = state.cartItems.reduce((sum, i) => sum + i.qty, 0);
  dom.cartCount.textContent = total;
  dom.cartCount.style.display = total > 0 ? 'flex' : 'none';
}

dom.cartBtn.addEventListener('click', () => {
  if (state.cartItems.length === 0) {
    showToast('🛒 Keranjang Anda masih kosong.', 'warning');
    return;
  }
  const items = state.cartItems.map(i => `• ${i.title} (x${i.qty}) - $${(i.price * i.qty).toFixed(2)}`).join('\n');
  const total = state.cartItems.reduce((sum, i) => sum + i.price * i.qty, 0);
  alert(`🛒 Keranjang Belanja:\n\n${items}\n\nTotal: $${total.toFixed(2)}`);
});

/* ============================================================
   WISHLIST
   ============================================================ */
function toggleWishlist(id, btn) {
  if (state.wishlist.has(id)) {
    state.wishlist.delete(id);
    if (btn) { btn.textContent = '🤍'; btn.classList.remove('active'); }
    showToast('💔 Dihapus dari wishlist.', 'default');
  } else {
    state.wishlist.add(id);
    if (btn) { btn.textContent = '❤️'; btn.classList.add('active'); }
    showToast('❤️ Ditambahkan ke wishlist!', 'success');
  }
}

/* ============================================================
   SEARCH
   ============================================================ */
function handleSearch() {
  state.searchQuery = dom.searchInput.value;
  state.activeCategory = 'all';
  state.currentPage = 1;
  applyFilters();
  updateActiveStates();
}

dom.searchBtn.addEventListener('click', handleSearch);
dom.searchInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') handleSearch();
});
dom.searchInput.addEventListener('input', debounce(() => {
  if (dom.searchInput.value === '') {
    state.searchQuery = '';
    applyFilters();
  }
}, 400));

/* ============================================================
   SORT
   ============================================================ */
dom.sortSelect.addEventListener('change', () => {
  state.sortMode = dom.sortSelect.value;
  applyFilters();
});

/* ============================================================
   RESET FILTER
   ============================================================ */
if (dom.resetFilterBtn) {
  dom.resetFilterBtn.addEventListener('click', () => {
    state.searchQuery = '';
    state.activeCategory = 'all';
    state.sortMode = 'default';
    dom.searchInput.value = '';
    dom.sortSelect.value = 'default';
    applyFilters();
    updateActiveStates();
  });
}

/* ============================================================
   TOAST
   ============================================================ */
function showToast(message, type = 'default') {
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  const icons = { success: '✅', error: '❌', warning: '⚠️', default: 'ℹ️' };
  toast.innerHTML = `<span class="toast-icon">${icons[type] || 'ℹ️'}</span><span>${message}</span>`;
  dom.toastContainer.appendChild(toast);

  setTimeout(() => {
    toast.style.animation = 'toastOut 0.3s ease forwards';
    setTimeout(() => toast.remove(), 300);
  }, 3000);
}

/* ============================================================
   LOADING STATE
   ============================================================ */
function showLoading(show) {
  state.loading = show;
  dom.loadingState.style.display = show ? 'block' : 'none';
  dom.productsContainer.style.display = show ? 'none' : 'flex';
}

/* ============================================================
   FOOTER STRIP
   ============================================================ */
if (dom.footerTopStrip) {
  dom.footerTopStrip.addEventListener('click', () => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  });
  dom.footerTopStrip.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ' ') window.scrollTo({ top: 0, behavior: 'smooth' });
  });
}



/* ============================================================
   BACK TO TOP
   ============================================================ */
window.addEventListener('scroll', debounce(() => {
  dom.backToTop.style.display = window.scrollY > 400 ? 'flex' : 'none';
}, 100));

dom.backToTop.addEventListener('click', () => {
  window.scrollTo({ top: 0, behavior: 'smooth' });
});

/* ============================================================
   UTILS
   ============================================================ */
function formatCategoryName(cat) {
  return cat.replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
}

function getCategoryAccentColor(cat) {
  const colors = {
    'beauty':              '#e91e8c',
    'fragrances':          '#9c27b0',
    'furniture':           '#e65100',
    'groceries':           '#2e7d32',
    'home-decoration':     '#00897b',
    'kitchen-accessories': '#1565c0',
    'laptops':             '#0277bd',
    'smartphones':         '#0288d1',
    'tablets':             '#0288d1',
    'mobile-accessories':  '#01579b',
    'mens-shirts':         '#37474f',
    'mens-shoes':          '#4e342e',
    'mens-watches':        '#5d4037',
    'skin-care':           '#c2185b',
    'sports-accessories':  '#f57f17',
    'sunglasses':          '#6a1b9a',
    'tops':                '#1a237e',
    'vehicle':             '#b71c1c',
    'motorcycle':          '#bf360c',
    'womens-bags':         '#ad1457',
    'womens-dresses':      '#d81b60',
    'womens-jewellery':    '#7b1fa2',
    'womens-shoes':        '#c62828',
    'womens-watches':      '#6a1b9a',
    'all':                 '#f0a500',
  };
  return colors[cat] || '#232f3e';
}

function getCategoryGradient(cat) {
  const gradients = {
    'beauty':              'linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)',
    'fragrances':          'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)',
    'furniture':           'linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)',
    'groceries':           'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)',
    'home-decoration':     'linear-gradient(135deg, #96fbc4 0%, #f9f586 100%)',
    'kitchen-accessories': 'linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%)',
    'laptops':             'linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%)',
    'smartphones':         'linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%)',
    'tablets':             'linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%)',
    'mobile-accessories':  'linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%)',
    'mens-shirts':         'linear-gradient(135deg, #c3cfe2 0%, #f5f7fa 100%)',
    'mens-shoes':          'linear-gradient(135deg, #b8c6db 0%, #f5f7fa 100%)',
    'mens-watches':        'linear-gradient(135deg, #d3cce3 0%, #e9e4f0 100%)',
    'skin-care':           'linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)',
    'sports-accessories':  'linear-gradient(135deg, #fddb92 0%, #d1fdff 100%)',
    'sunglasses':          'linear-gradient(135deg, #fbd3e9 0%, #bb1ccc 20%, #3a1c71 100%)',
    'tops':                'linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%)',
    'vehicle':             'linear-gradient(135deg, #a1c4fd 0%, #c2e9fb 100%)',
    'motorcycle':          'linear-gradient(135deg, #a1c4fd 0%, #c2e9fb 100%)',
    'womens-bags':         'linear-gradient(135deg, #fda085 0%, #f6d365 100%)',
    'womens-dresses':      'linear-gradient(135deg, #fd79a8 0%, #e17055 100%)',
    'womens-jewellery':    'linear-gradient(135deg, #fdcbf1 0%, #e6dee9 100%)',
    'womens-shoes':        'linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)',
    'womens-watches':      'linear-gradient(135deg, #d3cce3 0%, #e9e4f0 100%)',
    'all':                 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  };
  return gradients[cat] || 'linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%)';
}

function debounce(fn, delay) {
  let timer;
  return function (...args) {
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(this, args), delay);
  };
}

/* ============================================================
   INIT
   ============================================================ */
(function init() {
  createHeroSVGs();
  initSlider();
  updateCartBadge();
  fetchAllProducts();
})();

package ecommerce

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AdminUser model
type AdminUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Category model
type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// Brand model
type Brand struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
}

// ProductImage model
type ProductImage struct {
	ID          int    `json:"id"`
	ProductID   int    `json:"product_id"`
	ImageURL    string `json:"image_url"`
	IsThumbnail bool   `json:"is_thumbnail"`
}

// Product model
type Product struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Slug         string         `json:"slug"`
	BrandID      int            `json:"brand_id"`
	CategoryID   int            `json:"category_id"`
	Price        float64        `json:"price"`
	Stock        int            `json:"stock"`
	Thumbnail    string         `json:"thumbnail"`
	Description  string         `json:"description"`
	BrandName    string         `json:"brand_name"`
	CategoryName string         `json:"category_name"`
	BrandSlug    string         `json:"brand_slug"`
	CategorySlug string         `json:"category_slug"`
	Images       []ProductImage `json:"images"`
}

// FormatIDR formats a float64 price into Indonesian Rupiah format with dots as thousand separators (e.g. 21999000 -> "Rp 21.999.000")
func FormatIDR(amount float64) string {
	intPart := int64(amount)
	str := strconv.FormatInt(intPart, 10)

	n := len(str)
	if n <= 3 {
		return "Rp " + str
	}

	var parts []string
	remainder := n % 3
	if remainder > 0 {
		parts = append(parts, str[:remainder])
	}
	for i := remainder; i < n; i += 3 {
		parts = append(parts, str[i:i+3])
	}

	return "Rp " + strings.Join(parts, ".")
}

func (p Product) FormattedPrice() string {
	return FormatIDR(p.Price)
}

// Partner model
type Partner struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	CompanyName string    `json:"company_name"`
	Description string    `json:"description"`
	Password    string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProjectManager model
type ProjectManager struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	CompanyName string    `json:"company_name"`
	Description string    `json:"description"`
	Password    string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

// Banner model
type Banner struct {
	ID          int       `json:"id"`
	BadgeText   string    `json:"badge_text"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CtaText     string    `json:"cta_text"`
	CtaURL      string    `json:"cta_url"`
	IsActive    bool      `json:"is_active"`
	Icon        string    `json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
}

// OrderItem model
type OrderItem struct {
	ID          int     `json:"id"`
	OrderID     int     `json:"order_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

func (item OrderItem) FormattedPrice() string {
	return FormatIDR(item.Price)
}

func (item OrderItem) FormattedSubtotal() string {
	return FormatIDR(item.Subtotal)
}

// Order model
type Order struct {
	ID         int         `json:"id"`
	Email      string      `json:"email"`
	UserType   string      `json:"user_type"`
	UserName   string      `json:"user_name"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items"`
}

func (o Order) FormattedTotalPrice() string {
	return FormatIDR(o.TotalPrice)
}

func (o Order) FormattedDate() string {
	return o.CreatedAt.Format("02 Jan 2006, 15:04 WIB")
}

// CustomerUser for logged-in Partner / Project Manager
type CustomerUser struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	UserType string `json:"user_type"` // "partner" or "project_manager"
}

// AdminSearchResult represents search results grouped by category
type AdminSearchResult struct {
	Categories      []Category       `json:"categories,omitempty"`
	Brands          []Brand          `json:"brands,omitempty"`
	Products        []Product        `json:"products,omitempty"`
	Partners        []Partner        `json:"partners,omitempty"`
	ProjectManagers []ProjectManager `json:"project_managers,omitempty"`
	Banners         []Banner         `json:"banners,omitempty"`
	Orders          []Order          `json:"orders,omitempty"`
	TotalCount      int              `json:"total_count"`
}

// Stats summary for admin dashboard
type AdminStats struct {
	TotalCategories      int
	TotalBrands          int
	TotalProducts        int
	TotalPartners        int
	TotalProjectManagers int
	TotalBanners         int
	TotalOrders          int
	PendingOrders        int
	LowStockCount        int
}

// PageInfo struct for 10-item pagination
type PageInfo struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	TotalCount  int  `json:"total_count"`
	HasPrev     bool `json:"has_prev"`
	HasNext     bool `json:"has_next"`
	PrevPage    int  `json:"prev_page"`
	NextPage    int  `json:"next_page"`
}

func CalculatePageInfo(page, totalCount, pageSize int) PageInfo {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	return PageInfo{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
	}
}

// CreateSlug generates a clean URL slug from a title string
func CreateSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// GenerateToken creates a 64-character random hex token
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// MigrateTable initializes database tables if they do not exist
func MigrateTable(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS admin_sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
			expires_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS brands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			logo_url TEXT,
			description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			brand_id INTEGER NOT NULL REFERENCES brands(id) ON DELETE SET NULL,
			category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE SET NULL,
			price REAL NOT NULL DEFAULT 0,
			stock INTEGER NOT NULL DEFAULT 0,
			thumbnail TEXT,
			description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS partners (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			phone TEXT DEFAULT '',
			company_name TEXT DEFAULT '',
			description TEXT DEFAULT '',
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS project_managers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			phone TEXT DEFAULT '',
			company_name TEXT DEFAULT '',
			description TEXT DEFAULT '',
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS banners (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			badge_text TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			cta_text TEXT NOT NULL,
			cta_url TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			icon TEXT NOT NULL DEFAULT '🎁',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS product_images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			image_url TEXT NOT NULL,
			is_thumbnail INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS customer_sessions (
			token TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			user_type TEXT NOT NULL,
			expires_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			user_type TEXT NOT NULL,
			user_name TEXT NOT NULL,
			total_price REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id),
			product_name TEXT NOT NULL,
			price REAL NOT NULL,
			quantity INTEGER NOT NULL,
			subtotal REAL NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	// Seed default admin user if admin_users is empty
	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&count)
	if count == 0 {
		hashedPw, err := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
		if err == nil {
			_, _ = db.Exec("INSERT INTO admin_users (username, password, name) VALUES (?, ?, ?)",
				"admin", string(hashedPw), "Administrator Go-AI")
		}
	}

	// Seed default banners if banners table is empty
	var bannerCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM banners").Scan(&bannerCount)
	if bannerCount == 0 {
		_, _ = db.Exec(`INSERT INTO banners (badge_text, title, description, cta_text, cta_url, is_active, icon) VALUES
			('Promo Spesial Hari Ini', 'Belanja Hemat Hingga 70% OFF', 'Koleksi produk e-commerce terlengkap dengan jaminan kualitas 100% original dan bebas ongkir.', 'Belanja Sekarang →', '#product-section', 1, '🎁'),
			('Official Brand Store', 'Jaminan Produk 100% Original', 'Dapatkan garansi resmi dari mitra brand ternama dengan kemudahan pembayaran dan pengembalian barang.', 'Jelajahi Brand →', '#brand-section', 1, '🏷️'),
			('Layanan Ekstra Cepat', 'Pengiriman Cepat Ke Seluruh Indonesia', 'Pesanan diproses instan dengan integrasi kurir terpercaya dan fitur pelacakan waktu nyata.', 'Lihat Katalog →', '#product-section', 1, '🚀');`)
	}

	return nil
}

// --- ADMIN USER & SESSION DB OPERATIONS ---

func GetAdminUserByUsername(ctx context.Context, db *sql.DB, username string) (*AdminUser, error) {
	var u AdminUser
	err := db.QueryRowContext(ctx, "SELECT id, username, password, name FROM admin_users WHERE username = ?", username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Name)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateAdminSession(ctx context.Context, db *sql.DB, userID int) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.ExecContext(ctx, "INSERT INTO admin_sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

func ValidateAdminSession(ctx context.Context, db *sql.DB, token string) (*AdminUser, error) {
	if token == "" {
		return nil, errors.New("token kosong")
	}
	var u AdminUser
	var expiresAt time.Time
	query := `
		SELECT u.id, u.username, u.name, s.expires_at 
		FROM admin_sessions s
		JOIN admin_users u ON s.user_id = u.id
		WHERE s.token = ?
	`
	err := db.QueryRowContext(ctx, query, token).Scan(&u.ID, &u.Username, &u.Name, &expiresAt)
	if err != nil {
		return nil, err
	}
	if time.Now().After(expiresAt) {
		_ = DeleteAdminSession(ctx, db, token)
		return nil, errors.New("sesi kadaluarsa")
	}
	return &u, nil
}

func DeleteAdminSession(ctx context.Context, db *sql.DB, token string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM admin_sessions WHERE token = ?", token)
	return err
}

// --- CATEGORY DB OPERATIONS ---

func GetAllCategories(ctx context.Context, db *sql.DB) ([]Category, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name, slug, COALESCE(description, '') FROM categories ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func GetPaginatedCategories(ctx context.Context, db *sql.DB, page int) ([]Category, PageInfo, error) {
	var totalCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories").Scan(&totalCount)
	pageInfo := CalculatePageInfo(page, totalCount, 10)
	offset := (pageInfo.CurrentPage - 1) * 10

	rows, err := db.QueryContext(ctx, "SELECT id, name, slug, COALESCE(description, '') FROM categories ORDER BY id DESC LIMIT 10 OFFSET ?", offset)
	if err != nil {
		return nil, pageInfo, err
	}
	defer rows.Close()

	var list []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description); err != nil {
			return nil, pageInfo, err
		}
		list = append(list, c)
	}
	return list, pageInfo, nil
}

func GetCategoryByID(ctx context.Context, db *sql.DB, id int) (*Category, error) {
	var c Category
	err := db.QueryRowContext(ctx, "SELECT id, name, slug, COALESCE(description, '') FROM categories WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Description)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func SaveCategory(ctx context.Context, db *sql.DB, c *Category) error {
	if c.Slug == "" {
		c.Slug = CreateSlug(c.Name)
	}
	if c.ID > 0 {
		_, err := db.ExecContext(ctx, "UPDATE categories SET name = ?, slug = ?, description = ? WHERE id = ?",
			c.Name, c.Slug, c.Description, c.ID)
		return err
	}
	res, err := db.ExecContext(ctx, "INSERT INTO categories (name, slug, description) VALUES (?, ?, ?)",
		c.Name, c.Slug, c.Description)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		c.ID = int(id)
	}
	return nil
}

func DeleteCategory(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM categories WHERE id = ?", id)
	return err
}

// --- BRAND DB OPERATIONS ---

func GetAllBrands(ctx context.Context, db *sql.DB) ([]Brand, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name, slug, COALESCE(logo_url, ''), COALESCE(description, '') FROM brands ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Brand
	for rows.Next() {
		var b Brand
		if err := rows.Scan(&b.ID, &b.Name, &b.Slug, &b.LogoURL, &b.Description); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

func GetPaginatedBrands(ctx context.Context, db *sql.DB, page int) ([]Brand, PageInfo, error) {
	var totalCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM brands").Scan(&totalCount)
	pageInfo := CalculatePageInfo(page, totalCount, 10)
	offset := (pageInfo.CurrentPage - 1) * 10

	rows, err := db.QueryContext(ctx, "SELECT id, name, slug, COALESCE(logo_url, ''), COALESCE(description, '') FROM brands ORDER BY id DESC LIMIT 10 OFFSET ?", offset)
	if err != nil {
		return nil, pageInfo, err
	}
	defer rows.Close()

	var list []Brand
	for rows.Next() {
		var b Brand
		if err := rows.Scan(&b.ID, &b.Name, &b.Slug, &b.LogoURL, &b.Description); err != nil {
			return nil, pageInfo, err
		}
		list = append(list, b)
	}
	return list, pageInfo, nil
}

func GetBrandByID(ctx context.Context, db *sql.DB, id int) (*Brand, error) {
	var b Brand
	err := db.QueryRowContext(ctx, "SELECT id, name, slug, COALESCE(logo_url, ''), COALESCE(description, '') FROM brands WHERE id = ?", id).
		Scan(&b.ID, &b.Name, &b.Slug, &b.LogoURL, &b.Description)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func SaveBrand(ctx context.Context, db *sql.DB, b *Brand) error {
	if b.Slug == "" {
		b.Slug = CreateSlug(b.Name)
	}
	if b.ID > 0 {
		_, err := db.ExecContext(ctx, "UPDATE brands SET name = ?, slug = ?, logo_url = ?, description = ? WHERE id = ?",
			b.Name, b.Slug, b.LogoURL, b.Description, b.ID)
		return err
	}
	res, err := db.ExecContext(ctx, "INSERT INTO brands (name, slug, logo_url, description) VALUES (?, ?, ?, ?)",
		b.Name, b.Slug, b.LogoURL, b.Description)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		b.ID = int(id)
	}
	return nil
}

func DeleteBrand(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM brands WHERE id = ?", id)
	return err
}

// --- PRODUCT DB OPERATIONS ---

func GetAllProducts(ctx context.Context, db *sql.DB) ([]Product, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.brand_id, p.category_id, p.price, p.stock, 
		       COALESCE(p.thumbnail, ''), COALESCE(p.description, ''),
		       COALESCE(b.name, 'Tanpa Brand'), COALESCE(c.name, 'Tanpa Kategori'),
		       COALESCE(b.slug, ''), COALESCE(c.slug, '')
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.id DESC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.BrandID, &p.CategoryID, &p.Price, &p.Stock, &p.Thumbnail, &p.Description, &p.BrandName, &p.CategoryName, &p.BrandSlug, &p.CategorySlug); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func GetPaginatedProducts(ctx context.Context, db *sql.DB, page int) ([]Product, PageInfo, error) {
	var totalCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&totalCount)
	pageInfo := CalculatePageInfo(page, totalCount, 10)
	offset := (pageInfo.CurrentPage - 1) * 10

	query := `
		SELECT p.id, p.name, p.slug, p.brand_id, p.category_id, p.price, p.stock, 
		       COALESCE(p.thumbnail, ''), COALESCE(p.description, ''),
		       COALESCE(b.name, 'Tanpa Brand'), COALESCE(c.name, 'Tanpa Kategori'),
		       COALESCE(b.slug, ''), COALESCE(c.slug, '')
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.id DESC
		LIMIT 10 OFFSET ?
	`
	rows, err := db.QueryContext(ctx, query, offset)
	if err != nil {
		return nil, pageInfo, err
	}
	defer rows.Close()

	var list []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.BrandID, &p.CategoryID, &p.Price, &p.Stock, &p.Thumbnail, &p.Description, &p.BrandName, &p.CategoryName, &p.BrandSlug, &p.CategorySlug); err != nil {
			return nil, pageInfo, err
		}
		list = append(list, p)
	}
	return list, pageInfo, nil
}

func GetProductByID(ctx context.Context, db *sql.DB, id int) (*Product, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.brand_id, p.category_id, p.price, p.stock, 
		       COALESCE(p.thumbnail, ''), COALESCE(p.description, ''),
		       COALESCE(b.name, 'Tanpa Brand'), COALESCE(c.name, 'Tanpa Kategori'),
		       COALESCE(b.slug, ''), COALESCE(c.slug, '')
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = ?
	`
	var p Product
	err := db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.BrandID, &p.CategoryID, &p.Price, &p.Stock, &p.Thumbnail, &p.Description, &p.BrandName, &p.CategoryName, &p.BrandSlug, &p.CategorySlug,
	)
	if err != nil {
		return nil, err
	}
	p.Images, _ = GetProductImages(ctx, db, p.ID)
	if p.Thumbnail == "" && len(p.Images) > 0 {
		p.Thumbnail = p.Images[0].ImageURL
	}
	return &p, nil
}

func EnsureUniqueProductSlug(ctx context.Context, db *sql.DB, baseSlug string, currentID int) string {
	if baseSlug == "" {
		baseSlug = "produk"
	}
	slug := baseSlug
	counter := 1
	for {
		var existingID int
		err := db.QueryRowContext(ctx, "SELECT id FROM products WHERE slug = ? AND id != ?", slug, currentID).Scan(&existingID)
		if err == sql.ErrNoRows {
			return slug
		}
		if err != nil {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
}

func GetProductBySlug(ctx context.Context, db *sql.DB, slug string) (*Product, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.brand_id, p.category_id, p.price, p.stock, 
		       COALESCE(p.thumbnail, ''), COALESCE(p.description, ''),
		       COALESCE(b.name, 'Tanpa Brand'), COALESCE(c.name, 'Tanpa Kategori'),
		       COALESCE(b.slug, ''), COALESCE(c.slug, '')
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.slug = ?
	`
	var p Product
	err := db.QueryRowContext(ctx, query, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.BrandID, &p.CategoryID, &p.Price, &p.Stock, &p.Thumbnail, &p.Description, &p.BrandName, &p.CategoryName, &p.BrandSlug, &p.CategorySlug,
	)
	if err != nil {
		return nil, err
	}
	p.Images, _ = GetProductImages(ctx, db, p.ID)
	if p.Thumbnail == "" && len(p.Images) > 0 {
		p.Thumbnail = p.Images[0].ImageURL
	}
	return &p, nil
}

func GetRelatedProducts(ctx context.Context, db *sql.DB, categoryID int, currentID int, limit int) ([]Product, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.brand_id, p.category_id, p.price, p.stock, 
		       COALESCE(p.thumbnail, ''), COALESCE(p.description, ''),
		       COALESCE(b.name, 'Tanpa Brand'), COALESCE(c.name, 'Tanpa Kategori'),
		       COALESCE(b.slug, ''), COALESCE(c.slug, '')
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id != ? AND (p.category_id = ? OR ? = 0)
		ORDER BY p.id DESC LIMIT ?
	`
	rows, err := db.QueryContext(ctx, query, currentID, categoryID, categoryID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.BrandID, &p.CategoryID, &p.Price, &p.Stock, &p.Thumbnail, &p.Description, &p.BrandName, &p.CategoryName, &p.BrandSlug, &p.CategorySlug); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func SearchProducts(ctx context.Context, db *sql.DB, searchQuery string, categorySlug string) ([]Product, error) {
	searchQuery = strings.TrimSpace(searchQuery)
	likePattern := "%" + searchQuery + "%"

	query := `
		SELECT p.id, p.name, p.slug, p.brand_id, p.category_id, p.price, p.stock, 
		       COALESCE(p.thumbnail, ''), COALESCE(p.description, ''),
		       COALESCE(b.name, 'Tanpa Brand'), COALESCE(c.name, 'Tanpa Kategori'),
		       COALESCE(b.slug, ''), COALESCE(c.slug, '')
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE (
			? = '' OR 
			c.slug = ? OR 
			c.name LIKE ?
		) AND (
			? = '' OR 
			p.name LIKE ? OR 
			p.description LIKE ? OR 
			b.name LIKE ? OR 
			c.name LIKE ?
		)
		ORDER BY p.id DESC
	`
	rows, err := db.QueryContext(ctx, query, 
		categorySlug, categorySlug, "%"+categorySlug+"%",
		searchQuery, likePattern, likePattern, likePattern, likePattern,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.BrandID, &p.CategoryID, &p.Price, &p.Stock, &p.Thumbnail, &p.Description, &p.BrandName, &p.CategoryName, &p.BrandSlug, &p.CategorySlug); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func SaveProduct(ctx context.Context, db *sql.DB, p *Product) error {
	baseSlug := p.Slug
	if baseSlug == "" {
		baseSlug = CreateSlug(p.Name)
	} else {
		baseSlug = CreateSlug(baseSlug)
	}
	p.Slug = EnsureUniqueProductSlug(ctx, db, baseSlug, p.ID)

	if p.ID > 0 {
		_, err := db.ExecContext(ctx, `UPDATE products SET name = ?, slug = ?, brand_id = ?, category_id = ?, price = ?, stock = ?, thumbnail = ?, description = ? WHERE id = ?`,
			p.Name, p.Slug, p.BrandID, p.CategoryID, p.Price, p.Stock, p.Thumbnail, p.Description, p.ID)
		return err
	}
	res, err := db.ExecContext(ctx, `INSERT INTO products (name, slug, brand_id, category_id, price, stock, thumbnail, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.Slug, p.BrandID, p.CategoryID, p.Price, p.Stock, p.Thumbnail, p.Description)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		p.ID = int(id)
	}
	return nil
}

func DeleteProduct(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM products WHERE id = ?", id)
	return err
}

func GetAdminStats(ctx context.Context, db *sql.DB) (AdminStats, error) {
	var stats AdminStats
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories").Scan(&stats.TotalCategories)
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM brands").Scan(&stats.TotalBrands)
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&stats.TotalProducts)
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM partners").Scan(&stats.TotalPartners)
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM project_managers").Scan(&stats.TotalProjectManagers)
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM banners").Scan(&stats.TotalBanners)
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE stock <= 5").Scan(&stats.LowStockCount)
	return stats, nil
}

// --- BANNER DB OPERATIONS ---

func GetActiveBanners(ctx context.Context, db *sql.DB) ([]Banner, error) {
	query := `SELECT id, badge_text, title, description, cta_text, cta_url, is_active, icon, created_at FROM banners WHERE is_active = 1 ORDER BY id DESC`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Banner
	for rows.Next() {
		var b Banner
		var activeInt int
		if err := rows.Scan(&b.ID, &b.BadgeText, &b.Title, &b.Description, &b.CtaText, &b.CtaURL, &activeInt, &b.Icon, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.IsActive = (activeInt == 1)
		list = append(list, b)
	}
	return list, nil
}

func GetAllBanners(ctx context.Context, db *sql.DB) ([]Banner, error) {
	query := `SELECT id, badge_text, title, description, cta_text, cta_url, is_active, icon, created_at FROM banners ORDER BY id DESC`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Banner
	for rows.Next() {
		var b Banner
		var activeInt int
		if err := rows.Scan(&b.ID, &b.BadgeText, &b.Title, &b.Description, &b.CtaText, &b.CtaURL, &activeInt, &b.Icon, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.IsActive = (activeInt == 1)
		list = append(list, b)
	}
	return list, nil
}

func GetPaginatedBanners(ctx context.Context, db *sql.DB, page int) ([]Banner, PageInfo, error) {
	var totalCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM banners").Scan(&totalCount)
	pageInfo := CalculatePageInfo(page, totalCount, 10)
	offset := (pageInfo.CurrentPage - 1) * 10

	query := `SELECT id, badge_text, title, description, cta_text, cta_url, is_active, icon, created_at FROM banners ORDER BY id DESC LIMIT 10 OFFSET ?`
	rows, err := db.QueryContext(ctx, query, offset)
	if err != nil {
		return nil, pageInfo, err
	}
	defer rows.Close()

	var list []Banner
	for rows.Next() {
		var b Banner
		var activeInt int
		if err := rows.Scan(&b.ID, &b.BadgeText, &b.Title, &b.Description, &b.CtaText, &b.CtaURL, &activeInt, &b.Icon, &b.CreatedAt); err != nil {
			return nil, pageInfo, err
		}
		b.IsActive = (activeInt == 1)
		list = append(list, b)
	}
	return list, pageInfo, nil
}

func GetBannerByID(ctx context.Context, db *sql.DB, id int) (*Banner, error) {
	query := `SELECT id, badge_text, title, description, cta_text, cta_url, is_active, icon, created_at FROM banners WHERE id = ?`
	var b Banner
	var activeInt int
	err := db.QueryRowContext(ctx, query, id).Scan(&b.ID, &b.BadgeText, &b.Title, &b.Description, &b.CtaText, &b.CtaURL, &activeInt, &b.Icon, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	b.IsActive = (activeInt == 1)
	return &b, nil
}

func SaveBanner(ctx context.Context, db *sql.DB, b *Banner) error {
	activeInt := 0
	if b.IsActive {
		activeInt = 1
	}
	if b.Icon == "" {
		b.Icon = "🎁"
	}
	if b.ID > 0 {
		query := `UPDATE banners SET badge_text = ?, title = ?, description = ?, cta_text = ?, cta_url = ?, is_active = ?, icon = ? WHERE id = ?`
		_, err := db.ExecContext(ctx, query, b.BadgeText, b.Title, b.Description, b.CtaText, b.CtaURL, activeInt, b.Icon, b.ID)
		return err
	}

	query := `INSERT INTO banners (badge_text, title, description, cta_text, cta_url, is_active, icon) VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := db.ExecContext(ctx, query, b.BadgeText, b.Title, b.Description, b.CtaText, b.CtaURL, activeInt, b.Icon)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		b.ID = int(id)
	}
	return nil
}

func ToggleBannerActive(ctx context.Context, db *sql.DB, id int) error {
	query := `UPDATE banners SET is_active = CASE WHEN is_active = 1 THEN 0 ELSE 1 END WHERE id = ?`
	_, err := db.ExecContext(ctx, query, id)
	return err
}

func DeleteBanner(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM banners WHERE id = ?", id)
	return err
}

// --- PARTNER DB OPERATIONS ---

func GetAllPartners(ctx context.Context, db *sql.DB) ([]Partner, error) {
	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM partners ORDER BY id DESC`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Partner
	for rows.Next() {
		var p Partner
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.Phone, &p.CompanyName, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func GetPaginatedPartners(ctx context.Context, db *sql.DB, page int) ([]Partner, PageInfo, error) {
	var totalCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM partners").Scan(&totalCount)
	pageInfo := CalculatePageInfo(page, totalCount, 10)
	offset := (pageInfo.CurrentPage - 1) * 10

	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM partners ORDER BY id DESC LIMIT 10 OFFSET ?`
	rows, err := db.QueryContext(ctx, query, offset)
	if err != nil {
		return nil, pageInfo, err
	}
	defer rows.Close()

	var list []Partner
	for rows.Next() {
		var p Partner
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.Phone, &p.CompanyName, &p.Description, &p.CreatedAt); err != nil {
			return nil, pageInfo, err
		}
		list = append(list, p)
	}
	return list, pageInfo, nil
}

func GetPartnerByID(ctx context.Context, db *sql.DB, id int) (*Partner, error) {
	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM partners WHERE id = ?`
	var p Partner
	err := db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.Email, &p.Phone, &p.CompanyName, &p.Description, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetPartnerByEmail(ctx context.Context, db *sql.DB, email string) (*Partner, error) {
	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM partners WHERE email = ?`
	var p Partner
	err := db.QueryRowContext(ctx, query, email).Scan(&p.ID, &p.Name, &p.Email, &p.Phone, &p.CompanyName, &p.Description, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func SavePartner(ctx context.Context, db *sql.DB, p *Partner, plainPassword string) error {
	if p.ID > 0 {
		if plainPassword != "" {
			hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			_, err = db.ExecContext(ctx, `UPDATE partners SET name = ?, email = ?, phone = ?, company_name = ?, description = ?, password = ? WHERE id = ?`,
				p.Name, p.Email, p.Phone, p.CompanyName, p.Description, string(hashed), p.ID)
			return err
		}
		_, err := db.ExecContext(ctx, `UPDATE partners SET name = ?, email = ?, phone = ?, company_name = ?, description = ? WHERE id = ?`,
			p.Name, p.Email, p.Phone, p.CompanyName, p.Description, p.ID)
		return err
	}

	// Check if email exists in project_managers table (Rule: 1 email cannot be both Mitra and PM)
	var pmCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM project_managers WHERE LOWER(email) = LOWER(?)", p.Email).Scan(&pmCount)
	if pmCount > 0 {
		return fmt.Errorf("Email '%s' sudah terdaftar sebagai Project Manager! 1 email tidak boleh terdaftar di 2 tipe akun.", p.Email)
	}

	if plainPassword == "" {
		plainPassword = "password123"
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	res, err := db.ExecContext(ctx, `INSERT INTO partners (name, email, phone, company_name, description, password) VALUES (?, ?, ?, ?, ?, ?)`,
		p.Name, p.Email, p.Phone, p.CompanyName, p.Description, string(hashed))
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		p.ID = int(id)
	}
	return nil
}

func DeletePartner(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM partners WHERE id = ?", id)
	return err
}

func AuthenticatePartner(ctx context.Context, db *sql.DB, email, password string) (*Partner, error) {
	var p Partner
	var hashed string
	err := db.QueryRowContext(ctx, `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), password FROM partners WHERE email = ?`, email).
		Scan(&p.ID, &p.Name, &p.Email, &p.Phone, &p.CompanyName, &p.Description, &hashed)
	if err != nil {
		return nil, errors.New("Email atau password tidak ditemukan")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return nil, errors.New("Password tidak sesuai")
	}
	return &p, nil
}

// --- PROJECT MANAGER DB OPERATIONS ---

func GetAllProjectManagers(ctx context.Context, db *sql.DB) ([]ProjectManager, error) {
	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM project_managers ORDER BY id DESC`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProjectManager
	for rows.Next() {
		var pm ProjectManager
		if err := rows.Scan(&pm.ID, &pm.Name, &pm.Email, &pm.Phone, &pm.CompanyName, &pm.Description, &pm.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, pm)
	}
	return list, nil
}

func GetPaginatedProjectManagers(ctx context.Context, db *sql.DB, page int) ([]ProjectManager, PageInfo, error) {
	var totalCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM project_managers").Scan(&totalCount)
	pageInfo := CalculatePageInfo(page, totalCount, 10)
	offset := (pageInfo.CurrentPage - 1) * 10

	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM project_managers ORDER BY id DESC LIMIT 10 OFFSET ?`
	rows, err := db.QueryContext(ctx, query, offset)
	if err != nil {
		return nil, pageInfo, err
	}
	defer rows.Close()

	var list []ProjectManager
	for rows.Next() {
		var pm ProjectManager
		if err := rows.Scan(&pm.ID, &pm.Name, &pm.Email, &pm.Phone, &pm.CompanyName, &pm.Description, &pm.CreatedAt); err != nil {
			return nil, pageInfo, err
		}
		list = append(list, pm)
	}
	return list, pageInfo, nil
}

func GetProjectManagerByID(ctx context.Context, db *sql.DB, id int) (*ProjectManager, error) {
	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM project_managers WHERE id = ?`
	var pm ProjectManager
	err := db.QueryRowContext(ctx, query, id).Scan(&pm.ID, &pm.Name, &pm.Email, &pm.Phone, &pm.CompanyName, &pm.Description, &pm.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func GetProjectManagerByEmail(ctx context.Context, db *sql.DB, email string) (*ProjectManager, error) {
	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM project_managers WHERE email = ?`
	var pm ProjectManager
	err := db.QueryRowContext(ctx, query, email).Scan(&pm.ID, &pm.Name, &pm.Email, &pm.Phone, &pm.CompanyName, &pm.Description, &pm.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func SaveProjectManager(ctx context.Context, db *sql.DB, pm *ProjectManager, plainPassword string) error {
	if pm.ID > 0 {
		if plainPassword != "" {
			hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			_, err = db.ExecContext(ctx, `UPDATE project_managers SET name = ?, email = ?, phone = ?, company_name = ?, description = ?, password = ? WHERE id = ?`,
				pm.Name, pm.Email, pm.Phone, pm.CompanyName, pm.Description, string(hashed), pm.ID)
			return err
		}
		_, err := db.ExecContext(ctx, `UPDATE project_managers SET name = ?, email = ?, phone = ?, company_name = ?, description = ? WHERE id = ?`,
			pm.Name, pm.Email, pm.Phone, pm.CompanyName, pm.Description, pm.ID)
		return err
	}

	// Check if email exists in partners table (Rule: 1 email cannot be both Mitra and PM)
	var partnerCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM partners WHERE LOWER(email) = LOWER(?)", pm.Email).Scan(&partnerCount)
	if partnerCount > 0 {
		return fmt.Errorf("Email '%s' sudah terdaftar sebagai Mitra! 1 email tidak boleh terdaftar di 2 tipe akun.", pm.Email)
	}

	if plainPassword == "" {
		plainPassword = "password123"
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	res, err := db.ExecContext(ctx, `INSERT INTO project_managers (name, email, phone, company_name, description, password) VALUES (?, ?, ?, ?, ?, ?)`,
		pm.Name, pm.Email, pm.Phone, pm.CompanyName, pm.Description, string(hashed))
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		pm.ID = int(id)
	}
	return nil
}

// --- PRODUCT IMAGES DB OPERATIONS ---

func GetProductImages(ctx context.Context, db *sql.DB, productID int) ([]ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, is_thumbnail
		FROM product_images
		WHERE product_id = ?
		ORDER BY is_thumbnail DESC, id ASC
	`
	rows, err := db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []ProductImage
	for rows.Next() {
		var img ProductImage
		var isThumbInt int
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ImageURL, &isThumbInt); err != nil {
			return nil, err
		}
		img.IsThumbnail = (isThumbInt == 1)
		images = append(images, img)
	}
	return images, nil
}

func AddProductImage(ctx context.Context, db *sql.DB, productID int, imageURL string, isThumbnail bool) (int64, error) {
	var count int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_images WHERE product_id = ?", productID).Scan(&count)
	if count >= 20 {
		return 0, fmt.Errorf("maksimal 20 gambar per produk")
	}

	isThumbInt := 0
	if isThumbnail || count == 0 {
		isThumbInt = 1
	}

	res, err := db.ExecContext(ctx, "INSERT INTO product_images (product_id, image_url, is_thumbnail) VALUES (?, ?, ?)", productID, imageURL, isThumbInt)
	if err != nil {
		return 0, err
	}
	imgID, _ := res.LastInsertId()

	if isThumbInt == 1 {
		_, _ = db.ExecContext(ctx, "UPDATE products SET thumbnail = ? WHERE id = ?", imageURL, productID)
	}
	return imgID, nil
}

func SetProductThumbnail(ctx context.Context, db *sql.DB, productID int, imageID int) error {
	var imageURL string
	err := db.QueryRowContext(ctx, "SELECT image_url FROM product_images WHERE id = ? AND product_id = ?", imageID, productID).Scan(&imageURL)
	if err != nil {
		return fmt.Errorf("gambar tidak ditemukan: %w", err)
	}

	_, _ = db.ExecContext(ctx, "UPDATE product_images SET is_thumbnail = 0 WHERE product_id = ?", productID)
	_, _ = db.ExecContext(ctx, "UPDATE product_images SET is_thumbnail = 1 WHERE id = ?", imageID)
	_, _ = db.ExecContext(ctx, "UPDATE products SET thumbnail = ? WHERE id = ?", imageURL, productID)
	return nil
}

func DeleteProductImage(ctx context.Context, db *sql.DB, productID int, imageID int) (string, error) {
	var imageURL string
	var isThumbInt int
	err := db.QueryRowContext(ctx, "SELECT image_url, is_thumbnail FROM product_images WHERE id = ? AND product_id = ?", imageID, productID).Scan(&imageURL, &isThumbInt)
	if err != nil {
		return "", fmt.Errorf("gambar tidak ditemukan: %w", err)
	}

	_, err = db.ExecContext(ctx, "DELETE FROM product_images WHERE id = ?", imageID)
	if err != nil {
		return "", err
	}

	// If deleted image was thumbnail, promote another image to thumbnail if available
	if isThumbInt == 1 {
		var nextImgID int
		var nextURL string
		errNext := db.QueryRowContext(ctx, "SELECT id, image_url FROM product_images WHERE product_id = ? ORDER BY id ASC LIMIT 1", productID).Scan(&nextImgID, &nextURL)
		if errNext == nil {
			_, _ = db.ExecContext(ctx, "UPDATE product_images SET is_thumbnail = 1 WHERE id = ?", nextImgID)
			_, _ = db.ExecContext(ctx, "UPDATE products SET thumbnail = ? WHERE id = ?", nextURL, productID)
		} else {
			_, _ = db.ExecContext(ctx, "UPDATE products SET thumbnail = '' WHERE id = ?", productID)
		}
	}

	return imageURL, nil
}

func DeleteProjectManager(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM project_managers WHERE id = ?", id)
	return err
}

func AuthenticateProjectManager(ctx context.Context, db *sql.DB, email, password string) (*ProjectManager, error) {
	var pm ProjectManager
	var hashed string
	err := db.QueryRowContext(ctx, `SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), password FROM project_managers WHERE email = ?`, email).
		Scan(&pm.ID, &pm.Name, &pm.Email, &pm.Phone, &pm.CompanyName, &pm.Description, &hashed)
	if err != nil {
		return nil, errors.New("Email atau password tidak ditemukan")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return nil, errors.New("Password tidak sesuai")
	}
	return &pm, nil
}

// --- ORDERS & CUSTOMER AUTH DB OPERATIONS ---

func IsEmailRegisteredInAnyRole(ctx context.Context, db *sql.DB, email string) (bool, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return false, "", nil
	}

	var partnerCount int
	errP := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM partners WHERE LOWER(email) = ?", email).Scan(&partnerCount)
	if errP == nil && partnerCount > 0 {
		return true, "Mitra", nil
	}

	var pmCount int
	errPM := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM project_managers WHERE LOWER(email) = ?", email).Scan(&pmCount)
	if errPM == nil && pmCount > 0 {
		return true, "Project Manager", nil
	}

	return false, "", nil
}

func CreateCustomerSession(ctx context.Context, db *sql.DB, email string, userType string) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days session
	_, err = db.ExecContext(ctx, "INSERT INTO customer_sessions (token, email, user_type, expires_at) VALUES (?, ?, ?, ?)",
		token, email, userType, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

func ValidateCustomerSession(ctx context.Context, db *sql.DB, token string) (*CustomerUser, error) {
	if token == "" {
		return nil, errors.New("token kosong")
	}
	var cust CustomerUser
	var expiresAt time.Time
	err := db.QueryRowContext(ctx, "SELECT email, user_type, expires_at FROM customer_sessions WHERE token = ?", token).Scan(&cust.Email, &cust.UserType, &expiresAt)
	if err != nil {
		return nil, err
	}
	if time.Now().After(expiresAt) {
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_sessions WHERE token = ?", token)
		return nil, errors.New("sesi kadaluarsa")
	}

	// Fetch Name from corresponding table
	if strings.EqualFold(cust.UserType, "partner") || strings.EqualFold(cust.UserType, "Mitra") {
		_ = db.QueryRowContext(ctx, "SELECT name FROM partners WHERE LOWER(email) = ?", strings.ToLower(cust.Email)).Scan(&cust.Name)
	} else {
		_ = db.QueryRowContext(ctx, "SELECT name FROM project_managers WHERE LOWER(email) = ?", strings.ToLower(cust.Email)).Scan(&cust.Name)
	}
	if cust.Name == "" {
		cust.Name = cust.Email
	}

	return &cust, nil
}

func DeleteCustomerSession(ctx context.Context, db *sql.DB, token string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM customer_sessions WHERE token = ?", token)
	return err
}

func CreateOrder(ctx context.Context, db *sql.DB, email string, userType string, userName string, items []OrderItem) (int64, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("keranjang belanja kosong")
	}

	var totalPrice float64
	for i := range items {
		items[i].Subtotal = items[i].Price * float64(items[i].Quantity)
		totalPrice += items[i].Subtotal
	}

	res, err := db.ExecContext(ctx, `
		INSERT INTO orders (email, user_type, user_name, total_price, status)
		VALUES (?, ?, ?, ?, 'PENDING')
	`, email, userType, userName, totalPrice)
	if err != nil {
		return 0, fmt.Errorf("gagal membuat pesanan: %w", err)
	}

	orderID, _ := res.LastInsertId()

	for _, item := range items {
		_, errItem := db.ExecContext(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, price, quantity, subtotal)
			VALUES (?, ?, ?, ?, ?, ?)
		`, orderID, item.ProductID, item.ProductName, item.Price, item.Quantity, item.Subtotal)
		if errItem != nil {
			return orderID, errItem
		}
	}

	return orderID, nil
}

func GetAllOrders(ctx context.Context, db *sql.DB) ([]Order, error) {
	query := `
		SELECT id, email, user_type, user_name, total_price, status, created_at
		FROM orders
		ORDER BY id DESC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Email, &o.UserType, &o.UserName, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}

		// Fetch items for each order
		itemRows, errItems := db.QueryContext(ctx, `
			SELECT id, order_id, product_id, product_name, price, quantity, subtotal
			FROM order_items
			WHERE order_id = ?
			ORDER BY id ASC
		`, o.ID)
		if errItems == nil {
			for itemRows.Next() {
				var item OrderItem
				if errScan := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.Price, &item.Quantity, &item.Subtotal); errScan == nil {
					o.Items = append(o.Items, item)
				}
			}
			itemRows.Close()
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func CompleteOrder(ctx context.Context, db *sql.DB, orderID int) error {
	var status string
	err := db.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = ?", orderID).Scan(&status)
	if err != nil {
		return fmt.Errorf("pesanan tidak ditemukan: %w", err)
	}
	if status == "SELESAI" {
		return fmt.Errorf("pesanan sudah diselesaikan sebelumnya")
	}

	// Fetch items to deduct stock
	itemRows, errItems := db.QueryContext(ctx, "SELECT product_id, quantity FROM order_items WHERE order_id = ?", orderID)
	if errItems != nil {
		return errItems
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var prodID, qty int
		if errScan := itemRows.Scan(&prodID, &qty); errScan == nil && prodID > 0 && qty > 0 {
			_, _ = db.ExecContext(ctx, "UPDATE products SET stock = MAX(0, stock - ?) WHERE id = ?", qty, prodID)
		}
	}

	_, errUpdate := db.ExecContext(ctx, "UPDATE orders SET status = 'SELESAI' WHERE id = ?", orderID)
	return errUpdate
}

// --- ADMIN GLOBAL SEARCH ---

func SearchAdminGlobal(ctx context.Context, db *sql.DB, searchType string, queryStr string) (*AdminSearchResult, error) {
	queryStr = strings.TrimSpace(queryStr)
	likePattern := "%" + queryStr + "%"
	res := &AdminSearchResult{}

	searchType = strings.ToLower(strings.TrimSpace(searchType))
	if searchType == "" {
		searchType = "all"
	}

	// 1. Categories
	if searchType == "all" || searchType == "categories" {
		rows, err := db.QueryContext(ctx, "SELECT id, name, slug, COALESCE(description, '') FROM categories WHERE name LIKE ? OR description LIKE ? ORDER BY id DESC", likePattern, likePattern)
		if err == nil {
			for rows.Next() {
				var c Category
				if scanErr := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description); scanErr == nil {
					res.Categories = append(res.Categories, c)
				}
			}
			rows.Close()
		}
	}

	// 2. Brands
	if searchType == "all" || searchType == "brands" {
		rows, err := db.QueryContext(ctx, "SELECT id, name, slug, COALESCE(logo_url, ''), COALESCE(description, '') FROM brands WHERE name LIKE ? OR description LIKE ? ORDER BY id DESC", likePattern, likePattern)
		if err == nil {
			for rows.Next() {
				var b Brand
				if scanErr := rows.Scan(&b.ID, &b.Name, &b.Slug, &b.LogoURL, &b.Description); scanErr == nil {
					res.Brands = append(res.Brands, b)
				}
			}
			rows.Close()
		}
	}

	// 3. Products
	if searchType == "all" || searchType == "products" {
		products, err := SearchProducts(ctx, db, queryStr, "")
		if err == nil {
			res.Products = products
		}
	}

	// 4. Partners
	if searchType == "all" || searchType == "partners" {
		rows, err := db.QueryContext(ctx, "SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM partners WHERE name LIKE ? OR email LIKE ? OR company_name LIKE ? ORDER BY id DESC", likePattern, likePattern, likePattern)
		if err == nil {
			for rows.Next() {
				var p Partner
				if scanErr := rows.Scan(&p.ID, &p.Name, &p.Email, &p.Phone, &p.CompanyName, &p.Description, &p.CreatedAt); scanErr == nil {
					res.Partners = append(res.Partners, p)
				}
			}
			rows.Close()
		}
	}

	// 5. Project Managers
	if searchType == "all" || searchType == "project_managers" {
		rows, err := db.QueryContext(ctx, "SELECT id, name, email, COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(description, ''), created_at FROM project_managers WHERE name LIKE ? OR email LIKE ? OR company_name LIKE ? ORDER BY id DESC", likePattern, likePattern, likePattern)
		if err == nil {
			for rows.Next() {
				var pm ProjectManager
				if scanErr := rows.Scan(&pm.ID, &pm.Name, &pm.Email, &pm.Phone, &pm.CompanyName, &pm.Description, &pm.CreatedAt); scanErr == nil {
					res.ProjectManagers = append(res.ProjectManagers, pm)
				}
			}
			rows.Close()
		}
	}

	// 6. Banners
	if searchType == "all" || searchType == "banners" {
		rows, err := db.QueryContext(ctx, "SELECT id, badge_text, title, description, cta_text, cta_url, is_active, icon, created_at FROM banners WHERE title LIKE ? OR description LIKE ? OR badge_text LIKE ? ORDER BY id DESC", likePattern, likePattern, likePattern)
		if err == nil {
			for rows.Next() {
				var b Banner
				var activeInt int
				if scanErr := rows.Scan(&b.ID, &b.BadgeText, &b.Title, &b.Description, &b.CtaText, &b.CtaURL, &activeInt, &b.Icon, &b.CreatedAt); scanErr == nil {
					b.IsActive = (activeInt == 1)
					res.Banners = append(res.Banners, b)
				}
			}
			rows.Close()
		}
	}

	// 7. Orders
	if searchType == "all" || searchType == "orders" {
		allOrders, err := GetAllOrders(ctx, db)
		if err == nil {
			for _, o := range allOrders {
				queryLower := strings.ToLower(queryStr)
				if strings.Contains(strings.ToLower(o.Email), queryLower) ||
					strings.Contains(strings.ToLower(o.UserName), queryLower) ||
					strings.Contains(strings.ToLower(o.UserType), queryLower) ||
					strings.Contains(fmt.Sprintf("%d", o.ID), queryLower) {
					res.Orders = append(res.Orders, o)
				} else {
					for _, item := range o.Items {
						if strings.Contains(strings.ToLower(item.ProductName), queryLower) {
							res.Orders = append(res.Orders, o)
							break
						}
					}
				}
			}
		}
	}

	res.TotalCount = len(res.Categories) + len(res.Brands) + len(res.Products) +
		len(res.Partners) + len(res.ProjectManagers) + len(res.Banners) + len(res.Orders)

	return res, nil
}

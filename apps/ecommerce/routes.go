package ecommerce

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-ai/logger"
	"go-ai/render"
	"go-ai/router"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailPattern.MatchString(email)
}

func SaveUploadedImage(fileHeader *multipart.FileHeader, targetSubfolder string) (string, error) {
	if fileHeader == nil {
		return "", nil
	}

	// 1. Strict Size Limit: Max 2MB (2 * 1024 * 1024 bytes)
	if fileHeader.Size > 2*1024*1024 {
		return "", errors.New("Ukuran file maksimal 2MB!")
	}

	// 2. Strict Format Limit: JPG or PNG only
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", errors.New("Format file hanya boleh JPG atau PNG!")
	}

	// 3. Ensure target directory exists
	targetDir := filepath.Join("asset", targetSubfolder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", errors.New("Gagal membuat direktori penyimpanan asset")
	}

	// 4. Open uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return "", errors.New("Gagal membaca file gambar")
	}
	defer src.Close()

	// 5. Generate clean unique filename
	cleanFilename := CreateSlug(strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename)))
	if cleanFilename == "" {
		cleanFilename = "image"
	}
	newFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), cleanFilename, ext)
	dstPath := filepath.Join(targetDir, newFilename)

	// 6. Save file to disk
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", errors.New("Gagal menyimpan file gambar ke disk")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", errors.New("Gagal menyalin isi file gambar")
	}

	// 7. Return relative asset URL
	return "/" + filepath.ToSlash(dstPath), nil
}

// RegisterRoutes registers all home and admin routes for the e-commerce application.
func RegisterRoutes(r *router.Router, tmpl *render.Engine, log logger.Logger, db *sql.DB) {

	// Serve static upload assets (/asset/brands/... & /asset/products/...)
	r.Handle("GET /asset/", http.StripPrefix("/asset/", http.FileServer(http.Dir("asset"))))

	// Helper to check admin session from cookie
	getAdminUser := func(req *http.Request) *AdminUser {
		cookie, err := req.Cookie("admin_session")
		if err != nil || cookie.Value == "" {
			return nil
		}
		user, err := ValidateAdminSession(req.Context(), db, cookie.Value)
		if err != nil {
			return nil
		}
		return user
	}

	// Helper to require admin session (redirects if unauthenticated)
	requireAdmin := func(w http.ResponseWriter, req *http.Request) *AdminUser {
		user := getAdminUser(req)
		if user == nil {
			if req.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/admin/login")
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				http.Redirect(w, req, "/admin/login", http.StatusSeeOther)
			}
			return nil
		}
		return user
	}

	// 1. Public Home Storefront Route (GET /): Renders e-commerce homepage with Banners, Products, Categories & Brands
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		banners, err := GetActiveBanners(req.Context(), db)
		if err != nil {
			log.Error("Failed to fetch active banners for home", "error", err)
		}
		products, err := GetAllProducts(req.Context(), db)
		if err != nil {
			log.Error("Failed to fetch products for home", "error", err)
		}
		categories, err := GetAllCategories(req.Context(), db)
		if err != nil {
			log.Error("Failed to fetch categories for home", "error", err)
		}
		brands, err := GetAllBrands(req.Context(), db)
		if err != nil {
			log.Error("Failed to fetch brands for home", "error", err)
		}

		tmpl.Render(w, http.StatusOK, "store_home", map[string]interface{}{
			"Banners":    banners,
			"Products":   products,
			"Categories": categories,
			"Brands":     brands,
		})
	})

	// 2. Product Detail Route (GET /product/{slug}): Dedicated detail page for a product by unique slug
	r.Get("/product/{slug}", func(w http.ResponseWriter, req *http.Request) {
		slug := req.PathValue("slug")
		if slug == "" {
			http.Redirect(w, req, "/", http.StatusSeeOther)
			return
		}

		prod, err := GetProductBySlug(req.Context(), db, slug)
		if err != nil || prod == nil {
			http.NotFound(w, req)
			return
		}

		categories, _ := GetAllCategories(req.Context(), db)
		brands, _ := GetAllBrands(req.Context(), db)
		relatedProducts, _ := GetRelatedProducts(req.Context(), db, prod.CategoryID, prod.ID, 4)

		tmpl.Render(w, http.StatusOK, "store_product_detail", map[string]interface{}{
			"Product":         prod,
			"RelatedProducts": relatedProducts,
			"Categories":      categories,
			"Brands":          brands,
		})
	})

	// 3. Search Results Route (GET /search?q=...&cat=...): Dedicated search results page UI
	r.Get("/search", func(w http.ResponseWriter, req *http.Request) {
		searchQuery := strings.TrimSpace(req.URL.Query().Get("q"))
		categorySlug := strings.TrimSpace(req.URL.Query().Get("cat"))

		products, err := SearchProducts(req.Context(), db, searchQuery, categorySlug)
		if err != nil {
			log.Error("Failed to search products", "error", err)
		}

		categories, _ := GetAllCategories(req.Context(), db)
		brands, _ := GetAllBrands(req.Context(), db)

		tmpl.Render(w, http.StatusOK, "store_search", map[string]interface{}{
			"Products":     products,
			"SearchQuery":  searchQuery,
			"CategorySlug": categorySlug,
			"TotalCount":   len(products),
			"Categories":   categories,
			"Brands":       brands,
		})
	})

	// Fallback route for /static/img/no-image.png (returns 200 OK SVG placeholder)
	r.Get("/static/img/no-image.png", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 80 80"><rect width="80" height="80" fill="#f1f5f9"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" font-size="28">📦</text></svg>`))
	})

	getCustomerUser := func(req *http.Request) *CustomerUser {
		cookie, err := req.Cookie("customer_token")
		if err != nil || cookie.Value == "" {
			return nil
		}
		cust, errVal := ValidateCustomerSession(req.Context(), db, cookie.Value)
		if errVal != nil {
			return nil
		}
		return cust
	}

	// 4. Cart Page Route (GET /cart)
	r.Get("/cart", func(w http.ResponseWriter, req *http.Request) {
		categories, _ := GetAllCategories(req.Context(), db)
		cust := getCustomerUser(req)

		tmpl.Render(w, http.StatusOK, "store_cart", map[string]interface{}{
			"Categories": categories,
			"Customer":   cust,
		})
	})

	// 5. Customer Login / Register Routes (GET /login, POST /login, POST /register, GET /logout)
	r.Get("/login", func(w http.ResponseWriter, req *http.Request) {
		if getCustomerUser(req) != nil {
			redirect := req.URL.Query().Get("redirect")
			if redirect == "" {
				redirect = "/cart"
			}
			http.Redirect(w, req, redirect, http.StatusSeeOther)
			return
		}
		tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
			"Redirect": req.URL.Query().Get("redirect"),
		})
	})

	r.Post("/login", func(w http.ResponseWriter, req *http.Request) {
		email := strings.TrimSpace(req.FormValue("email"))
		password := strings.TrimSpace(req.FormValue("password"))
		userType := strings.TrimSpace(req.FormValue("user_type"))
		redirect := strings.TrimSpace(req.FormValue("redirect"))
		if redirect == "" {
			redirect = "/cart"
		}

		if email == "" || password == "" {
			tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
				"Error":    "Email dan password wajib diisi!",
				"Redirect": redirect,
			})
			return
		}

		var userName string
		if strings.EqualFold(userType, "partner") || strings.EqualFold(userType, "Mitra") {
			p, err := AuthenticatePartner(req.Context(), db, email, password)
			if err != nil {
				tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
					"Error":    err.Error(),
					"Redirect": redirect,
				})
				return
			}
			userName = p.Name
		} else {
			pm, err := AuthenticateProjectManager(req.Context(), db, email, password)
			if err != nil {
				tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
					"Error":    err.Error(),
					"Redirect": redirect,
				})
				return
			}
			userName = pm.Name
		}
		_ = userName

		token, err := CreateCustomerSession(req.Context(), db, email, userType)
		if err != nil {
			tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
				"Error":    "Gagal membuat sesi login: " + err.Error(),
				"Redirect": redirect,
			})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "customer_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
		})

		http.Redirect(w, req, redirect, http.StatusSeeOther)
	})

	r.Post("/register", func(w http.ResponseWriter, req *http.Request) {
		email := strings.TrimSpace(req.FormValue("email"))
		password := strings.TrimSpace(req.FormValue("password"))
		name := strings.TrimSpace(req.FormValue("name"))
		phone := strings.TrimSpace(req.FormValue("phone"))
		companyName := strings.TrimSpace(req.FormValue("company_name"))
		userType := strings.TrimSpace(req.FormValue("user_type"))
		redirect := strings.TrimSpace(req.FormValue("redirect"))
		if redirect == "" {
			redirect = "/cart"
		}

		if name == "" || email == "" || password == "" {
			tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
				"Error":    "Nama, email, dan password wajib diisi!",
				"Redirect": redirect,
			})
			return
		}

		// Enforce 1 email 1 role rule!
		exists, role, _ := IsEmailRegisteredInAnyRole(req.Context(), db, email)
		if exists {
			tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
				"Error":    fmt.Sprintf("Email '%s' sudah terdaftar sebagai %s! 1 email tidak boleh memiliki 2 tipe akun.", email, role),
				"Redirect": redirect,
			})
			return
		}

		if strings.EqualFold(userType, "partner") || strings.EqualFold(userType, "Mitra") {
			p := Partner{
				Name:        name,
				Email:       email,
				Phone:       phone,
				CompanyName: companyName,
			}
			if err := SavePartner(req.Context(), db, &p, password); err != nil {
				tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
					"Error":    err.Error(),
					"Redirect": redirect,
				})
				return
			}
		} else {
			pm := ProjectManager{
				Name:        name,
				Email:       email,
				Phone:       phone,
				CompanyName: companyName,
			}
			if err := SaveProjectManager(req.Context(), db, &pm, password); err != nil {
				tmpl.Render(w, http.StatusOK, "store_login", map[string]interface{}{
					"Error":    err.Error(),
					"Redirect": redirect,
				})
				return
			}
		}

		token, _ := CreateCustomerSession(req.Context(), db, email, userType)
		http.SetCookie(w, &http.Cookie{
			Name:     "customer_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
		})

		http.Redirect(w, req, redirect, http.StatusSeeOther)
	})

	r.Get("/logout", func(w http.ResponseWriter, req *http.Request) {
		if cookie, err := req.Cookie("customer_token"); err == nil && cookie.Value != "" {
			_ = DeleteCustomerSession(req.Context(), db, cookie.Value)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "customer_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(-1 * time.Hour),
		})
		http.Redirect(w, req, "/", http.StatusSeeOther)
	})

	// 6. Checkout Route (POST /checkout)
	r.Post("/checkout", func(w http.ResponseWriter, req *http.Request) {
		cust := getCustomerUser(req)
		if cust == nil {
			http.Error(w, "Unauthorized: Silakan login terlebih dahulu", http.StatusUnauthorized)
			return
		}

		var payload struct {
			Items []struct {
				ProductID   int     `json:"product_id"`
				ProductName string  `json:"product_name"`
				Price       float64 `json:"price"`
				Quantity    int     `json:"quantity"`
			} `json:"items"`
		}

		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil || len(payload.Items) == 0 {
			render.JSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   "Payload pesanan tidak valid",
			})
			return
		}

		var orderItems []OrderItem
		for _, item := range payload.Items {
			if item.Quantity <= 0 {
				item.Quantity = 1
			}
			orderItems = append(orderItems, OrderItem{
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
		}

		orderID, err := CreateOrder(req.Context(), db, cust.Email, cust.UserType, cust.Name, orderItems)
		if err != nil {
			render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		render.JSON(w, http.StatusOK, map[string]interface{}{
			"success":  true,
			"order_id": orderID,
		})
	})

	// --- AUTHENTICATION ROUTES ---

	// GET /admin/login
	r.Get("/admin/login", func(w http.ResponseWriter, req *http.Request) {
		if getAdminUser(req) != nil {
			http.Redirect(w, req, "/admin", http.StatusSeeOther)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_login", map[string]interface{}{})
	})

	// POST /admin/login
	r.Post("/admin/login", func(w http.ResponseWriter, req *http.Request) {
		username := strings.TrimSpace(req.FormValue("username"))
		password := strings.TrimSpace(req.FormValue("password"))

		if username == "" || password == "" {
			tmpl.Render(w, http.StatusOK, "admin_login", map[string]interface{}{
				"Error":    "Username dan password wajib diisi!",
				"Username": username,
			})
			return
		}

		user, err := GetAdminUserByUsername(req.Context(), db, username)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			tmpl.Render(w, http.StatusOK, "admin_login", map[string]interface{}{
				"Error":    "Username atau password salah!",
				"Username": username,
			})
			return
		}

		token, err := CreateAdminSession(req.Context(), db, user.ID)
		if err != nil {
			log.Error("Failed to create admin session", "error", err)
			http.Error(w, "Gagal membuat sesi", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(24 * time.Hour),
		})

		http.Redirect(w, req, "/admin", http.StatusSeeOther)
	})

	// GET /admin/logout & POST /admin/logout
	handleLogout := func(w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("admin_session")
		if err == nil && cookie.Value != "" {
			_ = DeleteAdminSession(req.Context(), db, cookie.Value)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
			Expires:  time.Now().Add(-1 * time.Hour),
		})
		http.Redirect(w, req, "/admin/login", http.StatusSeeOther)
	}

	r.Get("/admin/logout", handleLogout)
	r.Post("/admin/logout", handleLogout)

	// Helper to load dashboard payload
	getDashboardData := func(req *http.Request, adminUser *AdminUser) map[string]interface{} {
		stats, _ := GetAdminStats(req.Context(), db)
		products, _ := GetAllProducts(req.Context(), db)
		brands, _ := GetAllBrands(req.Context(), db)
		categories, _ := GetAllCategories(req.Context(), db)
		return map[string]interface{}{
			"ActiveTab":  "dashboard",
			"Stats":      stats,
			"Products":   products,
			"Brands":     brands,
			"Categories": categories,
			"AdminUser":  adminUser,
		}
	}

	// 2. Admin Dashboard Main Page (GET /admin)
	r.Get("/admin", func(w http.ResponseWriter, req *http.Request) {
		user := requireAdmin(w, req)
		if user == nil {
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", getDashboardData(req, user))
	})

	// Dashboard partial view
	r.Get("/admin/dashboard", func(w http.ResponseWriter, req *http.Request) {
		user := requireAdmin(w, req)
		if user == nil {
			return
		}
		data := getDashboardData(req, user)
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_dashboard", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// --- CATEGORIES ROUTES ---

	// GET /admin/categories
	r.Get("/admin/categories", func(w http.ResponseWriter, req *http.Request) {
		user := requireAdmin(w, req)
		if user == nil {
			return
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		categories, pageInfo, err := GetPaginatedCategories(req.Context(), db, page)
		if err != nil {
			log.Error("Failed to fetch categories", "error", err)
		}
		data := map[string]interface{}{
			"ActiveTab":  "categories",
			"Categories": categories,
			"PageInfo":   pageInfo,
			"AdminUser":  user,
		}
		if req.Header.Get("X-Page-Only") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_category_table", data)
			return
		}
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_categories", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// GET /admin/categories/modal
	r.Get("/admin/categories/modal", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		idStr := req.URL.Query().Get("id")
		var cat Category
		if idStr != "" {
			if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
				if c, err := GetCategoryByID(req.Context(), db, id); err == nil && c != nil {
					cat = *c
				}
			}
		}
		tmpl.Render(w, http.StatusOK, "admin_category_modal", map[string]interface{}{
			"Category": cat,
		})
	})

	// POST /admin/categories (Save/Update)
	r.Post("/admin/categories", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.FormValue("id"))
		name := strings.TrimSpace(req.FormValue("name"))
		slug := strings.TrimSpace(req.FormValue("slug"))
		description := strings.TrimSpace(req.FormValue("description"))

		if name == "" || slug == "" || description == "" {
			http.Error(w, "Semua field wajib diisi!", http.StatusBadRequest)
			return
		}

		cat := Category{
			ID:          id,
			Name:        name,
			Slug:        slug,
			Description: description,
		}
		if err := SaveCategory(req.Context(), db, &cat); err != nil {
			log.Error("Failed to save category", "error", err)
			http.Error(w, "Gagal menyimpan kategori", http.StatusInternalServerError)
			return
		}
		page, _ := strconv.Atoi(req.FormValue("page"))
		categories, pageInfo, _ := GetPaginatedCategories(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_category_table", map[string]interface{}{
			"Categories": categories,
			"PageInfo":   pageInfo,
		})
	})

	// DELETE /admin/categories
	r.Delete("/admin/categories", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if id > 0 {
			if err := DeleteCategory(req.Context(), db, id); err != nil {
				log.Error("Failed to delete category", "error", err)
			}
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		categories, pageInfo, _ := GetPaginatedCategories(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_category_table", map[string]interface{}{
			"Categories": categories,
			"PageInfo":   pageInfo,
		})
	})

	// --- BRANDS ROUTES ---

	// GET /admin/brands
	r.Get("/admin/brands", func(w http.ResponseWriter, req *http.Request) {
		user := requireAdmin(w, req)
		if user == nil {
			return
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		brands, pageInfo, err := GetPaginatedBrands(req.Context(), db, page)
		if err != nil {
			log.Error("Failed to fetch brands", "error", err)
		}
		data := map[string]interface{}{
			"ActiveTab": "brands",
			"Brands":    brands,
			"PageInfo":  pageInfo,
			"AdminUser": user,
		}
		if req.Header.Get("X-Page-Only") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_brand_table", data)
			return
		}
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_brands", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// GET /admin/brands/modal
	r.Get("/admin/brands/modal", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		idStr := req.URL.Query().Get("id")
		var brand Brand
		if idStr != "" {
			if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
				if b, err := GetBrandByID(req.Context(), db, id); err == nil && b != nil {
					brand = *b
				}
			}
		}
		tmpl.Render(w, http.StatusOK, "admin_brand_modal", map[string]interface{}{
			"Brand": brand,
		})
	})

	// POST /admin/brands (Save/Update with Image Upload)
	r.Post("/admin/brands", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		_ = req.ParseMultipartForm(10 << 20)

		id, _ := strconv.Atoi(req.FormValue("id"))
		name := strings.TrimSpace(req.FormValue("name"))
		slug := strings.TrimSpace(req.FormValue("slug"))
		logoURL := strings.TrimSpace(req.FormValue("logo_url"))
		description := strings.TrimSpace(req.FormValue("description"))

		// Process image file upload
		file, header, errFile := req.FormFile("logo_file")
		if errFile == nil && header != nil {
			_ = file.Close()
			savedPath, uploadErr := SaveUploadedImage(header, "brands")
			if uploadErr != nil {
				http.Error(w, uploadErr.Error(), http.StatusBadRequest)
				return
			}
			logoURL = savedPath
		}

		if logoURL == "" && id > 0 {
			if existing, err := GetBrandByID(req.Context(), db, id); err == nil && existing != nil {
				logoURL = existing.LogoURL
			}
		}

		if name == "" || slug == "" || logoURL == "" || description == "" {
			http.Error(w, "Semua field (termasuk Logo Brand) wajib diisi!", http.StatusBadRequest)
			return
		}

		brand := Brand{
			ID:          id,
			Name:        name,
			Slug:        slug,
			LogoURL:     logoURL,
			Description: description,
		}
		if err := SaveBrand(req.Context(), db, &brand); err != nil {
			log.Error("Failed to save brand", "error", err)
			http.Error(w, "Gagal menyimpan brand", http.StatusInternalServerError)
			return
		}
		page, _ := strconv.Atoi(req.FormValue("page"))
		brands, pageInfo, _ := GetPaginatedBrands(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_brand_table", map[string]interface{}{
			"Brands":   brands,
			"PageInfo": pageInfo,
		})
	})

	// DELETE /admin/brands
	r.Delete("/admin/brands", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if id > 0 {
			if err := DeleteBrand(req.Context(), db, id); err != nil {
				log.Error("Failed to delete brand", "error", err)
			}
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		brands, pageInfo, _ := GetPaginatedBrands(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_brand_table", map[string]interface{}{
			"Brands":   brands,
			"PageInfo": pageInfo,
		})
	})

	// --- PRODUCTS ROUTES ---

	// GET /admin/products
	r.Get("/admin/products", func(w http.ResponseWriter, req *http.Request) {
		user := requireAdmin(w, req)
		if user == nil {
			return
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		products, pageInfo, err := GetPaginatedProducts(req.Context(), db, page)
		if err != nil {
			log.Error("Failed to fetch products", "error", err)
		}
		data := map[string]interface{}{
			"ActiveTab": "products",
			"Products":  products,
			"PageInfo":  pageInfo,
			"AdminUser": user,
		}
		if req.Header.Get("X-Page-Only") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_product_table", data)
			return
		}
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_products", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// GET /admin/products/modal
	r.Get("/admin/products/modal", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		idStr := req.URL.Query().Get("id")
		var prod Product
		if idStr != "" {
			if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
				if p, err := GetProductByID(req.Context(), db, id); err == nil && p != nil {
					prod = *p
				}
			}
		}

		brands, _ := GetAllBrands(req.Context(), db)
		categories, _ := GetAllCategories(req.Context(), db)

		tmpl.Render(w, http.StatusOK, "admin_product_modal", map[string]interface{}{
			"Product":    prod,
			"Brands":     brands,
			"Categories": categories,
		})
	})

	// POST /admin/products (Save/Update with Multi-Image Upload)
	r.Post("/admin/products", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		_ = req.ParseMultipartForm(50 << 20) // 50MB max total form size

		id, _ := strconv.Atoi(req.FormValue("id"))
		name := strings.TrimSpace(req.FormValue("name"))
		slug := strings.TrimSpace(req.FormValue("slug"))
		brandID, _ := strconv.Atoi(req.FormValue("brand_id"))
		categoryID, _ := strconv.Atoi(req.FormValue("category_id"))
		priceStr := strings.TrimSpace(req.FormValue("price"))
		stockStr := strings.TrimSpace(req.FormValue("stock"))
		thumbnail := strings.TrimSpace(req.FormValue("thumbnail"))
		description := strings.TrimSpace(req.FormValue("description"))

		if thumbnail == "" && id > 0 {
			if existing, err := GetProductByID(req.Context(), db, id); err == nil && existing != nil {
				thumbnail = existing.Thumbnail
			}
		}

		price, errPrice := strconv.ParseFloat(priceStr, 64)
		stock, errStock := strconv.Atoi(stockStr)

		if name == "" || slug == "" || brandID <= 0 || categoryID <= 0 || priceStr == "" || errPrice != nil || price <= 0 || stockStr == "" || errStock != nil || stock < 0 || description == "" {
			http.Error(w, "Semua field wajib diisi dengan benar!", http.StatusBadRequest)
			return
		}

		prod := Product{
			ID:          id,
			Name:        name,
			Slug:        slug,
			BrandID:     brandID,
			CategoryID:  categoryID,
			Price:       price,
			Stock:       stock,
			Thumbnail:   thumbnail,
			Description: description,
		}

		if err := SaveProduct(req.Context(), db, &prod); err != nil {
			log.Error("Failed to save product", "error", err)
			http.Error(w, "Gagal menyimpan produk", http.StatusInternalServerError)
			return
		}

		// Process multiple image files upload (up to 20 images per product)
		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			var filesToUpload []*multipart.FileHeader
			if singleFiles, ok := req.MultipartForm.File["thumbnail_file"]; ok {
				filesToUpload = append(filesToUpload, singleFiles...)
			}
			if multiFiles, ok := req.MultipartForm.File["images"]; ok {
				filesToUpload = append(filesToUpload, multiFiles...)
			}

			for idx, fileHeader := range filesToUpload {
				if fileHeader == nil || fileHeader.Size == 0 {
					continue
				}
				savedPath, uploadErr := SaveUploadedImage(fileHeader, "products")
				if uploadErr != nil {
					log.Warn("Skipping image upload", "filename", fileHeader.Filename, "error", uploadErr)
					continue
				}
				isThumb := (prod.Thumbnail == "" && idx == 0)
				_, _ = AddProductImage(req.Context(), db, prod.ID, savedPath, isThumb)
			}
		}

		// Re-fetch product to ensure thumbnail is set if product_images exist
		if prod.Thumbnail == "" {
			if updated, err := GetProductByID(req.Context(), db, prod.ID); err == nil && updated != nil {
				prod.Thumbnail = updated.Thumbnail
			}
		}

		page, _ := strconv.Atoi(req.FormValue("page"))
		products, pageInfo, _ := GetPaginatedProducts(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_product_table", map[string]interface{}{
			"Products": products,
			"PageInfo": pageInfo,
		})
	})

	// POST /admin/products/images/set-thumbnail
	r.Post("/admin/products/images/set-thumbnail", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		productID, _ := strconv.Atoi(req.FormValue("product_id"))
		imageID, _ := strconv.Atoi(req.FormValue("image_id"))

		if productID <= 0 || imageID <= 0 {
			http.Error(w, "Parameter tidak valid", http.StatusBadRequest)
			return
		}

		if err := SetProductThumbnail(req.Context(), db, productID, imageID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		render.JSON(w, http.StatusOK, map[string]interface{}{"success": true})
	})

	// POST /admin/products/images/delete
	r.Post("/admin/products/images/delete", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		productID, _ := strconv.Atoi(req.FormValue("product_id"))
		imageID, _ := strconv.Atoi(req.FormValue("image_id"))

		if productID <= 0 || imageID <= 0 {
			http.Error(w, "Parameter tidak valid", http.StatusBadRequest)
			return
		}

		imageURL, err := DeleteProductImage(req.Context(), db, productID, imageID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if strings.HasPrefix(imageURL, "/asset/") {
			filePath := strings.TrimPrefix(imageURL, "/")
			_ = os.Remove(filePath)
		}

		render.JSON(w, http.StatusOK, map[string]interface{}{"success": true})
	})

	// DELETE /admin/products
	r.Delete("/admin/products", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if id > 0 {
			if err := DeleteProduct(req.Context(), db, id); err != nil {
				log.Error("Failed to delete product", "error", err)
			}
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		products, pageInfo, _ := GetPaginatedProducts(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_product_table", map[string]interface{}{
			"Products": products,
			"PageInfo": pageInfo,
		})
	})

	// --- BANNERS ROUTES ---

	// GET /admin/banners
	r.Get("/admin/banners", func(w http.ResponseWriter, req *http.Request) {
		user := requireAdmin(w, req)
		if user == nil {
			return
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		banners, pageInfo, err := GetPaginatedBanners(req.Context(), db, page)
		if err != nil {
			log.Error("Failed to fetch banners", "error", err)
		}
		data := map[string]interface{}{
			"ActiveTab": "banners",
			"Banners":   banners,
			"PageInfo":  pageInfo,
			"AdminUser": user,
		}
		if req.Header.Get("X-Page-Only") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_banner_table", data)
			return
		}
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_banners", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// GET /admin/orders
	r.Get("/admin/orders", func(w http.ResponseWriter, req *http.Request) {
		user := requireAdmin(w, req)
		if user == nil {
			return
		}
		orders, err := GetAllOrders(req.Context(), db)
		if err != nil {
			log.Error("Failed to fetch orders", "error", err)
		}
		data := map[string]interface{}{
			"ActiveTab": "orders",
			"Orders":    orders,
			"AdminUser": user,
		}
		if req.Header.Get("X-Page-Only") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_order_table", data)
			return
		}
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_orders", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// POST /admin/orders/complete
	r.Post("/admin/orders/complete", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		orderID, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if orderID <= 0 {
			http.Error(w, "ID pesanan tidak valid", http.StatusBadRequest)
			return
		}

		if err := CompleteOrder(req.Context(), db, orderID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		orders, _ := GetAllOrders(req.Context(), db)
		tmpl.Render(w, http.StatusOK, "admin_order_table", map[string]interface{}{
			"Orders": orders,
		})
	})

	// GET /admin/search
	r.Get("/admin/search", func(w http.ResponseWriter, req *http.Request) {
		user := requireAdmin(w, req)
		if user == nil {
			return
		}
		searchType := strings.TrimSpace(req.URL.Query().Get("type"))
		searchQuery := strings.TrimSpace(req.URL.Query().Get("q"))

		results, err := SearchAdminGlobal(req.Context(), db, searchType, searchQuery)
		if err != nil {
			log.Error("Failed to perform admin global search", "error", err)
		}

		data := map[string]interface{}{
			"ActiveTab":   "search",
			"SearchType":  searchType,
			"SearchQuery": searchQuery,
			"Results":     results,
			"AdminUser":   user,
		}

		if req.Header.Get("X-Page-Only") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_search_results", data)
			return
		}
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_search", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// GET /admin/banners/modal
	r.Get("/admin/banners/modal", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		idStr := req.URL.Query().Get("id")
		var b Banner
		b.IsActive = true
		b.Icon = "🎁"
		if idStr != "" {
			if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
				if banner, err := GetBannerByID(req.Context(), db, id); err == nil && banner != nil {
					b = *banner
				}
			}
		}
		tmpl.Render(w, http.StatusOK, "admin_banner_modal", map[string]interface{}{
			"Banner": b,
		})
	})

	// POST /admin/banners (Save/Update)
	r.Post("/admin/banners", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		_ = req.ParseMultipartForm(10 << 20)

		id, _ := strconv.Atoi(req.FormValue("id"))
		badgeText := strings.TrimSpace(req.FormValue("badge_text"))
		title := strings.TrimSpace(req.FormValue("title"))
		description := strings.TrimSpace(req.FormValue("description"))
		ctaText := strings.TrimSpace(req.FormValue("cta_text"))
		ctaURL := strings.TrimSpace(req.FormValue("cta_url"))
		isActive := req.FormValue("is_active") == "1" || req.FormValue("is_active") == "true" || req.FormValue("is_active") == "on"
		icon := strings.TrimSpace(req.FormValue("icon"))

		if badgeText == "" || title == "" || description == "" || ctaText == "" || ctaURL == "" {
			http.Error(w, "Semua field teks wajib diisi!", http.StatusBadRequest)
			return
		}
		if icon == "" {
			icon = "🎁"
		}

		b := Banner{
			ID:          id,
			BadgeText:   badgeText,
			Title:       title,
			Description: description,
			CtaText:     ctaText,
			CtaURL:      ctaURL,
			IsActive:    isActive,
			Icon:        icon,
		}

		if err := SaveBanner(req.Context(), db, &b); err != nil {
			log.Error("Failed to save banner", "error", err)
			http.Error(w, "Gagal menyimpan banner", http.StatusInternalServerError)
			return
		}

		page, _ := strconv.Atoi(req.FormValue("page"))
		banners, pageInfo, _ := GetPaginatedBanners(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_banner_table", map[string]interface{}{
			"Banners":  banners,
			"PageInfo": pageInfo,
		})
	})

	// POST /admin/banners/toggle (Quick Status Switch)
	r.Post("/admin/banners/toggle", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if id > 0 {
			if err := ToggleBannerActive(req.Context(), db, id); err != nil {
				log.Error("Failed to toggle banner status", "error", err)
			}
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		banners, pageInfo, _ := GetPaginatedBanners(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_banner_table", map[string]interface{}{
			"Banners":  banners,
			"PageInfo": pageInfo,
		})
	})

	// DELETE /admin/banners
	r.Delete("/admin/banners", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if id > 0 {
			if err := DeleteBanner(req.Context(), db, id); err != nil {
				log.Error("Failed to delete banner", "error", err)
			}
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		banners, pageInfo, _ := GetPaginatedBanners(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_banner_table", map[string]interface{}{
			"Banners":  banners,
			"PageInfo": pageInfo,
		})
	})

	// --- PUBLIC LOGIN ROUTES FOR PARTNER & PROJECT MANAGER ---

	// GET /login-partner
	r.Get("/login-partner", func(w http.ResponseWriter, req *http.Request) {
		tmpl.Render(w, http.StatusOK, "login_partner", map[string]interface{}{
			"Error": "",
		})
	})

	// POST /login-partner
	r.Post("/login-partner", func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		email := strings.TrimSpace(req.FormValue("email"))
		password := req.FormValue("password")

		partner, err := AuthenticatePartner(req.Context(), db, email, password)
		if err != nil {
			tmpl.Render(w, http.StatusUnauthorized, "login_partner", map[string]interface{}{
				"Error": err.Error(),
				"Email": email,
			})
			return
		}

		tmpl.Render(w, http.StatusOK, "login_partner", map[string]interface{}{
			"Success": "Berhasil masuk sebagai Partner. Selamat datang, " + partner.Name + "!",
			"Partner": partner,
		})
	})

	// GET /login-project-manager
	r.Get("/login-project-manager", func(w http.ResponseWriter, req *http.Request) {
		tmpl.Render(w, http.StatusOK, "login_pm", map[string]interface{}{
			"Error": "",
		})
	})

	// POST /login-project-manager
	r.Post("/login-project-manager", func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		email := strings.TrimSpace(req.FormValue("email"))
		password := req.FormValue("password")

		pm, err := AuthenticateProjectManager(req.Context(), db, email, password)
		if err != nil {
			tmpl.Render(w, http.StatusUnauthorized, "login_pm", map[string]interface{}{
				"Error": err.Error(),
				"Email": email,
			})
			return
		}

		tmpl.Render(w, http.StatusOK, "login_pm", map[string]interface{}{
			"Success": "Berhasil masuk sebagai Project Manager. Selamat datang, " + pm.Name + "!",
			"PM":      pm,
		})
	})

	// --- PUBLIC SIGNUP ROUTES FOR PARTNER & PROJECT MANAGER ---

	// GET /signup-partner
	r.Get("/signup-partner", func(w http.ResponseWriter, req *http.Request) {
		tmpl.Render(w, http.StatusOK, "signup_partner", map[string]interface{}{
			"Error": "",
		})
	})

	// POST /signup-partner
	r.Post("/signup-partner", func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(req.FormValue("name"))
		email := strings.TrimSpace(req.FormValue("email"))
		password := req.FormValue("password")
		confirmPassword := req.FormValue("confirm_password")

		if name == "" || email == "" || password == "" {
			tmpl.Render(w, http.StatusBadRequest, "signup_partner", map[string]interface{}{
				"Error": "Semua kolom wajib diisi!",
				"Name":  name,
				"Email": email,
			})
			return
		}

		if !isValidEmail(email) {
			tmpl.Render(w, http.StatusBadRequest, "signup_partner", map[string]interface{}{
				"Error": "Format email tidak valid. Gunakan format contoh@domain.com",
				"Name":  name,
				"Email": email,
			})
			return
		}

		if len(password) < 6 {
			tmpl.Render(w, http.StatusBadRequest, "signup_partner", map[string]interface{}{
				"Error": "Kata sandi minimal 6 karakter demi keamanan akun Anda.",
				"Name":  name,
				"Email": email,
			})
			return
		}

		if password != confirmPassword {
			tmpl.Render(w, http.StatusBadRequest, "signup_partner", map[string]interface{}{
				"Error": "Konfirmasi kata sandi tidak cocok dengan kata sandi yang dimasukkan.",
				"Name":  name,
				"Email": email,
			})
			return
		}

		if existing, err := GetPartnerByEmail(req.Context(), db, email); err == nil && existing != nil {
			tmpl.Render(w, http.StatusConflict, "signup_partner", map[string]interface{}{
				"Error": "Alamat email ini sudah terdaftar sebagai Partner. Silakan masuk.",
				"Name":  name,
				"Email": email,
			})
			return
		}

		partner := Partner{
			Name:  name,
			Email: email,
		}

		if err := SavePartner(req.Context(), db, &partner, password); err != nil {
			log.Error("Failed to register partner", "error", err)
			tmpl.Render(w, http.StatusInternalServerError, "signup_partner", map[string]interface{}{
				"Error": "Gagal mendaftarkan akun Partner. Silakan coba lagi.",
				"Name":  name,
				"Email": email,
			})
			return
		}

		tmpl.Render(w, http.StatusOK, "login_partner", map[string]interface{}{
			"Success": "Pendaftaran akun Partner berhasil! Silakan masuk dengan email dan kata sandi Anda.",
			"Email":   email,
		})
	})

	// GET /signup-pm
	r.Get("/signup-pm", func(w http.ResponseWriter, req *http.Request) {
		tmpl.Render(w, http.StatusOK, "signup_pm", map[string]interface{}{
			"Error": "",
		})
	})

	// POST /signup-pm
	r.Post("/signup-pm", func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(req.FormValue("name"))
		email := strings.TrimSpace(req.FormValue("email"))
		password := req.FormValue("password")
		confirmPassword := req.FormValue("confirm_password")

		if name == "" || email == "" || password == "" {
			tmpl.Render(w, http.StatusBadRequest, "signup_pm", map[string]interface{}{
				"Error": "Semua kolom wajib diisi!",
				"Name":  name,
				"Email": email,
			})
			return
		}

		if !isValidEmail(email) {
			tmpl.Render(w, http.StatusBadRequest, "signup_pm", map[string]interface{}{
				"Error": "Format email tidak valid. Gunakan format contoh@domain.com",
				"Name":  name,
				"Email": email,
			})
			return
		}

		if len(password) < 6 {
			tmpl.Render(w, http.StatusBadRequest, "signup_pm", map[string]interface{}{
				"Error": "Kata sandi minimal 6 karakter demi keamanan akun Anda.",
				"Name":  name,
				"Email": email,
			})
			return
		}

		if password != confirmPassword {
			tmpl.Render(w, http.StatusBadRequest, "signup_pm", map[string]interface{}{
				"Error": "Konfirmasi kata sandi tidak cocok dengan kata sandi yang dimasukkan.",
				"Name":  name,
				"Email": email,
			})
			return
		}

		if existing, err := GetProjectManagerByEmail(req.Context(), db, email); err == nil && existing != nil {
			tmpl.Render(w, http.StatusConflict, "signup_pm", map[string]interface{}{
				"Error": "Alamat email ini sudah terdaftar sebagai Project Manager. Silakan masuk.",
				"Name":  name,
				"Email": email,
			})
			return
		}

		pm := ProjectManager{
			Name:  name,
			Email: email,
		}

		if err := SaveProjectManager(req.Context(), db, &pm, password); err != nil {
			log.Error("Failed to register project manager", "error", err)
			tmpl.Render(w, http.StatusInternalServerError, "signup_pm", map[string]interface{}{
				"Error": "Gagal mendaftarkan akun Project Manager. Silakan coba lagi.",
				"Name":  name,
				"Email": email,
			})
			return
		}

		tmpl.Render(w, http.StatusOK, "login_pm", map[string]interface{}{
			"Success": "Pendaftaran akun Project Manager berhasil! Silakan masuk dengan email dan kata sandi Anda.",
			"Email":   email,
		})
	})

	// --- ADMIN PARTNERS MANAGEMENT ---

	// GET /admin/partners
	r.Get("/admin/partners", func(w http.ResponseWriter, req *http.Request) {
		adminUser := requireAdmin(w, req)
		if adminUser == nil {
			return
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		partners, pageInfo, _ := GetPaginatedPartners(req.Context(), db, page)
		data := map[string]interface{}{
			"AdminUser": adminUser,
			"ActiveTab": "partners",
			"Partners":  partners,
			"PageInfo":  pageInfo,
		}
		if req.Header.Get("X-Page-Only") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_partner_table", data)
			return
		}
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_partners", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// GET /admin/partners/modal
	r.Get("/admin/partners/modal", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		data := map[string]interface{}{}
		if id > 0 {
			partner, err := GetPartnerByID(req.Context(), db, id)
			if err == nil {
				data["Partner"] = partner
			}
		}
		tmpl.Render(w, http.StatusOK, "admin_partner_modal", data)
	})

	// POST /admin/partners
	r.Post("/admin/partners", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		id, _ := strconv.Atoi(req.FormValue("id"))
		name := strings.TrimSpace(req.FormValue("name"))
		email := strings.TrimSpace(req.FormValue("email"))
		phone := strings.TrimSpace(req.FormValue("phone"))
		companyName := strings.TrimSpace(req.FormValue("company_name"))
		description := strings.TrimSpace(req.FormValue("description"))
		password := req.FormValue("password")

		if name == "" || email == "" {
			http.Error(w, "Nama dan Email wajib diisi!", http.StatusBadRequest)
			return
		}

		partner := Partner{
			ID:          id,
			Name:        name,
			Email:       email,
			Phone:       phone,
			CompanyName: companyName,
			Description: description,
		}

		if err := SavePartner(req.Context(), db, &partner, password); err != nil {
			log.Error("Failed to save partner", "error", err)
			http.Error(w, "Gagal menyimpan data partner", http.StatusInternalServerError)
			return
		}

		page, _ := strconv.Atoi(req.FormValue("page"))
		partners, pageInfo, _ := GetPaginatedPartners(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_partner_table", map[string]interface{}{
			"Partners": partners,
			"PageInfo": pageInfo,
		})
	})

	// DELETE /admin/partners
	r.Delete("/admin/partners", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if id > 0 {
			_ = DeletePartner(req.Context(), db, id)
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		partners, pageInfo, _ := GetPaginatedPartners(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_partner_table", map[string]interface{}{
			"Partners": partners,
			"PageInfo": pageInfo,
		})
	})

	// --- ADMIN PROJECT MANAGERS MANAGEMENT ---

	// GET /admin/project-managers
	r.Get("/admin/project-managers", func(w http.ResponseWriter, req *http.Request) {
		adminUser := requireAdmin(w, req)
		if adminUser == nil {
			return
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		pms, pageInfo, _ := GetPaginatedProjectManagers(req.Context(), db, page)
		data := map[string]interface{}{
			"AdminUser":       adminUser,
			"ActiveTab":       "project_managers",
			"ProjectManagers": pms,
			"PageInfo":        pageInfo,
		}
		if req.Header.Get("X-Page-Only") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_project_manager_table", data)
			return
		}
		if req.Header.Get("HX-Request") == "true" {
			tmpl.Render(w, http.StatusOK, "admin_project_managers", data)
			return
		}
		tmpl.Render(w, http.StatusOK, "admin_index", data)
	})

	// GET /admin/project-managers/modal
	r.Get("/admin/project-managers/modal", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		data := map[string]interface{}{}
		if id > 0 {
			pm, err := GetProjectManagerByID(req.Context(), db, id)
			if err == nil {
				data["ProjectManager"] = pm
			}
		}
		tmpl.Render(w, http.StatusOK, "admin_project_manager_modal", data)
	})

	// POST /admin/project-managers
	r.Post("/admin/project-managers", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		id, _ := strconv.Atoi(req.FormValue("id"))
		name := strings.TrimSpace(req.FormValue("name"))
		email := strings.TrimSpace(req.FormValue("email"))
		phone := strings.TrimSpace(req.FormValue("phone"))
		companyName := strings.TrimSpace(req.FormValue("company_name"))
		description := strings.TrimSpace(req.FormValue("description"))
		password := req.FormValue("password")

		if name == "" || email == "" {
			http.Error(w, "Nama dan Email wajib diisi!", http.StatusBadRequest)
			return
		}

		pm := ProjectManager{
			ID:          id,
			Name:        name,
			Email:       email,
			Phone:       phone,
			CompanyName: companyName,
			Description: description,
		}

		if err := SaveProjectManager(req.Context(), db, &pm, password); err != nil {
			log.Error("Failed to save project manager", "error", err)
			http.Error(w, "Gagal menyimpan data project manager", http.StatusInternalServerError)
			return
		}

		page, _ := strconv.Atoi(req.FormValue("page"))
		pms, pageInfo, _ := GetPaginatedProjectManagers(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_project_manager_table", map[string]interface{}{
			"ProjectManagers": pms,
			"PageInfo":        pageInfo,
		})
	})

	// DELETE /admin/project-managers
	r.Delete("/admin/project-managers", func(w http.ResponseWriter, req *http.Request) {
		if requireAdmin(w, req) == nil {
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if id > 0 {
			_ = DeleteProjectManager(req.Context(), db, id)
		}
		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		pms, pageInfo, _ := GetPaginatedProjectManagers(req.Context(), db, page)
		tmpl.Render(w, http.StatusOK, "admin_project_manager_table", map[string]interface{}{
			"ProjectManagers": pms,
			"PageInfo":        pageInfo,
		})
	})
}

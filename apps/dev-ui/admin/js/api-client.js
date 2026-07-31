// api-client.js
// Penanganan otomatis interaksi REST API dengan Backend Go

const API_BASE_URL = 'http://localhost:8080/api';

/**
 * Mengambil token JWT dari localStorage
 */
function getToken() {
    return localStorage.getItem('admin_token');
}

/**
 * Fungsi fetch wrapper yang mengeksekusi HTTP request.
 * Secara otomatis menyuntikkan Authorization header jika token ada.
 */
export async function apiClient(endpoint, options = {}) {
    // Default config
    const config = {
        method: options.method || 'GET',
        headers: {
            'Content-Type': 'application/json',
            ...(options.headers || {})
        }
    };

    // Suntikan Bearer Token Otomatis
    const token = getToken();
    if (token) {
        config.headers['Authorization'] = `Bearer ${token}`;
    }

    if (options.body) {
        config.body = JSON.stringify(options.body);
    }

    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, config);

        // Cek jika error otentikasi (Token Expired)
        if (response.status === 401) {
            localStorage.removeItem('admin_token');
            localStorage.removeItem('admin_user');
            window.location.hash = '#login';
            throw new Error("Sesi kedaluwarsa, silakan login kembali.");
        }

        // Cek Forbidden Role (Bukan Admin)
        if (response.status === 403) {
            throw new Error("Akses ditolak. Anda tidak memiliki izin Admin.");
        }

        // Handle 204 No Content (common for DELETE)
        if (response.status === 204) {
            return null;
        }

        const data = await response.json();

        // Membungkus logic jika sukses tapi JSON memuat error = false sesuai standar OBSShop API
        if (!response.ok) {
            throw new Error(data.error || `HTTP Error ${response.status}`);
        }

        return data.data || data; // Return payload standard format
    } catch (error) {
        console.error(`API Error on ${endpoint}:`, error);
        throw error;
    }
}

/**
 * Sistem Toast Notification
 */
export function showToast(message, type = 'success') {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `<span class="toast-message">${message}</span>`;

    container.appendChild(toast);

    // Trigger animation frame
    setTimeout(() => toast.classList.add('toast-show'), 10);

    // Hilangkan setelah 3 detik
    setTimeout(() => {
        toast.classList.remove('toast-show');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

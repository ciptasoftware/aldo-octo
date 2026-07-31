// views/login.js
import { apiClient, showToast } from '../api-client.js';

export const LoginView = {
    async render(rootElement) {
        rootElement.innerHTML = `
            <div class="login-wrapper">
                <div class="login-card">
                    <div class="login-header">
                        <div style="width:50px;height:50px;background:var(--color-primary);border-radius:12px;margin: 0 auto 1rem;display:flex;align-items:center;justify-content:center;color:white;font-weight:bold;font-size:1.5rem">A</div>
                        <h2>Admin Login</h2>
                        <p>Masuk ke portal manajerial OBSShop</p>
                    </div>

                    <form id="frmAdminLogin">
                        <div class="form-group">
                            <label>Email Admin</label>
                            <input type="email" id="inEmail" class="form-control" placeholder="admin@obsshop.id" required>
                        </div>
                        <div class="form-group">
                            <label>Password</label>
                            <input type="password" id="inPassword" class="form-control" placeholder="••••••••" required>
                        </div>
                        <button type="submit" class="btn btn-primary" style="width:100%; margin-top:0.5rem; padding: 0.75rem;">Login ke Dashboard</button>
                    </form>
                </div>
            </div>
        `;

        document.getElementById('frmAdminLogin').addEventListener('submit', async (e) => {
            e.preventDefault();
            const email = document.getElementById('inEmail').value;
            const password = document.getElementById('inPassword').value;

            try {
                // Tembak standard /api/auth/login endpoint
                const res = await apiClient('/auth/login', {
                    method: 'POST',
                    body: { email, password }
                });

                // Cek apakan role admin
                if (res.user.role !== 'admin' && res.user.role !== 'brand_manager') {
                    showToast("Email Anda tidak memiliki akses Admin!", "error");
                    return;
                }

                // Sukses!
                localStorage.setItem('admin_token', res.token);
                localStorage.setItem('admin_user', JSON.stringify(res.user));
                
                showToast("Login Berhasil", "success");
                
                // Route ke main dashboard
                window.location.hash = '#dashboard';
            } catch (error) {
                showToast(error.message, "error");
            }
        });
    }
};

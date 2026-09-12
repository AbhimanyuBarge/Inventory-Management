<template>
  <div class="login-container">
    <div class="login-card m3-card">
      <div class="login-header">
        <div class="logo-circle">
          <Wrench class="w-8 h-8 text-white" />
        </div>
        <h2 class="app-title">SparePro Inventory</h2>
        <p class="app-desc">Spare Parts Management & Point of Sale</p>
      </div>

      <div v-if="error" class="error-alert">
        <AlertCircle class="w-5 h-5 flex-shrink-0" />
        <span>{{ error }}</span>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="m3-input-group">
          <label class="m3-label">Username</label>
          <div class="input-with-icon">
            <User class="input-icon" />
            <input
              v-model="form.username"
              type="text"
              class="m3-input pl-10"
              placeholder="e.g. admin or seller"
              required
            />
          </div>
        </div>

        <div class="m3-input-group">
          <label class="m3-label">Password</label>
          <div class="input-with-icon">
            <Lock class="input-icon" />
            <input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              class="m3-input pl-10"
              placeholder="Enter password"
              required
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="password-toggle"
            >
              <EyeOff v-if="showPassword" class="w-4 h-4" />
              <Eye v-else class="w-4 h-4" />
            </button>
          </div>
        </div>

        <button
          type="submit"
          class="m3-btn m3-btn-primary w-full login-btn"
          :disabled="loading"
        >
          <span v-if="loading">Signing in...</span>
          <span v-else>Sign In to Account</span>
          <ArrowRight v-if="!loading" class="w-4 h-4" />
        </button>
      </form>

      <!-- Quick Demo Access Chips -->
      <div class="demo-section">
        <p class="demo-label">Quick Demo Access:</p>
        <div class="demo-chips">
          <button
            type="button"
            @click="fillCredentials('admin', 'admin123')"
            class="demo-chip chip-admin"
          >
            <Shield class="w-3.5 h-3.5" />
            <span>Admin Demo</span>
          </button>
          <button
            type="button"
            @click="fillCredentials('seller', 'seller123')"
            class="demo-chip chip-seller"
          >
            <Store class="w-3.5 h-3.5" />
            <span>Seller Demo</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import {
  Wrench,
  User,
  Lock,
  Eye,
  EyeOff,
  ArrowRight,
  AlertCircle,
  Shield,
  Store,
} from 'lucide-vue-next';

const router = useRouter();
const authStore = useAuthStore();

const form = reactive({
  username: '',
  password: '',
});

const showPassword = ref(false);
const loading = ref(false);
const error = ref('');

const fillCredentials = (u, p) => {
  form.username = u;
  form.password = p;
  error.value = '';
};

const handleLogin = async () => {
  loading.value = true;
  error.value = '';
  try {
    await authStore.login(form.username, form.password);
    router.push('/');
  } catch (err) {
    error.value = err.message || 'Login failed';
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: radial-gradient(circle at 50% 20%, #E8F0FE 0%, #F8FAFC 80%);
}

.login-card {
  width: 100%;
  max-width: 440px;
  padding: 32px 28px;
  background: #FFFFFF;
  border-radius: var(--radius-xl);
  box-shadow: var(--md-sys-elevation-3);
  border: 1px solid var(--md-sys-color-surface-variant);
}

.login-header {
  text-align: center;
  margin-bottom: 28px;
}

.logo-circle {
  width: 60px;
  height: 60px;
  border-radius: var(--radius-lg);
  background: linear-gradient(135deg, #00629E, #0088DB);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  box-shadow: var(--md-sys-elevation-2);
}

.app-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--md-sys-color-on-background);
  letter-spacing: -0.02em;
}

.app-desc {
  font-size: 13px;
  color: var(--md-sys-color-outline);
  margin-top: 4px;
}

.error-alert {
  display: flex;
  align-items: center;
  gap: 10px;
  background-color: var(--md-sys-color-error-container);
  color: var(--md-sys-color-error);
  padding: 12px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 20px;
}

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 14px;
  width: 18px;
  height: 18px;
  color: var(--md-sys-color-outline);
  pointer-events: none;
}

.pl-10 {
  padding-left: 42px !important;
}

.password-toggle {
  position: absolute;
  right: 12px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--md-sys-color-outline);
  padding: 4px;
}

.login-btn {
  height: 48px;
  font-size: 15px;
  margin-top: 12px;
}

.w-full {
  width: 100%;
}

.demo-section {
  margin-top: 28px;
  padding-top: 20px;
  border-top: 1px dashed var(--md-sys-color-surface-variant);
}

.demo-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--md-sys-color-outline);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 10px;
  text-align: center;
}

.demo-chips {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.demo-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: var(--radius-full);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}

.chip-admin {
  background: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
}
.chip-admin:hover {
  filter: brightness(0.95);
  transform: translateY(-1px);
}

.chip-seller {
  background: var(--md-sys-color-success-container);
  color: var(--md-sys-color-on-success);
}
.chip-seller:hover {
  filter: brightness(0.95);
  transform: translateY(-1px);
}
</style>

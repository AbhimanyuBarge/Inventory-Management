<template>
  <div class="app-layout">
    <!-- Desktop Side Navigation Drawer -->
    <aside class="desktop-drawer">
      <div class="drawer-header">
        <div class="brand-logo">
          <div class="brand-icon">
            <Wrench class="w-6 h-6 text-white" />
          </div>
          <div class="brand-text">
            <span class="brand-title">SparePro</span>
            <span class="brand-subtitle">Inventory & POS</span>
          </div>
        </div>
      </div>

      <nav class="drawer-nav">
        <router-link to="/" class="nav-item" active-class="active" exact>
          <LayoutDashboard class="nav-icon" />
          <span>Dashboard</span>
        </router-link>

        <router-link to="/billing" class="nav-item" active-class="active">
          <Receipt class="nav-icon" />
          <span>POS Billing</span>
          <span class="pos-chip">Quick</span>
        </router-link>

        <router-link to="/bills" class="nav-item" active-class="active">
          <FileText class="nav-icon" />
          <span>Invoice History</span>
        </router-link>

        <router-link to="/parts" class="nav-item" active-class="active">
          <Package class="nav-icon" />
          <span>Spare Parts</span>
        </router-link>

        <router-link to="/racks" class="nav-item" active-class="active">
          <Grid3X3 class="nav-icon" />
          <span>Racks & Labels</span>
        </router-link>

        <router-link to="/customers" class="nav-item" active-class="active">
          <Users class="nav-icon" />
          <span>Customers</span>
        </router-link>

        <router-link v-if="authStore.isAdmin" to="/settings" class="nav-item" active-class="active">
          <Sliders class="nav-icon" />
          <span>Settings & Users</span>
        </router-link>
      </nav>

      <div class="drawer-footer">
        <div class="user-profile-badge">
          <div class="user-avatar" :class="authStore.isAdmin ? 'avatar-admin' : 'avatar-seller'">
            {{ (authStore.userName || 'U').charAt(0).toUpperCase() }}
          </div>
          <div class="user-info">
            <div class="user-name">{{ authStore.userName }}</div>
            <div class="user-role">{{ authStore.user?.role?.toUpperCase() }}</div>
          </div>
          <button @click="logout" class="logout-btn" title="Sign Out">
            <LogOut class="w-4 h-4" />
          </button>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <div class="main-content">
      <!-- Top App Bar -->
      <header class="top-app-bar">
        <div class="app-bar-left">
          <div class="mobile-brand">
            <div class="brand-icon-sm">
              <Wrench class="w-4 h-4 text-white" />
            </div>
            <span class="mobile-app-title">SparePro</span>
          </div>
          <h1 class="page-title">{{ currentRouteTitle }}</h1>
        </div>

        <div class="app-bar-actions">
          <div class="user-pill-mobile">
            <span class="role-badge" :class="authStore.isAdmin ? 'badge-admin' : 'badge-seller'">
              {{ authStore.user?.role }}
            </span>
            <button @click="logout" class="m3-btn-icon" title="Logout">
              <LogOut class="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      <!-- Page View Container -->
      <main class="page-body">
        <router-view />
      </main>

      <!-- Mobile Android Bottom Navigation Bar -->
      <nav class="mobile-bottom-nav">
        <router-link to="/" class="bottom-nav-item" active-class="active" exact>
          <LayoutDashboard class="bottom-icon" />
          <span>Home</span>
        </router-link>

        <router-link to="/billing" class="bottom-nav-item bottom-nav-highlight" active-class="active">
          <div class="fab-circle">
            <Receipt class="bottom-icon" />
          </div>
          <span>Bill POS</span>
        </router-link>

        <router-link to="/parts" class="bottom-nav-item" active-class="active">
          <Package class="bottom-icon" />
          <span>Parts</span>
        </router-link>

        <router-link to="/racks" class="bottom-nav-item" active-class="active">
          <Grid3X3 class="bottom-icon" />
          <span>Racks</span>
        </router-link>

        <router-link to="/bills" class="bottom-nav-item" active-class="active">
          <FileText class="bottom-icon" />
          <span>Bills</span>
        </router-link>

        <router-link v-if="authStore.isAdmin" to="/settings" class="bottom-nav-item" active-class="active">
          <Sliders class="bottom-icon" />
          <span>Admin</span>
        </router-link>
      </nav>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import {
  LayoutDashboard,
  Receipt,
  FileText,
  Package,
  Grid3X3,
  Users,
  Sliders,
  LogOut,
  Wrench,
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const currentRouteTitle = computed(() => {
  switch (route.name) {
    case 'Dashboard': return 'Store Dashboard';
    case 'Billing': return 'Point of Sale (POS)';
    case 'BillsHistory': return 'Invoice Records';
    case 'Parts': return 'Spare Parts Catalog';
    case 'Racks': return 'Rack Storage Locations';
    case 'Customers': return 'Customer Directory';
    case 'Settings': return 'System Settings & Users';
    default: return 'Inventory Management';
  }
});

const logout = () => {
  authStore.logout();
  router.push('/login');
};
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--md-sys-color-background);
}

/* Desktop Navigation Drawer */
.desktop-drawer {
  display: none;
  width: 260px;
  background: #FFFFFF;
  border-right: 1px solid var(--md-sys-color-surface-variant);
  flex-direction: column;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  height: 100vh;
  z-index: 30;
}

@media (min-width: 992px) {
  .desktop-drawer {
    display: flex;
  }
}

.drawer-header {
  padding: 24px 20px 16px;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-icon {
  width: 42px;
  height: 42px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, #00629E, #0088DB);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--md-sys-elevation-1);
}

.brand-title {
  display: block;
  font-size: 18px;
  font-weight: 800;
  color: var(--md-sys-color-on-background);
  letter-spacing: -0.02em;
}

.brand-subtitle {
  display: block;
  font-size: 11px;
  color: var(--md-sys-color-outline);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.drawer-nav {
  flex: 1;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: var(--radius-full);
  font-size: 14px;
  font-weight: 600;
  color: var(--md-sys-color-on-surface-variant);
  text-decoration: none;
  transition: all 0.18s ease;
}

.nav-item:hover {
  background-color: var(--md-sys-color-background);
  color: var(--md-sys-color-primary);
}

.nav-item.active {
  background-color: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
  font-weight: 700;
}

.nav-icon {
  width: 20px;
  height: 20px;
}

.pos-chip {
  margin-left: auto;
  font-size: 10px;
  background: #00629E;
  color: #FFF;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-weight: 700;
}

.drawer-footer {
  padding: 16px;
  border-top: 1px solid var(--md-sys-color-surface-variant);
}

.user-profile-badge {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--md-sys-color-background);
  border-radius: var(--radius-lg);
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 14px;
  color: #FFFFFF;
}

.avatar-admin {
  background: linear-gradient(135deg, #00629E, #0088DB);
}

.avatar-seller {
  background: linear-gradient(135deg, #1E8E3E, #34A853);
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 13px;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  font-size: 10px;
  color: var(--md-sys-color-outline);
  font-weight: 700;
}

.logout-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--md-sys-color-outline);
  padding: 6px;
  border-radius: 50%;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: var(--md-sys-color-error-container);
  color: var(--md-sys-color-error);
}

/* Top App Bar */
.top-app-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  background: #FFFFFF;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  position: sticky;
  top: 0;
  z-index: 20;
}

.app-bar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.mobile-brand {
  display: flex;
  align-items: center;
  gap: 8px;
}

@media (min-width: 992px) {
  .mobile-brand {
    display: none;
  }
}

.brand-icon-sm {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  background: #00629E;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mobile-app-title {
  font-weight: 800;
  font-size: 15px;
  color: #00629E;
}

.page-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--md-sys-color-on-background);
}

.user-pill-mobile {
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-badge {
  font-size: 11px;
  text-transform: uppercase;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: var(--radius-full);
}

.badge-admin {
  background: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
}

.badge-seller {
  background: var(--md-sys-color-success-container);
  color: var(--md-sys-color-on-success);
}

/* Page Body */
.page-body {
  padding: 16px;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

@media (min-width: 768px) {
  .page-body {
    padding: 24px;
  }
}

/* Android Mobile Bottom Navigation Bar */
.mobile-bottom-nav {
  display: flex;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: #FFFFFF;
  border-top: 1px solid var(--md-sys-color-surface-variant);
  z-index: 30;
  justify-content: space-around;
  align-items: center;
  box-shadow: 0 -2px 10px rgba(0,0,0,0.05);
}

@media (min-width: 992px) {
  .mobile-bottom-nav {
    display: none;
  }
}

.bottom-nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  color: var(--md-sys-color-outline);
  text-decoration: none;
  font-size: 11px;
  font-weight: 600;
  flex: 1;
  height: 100%;
  transition: color 0.15s ease;
}

.bottom-nav-item.active {
  color: var(--md-sys-color-primary);
  font-weight: 700;
}

.bottom-icon {
  width: 20px;
  height: 20px;
}
</style>

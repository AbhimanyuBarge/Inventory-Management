import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { guestOnly: true },
  },
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('../views/DashboardView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/billing',
    name: 'Billing',
    component: () => import('../views/BillingView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/bills',
    name: 'BillsHistory',
    component: () => import('../views/BillsHistoryView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/returns',
    name: 'ReturnsHistory',
    component: () => import('../views/ReturnsHistoryView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/returns/new',
    name: 'ProcessReturn',
    component: () => import('../views/ProcessReturnView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/items',
    name: 'Items',
    component: () => import('../views/ItemsView.vue'),
    alias: '/parts',
    meta: { requiresAuth: true },
  },
  {
    path: '/racks',
    name: 'Racks',
    component: () => import('../views/RacksView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/customers',
    name: 'Customers',
    component: () => import('../views/CustomersView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/vendors',
    name: 'Vendors',
    component: () => import('../views/VendorsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/SettingsView.vue'),
    meta: { requiresAuth: true, adminOnly: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore();

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return next({ name: 'Login' });
  }

  if (to.meta.guestOnly && authStore.isAuthenticated) {
    return next({ name: 'Dashboard' });
  }

  if (to.meta.adminOnly && !authStore.isAdmin) {
    return next({ name: 'Dashboard' });
  }

  next();
});

export default router;

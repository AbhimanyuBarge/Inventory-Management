import { defineStore } from 'pinia';
import api from '../api/client';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || null,
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    loading: false,
    error: null,
  }),

  getters: {
    isAuthenticated: (state) => !!state.token && !!state.user,
    isAdmin: (state) => state.user?.role === 'admin',
    isSeller: (state) => state.user?.role === 'seller',
    userName: (state) => state.user?.full_name || state.user?.username || 'User',
  },

  actions: {
    async login(username, password) {
      this.loading = true;
      this.error = null;
      try {
        const response = await api.post('/auth/login', { username, password });
        this.token = response.data.token;
        this.user = response.data.user;

        localStorage.setItem('token', this.token);
        localStorage.setItem('user', JSON.stringify(this.user));
        return response.data;
      } catch (err) {
        this.error = err.response?.data?.error || 'Login failed. Please check credentials.';
        throw new Error(this.error);
      } finally {
        this.loading = false;
      }
    },

    async fetchProfile() {
      if (!this.token) return;
      try {
        const res = await api.get('/auth/me');
        this.user = res.data;
        localStorage.setItem('user', JSON.stringify(this.user));
      } catch (err) {
        this.logout();
      }
    },

    logout() {
      this.token = null;
      this.user = null;
      localStorage.removeItem('token');
      localStorage.removeItem('user');
    },
  },
});

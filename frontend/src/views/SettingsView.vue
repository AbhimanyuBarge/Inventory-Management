<template>
  <div class="settings-page">
    <div class="settings-tabs">
      <button
        @click="activeTab = 'general'"
        class="tab-btn"
        :class="{ active: activeTab === 'general' }"
      >
        <Sliders class="w-4 h-4" />
        <span>Store & GST Settings</span>
      </button>

      <button
        @click="activeTab = 'users'"
        class="tab-btn"
        :class="{ active: activeTab === 'users' }"
      >
        <Users class="w-4 h-4" />
        <span>User & Seller Accounts</span>
      </button>
    </div>

    <!-- Tab 1: Store & GST Settings -->
    <div v-if="activeTab === 'general'" class="tab-content">
      <div class="m3-card settings-card">
        <div class="card-header">
          <div>
            <h3 class="card-title">Store Information & GST Configuration</h3>
            <p class="card-subtitle">These details appear on printed tax invoices and rack storage tags.</p>
          </div>
        </div>

        <form @submit.prevent="saveSettings" class="settings-form-grid">
          <div class="m3-input-group">
            <label class="m3-label">Store / Business Name *</label>
            <input
              v-model="settingsForm.shop_name"
              type="text"
              class="m3-input"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Default GST Rate (%) *</label>
            <input
              v-model.number="settingsForm.default_gst_rate"
              type="number"
              step="0.01"
              min="0"
              class="m3-input"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Business Contact Phone</label>
            <input
              v-model="settingsForm.shop_phone"
              type="text"
              class="m3-input"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Business Email Address</label>
            <input
              v-model="settingsForm.shop_email"
              type="email"
              class="m3-input"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">GSTIN / Tax ID Number</label>
            <input
              v-model="settingsForm.shop_gstin"
              type="text"
              class="m3-input"
              placeholder="e.g. 27AAACG0000A1Z5"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Invoice Prefix</label>
            <input
              v-model="settingsForm.invoice_prefix"
              type="text"
              class="m3-input"
              placeholder="e.g. INV"
            />
          </div>

          <div class="m3-input-group full-width">
            <label class="m3-label">Store Address (Appears on Invoice Header)</label>
            <textarea
              v-model="settingsForm.shop_address"
              class="m3-textarea"
            ></textarea>
          </div>

          <div class="full-width flex justify-end">
            <button type="submit" class="m3-btn m3-btn-primary" :disabled="savingSettings">
              <Check class="w-4 h-4" />
              <span>Save System Settings</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Tab 2: User & Seller Accounts -->
    <div v-if="activeTab === 'users'" class="tab-content">
      <div class="m3-card users-card">
        <div class="card-header flex-between">
          <div>
            <h3 class="card-title">User Roles & Sub-Accounts</h3>
            <p class="card-subtitle">Create and manage Seller / Staff accounts with restricted billing access.</p>
          </div>

          <button @click="openAddUserModal" class="m3-btn m3-btn-primary">
            <Plus class="w-4 h-4" />
            <span>Create User</span>
          </button>
        </div>

        <div class="table-responsive mt-4">
          <table class="m3-table">
            <thead>
              <tr>
                <th>User</th>
                <th>Username</th>
                <th>Role</th>
                <th class="text-center">Status</th>
                <th class="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id">
                <td>
                  <div class="font-bold">{{ u.full_name }}</div>
                </td>
                <td>
                  <span class="font-mono">{{ u.username }}</span>
                </td>
                <td>
                  <span
                    class="role-tag"
                    :class="u.role === 'admin' ? 'tag-admin' : 'tag-seller'"
                  >
                    {{ u.role.toUpperCase() }}
                  </span>
                </td>
                <td class="text-center">
                  <span class="status-indicator" :class="u.is_active ? 'active' : 'inactive'">
                    {{ u.is_active ? 'Active' : 'Disabled' }}
                  </span>
                </td>
                <td class="text-right">
                  <div class="flex gap-2 justify-end">
                    <button @click="openEditUserModal(u)" class="m3-btn-icon" title="Edit / Reset Password">
                      <Edit3 class="w-4 h-4 text-primary" />
                    </button>
                    <button
                      v-if="u.id !== authStore.user?.id"
                      @click="deleteUser(u)"
                      class="m3-btn-icon"
                      title="Delete User"
                    >
                      <Trash2 class="w-4 h-4 text-danger" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Add / Edit User Modal -->
    <div v-if="userModal.show" class="m3-modal-overlay" @click.self="userModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-4">
          {{ userModal.isEdit ? 'Edit User / Reset Password' : 'Create New User Account' }}
        </h3>

        <form @submit.prevent="saveUser">
          <div class="m3-input-group">
            <label class="m3-label">Full Name *</label>
            <input
              v-model="userForm.full_name"
              type="text"
              class="m3-input"
              placeholder="e.g. John Doe"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Username *</label>
            <input
              v-model="userForm.username"
              type="text"
              class="m3-input"
              placeholder="e.g. seller1"
              :disabled="userModal.isEdit"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">
              {{ userModal.isEdit ? 'New Password (leave blank to keep current)' : 'Password *' }}
            </label>
            <input
              v-model="userForm.password"
              type="password"
              class="m3-input"
              :placeholder="userModal.isEdit ? '••••••••' : 'Min 6 characters'"
              :required="!userModal.isEdit"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">User Role *</label>
            <select v-model="userForm.role" class="m3-select">
              <option value="seller">Seller (Sales & Billing Only)</option>
              <option value="admin">Admin (Full Access)</option>
            </select>
          </div>

          <div class="m3-input-group">
            <label class="flex items-center gap-2 cursor-pointer mt-2">
              <input
                v-model="userForm.is_active"
                type="checkbox"
                class="w-4 h-4 accent-primary"
              />
              <span class="text-sm font-bold">Account is Active & Enabled</span>
            </label>
          </div>

          <div class="flex gap-3 justify-end mt-6">
            <button type="button" @click="userModal.show = false" class="m3-btn m3-btn-outlined">
              Cancel
            </button>
            <button type="submit" class="m3-btn m3-btn-primary" :disabled="userModal.saving">
              {{ userModal.isEdit ? 'Update User' : 'Create Account' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue';
import { useAuthStore } from '../stores/auth';
import api from '../api/client';
import {
  Sliders,
  Users,
  Check,
  Plus,
  Edit3,
  Trash2,
} from 'lucide-vue-next';

const authStore = useAuthStore();
const activeTab = ref('general');

// Settings State
const settingsForm = reactive({
  shop_name: '',
  shop_address: '',
  shop_phone: '',
  shop_email: '',
  shop_gstin: '',
  default_gst_rate: 18.00,
  invoice_prefix: 'INV',
});
const savingSettings = ref(false);

// Users State
const users = ref([]);
const userModal = reactive({
  show: false,
  isEdit: false,
  saving: false,
  userId: null,
});

const userForm = reactive({
  username: '',
  password: '',
  full_name: '',
  role: 'seller',
  is_active: true,
});

const loadSettings = async () => {
  try {
    const res = await api.get('/settings');
    Object.assign(settingsForm, res.data);
  } catch (err) {
    console.error(err);
  }
};

const saveSettings = async () => {
  savingSettings.value = true;
  try {
    const res = await api.put('/settings', settingsForm);
    Object.assign(settingsForm, res.data);
    alert('Settings updated successfully!');
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to update settings');
  } finally {
    savingSettings.value = false;
  }
};

const loadUsers = async () => {
  try {
    const res = await api.get('/users');
    users.value = res.data;
  } catch (err) {
    console.error(err);
  }
};

const openAddUserModal = () => {
  userModal.isEdit = false;
  userModal.userId = null;
  userForm.username = '';
  userForm.password = '';
  userForm.full_name = '';
  userForm.role = 'seller';
  userForm.is_active = true;
  userModal.show = true;
};

const openEditUserModal = (u) => {
  userModal.isEdit = true;
  userModal.userId = u.id;
  userForm.username = u.username;
  userForm.password = '';
  userForm.full_name = u.full_name;
  userForm.role = u.role;
  userForm.is_active = u.is_active;
  userModal.show = true;
};

const saveUser = async () => {
  userModal.saving = true;
  try {
    if (userModal.isEdit) {
      await api.put(`/users/${userModal.userId}`, {
        full_name: userForm.full_name,
        role: userForm.role,
        is_active: userForm.is_active,
        password: userForm.password || undefined,
      });
    } else {
      await api.post('/users', userForm);
    }
    userModal.show = false;
    await loadUsers();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to save user');
  } finally {
    userModal.saving = false;
  }
};

const deleteUser = async (u) => {
  if (!confirm(`Delete user '${u.username}'?`)) return;
  try {
    await api.delete(`/users/${u.id}`);
    await loadUsers();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to delete user');
  }
};

onMounted(() => {
  loadSettings();
  loadUsers();
});
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.settings-tabs {
  display: flex;
  gap: 10px;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  padding-bottom: 8px;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border-radius: var(--radius-full);
  font-weight: 700;
  font-size: 13px;
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--md-sys-color-outline);
  transition: all 0.2s;
}

.tab-btn.active {
  background: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
}

.settings-card, .users-card {
  padding: 24px;
}

.card-header {
  margin-bottom: 20px;
}

.card-title {
  font-size: 17px;
  font-weight: 800;
  color: var(--md-sys-color-on-surface);
}

.card-subtitle {
  font-size: 13px;
  color: var(--md-sys-color-outline);
}

.settings-form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}
@media (min-width: 640px) {
  .settings-form-grid {
    grid-template-columns: 1fr 1fr;
  }
}

.full-width {
  grid-column: span 1;
}
@media (min-width: 640px) {
  .full-width {
    grid-column: span 2;
  }
}

.flex-between {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.table-responsive {
  overflow-x: auto;
}

.m3-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.m3-table th {
  background: #F8FAFC;
  padding: 12px 14px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--md-sys-color-outline);
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  text-align: left;
}

.m3-table td {
  padding: 12px 14px;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  vertical-align: middle;
}

.role-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: var(--radius-full);
}

.tag-admin {
  background: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
}

.tag-seller {
  background: var(--md-sys-color-success-container);
  color: var(--md-sys-color-on-success);
}

.status-indicator {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 600;
}
.status-indicator.active {
  background: #E6F4EA;
  color: #1E8E3E;
}
.status-indicator.inactive {
  background: #FCE8E6;
  color: #D93025;
}

.font-mono { font-family: var(--font-mono); }
.font-bold { font-weight: 700; }
.text-right { text-align: right; }
.text-center { text-align: center; }
.text-primary { color: var(--md-sys-color-primary); }
.text-danger { color: var(--md-sys-color-error); }
.flex { display: flex; }
.gap-2 { gap: 8px; }
.gap-3 { gap: 12px; }
.justify-end { justify-content: flex-end; }
.mt-4 { margin-top: 16px; }
.mt-6 { margin-top: 24px; }
.mb-4 { margin-bottom: 16px; }
.p-6 { padding: 24px; }
.text-lg { font-size: 18px; }
.text-sm { font-size: 13px; }
</style>

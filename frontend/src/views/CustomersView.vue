<template>
  <div class="customers-page">
    <div class="page-header-row">
      <div class="search-box">
        <Search class="search-icon" />
        <input
          v-model="searchQuery"
          @input="handleSearch"
          type="text"
          class="m3-input search-input"
          placeholder="Search customers by name or mobile number..."
        />
      </div>

      <button @click="openAddCustomerModal" class="m3-btn m3-btn-primary">
        <Plus class="w-4 h-4" />
        <span>Add Customer</span>
      </button>
    </div>

    <!-- Customers Table Card -->
    <div class="m3-card table-card">
      <div v-if="loading" class="loading-state">Loading customer records...</div>

      <div v-else-if="customers.length === 0" class="empty-state">
        <Users class="w-12 h-12 text-muted" />
        <p>No customers found.</p>
      </div>

      <div v-else class="table-responsive">
        <table class="m3-table">
          <thead>
            <tr>
              <th>Customer Name</th>
              <th>Contact Phone</th>
              <th>Email</th>
              <th>Address / Notes</th>
              <th class="text-center">Total Invoices</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in customers" :key="c.id">
              <td>
                <div class="cust-name-cell">
                  <div class="cust-avatar-circle">
                    {{ (c.name || 'C').charAt(0).toUpperCase() }}
                  </div>
                  <span class="cust-name-text">{{ c.name }}</span>
                </div>
              </td>
              <td>
                <span class="phone-text">{{ c.mobile || '-' }}</span>
              </td>
              <td>
                <span class="email-text">{{ c.email || '-' }}</span>
              </td>
              <td>
                <span class="address-text">{{ c.address || '-' }}</span>
              </td>
              <td class="text-center">
                <span class="m3-chip m3-chip-primary">
                  {{ c.total_bills || 0 }} Bills
                </span>
              </td>
              <td class="text-right">
                <div class="flex gap-2 justify-end">
                  <router-link
                    :to="`/bills?q=${encodeURIComponent(c.name)}`"
                    class="m3-btn m3-btn-tonal m3-btn-sm"
                    title="View Invoices"
                  >
                    <Receipt class="w-3.5 h-3.5" />
                    <span>Invoices</span>
                  </router-link>
                  <button @click="openEditCustomerModal(c)" class="m3-btn-icon" title="Edit">
                    <Edit3 class="w-4 h-4 text-primary" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Add / Edit Customer Modal -->
    <div v-if="custModal.show" class="m3-modal-overlay" @click.self="custModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-4">
          {{ custModal.isEdit ? 'Edit Customer' : 'Add Customer' }}
        </h3>

        <form @submit.prevent="saveCustomer">
          <div class="m3-input-group">
            <label class="m3-label">Customer Full Name *</label>
            <input
              v-model="custForm.name"
              type="text"
              class="m3-input"
              placeholder="e.g. Ramesh Kumar"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Mobile Number</label>
            <input
              v-model="custForm.mobile"
              type="text"
              class="m3-input"
              placeholder="e.g. 9823012345"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Email Address (Optional)</label>
            <input
              v-model="custForm.email"
              type="email"
              class="m3-input"
              placeholder="ramesh@gmail.com"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Address / Vehicle Details</label>
            <textarea
              v-model="custForm.address"
              class="m3-textarea"
              placeholder="Workshop address or vehicle model"
            ></textarea>
          </div>

          <div class="flex gap-3 justify-end mt-6">
            <button type="button" @click="custModal.show = false" class="m3-btn m3-btn-outlined">
              Cancel
            </button>
            <button type="submit" class="m3-btn m3-btn-primary" :disabled="custModal.saving">
              {{ custModal.isEdit ? 'Save Changes' : 'Create Customer' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue';
import api from '../api/client';
import {
  Search,
  Plus,
  Users,
  Receipt,
  Edit3,
} from 'lucide-vue-next';

const customers = ref([]);
const searchQuery = ref('');
const loading = ref(true);

const custModal = reactive({
  show: false,
  isEdit: false,
  saving: false,
  customerId: null,
});

const custForm = reactive({
  name: '',
  mobile: '',
  email: '',
  address: '',
});

const loadCustomers = async () => {
  loading.value = true;
  try {
    const q = searchQuery.value ? `?q=${encodeURIComponent(searchQuery.value)}` : '';
    const res = await api.get(`/customers${q}`);
    customers.value = res.data;
  } catch (err) {
    console.error('Failed to load customers:', err);
  } finally {
    loading.value = false;
  }
};

let debounce = null;
const handleSearch = () => {
  clearTimeout(debounce);
  debounce = setTimeout(loadCustomers, 250);
};

const openAddCustomerModal = () => {
  custModal.isEdit = false;
  custModal.customerId = null;
  custForm.name = '';
  custForm.mobile = '';
  custForm.email = '';
  custForm.address = '';
  custModal.show = true;
};

const openEditCustomerModal = (c) => {
  custModal.isEdit = true;
  custModal.customerId = c.id;
  custForm.name = c.name;
  custForm.mobile = c.mobile || '';
  custForm.email = c.email || '';
  custForm.address = c.address || '';
  custModal.show = true;
};

const saveCustomer = async () => {
  custModal.saving = true;
  try {
    if (custModal.isEdit) {
      await api.put(`/customers/${custModal.customerId}`, custForm);
    } else {
      await api.post('/customers', custForm);
    }
    custModal.show = false;
    await loadCustomers();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to save customer');
  } finally {
    custModal.saving = false;
  }
};

onMounted(() => {
  loadCustomers();
});
</script>

<style scoped>
.customers-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  flex: 1;
  max-width: 420px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 12px;
  width: 18px;
  height: 18px;
  color: var(--md-sys-color-outline);
}

.search-input {
  padding-left: 42px;
}

.table-card {
  padding: 0;
  overflow: hidden;
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
  padding: 14px 16px;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--md-sys-color-outline);
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  text-align: left;
}

.m3-table td {
  padding: 14px 16px;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  vertical-align: middle;
}

.cust-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.cust-avatar-circle {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 13px;
}

.cust-name-text {
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.phone-text {
  font-family: var(--font-mono);
  font-weight: 600;
}

.email-text, .address-text {
  color: var(--md-sys-color-outline);
}

.text-center { text-align: center; }
.text-right { text-align: right; }
.text-primary { color: var(--md-sys-color-primary); }
.text-muted { color: var(--md-sys-color-outline); }
.flex { display: flex; }
.gap-2 { gap: 8px; }
.gap-3 { gap: 12px; }
.justify-end { justify-content: flex-end; }
.p-6 { padding: 24px; }
.text-lg { font-size: 18px; }
.mb-4 { margin-bottom: 16px; }
.mt-6 { margin-top: 24px; }

.loading-state, .empty-state {
  text-align: center;
  padding: 48px 16px;
  color: var(--md-sys-color-outline);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
</style>

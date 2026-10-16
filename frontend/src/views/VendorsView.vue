<template>
  <div class="vendors-page">
    <div class="header-row">
      <div>
        <h2 class="section-title">Vendors & Suppliers</h2>
        <p class="section-subtitle">Manage wholesale parts suppliers, direct manufacturer contacts, and supplier inventory links.</p>
      </div>

      <div class="header-actions">
        <div class="search-box">
          <Search class="search-icon" />
          <input
            v-model="searchQuery"
            @input="handleSearch"
            type="text"
            class="m3-input search-input"
            placeholder="Search vendors by name, phone, email..."
          />
        </div>

        <button @click="openAddVendorModal" class="m3-btn m3-btn-primary">
          <Plus class="w-4 h-4" />
          <span>Add New Vendor</span>
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">Loading vendor partners...</div>

    <div v-else-if="vendors.length === 0" class="empty-state">
      <Truck class="w-12 h-12 text-muted" />
      <p>No vendors found. Add your first vendor partner above.</p>
    </div>

    <!-- Vendors Grid -->
    <div v-else class="vendors-grid">
      <div v-for="vendor in vendors" :key="vendor.id" class="vendor-card m3-card">
        <div class="vendor-card-top">
          <div class="vendor-avatar">
            <Truck class="w-5 h-5 text-primary" />
          </div>

          <div class="vendor-card-actions">
            <button @click="openEditVendorModal(vendor)" class="m3-btn-icon" title="Edit Vendor">
              <Edit3 class="w-4 h-4 text-primary" />
            </button>
            <button @click="deleteVendor(vendor)" class="m3-btn-icon" title="Delete Vendor">
              <Trash2 class="w-4 h-4 text-danger" />
            </button>
          </div>
        </div>

        <div class="vendor-info-body">
          <h3 class="vendor-title">{{ vendor.name }}</h3>

          <div class="vendor-contact-lines">
            <div v-if="vendor.contact_no" class="contact-line">
              <Phone class="w-3.5 h-3.5 text-muted" />
              <span>{{ vendor.contact_no }}</span>
            </div>
            <div v-else class="contact-line text-muted">
              <Phone class="w-3.5 h-3.5" />
              <span>No phone recorded</span>
            </div>

            <div v-if="vendor.email" class="contact-line">
              <Mail class="w-3.5 h-3.5 text-muted" />
              <span>{{ vendor.email }}</span>
            </div>

            <div v-if="vendor.address" class="contact-line">
              <MapPin class="w-3.5 h-3.5 text-muted" />
              <span>{{ vendor.address }}</span>
            </div>
          </div>
        </div>

        <div class="vendor-card-footer">
          <div class="items-count-badge">
            <Package class="w-4 h-4 text-primary" />
            <span><strong>{{ vendor.items_count || 0 }}</strong> Items Linked</span>
          </div>

          <router-link
            :to="`/items?vendor_id=${vendor.id}`"
            class="m3-btn m3-btn-tonal m3-btn-sm"
          >
            <span>View Items</span>
          </router-link>
        </div>
      </div>
    </div>

    <!-- Add / Edit Vendor Modal -->
    <div v-if="vendorModal.show" class="m3-modal-overlay" @click.self="vendorModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-4">
          {{ vendorModal.isEdit ? 'Edit Vendor Details' : 'Add New Vendor Partner' }}
        </h3>

        <form @submit.prevent="saveVendor" class="space-y-4">
          <div class="m3-input-group">
            <label class="m3-label">Vendor / Company Name *</label>
            <input
              v-model="vendorForm.name"
              type="text"
              class="m3-input"
              placeholder="e.g. Bosch India Distribution"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Contact / Mobile Number *</label>
            <input
              v-model="vendorForm.contact_no"
              type="text"
              class="m3-input"
              placeholder="e.g. +91 9876543210"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Email Address (Optional)</label>
            <input
              v-model="vendorForm.email"
              type="email"
              class="m3-input"
              placeholder="e.g. sales@bosch-auto.com"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Office / Warehouse Address</label>
            <textarea
              v-model="vendorForm.address"
              class="m3-textarea"
              placeholder="Supplier warehouse, city, state..."
              rows="3"
            ></textarea>
          </div>

          <div class="modal-actions">
            <button type="button" @click="vendorModal.show = false" class="m3-btn m3-btn-outlined">
              Cancel
            </button>
            <button type="submit" class="m3-btn m3-btn-primary" :disabled="vendorModal.saving">
              {{ vendorModal.isEdit ? 'Save Changes' : 'Create Vendor' }}
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
  Truck,
  Phone,
  Mail,
  MapPin,
  Package,
  Edit3,
  Trash2,
} from 'lucide-vue-next';

const vendors = ref([]);
const loading = ref(true);
const searchQuery = ref('');

const vendorModal = reactive({
  show: false,
  isEdit: false,
  saving: false,
  vendorId: null,
});

const vendorForm = reactive({
  name: '',
  contact_no: '',
  email: '',
  address: '',
});

const loadVendors = async () => {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    if (searchQuery.value.trim()) params.append('q', searchQuery.value.trim());

    const res = await api.get(`/vendors?${params.toString()}`);
    vendors.value = res.data;
  } catch (err) {
    console.error('Failed to load vendors:', err);
  } finally {
    loading.value = false;
  }
};

let debounce = null;
const handleSearch = () => {
  clearTimeout(debounce);
  debounce = setTimeout(loadVendors, 250);
};

const openAddVendorModal = () => {
  vendorModal.isEdit = false;
  vendorModal.vendorId = null;
  vendorForm.name = '';
  vendorForm.contact_no = '';
  vendorForm.email = '';
  vendorForm.address = '';
  vendorModal.show = true;
};

const openEditVendorModal = (vendor) => {
  vendorModal.isEdit = true;
  vendorModal.vendorId = vendor.id;
  vendorForm.name = vendor.name;
  vendorForm.contact_no = vendor.contact_no || '';
  vendorForm.email = vendor.email || '';
  vendorForm.address = vendor.address || '';
  vendorModal.show = true;
};

const saveVendor = async () => {
  if (!vendorForm.name.trim()) return;
  vendorModal.saving = true;
  try {
    if (vendorModal.isEdit) {
      await api.put(`/vendors/${vendorModal.vendorId}`, vendorForm);
    } else {
      await api.post('/vendors', vendorForm);
    }
    vendorModal.show = false;
    await loadVendors();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to save vendor');
  } finally {
    vendorModal.saving = false;
  }
};

const deleteVendor = async (vendor) => {
  if (vendor.items_count > 0) {
    alert(`Cannot delete '${vendor.name}' because ${vendor.items_count} item(s) are linked to them. Please reassign the items first.`);
    return;
  }

  if (!confirm(`Are you sure you want to delete vendor partner '${vendor.name}'?`)) return;
  try {
    await api.delete(`/vendors/${vendor.id}`);
    await loadVendors();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to delete vendor');
  }
};

onMounted(() => {
  loadVendors();
});
</script>

<style scoped>
.vendors-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  min-width: 260px;
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

.vendors-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.vendor-card {
  padding: 18px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 16px;
  transition: transform 0.2s, box-shadow 0.2s;
}

.vendor-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--md-sys-elevation-2);
}

.vendor-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.vendor-avatar {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-sm);
  background: var(--md-sys-color-primary-container);
  display: flex;
  align-items: center;
  justify-content: center;
}

.vendor-card-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.vendor-info-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.vendor-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.vendor-contact-lines {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.contact-line {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--md-sys-color-on-surface-variant);
}

.vendor-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid var(--md-sys-color-surface-variant);
  padding-top: 12px;
}

.items-count-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--md-sys-color-outline);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.space-y-4 > * + * {
  margin-top: 16px;
}

.loading-state, .empty-state {
  text-align: center;
  padding: 48px 16px;
  color: var(--md-sys-color-outline);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.p-6 { padding: 24px; }
.text-lg { font-size: 18px; }
.mb-4 { margin-bottom: 16px; }
.text-primary { color: var(--md-sys-color-primary); }
.text-danger { color: var(--md-sys-color-error); }
.text-muted { color: var(--md-sys-color-outline); }
</style>

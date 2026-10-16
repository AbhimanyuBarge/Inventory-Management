<template>
  <div class="items-page">
    <!-- Top Action & Filter Row -->
    <div class="header-controls">
      <div class="search-and-filters">
        <div class="search-box">
          <Search class="search-icon" />
          <input
            v-model="searchQuery"
            @input="handleFilter"
            type="text"
            class="m3-input search-input"
            placeholder="Search items by name, SKU, vendor..."
          />
        </div>

        <!-- Rack Filter -->
        <select v-model="selectedRackFilter" @change="handleFilter" class="m3-select rack-filter-select">
          <option value="">All Storage Racks</option>
          <option v-for="r in racks" :key="r.id" :value="r.id">
            {{ r.name }} ({{ r.number }})
          </option>
        </select>

        <!-- Vendor Filter -->
        <select v-model="selectedVendorFilter" @change="handleFilter" class="m3-select vendor-filter-select">
          <option value="">All Vendors / Suppliers</option>
          <option v-for="v in vendors" :key="v.id" :value="v.id">
            {{ v.name }}
          </option>
        </select>

        <!-- Low Stock Filter Toggle -->
        <button
          @click="toggleLowStockFilter"
          class="m3-btn"
          :class="lowStockOnly ? 'm3-btn-danger' : 'm3-btn-outlined'"
        >
          <AlertTriangle class="w-4 h-4" />
          <span>Low Stock Only</span>
        </button>
      </div>

      <button @click="openAddItemModal" class="m3-btn m3-btn-primary">
        <Plus class="w-4 h-4" />
        <span>Add New Item</span>
      </button>
    </div>

    <!-- Items Grid / Table -->
    <div class="m3-card table-card">
      <div v-if="loading" class="loading-state">Loading inventory items catalog...</div>

      <div v-else-if="items.length === 0" class="empty-state">
        <Package class="w-12 h-12 text-muted" />
        <p>No inventory items matching your filter.</p>
      </div>

      <div v-else class="table-responsive">
        <table class="m3-table">
          <thead>
            <tr>
              <th>SKU</th>
              <th>Item Name & Description</th>
              <th>Rack Location</th>
              <th>Vendor / Supplier</th>
              <th class="text-right">Price</th>
              <th class="text-center">Stock Level</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td>
                <span class="sku-pill">{{ item.sku }}</span>
              </td>
              <td>
                <div class="item-title">{{ item.name }}</div>
                <div class="item-desc">{{ item.description || 'No description' }}</div>
              </td>
              <td>
                <span v-if="item.rack" class="rack-badge">
                  <Grid3X3 class="w-3 h-3" />
                  <span>{{ item.rack.name }} ({{ item.rack.number }})</span>
                </span>
                <span v-else class="text-muted text-xs">Unassigned</span>
              </td>
              <td>
                <div v-if="item.vendor" class="vendor-cell">
                  <div class="vendor-name-badge">
                    <Truck class="w-3 h-3 text-primary" />
                    <span>{{ item.vendor.name }}</span>
                  </div>
                  <div v-if="item.vendor.contact_no" class="vendor-phone-sub">
                    <Phone class="w-2.5 h-2.5" />
                    <span>{{ item.vendor.contact_no }}</span>
                  </div>
                </div>
                <span v-else class="text-muted text-xs">Direct / In-house</span>
              </td>
              <td class="text-right font-bold">
                ₹{{ formatNumber(item.price) }}
              </td>
              <td class="text-center">
                <div
                  class="stock-pill"
                  :class="item.quantity > item.min_stock_alert ? 'stock-pill-ok' : 'stock-pill-low'"
                >
                  <span>{{ item.quantity }} units</span>
                  <span v-if="item.quantity <= item.min_stock_alert" class="low-alert-dot">⚠️</span>
                </div>
              </td>
              <td class="text-right">
                <div class="action-buttons-row">
                  <button @click="openStockAdjust(item)" class="m3-btn m3-btn-tonal m3-btn-sm" title="Adjust Stock">
                    <SlidersHorizontal class="w-3.5 h-3.5" />
                    <span>Stock</span>
                  </button>
                  <button @click="openEditItemModal(item)" class="m3-btn-icon" title="Edit Item">
                    <Edit3 class="w-4 h-4 text-primary" />
                  </button>
                  <button @click="deleteItem(item)" class="m3-btn-icon" title="Delete Item">
                    <Trash2 class="w-4 h-4 text-danger" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Add / Edit Item Modal -->
    <div v-if="itemModal.show" class="m3-modal-overlay" @click.self="itemModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-4">
          {{ itemModal.isEdit ? 'Edit Inventory Item' : 'Add New Inventory Item' }}
        </h3>

        <form @submit.prevent="saveItem" class="form-grid">
          <div class="m3-input-group">
            <label class="m3-label">Item SKU / Code *</label>
            <input
              v-model="itemForm.sku"
              type="text"
              class="m3-input"
              placeholder="e.g. ITEM-001"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Item Name *</label>
            <input
              v-model="itemForm.name"
              type="text"
              class="m3-input"
              placeholder="e.g. Front Ceramic Brake Pad Set"
              required
            />
          </div>

          <div class="m3-input-group full-width">
            <label class="m3-label">Description</label>
            <textarea
              v-model="itemForm.description"
              class="m3-textarea"
              placeholder="Specifications, compatibility, dimensions, etc."
            ></textarea>
          </div>

          <!-- Vendor Details Section with Auto-Population -->
          <div class="m3-input-group vendor-search-col">
            <label class="m3-label">Vendor / Supplier Name</label>
            <div class="relative-vendor-field">
              <input
                v-model="itemForm.vendor_name"
                @input="handleVendorInput(itemForm.vendor_name)"
                type="text"
                class="m3-input"
                placeholder="Type vendor name (auto-populates if existing)..."
                autocomplete="off"
              />
              <div v-if="vendorSuggestions.length > 0" class="vendor-suggestions-dropdown">
                <div
                  v-for="v in vendorSuggestions"
                  :key="v.id"
                  @click="selectVendorSuggestion(v)"
                  class="vendor-sugg-row"
                >
                  <div class="sugg-v-name">{{ v.name }}</div>
                  <div class="sugg-v-contact">{{ v.contact_no || 'No contact saved' }}</div>
                </div>
              </div>
            </div>
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Vendor Contact Number</label>
            <input
              v-model="itemForm.vendor_contact_no"
              type="text"
              class="m3-input"
              placeholder="e.g. +91 9876543210"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Selling Price (₹) *</label>
            <input
              v-model.number="itemForm.price"
              type="number"
              step="0.01"
              min="0"
              class="m3-input"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Cost Price (₹)</label>
            <input
              v-model.number="itemForm.cost_price"
              type="number"
              step="0.01"
              min="0"
              class="m3-input"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Current Stock Quantity *</label>
            <input
              v-model.number="itemForm.quantity"
              type="number"
              min="0"
              class="m3-input"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Low Stock Alert Threshold</label>
            <input
              v-model.number="itemForm.min_stock_alert"
              type="number"
              min="0"
              class="m3-input"
            />
          </div>

          <div class="m3-input-group full-width">
            <label class="m3-label">Assigned Storage Rack</label>
            <select v-model="itemForm.rack_id" class="m3-select">
              <option :value="null">-- No Rack Assigned --</option>
              <option v-for="r in racks" :key="r.id" :value="r.id">
                {{ r.name }} - #{{ r.number }} ({{ r.description || 'General' }})
              </option>
            </select>
          </div>

          <div class="modal-actions full-width">
            <button type="button" @click="itemModal.show = false" class="m3-btn m3-btn-outlined">
              Cancel
            </button>
            <button type="submit" class="m3-btn m3-btn-primary" :disabled="itemModal.saving">
              {{ itemModal.isEdit ? 'Save Changes' : 'Create Item' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Stock Adjust Modal -->
    <div v-if="stockModal.show" class="m3-modal-overlay" @click.self="stockModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-1">Adjust Stock: {{ stockModal.item?.name }}</h3>
        <p class="text-sm text-muted mb-4">Current Stock: <strong>{{ stockModal.item?.quantity }} units</strong></p>

        <div class="m3-input-group">
          <label class="m3-label">Adjustment Type</label>
          <select v-model="stockModal.type" class="m3-select">
            <option value="add">Add to Stock (+)</option>
            <option value="subtract">Subtract / Waste (-)</option>
            <option value="set">Set Exact Count (=)</option>
          </select>
        </div>

        <div class="m3-input-group">
          <label class="m3-label">Quantity</label>
          <input
            v-model.number="stockModal.amount"
            type="number"
            min="1"
            class="m3-input"
            required
          />
        </div>

        <div class="flex gap-3 justify-end mt-6">
          <button @click="stockModal.show = false" class="m3-btn m3-btn-outlined">Cancel</button>
          <button @click="submitStockAdjust" class="m3-btn m3-btn-primary" :disabled="stockModal.amount <= 0">
            Apply Adjustment
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import api from '../api/client';
import {
  Search,
  Plus,
  Package,
  Grid3X3,
  AlertTriangle,
  SlidersHorizontal,
  Edit3,
  Trash2,
  Truck,
  Phone,
} from 'lucide-vue-next';

const route = useRoute();
const items = ref([]);
const racks = ref([]);
const vendors = ref([]);
const vendorSuggestions = ref([]);
const loading = ref(true);

const searchQuery = ref('');
const selectedRackFilter = ref('');
const selectedVendorFilter = ref('');
const lowStockOnly = ref(route.query.low_stock === 'true');

const itemModal = reactive({
  show: false,
  isEdit: false,
  saving: false,
  itemId: null,
});

const itemForm = reactive({
  sku: '',
  name: '',
  description: '',
  price: 0,
  cost_price: 0,
  quantity: 0,
  min_stock_alert: 5,
  rack_id: null,
  vendor_id: null,
  vendor_name: '',
  vendor_contact_no: '',
  gst_applicable: true,
  gst_rate: 18,
});

const stockModal = reactive({
  show: false,
  item: null,
  type: 'add',
  amount: 1,
});

const loadItems = async () => {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    if (searchQuery.value) params.append('q', searchQuery.value);
    if (selectedRackFilter.value) params.append('rack_id', selectedRackFilter.value);
    if (selectedVendorFilter.value) params.append('vendor_id', selectedVendorFilter.value);
    if (lowStockOnly.value) params.append('low_stock', 'true');

    const res = await api.get(`/items?${params.toString()}`);
    items.value = res.data;
  } catch (err) {
    console.error('Failed to load items:', err);
  } finally {
    loading.value = false;
  }
};

const loadRacks = async () => {
  try {
    const res = await api.get('/racks');
    racks.value = res.data;
  } catch (err) {
    console.error(err);
  }
};

const loadVendors = async () => {
  try {
    const res = await api.get('/vendors');
    vendors.value = res.data;
  } catch (err) {
    console.error(err);
  }
};

let vendorDebounce = null;
const handleVendorInput = (q) => {
  clearTimeout(vendorDebounce);
  if (!q || q.trim().length < 1) {
    vendorSuggestions.value = [];
    itemForm.vendor_id = null;
    return;
  }

  vendorDebounce = setTimeout(async () => {
    try {
      const res = await api.get(`/vendors/search?q=${encodeURIComponent(q.trim())}`);
      vendorSuggestions.value = res.data;
      // If exact match exists, auto-fill contact
      const exact = res.data.find(v => v.name.toLowerCase() === q.trim().toLowerCase());
      if (exact) {
        itemForm.vendor_id = exact.id;
        itemForm.vendor_contact_no = exact.contact_no || itemForm.vendor_contact_no;
      }
    } catch (err) {
      console.error(err);
    }
  }, 200);
};

const selectVendorSuggestion = (v) => {
  itemForm.vendor_id = v.id;
  itemForm.vendor_name = v.name;
  itemForm.vendor_contact_no = v.contact_no || '';
  vendorSuggestions.value = [];
};

let debounce = null;
const handleFilter = () => {
  clearTimeout(debounce);
  debounce = setTimeout(loadItems, 250);
};

const toggleLowStockFilter = () => {
  lowStockOnly.value = !lowStockOnly.value;
  loadItems();
};

const openAddItemModal = () => {
  itemModal.isEdit = false;
  itemModal.itemId = null;
  itemForm.sku = '';
  itemForm.name = '';
  itemForm.description = '';
  itemForm.price = 0;
  itemForm.cost_price = 0;
  itemForm.quantity = 0;
  itemForm.min_stock_alert = 5;
  itemForm.rack_id = null;
  itemForm.vendor_id = null;
  itemForm.vendor_name = '';
  itemForm.vendor_contact_no = '';
  vendorSuggestions.value = [];
  itemModal.show = true;
};

const openEditItemModal = (item) => {
  itemModal.isEdit = true;
  itemModal.itemId = item.id;
  itemForm.sku = item.sku;
  itemForm.name = item.name;
  itemForm.description = item.description || '';
  itemForm.price = item.price;
  itemForm.cost_price = item.cost_price;
  itemForm.quantity = item.quantity;
  itemForm.min_stock_alert = item.min_stock_alert;
  itemForm.rack_id = item.rack_id;
  itemForm.vendor_id = item.vendor_id || null;
  itemForm.vendor_name = item.vendor ? item.vendor.name : '';
  itemForm.vendor_contact_no = item.vendor ? (item.vendor.contact_no || '') : '';
  vendorSuggestions.value = [];
  itemModal.show = true;
};

const saveItem = async () => {
  itemModal.saving = true;
  try {
    if (itemModal.isEdit) {
      await api.put(`/items/${itemModal.itemId}`, itemForm);
    } else {
      await api.post('/items', itemForm);
    }
    itemModal.show = false;
    await loadItems();
    await loadVendors();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to save item');
  } finally {
    itemModal.saving = false;
  }
};

const openStockAdjust = (item) => {
  stockModal.item = item;
  stockModal.type = 'add';
  stockModal.amount = 1;
  stockModal.show = true;
};

const submitStockAdjust = async () => {
  if (!stockModal.item) return;
  try {
    await api.post(`/items/${stockModal.item.id}/adjust`, {
      type: stockModal.type,
      amount: stockModal.amount,
    });
    stockModal.show = false;
    await loadItems();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to adjust stock');
  }
};

const deleteItem = async (item) => {
  if (!confirm(`Are you sure you want to delete item '${item.name}'?`)) return;
  try {
    await api.delete(`/items/${item.id}`);
    await loadItems();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to delete item');
  }
};

const formatNumber = (n) => {
  if (n === undefined || n === null) return '0.00';
  return Number(n).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

onMounted(() => {
  loadItems();
  loadRacks();
  loadVendors();
});
</script>

<style scoped>
.items-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.header-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.search-and-filters {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  min-width: 240px;
  flex: 1;
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

.rack-filter-select, .vendor-filter-select {
  max-width: 200px;
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

.sku-pill {
  font-family: var(--font-mono);
  font-weight: 700;
  font-size: 12px;
  background: #F0F4F8;
  color: var(--md-sys-color-primary);
  padding: 3px 7px;
  border-radius: var(--radius-xs);
}

.item-title {
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.item-desc {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.rack-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: var(--md-sys-color-secondary-container);
  color: var(--md-sys-color-on-secondary-container);
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 600;
}

.vendor-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.vendor-name-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 700;
  font-size: 12px;
  color: var(--md-sys-color-on-surface);
}

.vendor-phone-sub {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.stock-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  font-weight: 700;
  font-size: 12px;
}

.stock-pill-ok {
  background: var(--md-sys-color-success-container);
  color: var(--md-sys-color-on-success);
}

.stock-pill-low {
  background: var(--md-sys-color-error-container);
  color: var(--md-sys-color-error);
}

.action-buttons-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
}

/* Modal Form Grid */
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.full-width {
  grid-column: span 2;
}

.relative-vendor-field {
  position: relative;
}

.vendor-suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #FFFFFF;
  border-radius: var(--radius-sm);
  box-shadow: var(--md-sys-elevation-3);
  border: 1px solid var(--md-sys-color-surface-variant);
  z-index: 50;
  max-height: 180px;
  overflow-y: auto;
  margin-top: 4px;
}

.vendor-sugg-row {
  padding: 8px 12px;
  cursor: pointer;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  transition: background 0.15s;
}
.vendor-sugg-row:hover {
  background: var(--md-sys-color-primary-container);
}

.sugg-v-name {
  font-weight: 700;
  font-size: 12px;
  color: var(--md-sys-color-on-surface);
}

.sugg-v-contact {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}

.text-right { text-align: right; }
.text-center { text-align: center; }
.font-bold { font-weight: 700; }
.text-primary { color: var(--md-sys-color-primary); }
.text-danger { color: var(--md-sys-color-error); }
.text-muted { color: var(--md-sys-color-outline); }
.text-xs { font-size: 11px; }
.flex { display: flex; }
.gap-3 { gap: 12px; }
.justify-end { justify-content: flex-end; }
.mt-6 { margin-top: 24px; }
.p-6 { padding: 24px; }
.text-lg { font-size: 18px; }
.mb-1 { margin-bottom: 4px; }
.mb-4 { margin-bottom: 16px; }

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

<template>
  <div class="parts-page">
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
            placeholder="Search parts by name, SKU..."
          />
        </div>

        <!-- Rack Filter -->
        <select v-model="selectedRackFilter" @change="handleFilter" class="m3-select rack-filter-select">
          <option value="">All Storage Racks</option>
          <option v-for="r in racks" :key="r.id" :value="r.id">
            {{ r.name }} ({{ r.number }})
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

      <button @click="openAddPartModal" class="m3-btn m3-btn-primary">
        <Plus class="w-4 h-4" />
        <span>Add Spare Part</span>
      </button>
    </div>

    <!-- Parts Grid / Table -->
    <div class="m3-card table-card">
      <div v-if="loading" class="loading-state">Loading spare parts catalog...</div>

      <div v-else-if="parts.length === 0" class="empty-state">
        <Package class="w-12 h-12 text-muted" />
        <p>No spare parts matching your filter.</p>
      </div>

      <div v-else class="table-responsive">
        <table class="m3-table">
          <thead>
            <tr>
              <th>SKU</th>
              <th>Part Name & Description</th>
              <th>Rack Location</th>
              <th class="text-right">Price</th>
              <th class="text-center">Stock Level</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="part in parts" :key="part.id">
              <td>
                <span class="sku-pill">{{ part.sku }}</span>
              </td>
              <td>
                <div class="part-title">{{ part.name }}</div>
                <div class="part-desc">{{ part.description || 'No description' }}</div>
              </td>
              <td>
                <span v-if="part.rack" class="rack-badge">
                  <Grid3X3 class="w-3 h-3" />
                  <span>{{ part.rack.name }} ({{ part.rack.number }})</span>
                </span>
                <span v-else class="text-muted text-xs">Unassigned</span>
              </td>
              <td class="text-right font-bold">
                ₹{{ formatNumber(part.price) }}
              </td>
              <td class="text-center">
                <div
                  class="stock-pill"
                  :class="part.quantity > part.min_stock_alert ? 'stock-pill-ok' : 'stock-pill-low'"
                >
                  <span>{{ part.quantity }} units</span>
                  <span v-if="part.quantity <= part.min_stock_alert" class="low-alert-dot">⚠️</span>
                </div>
              </td>
              <td class="text-right">
                <div class="action-buttons-row">
                  <button @click="openStockAdjust(part)" class="m3-btn m3-btn-tonal m3-btn-sm" title="Adjust Stock">
                    <SlidersHorizontal class="w-3.5 h-3.5" />
                    <span>Stock</span>
                  </button>
                  <button @click="openEditPartModal(part)" class="m3-btn-icon" title="Edit">
                    <Edit3 class="w-4 h-4 text-primary" />
                  </button>
                  <button @click="deletePart(part)" class="m3-btn-icon" title="Delete">
                    <Trash2 class="w-4 h-4 text-danger" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Add / Edit Part Modal -->
    <div v-if="partModal.show" class="m3-modal-overlay" @click.self="partModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-4">
          {{ partModal.isEdit ? 'Edit Spare Part' : 'Add New Spare Part' }}
        </h3>

        <form @submit.prevent="savePart" class="form-grid">
          <div class="m3-input-group">
            <label class="m3-label">Part SKU / Code *</label>
            <input
              v-model="partForm.sku"
              type="text"
              class="m3-input"
              placeholder="e.g. SP-BRK-001"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Part Name *</label>
            <input
              v-model="partForm.name"
              type="text"
              class="m3-input"
              placeholder="e.g. Front Ceramic Brake Pad Set"
              required
            />
          </div>

          <div class="m3-input-group full-width">
            <label class="m3-label">Description</label>
            <textarea
              v-model="partForm.description"
              class="m3-textarea"
              placeholder="Specifications, compatibility, etc."
            ></textarea>
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Selling Price (₹) *</label>
            <input
              v-model.number="partForm.price"
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
              v-model.number="partForm.cost_price"
              type="number"
              step="0.01"
              min="0"
              class="m3-input"
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Current Stock Quantity *</label>
            <input
              v-model.number="partForm.quantity"
              type="number"
              min="0"
              class="m3-input"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Low Stock Alert Threshold</label>
            <input
              v-model.number="partForm.min_stock_alert"
              type="number"
              min="0"
              class="m3-input"
            />
          </div>

          <div class="m3-input-group full-width">
            <label class="m3-label">Assigned Storage Rack</label>
            <select v-model="partForm.rack_id" class="m3-select">
              <option :value="null">-- No Rack Assigned --</option>
              <option v-for="r in racks" :key="r.id" :value="r.id">
                {{ r.name }} - #{{ r.number }} ({{ r.description || 'General' }})
              </option>
            </select>
          </div>

          <div class="modal-actions full-width">
            <button type="button" @click="partModal.show = false" class="m3-btn m3-btn-outlined">
              Cancel
            </button>
            <button type="submit" class="m3-btn m3-btn-primary" :disabled="partModal.saving">
              {{ partModal.isEdit ? 'Save Changes' : 'Create Part' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Stock Adjust Modal -->
    <div v-if="stockModal.show" class="m3-modal-overlay" @click.self="stockModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-1">Adjust Stock: {{ stockModal.part?.name }}</h3>
        <p class="text-sm text-muted mb-4">Current Stock: <strong>{{ stockModal.part?.quantity }} units</strong></p>

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
} from 'lucide-vue-next';

const route = useRoute();
const parts = ref([]);
const racks = ref([]);
const loading = ref(true);

const searchQuery = ref('');
const selectedRackFilter = ref('');
const lowStockOnly = ref(route.query.low_stock === 'true');

const partModal = reactive({
  show: false,
  isEdit: false,
  saving: false,
  partId: null,
});

const partForm = reactive({
  sku: '',
  name: '',
  description: '',
  price: 0,
  cost_price: 0,
  quantity: 0,
  min_stock_alert: 5,
  rack_id: null,
  gst_applicable: true,
  gst_rate: 18,
});

const stockModal = reactive({
  show: false,
  part: null,
  type: 'add',
  amount: 1,
});

const loadParts = async () => {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    if (searchQuery.value) params.append('q', searchQuery.value);
    if (selectedRackFilter.value) params.append('rack_id', selectedRackFilter.value);
    if (lowStockOnly.value) params.append('low_stock', 'true');

    const res = await api.get(`/parts?${params.toString()}`);
    parts.value = res.data;
  } catch (err) {
    console.error('Failed to load parts:', err);
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

let debounce = null;
const handleFilter = () => {
  clearTimeout(debounce);
  debounce = setTimeout(loadParts, 250);
};

const toggleLowStockFilter = () => {
  lowStockOnly.value = !lowStockOnly.value;
  loadParts();
};

const openAddPartModal = () => {
  partModal.isEdit = false;
  partModal.partId = null;
  partForm.sku = '';
  partForm.name = '';
  partForm.description = '';
  partForm.price = 0;
  partForm.cost_price = 0;
  partForm.quantity = 0;
  partForm.min_stock_alert = 5;
  partForm.rack_id = null;
  partModal.show = true;
};

const openEditPartModal = (part) => {
  partModal.isEdit = true;
  partModal.partId = part.id;
  partForm.sku = part.sku;
  partForm.name = part.name;
  partForm.description = part.description || '';
  partForm.price = part.price;
  partForm.cost_price = part.cost_price;
  partForm.quantity = part.quantity;
  partForm.min_stock_alert = part.min_stock_alert;
  partForm.rack_id = part.rack_id;
  partModal.show = true;
};

const savePart = async () => {
  partModal.saving = true;
  try {
    if (partModal.isEdit) {
      await api.put(`/parts/${partModal.partId}`, partForm);
    } else {
      await api.post('/parts', partForm);
    }
    partModal.show = false;
    await loadParts();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to save part');
  } finally {
    partModal.saving = false;
  }
};

const openStockAdjust = (part) => {
  stockModal.part = part;
  stockModal.type = 'add';
  stockModal.amount = 1;
  stockModal.show = true;
};

const submitStockAdjust = async () => {
  if (!stockModal.part) return;
  try {
    await api.post(`/parts/${stockModal.part.id}/adjust`, {
      type: stockModal.type,
      amount: stockModal.amount,
    });
    stockModal.show = false;
    await loadParts();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to adjust stock');
  }
};

const deletePart = async (part) => {
  if (!confirm(`Are you sure you want to delete '${part.name}'?`)) return;
  try {
    await api.delete(`/parts/${part.id}`);
    await loadParts();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to delete part');
  }
};

const formatNumber = (n) => {
  if (n === undefined || n === null) return '0.00';
  return Number(n).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

onMounted(() => {
  loadParts();
  loadRacks();
});
</script>

<style scoped>
.parts-page {
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
  min-width: 260px;
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

.rack-filter-select {
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

.part-title {
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.part-desc {
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

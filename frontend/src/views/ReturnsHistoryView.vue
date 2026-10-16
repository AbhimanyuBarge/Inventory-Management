<template>
  <div class="returns-history-page">
    <div class="page-header-row">
      <div class="search-box">
        <Search class="search-icon" />
        <input
          v-model="searchQuery"
          @input="handleSearch"
          type="text"
          class="m3-input search-input"
          placeholder="Search by Return #, Original Invoice #, or Customer..."
        />
      </div>

      <router-link to="/returns/new" class="m3-btn m3-btn-danger">
        <RotateCcw class="w-4 h-4" />
        <span>Process New Return</span>
      </router-link>
    </div>

    <!-- Returns Table Card -->
    <div class="m3-card table-card">
      <div v-if="loading" class="loading-state">Loading return records...</div>

      <div v-else-if="returns.length === 0" class="empty-state">
        <RotateCcw class="w-12 h-12 text-muted" />
        <p>No return records found.</p>
      </div>

      <div v-else class="table-responsive">
        <table class="m3-table">
          <thead>
            <tr>
              <th>Return #</th>
              <th>Original Invoice</th>
              <th>Return Date & Time</th>
              <th>Customer</th>
              <th>Items Restocked</th>
              <th>Reason</th>
              <th class="text-right">Refund Amount</th>
              <th class="text-center">Credit Note PDF</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="ret in returns" :key="ret.id">
              <td>
                <span class="return-badge">{{ ret.return_number }}</span>
              </td>
              <td>
                <router-link :to="`/bills?q=${encodeURIComponent(ret.original_bill?.invoice_number || '')}`" class="orig-inv-link">
                  {{ ret.original_bill?.invoice_number || '-' }}
                </router-link>
              </td>
              <td>
                <div class="ret-date">{{ formatDate(ret.created_at) }}</div>
                <div class="ret-user">By: {{ ret.user?.full_name || 'Staff' }}</div>
              </td>
              <td>
                <div class="cust-name-val">{{ ret.customer?.name || 'Walk-in Customer' }}</div>
                <div class="cust-phone-val">{{ ret.customer?.mobile || '-' }}</div>
              </td>
              <td>
                <div class="items-summary-cell">
                  <span class="m3-chip m3-chip-primary">{{ ret.items?.length || 0 }} Items</span>
                  <div class="items-list-tooltip">
                    <span v-for="it in ret.items" :key="it.id" class="item-tag">
                      {{ it.item_sku || it.part_sku }} (x{{ it.quantity }})
                    </span>
                  </div>
                </div>
              </td>
              <td>
                <span class="reason-text">{{ ret.reason || 'General Return' }}</span>
              </td>
              <td class="text-right font-bold text-refund-amount">
                ₹{{ formatNumber(ret.total_refund_amount) }}
              </td>
              <td class="text-center">
                <a
                  :href="`/api/returns/${ret.id}/pdf`"
                  target="_blank"
                  class="m3-btn m3-btn-tonal m3-btn-sm"
                  title="Download & Print Credit Note PDF"
                >
                  <FileDown class="w-3.5 h-3.5 text-danger" />
                  <span>PDF</span>
                </a>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../api/client';
import {
  Search,
  RotateCcw,
  FileDown,
} from 'lucide-vue-next';

const returns = ref([]);
const searchQuery = ref('');
const loading = ref(true);

const loadReturns = async () => {
  loading.value = true;
  try {
    const q = searchQuery.value ? `?q=${encodeURIComponent(searchQuery.value)}` : '';
    const res = await api.get(`/returns${q}`);
    returns.value = res.data;
  } catch (err) {
    console.error('Failed to load returns:', err);
  } finally {
    loading.value = false;
  }
};

let debounce = null;
const handleSearch = () => {
  clearTimeout(debounce);
  debounce = setTimeout(loadReturns, 300);
};

const formatNumber = (n) => {
  if (n === undefined || n === null) return '0.00';
  return Number(n).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

const formatDate = (d) => {
  if (!d) return '';
  const date = new Date(d);
  return date.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
};

onMounted(() => {
  loadReturns();
});
</script>

<style scoped>
.returns-history-page {
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
  max-width: 440px;
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

.return-badge {
  font-family: var(--font-mono);
  font-weight: 700;
  color: var(--md-sys-color-error);
  background: var(--md-sys-color-error-container);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.orig-inv-link {
  font-family: var(--font-mono);
  font-weight: 700;
  font-size: 12px;
  color: var(--md-sys-color-primary);
  text-decoration: none;
}
.orig-inv-link:hover {
  text-decoration: underline;
}

.ret-date {
  font-weight: 600;
  color: var(--md-sys-color-on-surface);
}

.ret-user, .cust-phone-val {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.cust-name-val {
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.items-summary-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.items-list-tooltip {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.item-tag {
  font-size: 10px;
  font-family: var(--font-mono);
  background: #F0F4F8;
  padding: 1px 5px;
  border-radius: 3px;
  color: var(--md-sys-color-outline);
}

.reason-text {
  font-size: 12px;
  color: var(--md-sys-color-on-surface-variant);
}

.text-refund-amount {
  font-size: 14px;
  color: var(--md-sys-color-error);
}

.text-right { text-align: right; }
.text-center { text-align: center; }
.font-bold { font-weight: 700; }
.text-danger { color: var(--md-sys-color-error); }
.text-muted { color: var(--md-sys-color-outline); }

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

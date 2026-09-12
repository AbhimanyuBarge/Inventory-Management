<template>
  <div class="bills-history-page">
    <div class="page-header-row">
      <div class="search-box">
        <Search class="search-icon" />
        <input
          v-model="searchQuery"
          @input="handleSearch"
          type="text"
          class="m3-input search-input"
          placeholder="Search by invoice #, customer name or phone..."
        />
      </div>

      <router-link to="/billing" class="m3-btn m3-btn-primary">
        <Plus class="w-4 h-4" />
        <span>New Invoice</span>
      </router-link>
    </div>

    <!-- Invoices Table Card -->
    <div class="m3-card table-card">
      <div v-if="loading" class="loading-state">Loading invoices...</div>

      <div v-else-if="bills.length === 0" class="empty-state">
        <FileText class="w-12 h-12 text-muted" />
        <p>No invoices found.</p>
      </div>

      <div v-else class="table-responsive">
        <table class="m3-table">
          <thead>
            <tr>
              <th>Invoice #</th>
              <th>Date & Time</th>
              <th>Customer</th>
              <th>Customer Sequence</th>
              <th>Items</th>
              <th>GST Applied</th>
              <th class="text-right">Total Amount</th>
              <th class="text-center">Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="bill in bills" :key="bill.id">
              <td>
                <span class="invoice-badge">{{ bill.invoice_number }}</span>
              </td>
              <td>
                <div class="bill-date">{{ formatDate(bill.created_at) }}</div>
                <div class="bill-user">By: {{ bill.user?.full_name || 'Staff' }}</div>
              </td>
              <td>
                <div class="cust-name-val">{{ bill.customer?.name || 'Walk-in Customer' }}</div>
                <div class="cust-phone-val">{{ bill.customer?.mobile || '-' }}</div>
              </td>
              <td>
                <span class="seq-chip">
                  <User class="w-3 h-3" />
                  <span>Bill #{{ bill.customer_bill_sequence }}</span>
                </span>
              </td>
              <td>
                <span class="items-count-badge">{{ bill.items?.length || 0 }} items</span>
              </td>
              <td>
                <span v-if="bill.gst_applied" class="m3-chip m3-chip-primary">
                  {{ bill.gst_rate }}% GST (₹{{ formatNumber(bill.gst_amount) }})
                </span>
                <span v-else class="m3-chip m3-chip-warning">
                  No GST
                </span>
              </td>
              <td class="text-right font-bold text-amount">
                ₹{{ formatNumber(bill.total_amount) }}
              </td>
              <td class="text-center">
                <a
                  :href="`/api/bills/${bill.id}/pdf`"
                  target="_blank"
                  class="m3-btn m3-btn-tonal m3-btn-sm"
                  title="Download & Print PDF"
                >
                  <FileDown class="w-3.5 h-3.5" />
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
  Plus,
  FileText,
  FileDown,
  User,
} from 'lucide-vue-next';

const bills = ref([]);
const searchQuery = ref('');
const loading = ref(true);

const loadBills = async () => {
  loading.value = true;
  try {
    const q = searchQuery.value ? `?q=${encodeURIComponent(searchQuery.value)}` : '';
    const res = await api.get(`/bills${q}`);
    bills.value = res.data;
  } catch (err) {
    console.error('Failed to load invoices:', err);
  } finally {
    loading.value = false;
  }
};

let debounce = null;
const handleSearch = () => {
  clearTimeout(debounce);
  debounce = setTimeout(loadBills, 300);
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
  loadBills();
});
</script>

<style scoped>
.bills-history-page {
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

.invoice-badge {
  font-family: var(--font-mono);
  font-weight: 700;
  color: var(--md-sys-color-primary);
  background: var(--md-sys-color-primary-container);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.bill-date {
  font-weight: 600;
  color: var(--md-sys-color-on-surface);
}

.bill-user, .cust-phone-val {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.cust-name-val {
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.seq-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--md-sys-color-secondary-container);
  color: var(--md-sys-color-on-secondary-container);
  padding: 3px 8px;
  border-radius: var(--radius-full);
  font-size: 11px;
  font-weight: 700;
}

.items-count-badge {
  font-size: 12px;
  color: var(--md-sys-color-on-surface-variant);
}

.text-amount {
  font-size: 14px;
  color: var(--md-sys-color-on-background);
}

.text-right { text-align: right; }
.text-center { text-align: center; }
.font-bold { font-weight: 700; }
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

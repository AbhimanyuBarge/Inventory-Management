<template>
  <div class="dashboard-page">
    <!-- Top Welcome Banner -->
    <div class="welcome-bar">
      <div>
        <h2 class="welcome-title">Welcome back, {{ authStore.userName }} 👋</h2>
        <p class="welcome-subtitle">Here's the latest summary of your store inventory and sales.</p>
      </div>
      <router-link to="/billing" class="m3-btn m3-btn-primary new-sale-btn">
        <Receipt class="w-4 h-4" />
        <span>Create New Bill</span>
      </router-link>
    </div>

    <!-- KPI Metric Cards Grid -->
    <div class="kpi-grid">
      <!-- Total Items -->
      <div class="kpi-card m3-card">
        <div class="kpi-icon-wrap icon-blue">
          <Package class="w-6 h-6" />
        </div>
        <div class="kpi-content">
          <span class="kpi-label">Total Items</span>
          <div class="kpi-value">{{ stats.total_items_count || stats.total_parts_count || 0 }}</div>
          <span class="kpi-sub">{{ stats.total_stock_quantity || 0 }} units total in stock</span>
        </div>
      </div>

      <!-- Racks -->
      <div class="kpi-card m3-card">
        <div class="kpi-icon-wrap icon-purple">
          <Grid3X3 class="w-6 h-6" />
        </div>
        <div class="kpi-content">
          <span class="kpi-label">Storage Racks</span>
          <div class="kpi-value">{{ stats.total_racks_count || 0 }}</div>
          <span class="kpi-sub">Organized rack compartments</span>
        </div>
      </div>

      <!-- Low Stock Alerts -->
      <div class="kpi-card m3-card" :class="{ 'kpi-warning-border': stats.low_stock_count > 0 }">
        <div class="kpi-icon-wrap" :class="stats.low_stock_count > 0 ? 'icon-red' : 'icon-green'">
          <AlertTriangle class="w-6 h-6" />
        </div>
        <div class="kpi-content">
          <span class="kpi-label">Low Stock Alerts</span>
          <div class="kpi-value" :class="{ 'text-danger': stats.low_stock_count > 0 }">
            {{ stats.low_stock_count || 0 }}
          </div>
          <span class="kpi-sub" v-if="stats.low_stock_count > 0">Needs immediate replenishment</span>
          <span class="kpi-sub" v-else>All items well stocked</span>
        </div>
      </div>

      <!-- Today's Sales -->
      <div class="kpi-card m3-card">
        <div class="kpi-icon-wrap icon-green">
          <DollarSign class="w-6 h-6" />
        </div>
        <div class="kpi-content">
          <span class="kpi-label">Today's Sales</span>
          <div class="kpi-value">₹{{ formatNumber(stats.today_sales) }}</div>
          <span class="kpi-sub">{{ stats.today_bills_count || 0 }} bills generated today</span>
        </div>
      </div>

      <!-- Month's Sales & Net Revenue -->
      <div class="kpi-card m3-card">
        <div class="kpi-icon-wrap icon-orange">
          <TrendingUp class="w-6 h-6" />
        </div>
        <div class="kpi-content">
          <span class="kpi-label">This Month</span>
          <div class="kpi-value">₹{{ formatNumber(stats.month_sales) }}</div>
          <span class="kpi-sub">Net Revenue: <strong>₹{{ formatNumber(stats.net_revenue) }}</strong></span>
        </div>
      </div>

      <!-- Sales Returns & Refunds -->
      <div class="kpi-card m3-card">
        <div class="kpi-icon-wrap icon-crimson">
          <RotateCcw class="w-6 h-6" />
        </div>
        <div class="kpi-content">
          <span class="kpi-label">Returns & Refunds</span>
          <div class="kpi-value text-danger">₹{{ formatNumber(stats.total_refund_amount) }}</div>
          <span class="kpi-sub">{{ stats.total_returns_count || 0 }} return vouchers processed</span>
        </div>
      </div>

      <!-- Customers -->
      <div class="kpi-card m3-card">
        <div class="kpi-icon-wrap icon-teal">
          <Users class="w-6 h-6" />
        </div>
        <div class="kpi-content">
          <span class="kpi-label">Total Customers</span>
          <div class="kpi-value">{{ stats.total_customers_count || 0 }}</div>
          <span class="kpi-sub">Registered buyers</span>
        </div>
      </div>
    </div>

    <!-- Two-column Section: Low Stock List & Recent Invoices -->
    <div class="dashboard-columns">
      <!-- Low Stock Items Watchlist -->
      <div class="section-card m3-card">
        <div class="section-header">
          <div class="header-with-badge">
            <h3 class="section-title">Low Stock Alert List</h3>
            <span v-if="stats.low_stock_count > 0" class="m3-chip m3-chip-error">
              {{ stats.low_stock_count }} items
            </span>
          </div>
          <router-link to="/items?low_stock=true" class="link-btn">View All</router-link>
        </div>

        <div v-if="loading" class="loading-state">Loading dashboard data...</div>

        <div v-else-if="!stats.low_stock_items || stats.low_stock_items.length === 0" class="empty-state">
          <CheckCircle2 class="w-10 h-10 text-success" />
          <p>No low stock alerts. Inventory levels are healthy!</p>
        </div>

        <div v-else class="table-responsive">
          <table class="m3-table">
            <thead>
              <tr>
                <th>Item Details</th>
                <th>Rack</th>
                <th class="text-center">Stock</th>
                <th class="text-right">Action</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in stats.low_stock_items" :key="item.id">
                <td>
                  <div class="item-name">{{ item.name }}</div>
                  <div class="item-sku">{{ item.sku }}</div>
                </td>
                <td>
                  <span v-if="item.rack" class="rack-pill">
                    {{ item.rack.name }} ({{ item.rack.number }})
                  </span>
                  <span v-else class="text-muted">-</span>
                </td>
                <td class="text-center">
                  <span class="stock-badge-critical">{{ item.quantity }} left</span>
                </td>
                <td class="text-right">
                  <button @click="openQuickRestock(item)" class="m3-btn m3-btn-tonal m3-btn-sm">
                    + Restock
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Recent Bills -->
      <div class="section-card m3-card">
        <div class="section-header">
          <h3 class="section-title">Recent Invoices</h3>
          <router-link to="/bills" class="link-btn">View All</router-link>
        </div>

        <div v-if="loading" class="loading-state">Loading invoices...</div>

        <div v-else-if="!stats.recent_bills || stats.recent_bills.length === 0" class="empty-state">
          <Receipt class="w-10 h-10 text-muted" />
          <p>No invoices generated yet.</p>
        </div>

        <div v-else class="table-responsive">
          <table class="m3-table">
            <thead>
              <tr>
                <th>Invoice #</th>
                <th>Customer</th>
                <th>Payment</th>
                <th>Amount</th>
                <th class="text-right">PDF</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="bill in stats.recent_bills" :key="bill.id">
                <td>
                  <div class="invoice-num font-mono">{{ bill.invoice_number || bill.bill_number }}</div>
                  <div class="bill-time">{{ formatDate(bill.created_at) }}</div>
                </td>
                <td>
                  <div class="cust-name font-bold">{{ bill.customer?.name || bill.customer_name || 'Walk-in Customer' }}</div>
                  <div class="cust-seq text-xs text-muted">
                    {{ bill.customer?.mobile ? bill.customer.mobile : 'Bill #' + (bill.customer_bill_sequence || bill.customer_sequence_number || 1) }}
                  </div>
                </td>
                <td>
                  <span class="payment-badge">{{ (bill.payment_mode || 'cash').toUpperCase() }}</span>
                  <div class="text-xs text-muted mt-0.5">{{ bill.items?.length || 1 }} item(s)</div>
                </td>
                <td>
                  <div class="font-bold text-sm">₹{{ formatNumber(bill.total_amount || bill.final_total) }}</div>
                  <div class="text-xs text-muted" v-if="bill.gst_applied">Incl. GST</div>
                  <div class="text-xs text-muted" v-else>No GST</div>
                </td>
                <td class="text-right">
                  <a :href="`/api/bills/${bill.id}/pdf`" target="_blank" class="m3-btn-icon" title="View & Download PDF Invoice">
                    <FileDown class="w-4 h-4 text-primary" />
                  </a>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Quick Restock Modal -->
    <div v-if="restockModal.show" class="m3-modal-overlay" @click.self="restockModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-2">Restock: {{ restockModal.item?.name }}</h3>
        <p class="text-sm text-muted mb-4">Current Stock: <strong>{{ restockModal.item?.quantity }}</strong></p>

        <div class="m3-input-group">
          <label class="m3-label">Add Quantity</label>
          <input
            v-model.number="restockModal.amount"
            type="number"
            min="1"
            class="m3-input"
            placeholder="e.g. 10"
          />
        </div>

        <div class="flex gap-3 justify-end mt-6">
          <button @click="restockModal.show = false" class="m3-btn m3-btn-outlined">Cancel</button>
          <button @click="submitRestock" class="m3-btn m3-btn-primary" :disabled="restockModal.loading || restockModal.amount <= 0">
            Confirm Restock
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue';
import { useAuthStore } from '../stores/auth';
import api from '../api/client';
import {
  Package,
  Grid3X3,
  AlertTriangle,
  DollarSign,
  TrendingUp,
  Users,
  Receipt,
  CheckCircle2,
  FileText,
  FileDown,
  RotateCcw,
} from 'lucide-vue-next';

const authStore = useAuthStore();
const stats = ref({});
const loading = ref(true);

const restockModal = reactive({
  show: false,
  item: null,
  amount: 10,
  loading: false,
});

const loadStats = async () => {
  loading.value = true;
  try {
    const res = await api.get('/dashboard/stats');
    stats.value = res.data;
  } catch (err) {
    console.error('Failed to load dashboard stats:', err);
  } finally {
    loading.value = false;
  }
};

const openQuickRestock = (item) => {
  restockModal.item = item;
  restockModal.amount = 10;
  restockModal.show = true;
};

const submitRestock = async () => {
  if (!restockModal.item || restockModal.amount <= 0) return;
  restockModal.loading = true;
  try {
    await api.post(`/items/${restockModal.item.id}/adjust`, {
      type: 'add',
      amount: restockModal.amount,
    });
    restockModal.show = false;
    await loadStats();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to restock item');
  } finally {
    restockModal.loading = false;
  }
};

const formatNumber = (num) => {
  if (num === undefined || num === null) return '0.00';
  return Number(num).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

const formatDate = (d) => {
  if (!d) return '';
  const date = new Date(d);
  return date.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });
};

onMounted(() => {
  loadStats();
});
</script>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.welcome-bar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  justify-content: space-between;
  align-items: flex-start;
}
@media (min-width: 640px) {
  .welcome-bar {
    flex-direction: row;
    align-items: center;
  }
}

.welcome-title {
  font-size: 20px;
  font-weight: 800;
  color: var(--md-sys-color-on-background);
}

.welcome-subtitle {
  font-size: 13px;
  color: var(--md-sys-color-outline);
}

.new-sale-btn {
  box-shadow: var(--md-sys-elevation-2);
}

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
}

.kpi-card {
  padding: 20px;
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.kpi-warning-border {
  border-left: 4px solid var(--md-sys-color-error);
}

.kpi-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-blue { background: #E8F0FE; color: #00629E; }
.icon-purple { background: #F3E8FF; color: #7C3AED; }
.icon-green { background: #E6F4EA; color: #1E8E3E; }
.icon-red { background: #FCE8E6; color: #D93025; }
.icon-orange { background: #FEF7E0; color: #E37400; }
.icon-crimson { background: #FDE8E8; color: #A02028; }
.icon-teal { background: #E0F2F1; color: #00897B; }

.kpi-content {
  flex: 1;
}

.kpi-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--md-sys-color-outline);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.kpi-value {
  font-size: 24px;
  font-weight: 800;
  color: var(--md-sys-color-on-background);
  line-height: 1.2;
  margin: 2px 0 4px;
}

.text-danger {
  color: var(--md-sys-color-error);
}

.kpi-sub {
  font-size: 11px;
  color: var(--md-sys-color-outline);
  display: block;
}

/* Columns */
.dashboard-columns {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}
@media (min-width: 1024px) {
  .dashboard-columns {
    grid-template-columns: 1fr 1fr;
  }
}

.section-card {
  padding: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.header-with-badge {
  display: flex;
  align-items: center;
  gap: 10px;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--md-sys-color-on-background);
}

.link-btn {
  font-size: 13px;
  font-weight: 700;
  color: var(--md-sys-color-primary);
  text-decoration: none;
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
  text-align: left;
  padding: 10px 12px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--md-sys-color-outline);
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
}

.m3-table td {
  padding: 12px;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  vertical-align: middle;
}

.part-name, .invoice-num {
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.part-sku, .bill-time, .cust-seq {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.rack-pill {
  font-size: 11px;
  background: var(--md-sys-color-secondary-container);
  color: var(--md-sys-color-on-secondary-container);
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-weight: 600;
}

.stock-badge-critical {
  background: var(--md-sys-color-error-container);
  color: var(--md-sys-color-error);
  font-weight: 700;
  font-size: 11px;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
}

.text-center { text-align: center; }
.text-right { text-align: right; }
.font-bold { font-weight: 700; }
.text-primary { color: var(--md-sys-color-primary); }
.text-muted { color: var(--md-sys-color-outline); }
.text-success { color: var(--md-sys-color-success); }

.empty-state, .loading-state {
  text-align: center;
  padding: 36px 16px;
  color: var(--md-sys-color-outline);
  font-size: 13px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.text-xs { font-size: 11px; }
.font-mono { font-family: var(--font-mono, monospace); }
.mt-0\.5 { margin-top: 2px; }

.payment-badge {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.5px;
  padding: 2px 6px;
  border-radius: 4px;
  background: #E2E8F0;
  color: #334155;
  display: inline-block;
}

.p-6 { padding: 24px; }
.text-lg { font-size: 18px; }
.flex { display: flex; }
.gap-3 { gap: 12px; }
.justify-end { justify-content: flex-end; }
.mt-6 { margin-top: 24px; }
.mb-2 { margin-bottom: 8px; }
.mb-4 { margin-bottom: 16px; }
</style>

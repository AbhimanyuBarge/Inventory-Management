<template>
  <div class="process-return-page">
    <!-- Header -->
    <div class="header-row">
      <div>
        <h2 class="section-title">Process Sales Return & Refund</h2>
        <p class="section-subtitle">Select an original invoice, choose returned items, and issue a Credit Note with automatic inventory restock.</p>
      </div>

      <router-link to="/returns" class="m3-btn m3-btn-tonal">
        <History class="w-4 h-4" />
        <span>Return History</span>
      </router-link>
    </div>

    <!-- Step 1: Lookup Original Invoice -->
    <div class="m3-card lookup-box">
      <div class="title-row">
        <Search class="w-5 h-5 text-primary" />
        <h3 class="box-title">1. Search & Select Original Invoice</h3>
      </div>

      <div class="lookup-grid">
        <div class="search-input-wrapper">
          <input
            v-model="invoiceSearchInput"
            @input="handleInvoiceSearch"
            type="text"
            class="m3-input"
            placeholder="Type Invoice Number (e.g. INV-2026-0001) or Customer Phone..."
          />
          <button @click="handleInvoiceSearch" class="m3-btn m3-btn-primary search-action-btn">
            <Search class="w-4 h-4" />
            <span>Search</span>
          </button>
        </div>
      </div>

      <!-- Search Suggestions / Results -->
      <div v-if="invoiceSuggestions.length > 0" class="invoice-results-list">
        <div
          v-for="b in invoiceSuggestions"
          :key="b.id"
          @click="selectBill(b)"
          class="invoice-result-card"
          :class="{ active: selectedBill && selectedBill.id === b.id }"
        >
          <div class="inv-res-left">
            <div class="inv-badge">{{ b.invoice_number }}</div>
            <div class="inv-cust-name">{{ b.customer?.name || 'Walk-in Customer' }}</div>
            <div class="inv-date">{{ formatDate(b.created_at) }} • Bill #{{ b.customer_bill_sequence }}</div>
          </div>
          <div class="inv-res-right">
            <div class="inv-amount">₹{{ formatNumber(b.total_amount) }}</div>
            <span class="m3-chip m3-chip-primary">{{ b.items?.length || 0 }} Items</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 2: Selected Invoice & Return Items Selection -->
    <div v-if="selectedBill" class="return-workflow-grid">
      <!-- Left: Items Selection Checklist -->
      <div class="m3-card items-selection-card">
        <div class="bill-overview-banner">
          <div class="overview-item">
            <span class="lbl">Invoice Number:</span>
            <strong class="val text-primary">{{ selectedBill.invoice_number }}</strong>
          </div>
          <div class="overview-item">
            <span class="lbl">Customer:</span>
            <strong class="val">{{ selectedBill.customer?.name || 'Walk-in' }} ({{ selectedBill.customer?.mobile || '-' }})</strong>
          </div>
          <div class="overview-item">
            <span class="lbl">Date:</span>
            <span class="val">{{ formatDate(selectedBill.created_at) }}</span>
          </div>
          <div class="overview-item">
            <span class="lbl">GST Applied:</span>
            <span class="val">{{ selectedBill.gst_applied ? `${selectedBill.gst_rate}%` : 'None' }}</span>
          </div>
        </div>

        <h3 class="box-title mt-4 mb-3">2. Select Items to Return</h3>

        <div v-if="loadingReturnable" class="text-center py-6 text-muted">
          Checking returnable stock limits...
        </div>

        <div v-else-if="returnableItems.length === 0" class="text-center py-6 text-muted">
          No items found for this invoice.
        </div>

        <div v-else class="items-return-list">
          <div
            v-for="item in returnableItems"
            :key="item.bill_item_id"
            class="return-item-row"
            :class="{ 'fully-returned': item.returnable_quantity <= 0, selected: returnSelections[item.bill_item_id]?.selected }"
          >
            <!-- Checkbox toggle -->
            <div class="item-checkbox-col">
              <input
                type="checkbox"
                :disabled="item.returnable_quantity <= 0"
                v-model="returnSelections[item.bill_item_id].selected"
                class="m3-checkbox"
              />
            </div>

            <!-- Item Info -->
            <div class="item-info-col">
              <div class="item-header">
                <span class="item-sku">{{ item.item_sku || item.part_sku }}</span>
                <span class="item-name">{{ item.item_name || item.part_name }}</span>
              </div>
              <div class="item-meta">
                <span>Location: {{ item.rack_info || 'General' }}</span>
                <span>• Unit Price: ₹{{ formatNumber(item.unit_price) }}</span>
              </div>
              <div class="item-stock-history">
                <span>Purchased: <strong>{{ item.original_quantity }}</strong></span>
                <span>Already Returned: <strong>{{ item.returned_quantity }}</strong></span>
                <span class="remaining-badge" :class="item.returnable_quantity > 0 ? 'badge-avail' : 'badge-none'">
                  Available to Return: {{ item.returnable_quantity }}
                </span>
              </div>
            </div>

            <!-- Return Quantity Selector -->
            <div v-if="item.returnable_quantity > 0" class="item-qty-col">
              <span class="qty-label">Return Qty:</span>
              <div class="qty-stepper">
                <button
                  type="button"
                  @click="decrementReturnQty(item.bill_item_id)"
                  :disabled="!returnSelections[item.bill_item_id].selected || returnSelections[item.bill_item_id].quantity <= 1"
                  class="qty-btn"
                >
                  <Minus class="w-3.5 h-3.5" />
                </button>
                <span class="qty-display">{{ returnSelections[item.bill_item_id].quantity }}</span>
                <button
                  type="button"
                  @click="incrementReturnQty(item.bill_item_id, item.returnable_quantity)"
                  :disabled="!returnSelections[item.bill_item_id].selected || returnSelections[item.bill_item_id].quantity >= item.returnable_quantity"
                  class="qty-btn"
                >
                  <Plus class="w-3.5 h-3.5" />
                </button>
              </div>

              <div class="item-refund-sum">
                ₹{{ formatNumber(item.unit_price * (returnSelections[item.bill_item_id].selected ? returnSelections[item.bill_item_id].quantity : 0)) }}
              </div>
            </div>

            <div v-else class="fully-returned-notice">
              All Units Returned
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Refund Summary & Confirmation Card -->
      <div class="m3-card refund-summary-card">
        <h3 class="box-title mb-4">3. Refund Summary</h3>

        <div class="m3-input-group">
          <label class="m3-label">Return Reason *</label>
          <select v-model="returnForm.reason" class="m3-select">
            <option value="Customer Return / Changed Mind">Customer Return / Changed Mind</option>
            <option value="Defective / Damaged Item">Defective / Damaged Item</option>
            <option value="Wrong Item Fitment / Model">Wrong Item Fitment / Model</option>
            <option value="Exchange for Alternative">Exchange for Alternative</option>
            <option value="Other">Other Reason</option>
          </select>
        </div>

        <div class="m3-input-group">
          <label class="m3-label">Refund Payment Mode</label>
          <select v-model="returnForm.refund_mode" class="m3-select">
            <option value="cash">Cash Refund</option>
            <option value="upi">UPI / Online Transfer</option>
            <option value="store_credit">Store Credit Voucher</option>
            <option value="card">Card Reversal</option>
          </select>
        </div>

        <div class="m3-input-group">
          <label class="m3-label">Additional Notes (Optional)</label>
          <textarea
            v-model="returnForm.notes"
            class="m3-textarea"
            placeholder="Remarks or inspector notes..."
          ></textarea>
        </div>

        <div class="calc-divider"></div>

        <!-- Calculations -->
        <div class="calc-breakdown">
          <div class="calc-row">
            <span>Selected Items:</span>
            <strong>{{ selectedItemsCount }} items</strong>
          </div>
          <div class="calc-row">
            <span>Refund Subtotal:</span>
            <strong>₹{{ formatNumber(calculatedRefundSubtotal) }}</strong>
          </div>
          <div v-if="selectedBill.gst_applied" class="calc-row">
            <span>GST Refund ({{ selectedBill.gst_rate }}%):</span>
            <strong class="text-primary">+ ₹{{ formatNumber(calculatedGSTRefund) }}</strong>
          </div>

          <div class="total-refund-box">
            <span class="total-lbl">Total Refund Amount</span>
            <span class="total-val">₹{{ formatNumber(calculatedTotalRefund) }}</span>
          </div>
        </div>

        <button
          @click="submitReturn"
          class="m3-btn m3-btn-danger w-full submit-return-btn"
          :disabled="selectedItemsCount === 0 || submitting"
        >
          <RotateCcw class="w-5 h-5" />
          <span v-if="submitting">Processing Return...</span>
          <span v-else>Confirm Return & Generate Credit Note</span>
        </button>
      </div>
    </div>

    <!-- Success Modal -->
    <div v-if="successModal.show" class="m3-modal-overlay">
      <div class="m3-modal-card p-6 text-center">
        <div class="success-icon-circle">
          <CheckCircle2 class="w-12 h-12 text-success" />
        </div>
        <h3 class="text-xl font-bold mt-3">Return Processed Successfully!</h3>
        <p class="text-sm text-muted mt-1">Credit Note #: <strong>{{ successModal.returnBill?.return_number }}</strong></p>

        <div class="success-summary-box">
          <div class="summary-pill">
            <span class="pill-label">Original Invoice</span>
            <span class="pill-val">{{ successModal.returnBill?.original_bill?.invoice_number }}</span>
          </div>
          <div class="summary-pill">
            <span class="pill-label">Total Refunded</span>
            <span class="pill-val">₹{{ formatNumber(successModal.returnBill?.total_refund_amount) }}</span>
          </div>
        </div>

        <div class="flex flex-col gap-3 mt-6">
          <a
            :href="`/api/returns/${successModal.returnBill?.id}/pdf`"
            target="_blank"
            class="m3-btn m3-btn-primary w-full"
          >
            <FileDown class="w-4 h-4" />
            <span>Download & Print Credit Note (PDF)</span>
          </a>
          <router-link to="/returns" class="m3-btn m3-btn-outlined w-full">
            View Return History
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import api from '../api/client';
import {
  Search,
  History,
  RotateCcw,
  Minus,
  Plus,
  CheckCircle2,
  FileDown,
} from 'lucide-vue-next';

const route = useRoute();

const invoiceSearchInput = ref('');
const invoiceSuggestions = ref([]);
const selectedBill = ref(null);
const returnableItems = ref([]);
const loadingReturnable = ref(false);

const returnSelections = reactive({});
const returnForm = reactive({
  reason: 'Customer Return / Changed Mind',
  refund_mode: 'cash',
  notes: '',
});

const submitting = ref(false);
const successModal = reactive({
  show: false,
  returnBill: null,
});

let searchDebounce = null;
const handleInvoiceSearch = () => {
  clearTimeout(searchDebounce);
  const q = invoiceSearchInput.value.trim();
  if (!q) {
    invoiceSuggestions.value = [];
    return;
  }
  searchDebounce = setTimeout(async () => {
    try {
      const res = await api.get(`/bills?q=${encodeURIComponent(q)}`);
      invoiceSuggestions.value = res.data;
    } catch (err) {
      console.error(err);
    }
  }, 250);
};

const selectBill = async (bill) => {
  selectedBill.value = bill;
  invoiceSuggestions.value = [];
  invoiceSearchInput.value = bill.invoice_number;
  await fetchReturnableItems(bill.id);
};

const fetchReturnableItems = async (billId) => {
  loadingReturnable.value = true;
  try {
    const res = await api.get(`/bills/${billId}/returnable`);
    returnableItems.value = res.data.items || [];

    // Reset return selections
    Object.keys(returnSelections).forEach((key) => delete returnSelections[key]);
    returnableItems.value.forEach((item) => {
      returnSelections[item.bill_item_id] = {
        selected: item.returnable_quantity > 0,
        quantity: item.returnable_quantity > 0 ? 1 : 0,
      };
    });
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to load bill items');
  } finally {
    loadingReturnable.value = false;
  }
};

const incrementReturnQty = (billItemId, maxQty) => {
  if (returnSelections[billItemId].quantity < maxQty) {
    returnSelections[billItemId].quantity++;
  }
};

const decrementReturnQty = (billItemId) => {
  if (returnSelections[billItemId].quantity > 1) {
    returnSelections[billItemId].quantity--;
  }
};

// Computations
const selectedItemsCount = computed(() => {
  return Object.values(returnSelections).filter((s) => s.selected && s.quantity > 0).length;
});

const calculatedRefundSubtotal = computed(() => {
  let subtotal = 0;
  returnableItems.value.forEach((item) => {
    const sel = returnSelections[item.bill_item_id];
    if (sel && sel.selected && sel.quantity > 0) {
      subtotal += item.unit_price * sel.quantity;
    }
  });
  return subtotal;
});

const calculatedGSTRefund = computed(() => {
  if (!selectedBill.value || !selectedBill.value.gst_applied) return 0;
  return Math.round((calculatedRefundSubtotal.value * (selectedBill.value.gst_rate / 100)) * 100) / 100;
});

const calculatedTotalRefund = computed(() => {
  return Math.round((calculatedRefundSubtotal.value + calculatedGSTRefund.value) * 100) / 100;
});

// Submit Return
const submitReturn = async () => {
  if (selectedItemsCount.value === 0) {
    alert('Please select at least one item to return.');
    return;
  }

  const itemsPayload = [];
  returnableItems.value.forEach((item) => {
    const sel = returnSelections[item.bill_item_id];
    if (sel && sel.selected && sel.quantity > 0) {
      itemsPayload.push({
        bill_item_id: item.bill_item_id,
        quantity: sel.quantity,
      });
    }
  });

  submitting.value = true;
  try {
    const payload = {
      original_bill_id: selectedBill.value.id,
      reason: returnForm.reason,
      refund_mode: returnForm.refund_mode,
      notes: returnForm.notes,
      items: itemsPayload,
    };

    const res = await api.post('/returns', payload);
    successModal.returnBill = res.data;
    successModal.show = true;
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to process return');
  } finally {
    submitting.value = false;
  }
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

onMounted(async () => {
  const billIdFromQuery = route.query.bill_id;
  if (billIdFromQuery) {
    try {
      const res = await api.get(`/bills/${billIdFromQuery}`);
      selectBill(res.data);
    } catch (err) {
      console.error(err);
    }
  }
});
</script>

<style scoped>
.process-return-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.section-title {
  font-size: 18px;
  font-weight: 800;
  color: var(--md-sys-color-on-background);
}

.section-subtitle {
  font-size: 13px;
  color: var(--md-sys-color-outline);
}

.lookup-box {
  padding: 20px;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.box-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.search-input-wrapper {
  display: flex;
  gap: 10px;
}

.search-action-btn {
  padding: 0 20px;
}

.invoice-results-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
  margin-top: 16px;
  max-height: 220px;
  overflow-y: auto;
}

.invoice-result-card {
  padding: 12px 14px;
  background: #FFFFFF;
  border: 1px solid var(--md-sys-color-surface-variant);
  border-radius: var(--radius-md);
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: all 0.2s;
}

.invoice-result-card:hover {
  border-color: var(--md-sys-color-primary);
  background: var(--md-sys-color-background);
}

.invoice-result-card.active {
  border-color: var(--md-sys-color-primary);
  background: var(--md-sys-color-primary-container);
}

.inv-badge {
  font-family: var(--font-mono);
  font-weight: 700;
  font-size: 12px;
  color: var(--md-sys-color-primary);
}

.inv-cust-name {
  font-weight: 700;
  font-size: 13px;
  margin: 2px 0;
}

.inv-date {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.inv-amount {
  font-weight: 800;
  font-size: 14px;
  text-align: right;
  margin-bottom: 4px;
}

/* Return Workflow */
.return-workflow-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}
@media (min-width: 1024px) {
  .return-workflow-grid {
    grid-template-columns: 1fr 380px;
  }
}

.items-selection-card {
  padding: 20px;
}

.bill-overview-banner {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
  background: var(--md-sys-color-background);
  padding: 12px 16px;
  border-radius: var(--radius-md);
  border: 1px solid var(--md-sys-color-surface-variant);
}

.overview-item {
  display: flex;
  flex-direction: column;
}

.overview-item .lbl {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.overview-item .val {
  font-size: 13px;
  font-weight: 700;
}

.items-return-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.return-item-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px;
  border: 1px solid var(--md-sys-color-surface-variant);
  border-radius: var(--radius-md);
  background: #FFFFFF;
  transition: all 0.2s;
}

.return-item-row.selected {
  border-color: var(--md-sys-color-error);
  background: #FFF9F9;
}

.return-item-row.fully-returned {
  opacity: 0.5;
  background: #F8FAFC;
}

.m3-checkbox {
  width: 20px;
  height: 20px;
  accent-color: var(--md-sys-color-error);
  cursor: pointer;
}

.item-info-col {
  flex: 1;
}

.item-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.item-sku {
  font-family: var(--font-mono);
  font-size: 11px;
  background: #F0F4F8;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 700;
}

.item-name {
  font-weight: 700;
  font-size: 13px;
}

.item-meta {
  font-size: 11px;
  color: var(--md-sys-color-outline);
  margin: 3px 0;
}

.item-stock-history {
  display: flex;
  gap: 12px;
  font-size: 11px;
  flex-wrap: wrap;
}

.remaining-badge {
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 700;
}
.badge-avail { background: #E6F4EA; color: #1E8E3E; }
.badge-none { background: #FCE8E6; color: #D93025; }

.item-qty-col {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.qty-label {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.qty-stepper {
  display: flex;
  align-items: center;
  border: 1px solid var(--md-sys-color-outline-variant);
  border-radius: var(--radius-sm);
  background: #FFF;
}

.qty-btn {
  background: none;
  border: none;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}
.qty-btn:disabled { opacity: 0.3; cursor: not-allowed; }

.qty-display {
  padding: 0 8px;
  font-weight: 700;
  font-size: 13px;
}

.item-refund-sum {
  font-weight: 800;
  font-size: 14px;
  color: var(--md-sys-color-error);
}

.fully-returned-notice {
  font-size: 11px;
  color: var(--md-sys-color-error);
  font-weight: 700;
  padding: 6px 10px;
  background: var(--md-sys-color-error-container);
  border-radius: var(--radius-sm);
}

/* Summary Card */
.refund-summary-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: fit-content;
}

.calc-divider {
  height: 1px;
  background: var(--md-sys-color-surface-variant);
  margin: 6px 0;
}

.calc-breakdown {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.calc-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.total-refund-box {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  background: var(--md-sys-color-error-container);
  color: var(--md-sys-color-error);
  padding: 12px 14px;
  border-radius: var(--radius-md);
  margin-top: 8px;
}

.total-lbl {
  font-weight: 700;
  font-size: 13px;
}

.total-val {
  font-weight: 800;
  font-size: 20px;
}

.submit-return-btn {
  height: 48px;
  margin-top: 10px;
  font-size: 14px;
}

.w-full { width: 100%; }
.mt-4 { margin-top: 16px; }
.mb-3 { margin-bottom: 12px; }
.mb-4 { margin-bottom: 16px; }
.py-6 { padding-top: 24px; padding-bottom: 24px; }
.text-center { text-align: center; }
.text-muted { color: var(--md-sys-color-outline); }
.text-primary { color: var(--md-sys-color-primary); }
.text-success { color: var(--md-sys-color-success); }
.text-xl { font-size: 20px; }

.success-icon-circle {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: #E6F4EA;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}

.success-summary-box {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 16px;
  background: var(--md-sys-color-background);
  padding: 12px;
  border-radius: var(--radius-md);
}

.summary-pill {
  display: flex;
  flex-direction: column;
}

.pill-label {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.pill-val {
  font-size: 15px;
  font-weight: 800;
  color: var(--md-sys-color-on-background);
}
</style>

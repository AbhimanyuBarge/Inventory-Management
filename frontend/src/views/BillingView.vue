<template>
  <div class="billing-page">
    <div class="billing-layout">
      <!-- Left Column: Customer & Part Selector -->
      <div class="billing-main">
        <!-- Step 1: Customer Details Card (Auto-fill by name or phone) -->
        <div class="m3-card section-box">
          <div class="section-title-row">
            <div class="title-with-icon">
              <UserCheck class="w-5 h-5 text-primary" />
              <h3 class="box-title">Customer Information</h3>
            </div>
            <span v-if="selectedCustomer" class="m3-chip m3-chip-primary">
              Returning Customer (Past Bills: {{ selectedCustomer.total_bills || 0 }})
            </span>
          </div>

          <div class="customer-inputs-grid">
            <!-- Search & Auto-populate Phone / Name -->
            <div class="m3-input-group relative">
              <label class="m3-label">Customer Mobile</label>
              <div class="input-container">
                <Phone class="field-icon" />
                <input
                  v-model="customerForm.mobile"
                  @input="handleCustomerSearch(customerForm.mobile)"
                  type="text"
                  class="m3-input field-padded"
                  placeholder="e.g. 9823012345"
                />
              </div>

              <!-- Customer Auto-suggest dropdown -->
              <div v-if="customerSuggestions.length > 0" class="suggestions-popup">
                <div
                  v-for="c in customerSuggestions"
                  :key="c.id"
                  @click="selectExistingCustomer(c)"
                  class="suggestion-item"
                >
                  <div class="sugg-name">{{ c.name }}</div>
                  <div class="sugg-phone">{{ c.mobile || 'No Phone' }} • Bills: {{ c.total_bills || 0 }}</div>
                </div>
              </div>
            </div>

            <div class="m3-input-group relative">
              <label class="m3-label">Customer Name *</label>
              <div class="input-container">
                <User class="field-icon" />
                <input
                  v-model="customerForm.name"
                  @input="handleCustomerSearch(customerForm.name)"
                  type="text"
                  class="m3-input field-padded"
                  placeholder="Full Name (Mandatory)"
                  required
                />
              </div>
            </div>

            <div class="m3-input-group">
              <label class="m3-label">Customer Email (Optional)</label>
              <input
                v-model="customerForm.email"
                type="email"
                class="m3-input"
                placeholder="customer@email.com"
              />
            </div>

            <div class="m3-input-group">
              <label class="m3-label">Address / Vehicle No.</label>
              <input
                v-model="customerForm.address"
                type="text"
                class="m3-input"
                placeholder="City / Car Model / Reg No."
              />
            </div>
          </div>
        </div>

        <!-- Step 2: Search and Add Spare Parts -->
        <div class="m3-card section-box">
          <div class="section-title-row">
            <div class="title-with-icon">
              <Search class="w-5 h-5 text-primary" />
              <h3 class="box-title">Search & Add Spare Parts</h3>
            </div>
            <span class="text-sm text-muted">Click an item to add to bill</span>
          </div>

          <div class="parts-search-bar">
            <Search class="search-icon" />
            <input
              v-model="partsSearchQuery"
              @input="filterParts"
              type="text"
              class="m3-input search-input"
              placeholder="Search by part name, SKU, or rack..."
            />
          </div>

          <!-- Quick Part Selection Chips/Grid -->
          <div class="parts-selector-list">
            <div
              v-for="part in filteredParts"
              :key="part.id"
              @click="addItemToCart(part)"
              class="part-item-card"
              :class="{ 'out-of-stock': part.quantity <= 0 }"
            >
              <div class="part-item-left">
                <div class="part-sku-badge">{{ part.sku }}</div>
                <div class="part-name-text">{{ part.name }}</div>
                <div class="part-location-badge">
                  <Grid3X3 class="w-3 h-3" />
                  <span>{{ part.rack ? `${part.rack.name} (${part.rack.number})` : 'Unassigned' }}</span>
                </div>
              </div>

              <div class="part-item-right">
                <div class="part-price-tag">₹{{ formatNumber(part.price) }}</div>
                <div
                  class="stock-status-chip"
                  :class="part.quantity > 5 ? 'stock-ok' : part.quantity > 0 ? 'stock-low' : 'stock-zero'"
                >
                  {{ part.quantity }} in stock
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column: Cart & Checkout Summary -->
      <div class="billing-sidebar">
        <div class="m3-card cart-box">
          <div class="cart-header">
            <div class="title-with-icon">
              <ShoppingCart class="w-5 h-5 text-primary" />
              <h3 class="box-title">Current Bill Cart</h3>
            </div>
            <button v-if="cart.length > 0" @click="clearCart" class="clear-btn">
              Clear All
            </button>
          </div>

          <!-- Cart Items List -->
          <div v-if="cart.length === 0" class="empty-cart-state">
            <ShoppingCart class="w-12 h-12 text-muted" />
            <p>Your cart is currently empty.</p>
            <span>Search or tap spare parts on the left to add items.</span>
          </div>

          <div v-else class="cart-items-scroll">
            <div v-for="(item, idx) in cart" :key="item.spare_part_id" class="cart-row">
              <div class="cart-row-details">
                <div class="cart-item-name">{{ item.name }}</div>
                <div class="cart-item-meta">
                  <span>{{ item.sku }}</span>
                  <span v-if="item.rack_info">• {{ item.rack_info }}</span>
                </div>
                <div class="cart-item-unit-price">₹{{ formatNumber(item.unit_price) }} each</div>
              </div>

              <div class="cart-row-actions">
                <!-- Quantity Stepper -->
                <div class="qty-stepper">
                  <button @click="decreaseQty(idx)" class="qty-btn" :disabled="item.quantity <= 1">
                    <Minus class="w-3.5 h-3.5" />
                  </button>
                  <span class="qty-value">{{ item.quantity }}</span>
                  <button @click="increaseQty(idx)" class="qty-btn" :disabled="item.quantity >= item.max_stock">
                    <Plus class="w-3.5 h-3.5" />
                  </button>
                </div>

                <div class="cart-item-total">
                  ₹{{ formatNumber(item.unit_price * item.quantity) }}
                </div>

                <button @click="removeItem(idx)" class="cart-remove-btn" title="Remove">
                  <Trash2 class="w-4 h-4 text-danger" />
                </button>
              </div>
            </div>
          </div>

          <!-- Bill Calculation & GST Controls -->
          <div class="cart-footer">
            <div class="calc-row">
              <span>Subtotal:</span>
              <span class="font-bold">₹{{ formatNumber(cartSubtotal) }}</span>
            </div>

            <!-- GST Toggle Checkbox -->
            <div class="gst-toggle-card">
              <label class="gst-checkbox-label">
                <input
                  v-model="addGST"
                  type="checkbox"
                  class="gst-checkbox"
                />
                <div class="gst-label-text">
                  <span class="gst-title">Add GST ({{ defaultGSTRate }}%)</span>
                  <span class="gst-sub">Applies government tax to invoice</span>
                </div>
              </label>
              <span v-if="addGST" class="gst-amount-text">+ ₹{{ formatNumber(gstAmount) }}</span>
            </div>

            <!-- Discount Row -->
            <div class="calc-row align-center">
              <span>Discount (₹):</span>
              <input
                v-model.number="discount"
                type="number"
                min="0"
                class="m3-input discount-input"
                placeholder="0.00"
              />
            </div>

            <!-- Payment Mode -->
            <div class="calc-row align-center">
              <span>Payment Mode:</span>
              <select v-model="paymentMode" class="m3-select payment-select">
                <option value="cash">Cash</option>
                <option value="upi">UPI / Online</option>
                <option value="card">Card (Debit/Credit)</option>
                <option value="credit">Credit / Account</option>
              </select>
            </div>

            <div class="divider"></div>

            <!-- Grand Total -->
            <div class="grand-total-row">
              <div class="total-label">Grand Total</div>
              <div class="total-amount">₹{{ formatNumber(grandTotal) }}</div>
            </div>

            <!-- Action Button -->
            <button
              @click="submitBill"
              class="m3-btn m3-btn-primary checkout-btn"
              :disabled="cart.length === 0 || !customerForm.name || submitting"
            >
              <Printer class="w-5 h-5" />
              <span v-if="submitting">Processing & Printing...</span>
              <span v-else>Generate & Print Invoice</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Bill Success Modal -->
    <div v-if="successModal.show" class="m3-modal-overlay">
      <div class="m3-modal-card p-6 text-center">
        <div class="success-icon-circle">
          <CheckCircle2 class="w-12 h-12 text-success" />
        </div>
        <h3 class="text-xl font-bold mt-3">Invoice Generated Successfully!</h3>
        <p class="text-sm text-muted mt-1">Invoice Number: <strong>{{ successModal.bill?.invoice_number }}</strong></p>

        <div class="success-summary-box">
          <div class="summary-pill">
            <span class="pill-label">Customer Sequence</span>
            <span class="pill-val">Bill #{{ successModal.bill?.customer_bill_sequence }}</span>
          </div>
          <div class="summary-pill">
            <span class="pill-label">Total Amount</span>
            <span class="pill-val">₹{{ formatNumber(successModal.bill?.total_amount) }}</span>
          </div>
        </div>

        <div class="flex flex-col gap-3 mt-6">
          <a
            :href="`/api/bills/${successModal.bill?.id}/pdf`"
            target="_blank"
            class="m3-btn m3-btn-primary w-full"
          >
            <FileDown class="w-4 h-4" />
            <span>Open & Print PDF Invoice</span>
          </a>
          <button @click="resetAfterSuccess" class="m3-btn m3-btn-outlined w-full">
            Start Next Bill
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import api from '../api/client';
import {
  UserCheck,
  Phone,
  User,
  Search,
  ShoppingCart,
  Minus,
  Plus,
  Trash2,
  Printer,
  Grid3X3,
  CheckCircle2,
  FileDown,
} from 'lucide-vue-next';

// State
const availableParts = ref([]);
const filteredParts = ref([]);
const partsSearchQuery = ref('');
const defaultGSTRate = ref(18.00);

const customerForm = reactive({
  id: null,
  name: '',
  mobile: '',
  email: '',
  address: '',
});

const selectedCustomer = ref(null);
const customerSuggestions = ref([]);

const cart = ref([]);
const addGST = ref(true);
const discount = ref(0);
const paymentMode = ref('cash');
const submitting = ref(false);

const successModal = reactive({
  show: false,
  bill: null,
});

// Fetch inventory parts & store settings
const loadData = async () => {
  try {
    const [partsRes, settingsRes] = await Promise.all([
      api.get('/parts'),
      api.get('/settings'),
    ]);
    availableParts.value = partsRes.data;
    filteredParts.value = partsRes.data;
    if (settingsRes.data && settingsRes.data.default_gst_rate) {
      defaultGSTRate.value = settingsRes.data.default_gst_rate;
    }
  } catch (err) {
    console.error('Failed to load parts:', err);
  }
};

// Filter parts search
const filterParts = () => {
  const q = partsSearchQuery.value.toLowerCase().trim();
  if (!q) {
    filteredParts.value = availableParts.value;
    return;
  }
  filteredParts.value = availableParts.value.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.sku.toLowerCase().includes(q) ||
      (p.rack && (p.rack.name.toLowerCase().includes(q) || p.rack.number.toLowerCase().includes(q)))
  );
};

// Search existing customers
let searchDebounce = null;
const handleCustomerSearch = (query) => {
  if (selectedCustomer.value && selectedCustomer.value.name !== customerForm.name) {
    selectedCustomer.value = null;
    customerForm.id = null;
  }

  clearTimeout(searchDebounce);
  if (!query || query.length < 2) {
    customerSuggestions.value = [];
    return;
  }

  searchDebounce = setTimeout(async () => {
    try {
      const res = await api.get(`/customers/search?q=${encodeURIComponent(query)}`);
      customerSuggestions.value = res.data;
    } catch (err) {
      console.error(err);
    }
  }, 250);
};

const selectExistingCustomer = (c) => {
  selectedCustomer.value = c;
  customerForm.id = c.id;
  customerForm.name = c.name;
  customerForm.mobile = c.mobile || '';
  customerForm.email = c.email || '';
  customerForm.address = c.address || '';
  customerSuggestions.value = [];
};

// Cart Actions
const addItemToCart = (part) => {
  if (part.quantity <= 0) {
    alert('This part is currently out of stock!');
    return;
  }

  const existing = cart.value.find((i) => i.spare_part_id === part.id);
  if (existing) {
    if (existing.quantity < part.quantity) {
      existing.quantity++;
    } else {
      alert(`Cannot add more. Available stock limit is ${part.quantity}.`);
    }
  } else {
    cart.value.push({
      spare_part_id: part.id,
      name: part.name,
      sku: part.sku,
      rack_info: part.rack ? `${part.rack.name} (${part.rack.number})` : '',
      unit_price: part.price,
      quantity: 1,
      max_stock: part.quantity,
    });
  }
};

const increaseQty = (idx) => {
  const item = cart.value[idx];
  if (item.quantity < item.max_stock) {
    item.quantity++;
  }
};

const decreaseQty = (idx) => {
  const item = cart.value[idx];
  if (item.quantity > 1) {
    item.quantity--;
  }
};

const removeItem = (idx) => {
  cart.value.splice(idx, 1);
};

const clearCart = () => {
  cart.value = [];
};

// Calculations
const cartSubtotal = computed(() => {
  return cart.value.reduce((sum, item) => sum + item.unit_price * item.quantity, 0);
});

const gstAmount = computed(() => {
  if (!addGST.value) return 0;
  return Math.round((cartSubtotal.value * (defaultGSTRate.value / 100)) * 100) / 100;
});

const grandTotal = computed(() => {
  const disc = Number(discount.value) || 0;
  const tot = cartSubtotal.value + gstAmount.value - disc;
  return tot > 0 ? tot : 0;
});

// Submit and generate bill
const submitBill = async () => {
  if (!customerForm.name.trim()) {
    alert('Customer Name is mandatory.');
    return;
  }
  if (cart.value.length === 0) {
    alert('Please add at least one item to the bill.');
    return;
  }

  submitting.value = true;
  try {
    const payload = {
      customer_id: customerForm.id || null,
      name: customerForm.name.trim(),
      mobile: customerForm.mobile.trim(),
      email: customerForm.email.trim(),
      address: customerForm.address.trim(),
      gst_applied: addGST.value,
      gst_rate: defaultGSTRate.value,
      discount: Number(discount.value) || 0,
      payment_mode: paymentMode.value,
      items: cart.value.map((i) => ({
        spare_part_id: i.spare_part_id,
        quantity: i.quantity,
        unit_price: i.unit_price,
      })),
    };

    const res = await api.post('/bills', payload);
    successModal.bill = res.data;
    successModal.show = true;

    // Reload parts to reflect updated stock levels
    await loadData();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to create bill');
  } finally {
    submitting.value = false;
  }
};

const resetAfterSuccess = () => {
  successModal.show = false;
  successModal.bill = null;
  customerForm.id = null;
  customerForm.name = '';
  customerForm.mobile = '';
  customerForm.email = '';
  customerForm.address = '';
  selectedCustomer.value = null;
  cart.value = [];
  discount.value = 0;
};

const formatNumber = (n) => {
  if (n === undefined || n === null) return '0.00';
  return Number(n).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

onMounted(() => {
  loadData();
});
</script>

<style scoped>
.billing-layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}
@media (min-width: 1024px) {
  .billing-layout {
    grid-template-columns: 1fr 400px;
  }
}

.section-box {
  padding: 20px;
  margin-bottom: 20px;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.title-with-icon {
  display: flex;
  align-items: center;
  gap: 8px;
}

.box-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--md-sys-color-on-background);
}

.customer-inputs-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
@media (min-width: 640px) {
  .customer-inputs-grid {
    grid-template-columns: 1fr 1fr;
  }
}

.relative {
  position: relative;
}

.input-container {
  position: relative;
  display: flex;
  align-items: center;
}

.field-icon {
  position: absolute;
  left: 12px;
  width: 18px;
  height: 18px;
  color: var(--md-sys-color-outline);
  pointer-events: none;
}

.field-padded {
  padding-left: 38px !important;
}

.suggestions-popup {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #FFFFFF;
  border-radius: var(--radius-md);
  box-shadow: var(--md-sys-elevation-3);
  border: 1px solid var(--md-sys-color-surface-variant);
  z-index: 50;
  max-height: 200px;
  overflow-y: auto;
  margin-top: 4px;
}

.suggestion-item {
  padding: 10px 14px;
  cursor: pointer;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  transition: background 0.15s;
}
.suggestion-item:hover {
  background: var(--md-sys-color-primary-container);
}

.sugg-name {
  font-weight: 700;
  font-size: 13px;
  color: var(--md-sys-color-on-surface);
}

.sugg-phone {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

/* Parts Selector */
.parts-search-bar {
  position: relative;
  margin-bottom: 16px;
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

.parts-selector-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
  max-height: 480px;
  overflow-y: auto;
  padding-right: 4px;
}

.part-item-card {
  padding: 12px 14px;
  background: #FFFFFF;
  border: 1px solid var(--md-sys-color-surface-variant);
  border-radius: var(--radius-md);
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.part-item-card:hover {
  border-color: var(--md-sys-color-primary);
  box-shadow: var(--md-sys-elevation-1);
  transform: translateY(-2px);
}

.part-item-card.out-of-stock {
  opacity: 0.5;
  cursor: not-allowed;
}

.part-sku-badge {
  font-size: 10px;
  font-family: var(--font-mono);
  background: #F0F4F8;
  color: var(--md-sys-color-primary);
  padding: 2px 6px;
  border-radius: 4px;
  display: inline-block;
  font-weight: 700;
}

.part-name-text {
  font-weight: 700;
  font-size: 13px;
  margin: 3px 0;
  color: var(--md-sys-color-on-surface);
  line-height: 1.3;
}

.part-location-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.part-item-right {
  text-align: right;
  flex-shrink: 0;
}

.part-price-tag {
  font-weight: 800;
  font-size: 14px;
  color: var(--md-sys-color-on-background);
}

.stock-status-chip {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  margin-top: 4px;
  display: inline-block;
}

.stock-ok { background: #E6F4EA; color: #1E8E3E; }
.stock-low { background: #FEF7E0; color: #E37400; }
.stock-zero { background: #FCE8E6; color: #D93025; }

/* Cart Box */
.cart-box {
  padding: 20px;
  display: flex;
  flex-direction: column;
  position: sticky;
  top: 80px;
}

.cart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
}

.clear-btn {
  background: none;
  border: none;
  color: var(--md-sys-color-error);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.empty-cart-state {
  text-align: center;
  padding: 40px 16px;
  color: var(--md-sys-color-outline);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.cart-items-scroll {
  max-height: 280px;
  overflow-y: auto;
  padding: 10px 0;
}

.cart-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 0;
  border-bottom: 1px dashed var(--md-sys-color-surface-variant);
}

.cart-item-name {
  font-weight: 700;
  font-size: 13px;
}

.cart-item-meta {
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.cart-item-unit-price {
  font-size: 11px;
  color: var(--md-sys-color-primary);
  font-weight: 600;
}

.cart-row-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
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
  color: var(--md-sys-color-on-surface);
}

.qty-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.qty-value {
  padding: 0 8px;
  font-weight: 700;
  font-size: 13px;
}

.cart-item-total {
  font-weight: 700;
  font-size: 14px;
}

.cart-remove-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
}

.cart-footer {
  margin-top: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.calc-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.align-center {
  align-items: center;
}

.gst-toggle-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: var(--md-sys-color-background);
  border-radius: var(--radius-md);
  border: 1px solid var(--md-sys-color-surface-variant);
}

.gst-checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.gst-checkbox {
  width: 18px;
  height: 18px;
  accent-color: var(--md-sys-color-primary);
  cursor: pointer;
}

.gst-title {
  display: block;
  font-size: 13px;
  font-weight: 700;
}

.gst-sub {
  display: block;
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.gst-amount-text {
  font-weight: 700;
  font-size: 13px;
  color: var(--md-sys-color-primary);
}

.discount-input, .payment-select {
  width: 110px;
  padding: 6px 10px;
  font-size: 13px;
}

.divider {
  height: 1px;
  background: var(--md-sys-color-surface-variant);
  margin: 6px 0;
}

.grand-total-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.total-label {
  font-size: 15px;
  font-weight: 700;
}

.total-amount {
  font-size: 24px;
  font-weight: 800;
  color: var(--md-sys-color-primary);
}

.checkout-btn {
  height: 48px;
  width: 100%;
  font-size: 15px;
}

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
  font-weight: 600;
}

.pill-val {
  font-size: 16px;
  font-weight: 800;
  color: var(--md-sys-color-on-background);
}

.w-full { width: 100%; }
.text-danger { color: var(--md-sys-color-error); }
.text-success { color: var(--md-sys-color-success); }
.text-muted { color: var(--md-sys-color-outline); }
.font-bold { font-weight: 700; }
</style>

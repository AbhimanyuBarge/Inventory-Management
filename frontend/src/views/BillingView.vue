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
            <!-- 1. Customer Name (First) -->
            <div class="m3-input-group relative">
              <label class="m3-label">Customer Name *</label>
              <div class="input-container">
                <User class="field-icon" />
                <input
                  v-model="customerForm.name"
                  @input="handleCustomerSearch(customerForm.name, 'name')"
                  type="text"
                  class="m3-input field-padded"
                  placeholder="Full Name (Mandatory)"
                  required
                />
              </div>

              <!-- Customer Auto-suggest dropdown when searching by name -->
              <div v-if="customerSuggestions.length > 0 && activeSearchField === 'name'" class="suggestions-popup">
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

            <!-- 2. Customer Mobile (Second) -->
            <div class="m3-input-group relative">
              <label class="m3-label">Customer Mobile</label>
              <div class="input-container">
                <Phone class="field-icon" />
                <input
                  v-model="customerForm.mobile"
                  @input="handleCustomerSearch(customerForm.mobile, 'mobile')"
                  type="text"
                  class="m3-input field-padded"
                  placeholder="e.g. 9823012345"
                />
              </div>

              <!-- Customer Auto-suggest dropdown when searching by phone -->
              <div v-if="customerSuggestions.length > 0 && activeSearchField === 'mobile'" class="suggestions-popup">
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

        <!-- Step 2: Search and Add Inventory Items -->
        <div class="m3-card section-box">
          <div class="section-title-row">
            <div class="title-with-icon">
              <Search class="w-5 h-5 text-primary" />
              <h3 class="box-title">Search & Add Inventory Items</h3>
            </div>
            <span class="text-sm text-muted">Click an item to add to bill</span>
          </div>

          <div class="items-search-bar">
            <Search class="search-icon" />
            <input
              v-model="itemsSearchQuery"
              @input="filterItems"
              type="text"
              class="m3-input search-input"
              placeholder="Search by item name, SKU, or rack..."
            />
          </div>

          <!-- Quick Item Selection Chips/Grid -->
          <div class="items-selector-list">
            <div
              v-for="item in filteredItems"
              :key="item.id"
              @click="addItemToCart(item)"
              class="item-card-tile"
              :class="{ 'out-of-stock': item.quantity <= 0 }"
            >
              <div class="item-card-left">
                <div class="item-sku-badge">{{ item.sku }}</div>
                <div class="item-name-text">{{ item.name }}</div>
                <div class="item-location-badge">
                  <Grid3X3 class="w-3 h-3" />
                  <span>{{ item.rack ? `${item.rack.name} (${item.rack.number})` : 'Unassigned' }}</span>
                </div>
              </div>

              <div class="item-card-right">
                <div class="item-price-tag">₹{{ formatNumber(item.price) }}</div>
                <div
                  class="stock-status-chip"
                  :class="item.quantity > 5 ? 'stock-ok' : item.quantity > 0 ? 'stock-low' : 'stock-zero'"
                >
                  {{ item.quantity }} in stock
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
            <span>Search or tap inventory items on the left to add to bill.</span>
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

            <!-- Action Button: Open Confirmation & Summary Modal -->
            <button
              @click="openConfirmationModal"
              class="m3-btn m3-btn-primary checkout-btn"
              :disabled="cart.length === 0 || !customerForm.name || submitting"
            >
              <Printer class="w-5 h-5" />
              <span>Generate & Print Invoice</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Invoice Confirmation & Summary Modal -->
    <div v-if="confirmModal.show" class="m3-modal-overlay">
      <div class="m3-modal-card confirm-modal-card">
        <!-- Modal Header -->
        <div class="confirm-modal-header">
          <div class="flex items-center gap-3">
            <div class="confirm-header-icon">
              <Receipt class="w-6 h-6 text-primary" />
            </div>
            <div>
              <h3 class="text-lg font-bold">Invoice Summary & Confirmation</h3>
              <p class="text-xs text-muted">Please review invoice details before confirming and printing</p>
            </div>
          </div>
          <button @click="confirmModal.show = false" class="modal-close-btn" :disabled="submitting">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="confirm-modal-body">
          <!-- Customer Info Box -->
          <div class="confirm-section-box">
            <div class="confirm-section-title">
              <UserCheck class="w-4 h-4 text-primary" />
              <span>Customer Details</span>
              <span v-if="selectedCustomer" class="m3-chip m3-chip-primary text-xs ml-auto">
                Returning Customer
              </span>
            </div>
            <div class="customer-summary-grid">
              <div class="summary-field">
                <span class="field-lbl">Name:</span>
                <span class="field-val font-bold">{{ customerForm.name }}</span>
              </div>
              <div class="summary-field">
                <span class="field-lbl">Mobile:</span>
                <span class="field-val">{{ customerForm.mobile || '—' }}</span>
              </div>
              <div v-if="customerForm.email" class="summary-field">
                <span class="field-lbl">Email:</span>
                <span class="field-val">{{ customerForm.email }}</span>
              </div>
              <div v-if="customerForm.address" class="summary-field">
                <span class="field-lbl">Address/Vehicle:</span>
                <span class="field-val">{{ customerForm.address }}</span>
              </div>
            </div>
          </div>

          <!-- Items Summary Table -->
          <div class="confirm-section-box">
            <div class="confirm-section-title">
              <ShoppingCart class="w-4 h-4 text-primary" />
              <span>Items Breakdown ({{ cart.length }} items, {{ totalCartUnits }} units)</span>
            </div>
            <div class="confirm-items-table-wrapper">
              <table class="confirm-items-table">
                <thead>
                  <tr>
                    <th>Item & SKU</th>
                    <th>Rack</th>
                    <th class="text-center">Qty</th>
                    <th class="text-right">Price</th>
                    <th class="text-right">Total</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in cart" :key="item.spare_part_id">
                    <td>
                      <div class="font-bold text-sm">{{ item.name }}</div>
                      <div class="text-xs text-muted font-mono">{{ item.sku }}</div>
                    </td>
                    <td>
                      <span class="rack-badge text-xs">{{ item.rack_info || 'Unassigned' }}</span>
                    </td>
                    <td class="text-center font-bold">{{ item.quantity }}</td>
                    <td class="text-right">₹{{ formatNumber(item.unit_price) }}</td>
                    <td class="text-right font-bold">₹{{ formatNumber(item.unit_price * item.quantity) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Financial Breakdown Summary -->
          <div class="confirm-calc-box">
            <div class="confirm-calc-row">
              <span>Items Subtotal:</span>
              <span class="font-bold">₹{{ formatNumber(cartSubtotal) }}</span>
            </div>

            <div class="confirm-calc-row">
              <span>GST ({{ addGST ? defaultGSTRate + '%' : 'Exempt / Not Added' }}):</span>
              <span :class="addGST ? 'text-primary font-bold' : 'text-muted'">
                {{ addGST ? '+ ₹' + formatNumber(gstAmount) : '₹0.00' }}
              </span>
            </div>

            <div v-if="discount > 0" class="confirm-calc-row">
              <span>Discount:</span>
              <span class="text-danger font-bold">- ₹{{ formatNumber(discount) }}</span>
            </div>

            <div class="confirm-calc-row">
              <span>Payment Mode:</span>
              <span class="payment-mode-pill">{{ paymentMode.toUpperCase() }}</span>
            </div>

            <div class="divider"></div>

            <div class="confirm-grand-total-row">
              <span class="total-title">Total Invoice Amount:</span>
              <span class="total-value">₹{{ formatNumber(grandTotal) }}</span>
            </div>
          </div>
        </div>

        <!-- Modal Footer Actions -->
        <div class="confirm-modal-footer">
          <button
            @click="confirmModal.show = false"
            class="m3-btn m3-btn-outlined"
            :disabled="submitting"
          >
            Edit Bill
          </button>
          <button
            @click="confirmAndSubmitBill"
            class="m3-btn m3-btn-primary confirm-submit-btn"
            :disabled="submitting"
          >
            <Printer class="w-5 h-5" />
            <span v-if="submitting">Generating & Printing...</span>
            <span v-else>Confirm & Print Invoice</span>
          </button>
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
  X,
  Receipt,
  CreditCard,
} from 'lucide-vue-next';

// State
const availableItems = ref([]);
const filteredItems = ref([]);
const itemsSearchQuery = ref('');
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
const activeSearchField = ref('name');

const cart = ref([]);
const addGST = ref(true);
const discount = ref(0);
const paymentMode = ref('cash');
const submitting = ref(false);

const confirmModal = reactive({
  show: false,
});

const successModal = reactive({
  show: false,
  bill: null,
});

// Fetch inventory items & store settings
const loadData = async () => {
  try {
    const [itemsRes, settingsRes] = await Promise.all([
      api.get('/items'),
      api.get('/settings'),
    ]);
    availableItems.value = itemsRes.data;
    filteredItems.value = itemsRes.data;
    if (settingsRes.data && settingsRes.data.default_gst_rate) {
      defaultGSTRate.value = settingsRes.data.default_gst_rate;
    }
  } catch (err) {
    console.error('Failed to load items:', err);
  }
};

// Filter items search
const filterItems = () => {
  const q = itemsSearchQuery.value.toLowerCase().trim();
  if (!q) {
    filteredItems.value = availableItems.value;
    return;
  }
  filteredItems.value = availableItems.value.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.sku.toLowerCase().includes(q) ||
      (p.rack && (p.rack.name.toLowerCase().includes(q) || p.rack.number.toLowerCase().includes(q)))
  );
};

// Search existing customers
let searchDebounce = null;
const handleCustomerSearch = (query, field = 'name') => {
  activeSearchField.value = field;
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
const addItemToCart = (item) => {
  if (item.quantity <= 0) {
    alert('This item is currently out of stock!');
    return;
  }

  const existing = cart.value.find((i) => (i.item_id === item.id || i.spare_part_id === item.id));
  if (existing) {
    if (existing.quantity < item.quantity) {
      existing.quantity++;
    } else {
      alert(`Cannot add more. Available stock limit is ${item.quantity}.`);
    }
  } else {
    cart.value.push({
      item_id: item.id,
      spare_part_id: item.id,
      name: item.name,
      sku: item.sku,
      rack_info: item.rack ? `${item.rack.name} (${item.rack.number})` : '',
      unit_price: item.price,
      quantity: 1,
      max_stock: item.quantity,
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

const totalCartUnits = computed(() => {
  return cart.value.reduce((sum, item) => sum + item.quantity, 0);
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

// Open Invoice Summary & Confirmation Modal
const openConfirmationModal = () => {
  if (!customerForm.name.trim()) {
    alert('Customer Name is mandatory before generating invoice.');
    return;
  }
  if (cart.value.length === 0) {
    alert('Please add at least one item to the bill.');
    return;
  }
  confirmModal.show = true;
};

// Submit and generate bill with automatic print trigger
const confirmAndSubmitBill = async () => {
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
        item_id: i.item_id || i.spare_part_id,
        spare_part_id: i.item_id || i.spare_part_id,
        quantity: i.quantity,
        unit_price: i.unit_price,
      })),
    };

    const res = await api.post('/bills', payload);
    const createdBill = res.data;

    // Close confirmation modal
    confirmModal.show = false;

    // Open Success Modal
    successModal.bill = createdBill;
    successModal.show = true;

    // Automatically open/print PDF Invoice in new tab
    if (createdBill && createdBill.id) {
      window.open(`/api/bills/${createdBill.id}/pdf`, '_blank');
    }

    // Reload items to reflect updated stock levels
    await loadData();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to create invoice');
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

.billing-main {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-box {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  padding-bottom: 12px;
}

.title-with-icon {
  display: flex;
  align-items: center;
  gap: 8px;
}

.box-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
}

.customer-form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.full-col {
  grid-column: span 2;
}

.customer-search-relative {
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

.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #FFFFFF;
  border-radius: var(--radius-sm);
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

/* Items Selector */
.items-search-bar {
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

.items-selector-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
  max-height: 480px;
  overflow-y: auto;
  padding-right: 4px;
}

.item-card-tile {
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

.item-card-tile:hover {
  border-color: var(--md-sys-color-primary);
  box-shadow: var(--md-sys-elevation-1);
  transform: translateY(-2px);
}

.item-card-tile.out-of-stock {
  opacity: 0.5;
  cursor: not-allowed;
}

.item-card-left {
  flex: 1;
  min-width: 0;
}

.item-sku-badge {
  font-size: 10px;
  font-family: var(--font-mono);
  background: #F0F4F8;
  color: var(--md-sys-color-primary);
  padding: 2px 6px;
  border-radius: 4px;
  display: inline-block;
  font-weight: 700;
}

.item-name-text {
  font-weight: 700;
  font-size: 13px;
  margin: 3px 0;
  color: var(--md-sys-color-on-surface);
  line-height: 1.3;
}

.item-location-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--md-sys-color-outline);
}

.item-card-right {
  text-align: right;
  flex-shrink: 0;
}

.item-price-tag {
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
.text-xs { font-size: 11px; }
.text-sm { font-size: 13px; }
.text-lg { font-size: 18px; }
.text-xl { font-size: 20px; }
.text-center { text-align: center; }
.text-right { text-align: right; }
.ml-auto { margin-left: auto; }
.items-center { align-items: center; }
.gap-3 { gap: 12px; }

/* Confirmation & Summary Modal */
.confirm-modal-card {
  max-width: 620px;
  background: #FFFFFF;
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
}

.confirm-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--md-sys-color-surface-variant);
  background: var(--md-sys-color-surface);
}

.confirm-header-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #E1F0FA;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.modal-close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--md-sys-color-outline);
  padding: 6px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}
.modal-close-btn:hover {
  background: var(--md-sys-color-surface-variant);
  color: var(--md-sys-color-on-surface);
}

.confirm-modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-height: calc(85vh - 140px);
  overflow-y: auto;
}

.confirm-section-box {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-md);
  padding: 14px;
}

.confirm-section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 700;
  color: var(--md-sys-color-on-surface);
  margin-bottom: 10px;
}

.customer-summary-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  font-size: 13px;
}

.summary-field {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.field-lbl {
  font-size: 11px;
  color: var(--md-sys-color-outline);
  font-weight: 600;
}

.field-val {
  color: var(--md-sys-color-on-surface);
  word-break: break-word;
}

.confirm-items-table-wrapper {
  max-height: 180px;
  overflow-y: auto;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  background: #FFFFFF;
}

.confirm-items-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.confirm-items-table th {
  background: #F1F5F9;
  padding: 8px 12px;
  text-align: left;
  font-weight: 700;
  color: #475569;
  border-bottom: 1px solid #E2E8F0;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.confirm-items-table td {
  padding: 8px 12px;
  border-bottom: 1px solid #F1F5F9;
  vertical-align: middle;
}

.confirm-items-table tr:last-child td {
  border-bottom: none;
}

.rack-badge {
  background: #EFF6FF;
  color: #1E40AF;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
}

.confirm-calc-box {
  background: #F0FDF4;
  border: 1px solid #BBF7D0;
  border-radius: var(--radius-md);
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.confirm-calc-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: #334155;
}

.payment-mode-pill {
  background: #E2E8F0;
  color: #1E293B;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 700;
  font-size: 11px;
  letter-spacing: 0.5px;
}

.confirm-grand-total-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding-top: 4px;
}

.total-title {
  font-size: 15px;
  font-weight: 800;
  color: #166534;
}

.total-value {
  font-size: 22px;
  font-weight: 900;
  color: #15803D;
}

.confirm-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--md-sys-color-surface-variant);
  background: var(--md-sys-color-surface);
}

.confirm-submit-btn {
  padding: 10px 20px;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 4px 12px rgba(0, 98, 158, 0.25);
}
</style>

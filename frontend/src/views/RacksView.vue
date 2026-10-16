<template>
  <div class="racks-page">
    <div class="header-row">
      <div>
        <h2 class="section-title">Storage Racks & Compartments</h2>
        <p class="section-subtitle">Organize inventory items by rack location and print physical warehouse labels.</p>
      </div>

      <button @click="openAddRackModal" class="m3-btn m3-btn-primary">
        <Plus class="w-4 h-4" />
        <span>Create Rack</span>
      </button>
    </div>

    <div v-if="loading" class="loading-state">Loading rack compartments...</div>

    <div v-else-if="racks.length === 0" class="empty-state">
      <Grid3X3 class="w-12 h-12 text-muted" />
      <p>No storage racks created yet.</p>
    </div>

    <!-- Racks Grid -->
    <div v-else class="racks-grid">
      <div v-for="rack in racks" :key="rack.id" class="rack-card m3-card">
        <div class="rack-card-top">
          <div class="rack-tag-pill">
            <Tag class="w-3.5 h-3.5" />
            <span>#{{ rack.number }}</span>
          </div>

          <div class="rack-card-actions">
            <button @click="openEditRackModal(rack)" class="m3-btn-icon" title="Edit Rack">
              <Edit3 class="w-4 h-4 text-primary" />
            </button>
            <button @click="deleteRack(rack)" class="m3-btn-icon" title="Delete Rack">
              <Trash2 class="w-4 h-4 text-danger" />
            </button>
          </div>
        </div>

        <div class="rack-main-info">
          <h3 class="rack-title">{{ rack.name }}</h3>
          <p class="rack-desc">{{ rack.description || 'General item storage compartment' }}</p>
        </div>

        <div class="rack-footer">
          <div class="items-counter">
            <Package class="w-4 h-4 text-primary" />
            <span><strong>{{ rack.items_count || rack.parts_count || 0 }}</strong> Items Stored</span>
          </div>

          <!-- Print Label PDF Action -->
          <a
            :href="`/api/racks/${rack.id}/label-pdf`"
            target="_blank"
            class="m3-btn m3-btn-tonal m3-btn-sm print-label-btn"
            title="Generate & Print Rack Tag Label"
          >
            <Printer class="w-3.5 h-3.5" />
            <span>Print Label</span>
          </a>
        </div>
      </div>
    </div>

    <!-- Add / Edit Rack Modal -->
    <div v-if="rackModal.show" class="m3-modal-overlay" @click.self="rackModal.show = false">
      <div class="m3-modal-card p-6">
        <h3 class="text-lg font-bold mb-4">
          {{ rackModal.isEdit ? 'Edit Storage Rack' : 'Create New Storage Rack' }}
        </h3>

        <form @submit.prevent="saveRack">
          <div class="m3-input-group">
            <label class="m3-label">Rack Name *</label>
            <input
              v-model="rackForm.name"
              type="text"
              class="m3-input"
              placeholder="e.g. Rack A, Shelf B, Bin 3"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Rack Number / Section Code *</label>
            <input
              v-model="rackForm.number"
              type="text"
              class="m3-input"
              placeholder="e.g. A1, 04, TOP-1"
              required
            />
          </div>

          <div class="m3-input-group">
            <label class="m3-label">Description / Category (Optional)</label>
            <textarea
              v-model="rackForm.description"
              class="m3-textarea"
              placeholder="e.g. Braking and suspension components"
            ></textarea>
          </div>

          <div class="flex gap-3 justify-end mt-6">
            <button type="button" @click="rackModal.show = false" class="m3-btn m3-btn-outlined">
              Cancel
            </button>
            <button type="submit" class="m3-btn m3-btn-primary" :disabled="rackModal.saving">
              {{ rackModal.isEdit ? 'Save Changes' : 'Create Rack' }}
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
  Plus,
  Grid3X3,
  Tag,
  Package,
  Printer,
  Edit3,
  Trash2,
} from 'lucide-vue-next';

const racks = ref([]);
const loading = ref(true);

const rackModal = reactive({
  show: false,
  isEdit: false,
  saving: false,
  rackId: null,
});

const rackForm = reactive({
  name: '',
  number: '',
  description: '',
});

const loadRacks = async () => {
  loading.value = true;
  try {
    const res = await api.get('/racks');
    racks.value = res.data;
  } catch (err) {
    console.error('Failed to load racks:', err);
  } finally {
    loading.value = false;
  }
};

const openAddRackModal = () => {
  rackModal.isEdit = false;
  rackModal.rackId = null;
  rackForm.name = '';
  rackForm.number = '';
  rackForm.description = '';
  rackModal.show = true;
};

const openEditRackModal = (rack) => {
  rackModal.isEdit = true;
  rackModal.rackId = rack.id;
  rackForm.name = rack.name;
  rackForm.number = rack.number;
  rackForm.description = rack.description || '';
  rackModal.show = true;
};

const saveRack = async () => {
  rackModal.saving = true;
  try {
    if (rackModal.isEdit) {
      await api.put(`/racks/${rackModal.rackId}`, rackForm);
    } else {
      await api.post('/racks', rackForm);
    }
    rackModal.show = false;
    await loadRacks();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to save rack');
  } finally {
    rackModal.saving = false;
  }
};

const deleteRack = async (rack) => {
  if (!confirm(`Delete rack '${rack.name} - ${rack.number}'? (Assigned items will be unlinked)`)) return;
  try {
    await api.delete(`/racks/${rack.id}`);
    await loadRacks();
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to delete rack');
  }
};

onMounted(() => {
  loadRacks();
});
</script>

<style scoped>
.racks-page {
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

.racks-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.rack-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  border-left: 4px solid var(--md-sys-color-primary);
  min-height: 180px;
}

.rack-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rack-tag-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--md-sys-color-primary-container);
  color: var(--md-sys-color-on-primary-container);
  padding: 4px 10px;
  border-radius: var(--radius-full);
  font-family: var(--font-mono);
  font-weight: 700;
  font-size: 12px;
}

.rack-card-actions {
  display: flex;
  gap: 4px;
}

.rack-main-info {
  margin: 14px 0;
}

.rack-title {
  font-size: 17px;
  font-weight: 800;
  color: var(--md-sys-color-on-surface);
}

.rack-desc {
  font-size: 12px;
  color: var(--md-sys-color-outline);
  margin-top: 4px;
}

.rack-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 14px;
  border-top: 1px solid var(--md-sys-color-surface-variant);
}

.parts-counter {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--md-sys-color-on-surface-variant);
}

.print-label-btn {
  text-decoration: none;
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
.flex { display: flex; }
.gap-3 { gap: 12px; }
.justify-end { justify-content: flex-end; }
.mt-6 { margin-top: 24px; }
.text-primary { color: var(--md-sys-color-primary); }
.text-danger { color: var(--md-sys-color-error); }
</style>

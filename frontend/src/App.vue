<template>
  <div id="app">
    <router-view v-if="isLoginPage" />
    <AppLayout v-else>
      <router-view />
    </AppLayout>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import AppLayout from './components/AppLayout.vue';

const route = useRoute();
const isLoginPage = computed(() => route.name === 'Login');
</script>

<style>
#app {
  min-height: 100vh;
}
</style>

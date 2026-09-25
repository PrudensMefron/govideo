<script setup lang="ts">
import { provide } from "vue";
import { createGoVideo } from "./composables/useGoVideo";
import AppIcon from "./components/AppIcon.vue";
import HomeView from "./views/HomeView.vue";
import ActivitiesView from "./views/ActivitiesView.vue";
import MediaDialog from "./components/MediaDialog.vue";
import SettingsDialog from "./components/SettingsDialog.vue";
const state = createGoVideo();
provide("govideo", state);
const { page, jobs, active, notice, openSettings } = state;
</script>
<template>
  <div class="gv-shell">
    <header class="navbar bg-base-100 gv-navbar">
      <button class="btn btn-ghost gv-brand" @click="page = 'home'">
        <AppIcon name="play" />GoVideo
      </button>
      <nav class="join" aria-label="Principal">
        <button
          class="btn btn-ghost join-item"
          :class="{ 'btn-active': page === 'home' }"
          :aria-current="page === 'home' ? 'page' : undefined"
          @click="page = 'home'"
        >
          <AppIcon name="home" />Início
        </button>
        <button
          class="btn btn-ghost join-item"
          :class="{ 'btn-active': page === 'activity' }"
          :aria-current="page === 'activity' ? 'page' : undefined"
          @click="page = 'activity'"
        >
          <AppIcon name="activity" />Atividades
          <span v-if="jobs.some(active)" class="badge badge-primary badge-sm">{{
            jobs.filter(active).length
          }}</span>
        </button>
      </nav>
      <button
        class="btn btn-ghost btn-square"
        aria-label="Abrir configurações"
        @click="openSettings"
      >
        <AppIcon name="settings" />
      </button>
    </header>
    <div v-if="notice" class="alert alert-warning gv-notice" role="status">
      <AppIcon name="warning" /><span>{{ notice }}</span
      ><button
        class="btn btn-ghost btn-sm btn-square"
        aria-label="Fechar aviso"
        @click="notice = ''"
      >
        ×
      </button>
    </div>
    <HomeView v-if="page === 'home'" />
    <ActivitiesView v-else />
    <MediaDialog />
    <SettingsDialog />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import AppIcon from "./AppIcon.vue";
const props = defineProps<{ src: string }>();
const emit = defineEmits<{
  ready: [element: HTMLAudioElement | undefined];
  error: [];
}>();
const audio = ref<HTMLAudioElement>();
const playing = ref(false),
  loading = ref(true),
  failed = ref(false);
const current = ref(0),
  duration = ref(0),
  volume = ref(1);
const usable = computed(() => !failed.value && duration.value > 0);
function time(seconds: number) {
  if (!Number.isFinite(seconds)) return "0:00";
  return `${Math.floor(seconds / 60)}:${String(Math.floor(seconds % 60)).padStart(2, "0")}`;
}
function metadata() {
  const value = audio.value?.duration || 0;
  duration.value = Number.isFinite(value) ? value : 0;
  loading.value = false;
}
function error() {
  loading.value = false;
  failed.value = true;
  playing.value = false;
  emit("error");
}
async function toggle() {
  if (!audio.value) return;
  if (!audio.value.paused) {
    audio.value.pause();
    return;
  }
  try {
    await audio.value.play();
  } catch (e) {
    // Changing sources or leaving the editor legitimately interrupts play().
    if (!(e instanceof DOMException && e.name === "AbortError")) error();
  }
}
function seek(event: Event) {
  const value = Number((event.target as HTMLInputElement).value);
  if (audio.value && usable.value) {
    audio.value.currentTime = value;
    current.value = value;
  }
}
function changeVolume(event: Event) {
  volume.value = Number((event.target as HTMLInputElement).value);
  if (audio.value) audio.value.volume = volume.value;
}
watch(
  () => props.src,
  () => {
    playing.value = false;
    loading.value = true;
    failed.value = false;
    current.value = 0;
    duration.value = 0;
  },
);
onMounted(() => emit("ready", audio.value));
onBeforeUnmount(() => {
  audio.value?.pause();
  emit("ready", undefined);
});
</script>

<template>
  <div
    class="gv-audio-player"
    role="group"
    aria-label="Prévia da música recortada"
    :aria-busy="loading"
  >
    <audio
      ref="audio"
      :src="src"
      preload="metadata"
      @loadedmetadata="metadata"
      @durationchange="metadata"
      @canplay="loading = false"
      @waiting="loading = true"
      @playing="
        loading = false;
        playing = true;
      "
      @play="playing = true"
      @pause="playing = false"
      @ended="playing = false"
      @timeupdate="current = audio?.currentTime || 0"
      @error="error"
    />
    <button
      type="button"
      class="btn btn-primary btn-circle"
      :disabled="!usable"
      :aria-label="playing ? 'Pausar prévia' : 'Reproduzir prévia'"
      @click="toggle"
    >
      <span
        v-if="loading"
        class="loading loading-spinner loading-sm"
        aria-hidden="true"
      /><AppIcon v-else :name="playing ? 'pause' : 'play'" />
    </button>
    <div class="gv-audio-track">
      <div class="gv-audio-time">
        <span>{{
          failed
            ? "Prévia indisponível"
            : loading
              ? "Carregando áudio…"
              : playing
                ? "Reproduzindo"
                : "Prévia do recorte"
        }}</span
        ><span>{{ time(current) }} / {{ time(duration) }}</span>
      </div>
      <input
        type="range"
        class="range range-primary range-xs"
        min="0"
        :max="duration || 1"
        step="0.1"
        :value="current"
        :disabled="!usable"
        aria-label="Posição na música"
        :aria-valuetext="`${time(current)} de ${time(duration)}`"
        @input="seek"
      />
    </div>
    <label class="gv-audio-volume"
      ><span>Volume</span
      ><input
        type="range"
        class="range range-primary range-xs"
        min="0"
        max="1"
        step="0.05"
        :value="volume"
        :aria-valuetext="`${Math.round(volume * 100)}%`"
        @input="changeVolume"
    /></label>
  </div>
</template>

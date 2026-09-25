import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { CancellablePromise } from "@wailsio/runtime";
import * as api from "../../bindings/github.com/PrudensMefron/govideo/desktopservice";
import type { AudioFile } from "../../bindings/github.com/PrudensMefron/govideo/internal/core/models";
import type { SavedAudio } from "../../bindings/github.com/PrudensMefron/govideo/internal/app/models";
import type { AudioPreviewDTO } from "../../bindings/github.com/PrudensMefron/govideo/models";

export function useAudioTrim() {
  const library = ref<AudioFile[]>([]),
    loadingLibrary = ref(true);
  const source = ref<AudioFile | null>(null),
    selectedPath = ref("");
  const startSeconds = ref<number | string>(0),
    endSeconds = ref<number | string>(0);
  const preview = ref<AudioPreviewDTO | null>(null),
    saved = ref<SavedAudio | null>(null);
  const busy = ref<"inspect" | "preview" | "save" | null>(null),
    error = ref("");
  const player = ref<HTMLAudioElement>(),
    playbackError = ref("");
  let pending: CancellablePromise<unknown> | undefined;
  let cleanup: Promise<void> = Promise.resolve();
  let disposed = false,
    cancelled = false;

  const remaining = computed(
    () =>
      (source.value?.durationSeconds || 0) -
      Number(startSeconds.value) -
      Number(endSeconds.value),
  );
  const validation = computed(() => {
    if (!source.value) return "Selecione uma música para começar.";
    const values = [startSeconds.value, endSeconds.value];
    if (
      values.some(
        (v) => v === "" || !Number.isFinite(Number(v)) || Number(v) < 0,
      )
    )
      return "Informe segundos válidos e não negativos nos dois campos.";
    if (values.every((v) => Number(v) === 0))
      return "Informe quantos segundos cortar do início ou do final.";
    if (remaining.value < 0.1)
      return "Os cortes precisam deixar pelo menos 0,1 segundo de música.";
    return "";
  });
  const previewURL = computed(() =>
    preview.value?.playbackURL || "",
  );
  const formatTime = (seconds: number) => {
    if (!Number.isFinite(seconds) || seconds < 0) return "—";
    const tenths = Math.round(seconds * 10);
    return `${Math.floor(tenths / 600)}:${String(Math.floor(tenths / 10) % 60).padStart(2, "0")}${tenths % 10 ? `,${tenths % 10}` : ""}`;
  };

  function releasePlayer() {
    player.value?.pause();
    player.value?.removeAttribute("src");
    player.value?.load();
  }
  function clearPreview() {
    releasePlayer();
    const old = preview.value;
    preview.value = null;
    playbackError.value = "";
    if (old) cleanup = cleanup.then(async () => { await api.DiscardAudioPreview(old.id); }).catch(() => {});
  }
  watch([startSeconds, endSeconds], () => {
    clearPreview();
    saved.value = null;
    error.value = "";
  });

  async function loadLibrary() {
    loadingLibrary.value = true;
    try {
      library.value = (await api.ListAudioArtifacts()) || [];
    } catch (e) {
      error.value = String(e);
    } finally {
      loadingLibrary.value = false;
    }
  }

  async function selectSource(path: string) {
    if (busy.value || !path) return;
    clearPreview();
    saved.value = null;
    source.value = null;
    selectedPath.value = path;
    startSeconds.value = endSeconds.value = 0;
    busy.value = "inspect";
    error.value = "";
    try {
      pending = api.InspectAudio(path);
      const file = (await pending) as AudioFile;
      if (!disposed) source.value = file;
    } catch (e) {
      if (!disposed) error.value = String(e);
    } finally {
      pending = undefined;
      busy.value = null;
    }
  }
  async function chooseFile() {
    if (busy.value) return;
    try {
      const path = await api.SelectInputAudio();
      if (path && !disposed) await selectSource(path);
    } catch (e) {
      error.value = String(e);
    }
  }
  async function generatePreview() {
    if (busy.value) return;
    if (validation.value) {
      error.value = validation.value;
      return;
    }
    clearPreview();
    error.value = "";
    saved.value = null;
    cancelled = false;
    busy.value = "preview";
    try {
      // Discard and render share a backend operation lock. Never race them.
      await cleanup;
      if (disposed || cancelled) return;
      pending = api.CreateAudioPreview({
        path: source.value!.path,
        startSeconds: Number(startSeconds.value),
        endSeconds: Number(endSeconds.value),
      });
      const result = (await pending) as AudioPreviewDTO;
      if (disposed || cancelled) {
        await api.DiscardAudioPreview(result.id);
        return;
      }
      preview.value = result;
    } catch (e) {
      if (!disposed && !cancelled) error.value = String(e);
    } finally {
      pending = undefined;
      busy.value = null;
    }
  }
  async function cancelPreview() {
    cancelled = true;
    try {
      await pending?.cancel();
    } catch {
      /* Original call reports cancellation. */
    }
  }
  async function savePreview(overwrite: boolean) {
    if (!preview.value || busy.value) return;
    busy.value = "save";
    error.value = "";
    releasePlayer();
    try {
      // Saving publishes an already prepared file; do not cancel a filesystem commit on navigation.
      const result = await api.SaveAudioPreview({
        previewID: preview.value.id,
        overwrite,
      });
      saved.value = result;
      preview.value = null;
      if (overwrite) {
        source.value = null; // Reinspect before cutting the replaced source again.
        selectedPath.value = "";
      }
      await loadLibrary();
    } catch (e) {
      error.value = String(e);
      if (player.value && previewURL.value) {
        player.value.src = previewURL.value;
        player.value.load();
      }
    } finally {
      busy.value = null;
    }
  }
  async function revealSaved() {
    if (!saved.value) return;
    try {
      await api.OpenArtifactLocation(saved.value.path);
    } catch (e) {
      error.value = String(e);
    }
  }
  async function playSaved() {
    if (!saved.value) return;
    try {
      await api.OpenArtifact(saved.value.path);
    } catch (e) {
      error.value = String(e);
    }
  }
  onMounted(loadLibrary);
  onBeforeUnmount(() => {
    disposed = true;
    if (pending) void Promise.resolve(pending.cancel()).catch(() => {});
    clearPreview();
  });
  return {
    library,
    loadingLibrary,
    source,
    selectedPath,
    startSeconds,
    endSeconds,
    preview,
    saved,
    busy,
    error,
    player,
    playbackError,
    remaining,
    validation,
    previewURL,
    formatTime,
    selectSource,
    chooseFile,
    generatePreview,
    cancelPreview,
    savePreview,
    revealSaved,
    playSaved,
  };
}

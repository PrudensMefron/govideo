import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  inject,
} from "vue";
import { Events } from "@wailsio/runtime";
import * as api from "../../bindings/github.com/PrudensMefron/govideo/desktopservice";
import {
  AudioFormat,
  AudioQuality,
  Media as CoreMedia,
  OutputMode,
} from "../../bindings/github.com/PrudensMefron/govideo/internal/core/models";
import type { Status as UpdateStatus } from "../../bindings/github.com/PrudensMefron/govideo/internal/desktop/update/models";

type Media = CoreMedia;
type Job = {
  id: string;
  kind: string;
  inputPath?: string;
  media?: Media;
  mode: string;
  status: string;
  progress: {
    phase: string;
    fraction?: number | null;
    completedBytes?: number;
    totalBytes?: number;
    speedBytesPerSecond?: number;
    etaSeconds?: number;
  };
  artifacts?: { path: string; available: boolean }[];
  failure?: { message: string };
  createdAt: string;
  updatedAt?: string;
};
type Settings = {
  videoDirectory: string;
  musicDirectory: string;
  conversionDirectory: string;
  theme: string;
};
type Dep = {
  name: string;
  state: string;
  version?: string;
  path?: string;
  error?: string;
  license?: string;
};
export function createGoVideo() {
  const page = ref<"home" | "activity">("home"),
    mode = ref<"download" | "convert">("download"),
    url = ref(""),
    busy = ref(false),
    error = ref(""),
    notice = ref(""),
    media = ref<Media | null>(null),
    jobs = ref<Job[]>([]),
    deps = ref<Dep[]>([]),
    files = ref<string[]>([]),
    settings = ref<Settings>({
      videoDirectory: "",
      musicDirectory: "",
      conversionDirectory: "",
      theme: "system",
    }),
    update = ref<UpdateStatus | null>(null),
    checkingUpdate = ref(false),
    updateError = ref(""),
    output = ref(OutputMode.OutputVideo),
    height = ref(0),
    audioFormat = ref(AudioFormat.AudioMP3),
    audioQuality = ref(AudioQuality.AudioStandard),
    mediaDialog = ref<HTMLDialogElement>(),
    settingsDialog = ref<HTMLDialogElement>(),
    mediaTrigger = ref<HTMLElement | null>(null),
    settingsTrigger = ref<HTMLElement | null>(null),
    jobAction = ref<string | null>(null);
  const savingSettings = ref(false),
    settingsError = ref("");
  let settingsBeforeEdit: Settings | null = null;
  const active = (j: Job) =>
    ["pending", "analyzing", "ready", "downloading", "processing"].includes(
      j.status,
    );
  const qualities = computed(() =>
    [
      0,
      ...new Set(
        (media.value?.formats || [])
          .filter((f) => f.hasVideo && f.height > 0 && f.height <= 2160)
          .map((f) => f.height),
      ),
    ].sort((a, b) => b - a),
  );
  const phases: Record<string, string> = {
    preparing_dependencies: "Preparando dependências",
    queued: "Na fila",
    analyzing: "Analisando",
    downloading: "Baixando",
    merging: "Combinando faixas",
    extracting_audio: "Extraindo áudio",
    transcoding: "Convertendo",
    finalizing: "Finalizando",
  };
  const phase = (j: Job) => phases[j.progress?.phase] || j.status,
    title = (j: Job) =>
      j.media?.title || j.inputPath?.split(/[\\/]/).pop() || "Mídia";
  const bytes = (n = 0) =>
    n
      ? `${(n / (n >= 1e9 ? 1e9 : 1e6)).toFixed(1)} ${n >= 1e9 ? "GB" : "MB"}`
      : "—";
  const time = (ns = 0) => {
    const s = Math.round(ns / 1e9);
    return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;
  };
  const eta = (seconds = 0) =>
    seconds > 0
      ? `${Math.floor(seconds / 60)} min ${String(seconds % 60).padStart(2, "0")} s`
      : "Calculando…";
  const progressValue = (j: Job) =>
    j.status === "completed"
      ? 1
      : Number.isFinite(j.progress?.fraction)
        ? Math.max(0, Math.min(1, j.progress.fraction!))
        : null;
  const progressText = (j: Job) =>
    progressValue(j) === null
      ? "Progresso em andamento"
      : `${Math.round((progressValue(j) || 0) * 100)}% concluído`;
  const date = (value: string) =>
    new Intl.DateTimeFormat("pt-BR", {
      hour: "2-digit",
      minute: "2-digit",
    }).format(new Date(value));
  async function refresh() {
    try {
      jobs.value = (await api.ListJobs(0, 50)) as Job[];
      deps.value = (await api.ListDependencies()) as Dep[];
      settings.value = (await api.GetSettings()) as Settings;
      update.value = (await api.GetUpdateStatus()) as UpdateStatus;
    } catch {
      notice.value =
        "Não foi possível atualizar os dados agora. Tente novamente.";
    }
  }
  async function refreshUpdate() {
    try {
      update.value = (await api.GetUpdateStatus()) as UpdateStatus;
    } catch {}
  }
  async function checkUpdates() {
    checkingUpdate.value = true;
    updateError.value = "";
    try {
      update.value = (await api.CheckForUpdates()) as UpdateStatus;
    } catch (e) {
      updateError.value = String(e);
      await refreshUpdate();
    } finally {
      checkingUpdate.value = false;
    }
  }
  function openSettings(trigger: Event) {
    settingsBeforeEdit = { ...settings.value };
    settingsError.value = "";
    settingsTrigger.value = trigger.currentTarget as HTMLElement;
    settingsDialog.value?.showModal();
  }
  function restoreFocus(target: "media" | "settings") {
    if (target === "settings" && settingsBeforeEdit) {
      settings.value = settingsBeforeEdit;
      settingsBeforeEdit = null;
    }
    nextTick(() =>
      target === "media"
        ? mediaTrigger.value?.focus()
        : settingsTrigger.value?.focus(),
    );
  }
  async function analyze(e: Event) {
    mediaTrigger.value = (e.currentTarget as HTMLElement).querySelector(
      "button",
    );
    error.value = "";
    if (!url.value.trim()) {
      error.value = "Cole um endereço válido para analisar.";
      return;
    }
    busy.value = true;
    try {
      media.value = (await api.AnalyzeURL({ url: url.value })) as Media;
      if (media.value.drm) {
        error.value = "Esta mídia possui proteção DRM e não pode ser baixada.";
        return;
      }
      mediaDialog.value?.showModal();
    } catch (e) {
      error.value = String(e);
    } finally {
      busy.value = false;
    }
  }
  async function download() {
    if (!media.value) return;
    error.value = "";
    busy.value = true;
    try {
      await api.StartDownload({
        url: url.value,
        media: media.value,
        mode: output.value,
        maxHeight: height.value,
        audioFormat: audioFormat.value,
        audioQuality: audioQuality.value,
        destination:
          output.value === "audio"
            ? settings.value.musicDirectory
            : settings.value.videoDirectory,
      });
      mediaDialog.value?.close();
      page.value = "activity";
      await refresh();
    } catch (e) {
      error.value = String(e);
    } finally {
      busy.value = false;
    }
  }
  async function choose() {
    try {
      files.value = [...new Set((await api.SelectInputVideos()) || [])];
    } catch (e) {
      error.value = String(e);
    }
  }
  async function convert() {
    busy.value = true;
    try {
      await api.StartLocalConversions({
        paths: files.value,
        audioFormat: audioFormat.value,
        audioQuality: audioQuality.value,
        destination: settings.value.conversionDirectory,
      });
      files.value = [];
      page.value = "activity";
      await refresh();
    } catch (e) {
      error.value = String(e);
    } finally {
      busy.value = false;
    }
  }
  async function folder(k: keyof Settings) {
    const p = await api.SelectDestination(String(k));
    if (p) settings.value = { ...settings.value, [k]: p };
  }
  async function save() {
    if (savingSettings.value) return;
    savingSettings.value = true;
    settingsError.value = "";
    try {
      settings.value = await api.UpdateSettings(settings.value);
      settingsBeforeEdit = null;
      applyTheme();
      settingsDialog.value?.close();
    } catch (e) {
      settingsError.value = String(e);
    } finally {
      savingSettings.value = false;
    }
  }
  async function runJobAction(kind: "cancel" | "retry", id: string) {
    jobAction.value = `${kind}:${id}`;
    try {
      kind === "cancel" ? await api.CancelJob(id) : await api.RetryJob(id);
      await refresh();
    } catch (e) {
      notice.value = String(e);
    } finally {
      jobAction.value = null;
    }
  }
  function applyTheme() {
    document.documentElement.dataset.theme =
      settings.value.theme === "system"
        ? matchMedia("(prefers-color-scheme: dark)").matches
          ? "govideo-dark"
          : "govideo-light"
        : `govideo-${settings.value.theme}`;
    try {
      localStorage.setItem("govideo-theme", settings.value.theme);
    } catch {}
  }
  let off: (() => void)[] = [];
  onMounted(async () => {
    await refresh();
    applyTheme();
    off = [
      Events.On("job:updated", (e) => {
        const j = e.data as Job,
          i = jobs.value.findIndex((x) => x.id === j.id);
        i < 0 ? jobs.value.unshift(j) : (jobs.value[i] = j);
      }),
      Events.On("dependency:updated", (e) => {
        const d = e.data as Dep,
          i = deps.value.findIndex((x) => x.name === d.name);
        i < 0 ? deps.value.push(d) : (deps.value[i] = d);
      }),
      Events.On("update:updated", (e) => {
        update.value = e.data as UpdateStatus;
      }),
      Events.On("wails:updater:update-available", () => refreshUpdate()),
      Events.On("wails:updater:no-update", () => refreshUpdate()),
      Events.On("wails:updater:error", () => refreshUpdate()),
      Events.On("files:dropped", (e) => {
        files.value = [...new Set([...files.value, ...(e.data as string[])])];
        mode.value = "convert";
      }),
    ];
  });
  onBeforeUnmount(() => off.forEach((f) => f()));

  return {
    page,
    mode,
    url,
    busy,
    error,
    notice,
    media,
    jobs,
    deps,
    files,
    settings,
    update,
    checkingUpdate,
    updateError,
    output,
    height,
    audioFormat,
    audioQuality,
    mediaDialog,
    settingsDialog,
    jobAction,
    active,
    qualities,
    phase,
    title,
    bytes,
    time,
    eta,
    progressValue,
    progressText,
    date,
    analyze,
    download,
    choose,
    convert,
    folder,
    save,
    runJobAction,
    openSettings,
    restoreFocus,
    checkUpdates,
    savingSettings,
    settingsError,
  };
}
export type GoVideoState = ReturnType<typeof createGoVideo>;
export function useGoVideo() {
  const state = inject<GoVideoState>("govideo");
  if (!state) throw new Error("GoVideo provider missing");
  return state;
}

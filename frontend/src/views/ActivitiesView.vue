<script setup lang="ts">
import { nextTick, ref } from "vue";
import AppIcon from "../components/AppIcon.vue";
import { useGoVideo } from "../composables/useGoVideo";
const {
  page,
  jobs,
  jobAction,
  active,
  phase,
  title,
  bytes,
  eta,
  progressValue,
  progressText,
  date,
  runJobAction,
  notice,
} = useGoVideo();
import * as api from "../../bindings/github.com/PrudensMefron/govideo/desktopservice";
type Activity = (typeof jobs.value)[number];
const filesDialog = ref<HTMLDialogElement>();
const choices = ref<NonNullable<Activity["artifacts"]>>([]);
const openingPath = ref("");
const activityTitle = ref<HTMLElement>();
let trigger: HTMLElement | null = null;
const outputs = (j: Activity) =>
  j.status === "completed"
    ? (j.artifacts || []).filter((a) => a.available)
    : [];
async function openFile(path: string) {
  if (openingPath.value) return;
  openingPath.value = path;
  try {
    await api.OpenArtifact(path);
    filesDialog.value?.close();
  } catch (e) {
    filesDialog.value?.close();
    notice.value = String(e);
  } finally {
    openingPath.value = "";
  }
}
function openJob(j: Activity, event: MouseEvent) {
  const files = outputs(j);
  if (!files.length || openingPath.value) return;
  if (files.length === 1) {
    void openFile(files[0].path);
    return;
  }
  trigger =
    (event.currentTarget as HTMLElement).querySelector<HTMLButtonElement>(
      ".gv-open-media",
    ) || (event.currentTarget as HTMLElement);
  choices.value = files;
  filesDialog.value?.showModal();
}
function rowClick(j: Activity, event: MouseEvent) {
  if (
    (event.target as HTMLElement).closest("button, a, input, select") ||
    window.getSelection()?.toString()
  )
    return;
  openJob(j, event);
}
async function reveal(path: string) {
  try {
    await api.OpenArtifactLocation(path);
  } catch (e) {
    notice.value = String(e);
  }
}
async function removeUnavailable(id: string) {
  if (await runJobAction("prune", id)) {
    await nextTick();
    activityTitle.value?.focus();
  }
}
</script>
<template>
  <main class="activities" aria-labelledby="activity-title">
    <div class="title">
      <div>
        <span class="eyebrow">CENTRAL DE ATIVIDADES</span>
        <h1 id="activity-title" ref="activityTitle" tabindex="-1">
          Acompanhe o que está acontecendo.
        </h1>
        <p>
          Fila, download, processamento e arquivos concluídos em um só lugar.
        </p>
      </div>
      <span class="badge badge-outline"
        >{{ jobs.length }} {{ jobs.length === 1 ? "job" : "jobs" }}</span
      >
    </div>
    <div v-if="!jobs.length" class="card bg-base-100 empty">
      <i><AppIcon name="activity" /></i><b>Nenhuma atividade por enquanto</b
      ><span
        >Seus downloads e conversões aparecerão aqui assim que começarem.</span
      ><button class="btn btn-outline" @click="page = 'home'">
        Iniciar uma tarefa
      </button>
    </div>
    <ul v-else class="list bg-base-100 jobs" aria-label="Lista de atividades">
      <li
        class="list-row gv-job-row"
        v-for="j in jobs"
        :key="j.id"
        :class="[`job-${j.status}`, { 'gv-playable': outputs(j).length > 0 }]"
        @click="rowClick(j, $event)"
      >
        <i class="jobicon"
          ><AppIcon :name="j.kind === 'download' ? 'download' : 'audio'"
        /></i>
        <div class="info">
          <button
            v-if="outputs(j).length"
            class="job-name gv-open-media"
            :title="`Abrir ${title(j)} no aplicativo padrão`"
            :aria-label="`Abrir ${title(j)}`"
            :disabled="!!openingPath"
            @click.stop="openJob(j, $event)"
          >
            <AppIcon name="play" /><span>{{ title(j) }}</span>
          </button>
          <b v-else class="job-name" :title="title(j)">{{ title(j) }}</b>
          <span
            >{{
              j.kind === "download"
                ? "Download online"
                : j.kind === "audio_trim"
                  ? "Recorte de áudio"
                  : "Conversão local"
            }}
            · iniciado {{ date(j.createdAt) }}</span
          ><small v-if="j.failure" class="text-error"
            ><AppIcon name="warning" />{{ j.failure.message }}</small
          >
          <small
            v-if="
              j.status === 'completed' &&
              j.artifacts?.some((a) => !a.available)
            "
            class="text-warning"
            ><AppIcon name="warning" />{{
              outputs(j).length
                ? "Parte dos arquivos está indisponível"
                : "Arquivo indisponível"
            }}</small
          >
        </div>
        <div class="gv-job-progress">
          <div
            v-if="active(j) || j.status === 'completed'"
            class="radial-progress job-radial"
            :class="{
              complete: j.status === 'completed',
              indeterminate: progressValue(j) === null && active(j),
            }"
            :style="{
              '--value':
                progressValue(j) === null && active(j)
                  ? 25
                  : Math.round((progressValue(j) || 0) * 100),
              '--size': '3.15rem',
              '--thickness': '4px',
            }"
            role="progressbar"
            :aria-label="`${title(j)}: ${progressText(j)}`"
            :aria-valuemin="0"
            :aria-valuemax="100"
            :aria-valuenow="
              progressValue(j) === null
                ? undefined
                : Math.round((progressValue(j) || 0) * 100)
            "
          >
            <span>{{
              progressValue(j) === null
                ? "…"
                : `${Math.round((progressValue(j) || 0) * 100)}%`
            }}</span>
          </div>
          <span
            v-else
            class="jobicon"
            :class="
              j.status === 'failed' ? 'text-error' : 'text-base-content/60'
            "
            aria-hidden="true"
            ><AppIcon :name="j.status === 'failed' ? 'warning' : 'close'"
          /></span>
          <div class="gv-job-details">
            <div>
              <span class="badge badge-outline gv-status" :class="j.status">{{
                j.status === "completed"
                  ? "Concluído"
                  : j.status === "failed"
                    ? "Falhou"
                    : j.status === "cancelled"
                      ? "Cancelado"
                      : phase(j)
              }}</span>
            </div>
            <small
              v-if="
                active(j) &&
                (j.progress?.speedBytesPerSecond || j.progress?.etaSeconds)
              "
              ><span v-if="j.progress?.speedBytesPerSecond"
                >{{ bytes(j.progress.speedBytesPerSecond) }}/s</span
              ><span v-if="j.progress?.etaSeconds">
                · cerca de {{ eta(j.progress.etaSeconds) }}</span
              ></small
            ><small v-else-if="active(j)"
              >O tempo restante aparece quando estiver disponível.</small
            >
          </div>
        </div>
        <div class="actions">
          <button
            v-if="active(j)"
            class="btn btn-ghost btn-sm text-error"
            :disabled="jobAction === `cancel:${j.id}`"
            @click="runJobAction('cancel', j.id)"
          >
            {{
              jobAction === `cancel:${j.id}` ? "Cancelando…" : "Cancelar"
            }}</button
          ><button
            v-if="['failed', 'cancelled'].includes(j.status)"
            class="btn btn-outline btn-sm"
            :disabled="jobAction === `retry:${j.id}`"
            @click="runJobAction('retry', j.id)"
          >
            <AppIcon name="refresh" />{{
              jobAction === `retry:${j.id}` ? "Preparando…" : "Tentar novamente"
            }}</button
          ><button
            v-for="a in j.artifacts?.filter((x) => x.available)"
            :key="a.path"
            class="btn btn-outline btn-sm"
            @click.stop="reveal(a.path)"
            :title="a.path"
          >
            <AppIcon name="folder" />{{
              (j.artifacts?.length || 0) > 1
                ? `Pasta · ${a.kind === "audio" ? "música" : "vídeo"}`
                : "Abrir pasta"
            }}
          </button>
          <button
            v-if="
              j.status === 'completed' &&
              j.artifacts?.some((a) => !a.available)
            "
            class="btn btn-ghost btn-sm text-error"
            :disabled="!!jobAction"
            @click.stop="removeUnavailable(j.id)"
          >
            <AppIcon name="close" />{{
              jobAction === `prune:${j.id}`
                ? "Removendo…"
                : outputs(j).length
                  ? "Limpar ausentes"
                  : "Remover da lista"
            }}
          </button>
        </div>
      </li>
    </ul>
    <dialog
      ref="filesDialog"
      class="modal"
      aria-labelledby="artifact-choice-title"
      @close="trigger?.focus()"
    >
      <div class="modal-box">
        <h2 id="artifact-choice-title">Qual arquivo você quer abrir?</h2>
        <p class="my-3">
          Esta atividade gerou mais de um arquivo. Escolha para abrir no
          aplicativo padrão.
        </p>
        <div class="gv-artifact-choices">
          <button
            v-for="a in choices"
            :key="a.path"
            class="btn btn-outline"
            :disabled="!!openingPath"
            @click="openFile(a.path)"
          >
            <AppIcon :name="a.kind === 'audio' ? 'audio' : 'video'" /><span
              >{{ a.kind === "audio" ? "Música" : "Vídeo" }} ·
              {{ a.path.split(/[\\/]/).pop() }}</span
            >
          </button>
        </div>
        <div class="modal-action">
          <button class="btn btn-ghost" autofocus @click="filesDialog?.close()">
            Fechar
          </button>
        </div>
      </div>
    </dialog>
  </main>
</template>

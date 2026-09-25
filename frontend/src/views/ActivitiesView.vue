<script setup lang="ts">
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
} = useGoVideo();
import * as api from "../../bindings/github.com/PrudensMefron/govideo/desktopservice";
</script>
<template>
  <main class="activities" aria-labelledby="activity-title">
    <div class="title">
      <div>
        <span class="eyebrow">CENTRAL DE ATIVIDADES</span>
        <h1 id="activity-title">Acompanhe o que está acontecendo.</h1>
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
        :class="`job-${j.status}`"
      >
        <i class="jobicon"
          ><AppIcon :name="j.kind === 'download' ? 'download' : 'audio'"
        /></i>
        <div class="info">
          <b class="job-name" :title="title(j)">{{ title(j) }}</b
          ><span
            >{{
              j.kind === "download" ? "Download online" : "Conversão local"
            }}
            · iniciado {{ date(j.createdAt) }}</span
          ><small v-if="j.failure" class="text-error"
            ><AppIcon name="warning" />{{ j.failure.message }}</small
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
            @click="api.OpenArtifactLocation(a.path)"
          >
            <AppIcon name="folder" />Abrir pasta
          </button>
        </div>
      </li>
    </ul>
  </main>
</template>

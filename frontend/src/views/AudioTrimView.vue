<script setup lang="ts">
import { ref } from "vue";
import AppIcon from "../components/AppIcon.vue";
import { useAudioTrim } from "../composables/useAudioTrim";
const {
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
} = useAudioTrim();
const overwriteDialog = ref<HTMLDialogElement>();
const overwriteTrigger = ref<HTMLButtonElement>();
async function confirmOverwrite() {
  overwriteDialog.value?.close();
  await savePreview(true);
}
</script>

<template>
  <section class="workspace gv-trim" aria-labelledby="trim-title">
    <span class="eyebrow">RECORTE DE ÁUDIO</span>
    <h1 id="trim-title">Fique só com a parte que importa.</h1>
    <p>
      Corte a introdução ou o final da música. Ouça a prévia antes de salvar.
    </p>

    <div class="card bg-base-100 gv-trim-panel">
      <h2>Escolha uma música</h2>
      <div class="gv-trim-source">
        <label for="trim-library"
          >Músicas em Atividades
          <select
            id="trim-library"
            class="select w-full"
            :value="selectedPath"
            :disabled="!!busy || loadingLibrary || !library.length"
            @change="selectSource(($event.target as HTMLSelectElement).value)"
          >
            <option value="" disabled>
              {{
                loadingLibrary
                  ? "Carregando músicas…"
                  : library.length
                    ? "Selecione uma música"
                    : "Nenhuma música salva por aqui"
              }}
            </option>
            <option
              v-if="
                selectedPath && !library.some((f) => f.path === selectedPath)
              "
              :value="selectedPath"
            >
              {{ selectedPath.split(/[\\/]/).pop() }}
            </option>
            <option v-for="file in library" :key="file.path" :value="file.path">
              {{ file.name }}
            </option>
          </select>
        </label>
        <span class="gv-trim-or">ou</span>
        <button class="btn btn-outline" :disabled="!!busy" @click="chooseFile">
          <AppIcon name="folder" />Escolher do computador
        </button>
      </div>
      <p class="gv-trim-help">MP3, M4A, Opus, WAV, FLAC, Ogg e AAC.</p>
      <p v-if="busy === 'inspect'" class="gv-trim-loading" role="status">
        <span class="loading loading-spinner loading-sm" />Verificando o arquivo
        e sua duração…
      </p>
    </div>

    <div v-if="error" class="alert alert-error" role="alert">
      <AppIcon name="warning" /><span>{{ error }}</span>
    </div>

    <div v-if="source" class="card bg-base-100 gv-trim-panel">
      <div class="gv-trim-heading">
        <div class="min-w-0">
          <h2 class="gv-trim-filename">{{ source.name }}</h2>
          <p class="gv-trim-help">
            {{ source.format.toUpperCase() }} · Original
            {{ formatTime(source.durationSeconds) }}
          </p>
        </div>
        <AppIcon name="audio" />
      </div>
      <form @submit.prevent="generatePreview" novalidate>
        <fieldset :disabled="!!busy" class="gv-trim-fields">
          <legend class="sr">Definir cortes em segundos</legend>
          <label for="trim-start"
            >Cortar do início
            <div class="gv-trim-input">
              <input
                id="trim-start"
                v-model="startSeconds"
                type="number"
                min="0"
                :max="source.durationSeconds"
                step="0.1"
                inputmode="decimal"
                class="input w-full"
                aria-describedby="trim-start-help"
              /><span>segundos</span>
            </div>
            <small id="trim-start-help"
              >Contados a partir do começo da música.</small
            >
          </label>
          <label for="trim-end"
            >Cortar do final
            <div class="gv-trim-input">
              <input
                id="trim-end"
                v-model="endSeconds"
                type="number"
                min="0"
                :max="source.durationSeconds"
                step="0.1"
                inputmode="decimal"
                class="input w-full"
                aria-describedby="trim-end-help"
              /><span>segundos</span>
            </div>
            <small id="trim-end-help"
              >Contados de trás para frente, a partir do fim.</small
            >
          </label>
        </fieldset>
        <div class="gv-trim-interval" aria-live="polite">
          <template v-if="!validation"
            ><span
              >Trecho mantido
              <strong
                >{{ formatTime(Number(startSeconds)) }} →
                {{
                  formatTime(source.durationSeconds - Number(endSeconds))
                }}</strong
              ></span
            ><span
              >Duração final <strong>{{ formatTime(remaining) }}</strong></span
            ></template
          >
          <span v-else>{{ validation }}</span>
        </div>
        <div class="gv-trim-toolbar">
          <button class="btn btn-primary" type="submit" :disabled="!!busy">
            <span
              v-if="busy === 'preview'"
              class="loading loading-spinner loading-sm"
            /><AppIcon v-else name="scissors" />{{
              busy === "preview"
                ? "Gerando prévia…"
                : preview
                  ? "Gerar novamente"
                  : "Gerar prévia"
            }}
          </button>
          <button
            v-if="busy === 'preview'"
            class="btn btn-ghost"
            type="button"
            @click="cancelPreview"
          >
            Cancelar
          </button>
          <span class="gv-trim-help" role="status">{{
            busy === "preview"
              ? "Preparando o áudio para você ouvir."
              : "O original permanece intacto durante o teste."
          }}</span>
        </div>
      </form>
    </div>

    <div v-if="preview" class="card bg-base-100 gv-trim-panel">
      <div class="gv-trim-heading">
        <h2>Ouça o resultado</h2>
        <span class="badge badge-success badge-outline">{{
          formatTime(preview.durationSeconds)
        }}</span>
      </div>
      <p class="gv-trim-help">
        Use o player para conferir o começo e avançar até o final.
      </p>
      <audio
        ref="player"
        :key="preview.id"
        :src="previewURL"
        controls
        preload="metadata"
        aria-label="Prévia da música recortada"
        @error="
          playbackError =
            'Não foi possível reproduzir a prévia. Gere-a novamente para tentar outra vez.'
        "
      ></audio>
      <p v-if="playbackError" class="text-error" role="alert">
        {{ playbackError }}
      </p>
      <p class="gv-trim-help">
        A cópia será salva na pasta da música, com “- recortada” no nome. O
        formato é mantido; áudio comprimido é recodificado para aplicar o corte.
      </p>
      <div class="gv-trim-toolbar">
        <button
          class="btn btn-primary"
          :disabled="!!busy"
          @click="savePreview(false)"
        >
          <span
            v-if="busy === 'save'"
            class="loading loading-spinner loading-sm"
          /><AppIcon v-else name="check" />{{
            busy === "save" ? "Salvando…" : "Salvar novo arquivo"
          }}
        </button>
        <button
          ref="overwriteTrigger"
          class="btn btn-outline"
          :disabled="!!busy"
          @click="overwriteDialog?.showModal()"
        >
          Substituir original…
        </button>
      </div>
    </div>

    <div v-if="saved" class="alert alert-success gv-trim-saved" role="status">
      <AppIcon name="check" />
      <div>
        <strong>{{
          saved.overwritten ? "Original substituído." : "Nova música salva."
        }}</strong>
        <p>{{ saved.path }}</p>
      </div>
      <div class="gv-trim-toolbar">
        <button class="btn btn-sm btn-outline" @click="playSaved">
          <AppIcon name="play" />Reproduzir</button
        ><button class="btn btn-sm btn-outline" @click="revealSaved">
          <AppIcon name="folder" />Abrir pasta
        </button>
      </div>
    </div>

    <dialog
      ref="overwriteDialog"
      class="modal"
      aria-labelledby="overwrite-title"
      @close="overwriteTrigger?.focus()"
    >
      <div class="modal-box">
        <h2 id="overwrite-title">Substituir o arquivo original?</h2>
        <p class="my-4 break-words">
          {{ source?.name }} será substituído pela versão que você ouviu. Os
          trechos removidos não poderão ser recuperados por aqui.
        </p>
        <div class="modal-action flex-wrap">
          <button
            class="btn btn-outline"
            autofocus
            @click="overwriteDialog?.close()"
          >
            Voltar</button
          ><button class="btn btn-error" @click="confirmOverwrite">
            Substituir original
          </button>
        </div>
      </div>
    </dialog>
  </section>
</template>

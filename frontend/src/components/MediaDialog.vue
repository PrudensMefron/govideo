<script setup lang="ts">
import AppIcon from "../components/AppIcon.vue";
import { useGoVideo } from "../composables/useGoVideo";
const {
  busy,
  error,
  media,
  output,
  height,
  audioFormat,
  audioQuality,
  mediaDialog,
  qualities,
  bytes,
  time,
  download,
  restoreFocus,
} = useGoVideo();
</script>
<template>
  <dialog
    class="modal"
    ref="mediaDialog"
    aria-labelledby="media-dialog-title"
    @close="restoreFocus('media')"
  >
    <div class="modal-box gv-modal-box">
      <div class="dialoghead">
        <div>
          <span class="eyebrow">PRÉVIA DO DOWNLOAD</span>
          <h2 id="media-dialog-title">Mídia analisada</h2>
        </div>
        <button
          class="btn btn-square btn-ghost"
          aria-label="Fechar prévia"
          autofocus
          @click="mediaDialog?.close()"
        >
          <AppIcon name="close" />
        </button>
      </div>
      <div v-if="media" class="dialogbody">
        <div v-if="error" class="alert alert-error" role="alert">
          {{ error }}
        </div>
        <div class="summary">
          <img
            v-if="media.thumbnailURL"
            :src="media.thumbnailURL"
            :alt="`Miniatura de ${media.title}`"
          />
          <div>
            <h3>{{ media.title }}</h3>
            <p>
              <span>{{ time(media.duration) }}</span
              ><span v-if="media.sizeEstimated"
                >Estimativa: {{ bytes(media.sizeBytes) }}</span
              >
            </p>
          </div>
        </div>
        <fieldset>
          <legend>Formato de saída</legend>
          <p class="field-help">
            Você pode baixar vídeo, música ou os dois em uma única tarefa.
          </p>
          <div class="choices">
            <label
              v-for="m in [
                { v: 'video', l: 'Vídeo', i: 'video' },
                { v: 'audio', l: 'Música', i: 'audio' },
                { v: 'both', l: 'Vídeo e música', i: 'download' },
              ]"
              :key="m.v"
              ><input
                v-model="output"
                class="radio radio-primary"
                type="radio"
                :value="m.v"
              /><span
                ><AppIcon :name="m.i as 'video' | 'audio' | 'download'" /><b>{{
                  m.l
                }}</b></span
              ></label
            >
          </div>
        </fieldset>
        <div class="formgrid">
          <label v-if="output !== 'audio'"
            >Resolução<select class="select w-full" v-model="height">
              <option v-for="q in qualities" :key="q" :value="q">
                {{ q ? `${q}p` : "Melhor disponível" }}
              </option>
            </select></label
          ><label v-if="output !== 'video'"
            >Formato de áudio<select
              class="select w-full"
              v-model="audioFormat"
            >
              <option>mp3</option>
              <option>m4a</option>
              <option>opus</option>
            </select></label
          ><label v-if="output !== 'video'"
            >Qualidade<select class="select w-full" v-model="audioQuality">
              <option value="economy">Econômica</option>
              <option value="standard">Padrão</option>
              <option value="high">Alta</option>
            </select></label
          >
        </div>
      </div>
      <div class="dialogactions">
        <p>Você poderá acompanhar e cancelar a tarefa em Atividades.</p>
        <div>
          <button class="btn btn-outline" @click="mediaDialog?.close()">
            Voltar</button
          ><button class="btn btn-primary" :disabled="busy" @click="download">
            <AppIcon name="download" />{{
              busy ? "Iniciando…" : "Iniciar download"
            }}
          </button>
        </div>
      </div>
    </div>
  </dialog>
</template>

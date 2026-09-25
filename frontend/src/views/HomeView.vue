<script setup lang="ts">
import AppIcon from "../components/AppIcon.vue";
import { useGoVideo } from "../composables/useGoVideo";
const {
  mode,
  url,
  busy,
  error,
  deps,
  files,
  audioFormat,
  audioQuality,
  analyze,
  choose,
  convert,
} = useGoVideo();
</script>
<template>
  <main>
    <div class="join segments" role="group" aria-label="Tipo de operação">
      <button
        class="btn join-item"
        :aria-pressed="mode === 'download'"
        @click="mode = 'download'"
      >
        <AppIcon name="download" />Baixar da internet
      </button>
      <button
        class="btn join-item"
        :aria-pressed="mode === 'convert'"
        @click="mode = 'convert'"
      >
        <AppIcon name="audio" />Converter vídeo em música
      </button>
    </div>

    <section
      v-if="mode === 'download'"
      class="workspace"
      aria-labelledby="download-title"
    >
      <span class="eyebrow">DOWNLOAD ONLINE</span>
      <h1 id="download-title">Seu vídeo, no formato que você escolher.</h1>
      <p>
        Analise um link antes de baixar. Escolha resolução, áudio e pasta de
        destino com segurança.
      </p>
      <form class="url" novalidate @submit.prevent="analyze">
        <AppIcon name="link" />
        <label class="sr" for="url">Endereço da mídia</label>
        <input
          class="input w-full"
          id="url"
          v-model="url"
          type="url"
          inputmode="url"
          autocomplete="url"
          placeholder="Cole um link, como https://…"
          :aria-describedby="error ? 'url-error' : undefined"
          @input="error = ''"
        />
        <button class="btn btn-primary" :disabled="busy">
          <span
            v-if="busy"
            class="loading loading-spinner loading-sm"
            aria-hidden="true"
          ></span
          >{{ busy ? "Analisando" : "Analisar" }}
        </button>
      </form>
      <p
        v-if="error"
        id="url-error"
        class="text-error inline-error"
        role="alert"
      >
        <AppIcon name="warning" />{{ error }}
      </p>
      <div class="card bg-base-100 empty intro-empty">
        <i><AppIcon name="link" /></i><b>Comece por um link</b
        ><span
          >A prévia mostrará título, duração e opções reais antes do
          download.</span
        >
        <small
          v-if="deps.find((d) => d.name === 'yt-dlp')?.state === 'installing'"
          class="dependency-note"
          ><span class="loading loading-spinner loading-sm"></span>Preparando
          ferramenta de análise…</small
        >
      </div>
    </section>

    <section
      v-else
      class="workspace"
      data-file-drop-target
      aria-labelledby="convert-title"
    >
      <span class="eyebrow">CONVERSÃO LOCAL</span>
      <h1 id="convert-title">Transforme seus vídeos em música.</h1>
      <p>
        Arraste arquivos ou selecione-os no computador. Cada vídeo entra na fila
        separadamente.
      </p>
      <div
        class="card bg-base-100 drop"
        data-file-drop-target
        aria-describedby="drop-help"
      >
        <i><AppIcon name="upload" /></i><b>Solte seus vídeos aqui</b
        ><span id="drop-help">MP4, MKV, WebM, MOV, AVI ou M4V</span>
        <button class="btn btn-outline" @click="choose">
          <AppIcon name="folder" />Selecionar vídeos
        </button>
      </div>
      <ul v-if="files.length" class="files" aria-label="Arquivos selecionados">
        <li v-for="f in files" :key="f">
          <i><AppIcon name="file" /></i><b>{{ f.split(/[\\/]/).pop() }}</b
          ><span>Pronto para converter</span
          ><button
            class="btn btn-square btn-ghost"
            :aria-label="`Remover ${f.split(/[\\/]/).pop()}`"
            @click="files = files.filter((x) => x !== f)"
          >
            <AppIcon name="close" />
          </button>
        </li>
      </ul>
      <div class="options" :class="{ disabled: !files.length }">
        <label
          >Formato<select class="select w-full" v-model="audioFormat">
            <option>mp3</option>
            <option>m4a</option>
            <option>opus</option>
          </select></label
        >
        <label
          >Qualidade<select class="select w-full" v-model="audioQuality">
            <option value="economy">Econômica</option>
            <option value="standard">Padrão</option>
            <option value="high">Alta</option>
          </select></label
        >
        <button
          class="btn btn-primary"
          :disabled="!files.length || busy"
          @click="convert"
        >
          <AppIcon name="audio" />{{
            busy
              ? "Enviando à fila"
              : `Converter ${files.length} ${files.length === 1 ? "vídeo" : "vídeos"}`
          }}
        </button>
      </div>
    </section>
  </main>
</template>

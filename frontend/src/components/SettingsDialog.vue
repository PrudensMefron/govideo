<script setup lang="ts">
import AppIcon from "../components/AppIcon.vue";
import { useGoVideo } from "../composables/useGoVideo";
const dependencyLabels: Record<string, string> = {
  missing: "Não instalada",
  installing: "Instalando",
  ready: "Pronta",
  invalid: "Inválida",
  failed: "Falhou",
};
const {
  deps,
  settings,
  update,
  checkingUpdate,
  updateError,
  settingsDialog,
  folder,
  save,
  restoreFocus,
  checkUpdates,
  savingSettings,
  settingsError,
} = useGoVideo();
import * as api from "../../bindings/github.com/PrudensMefron/govideo/desktopservice";
</script>
<template>
  <dialog
    class="modal"
    ref="settingsDialog"
    aria-labelledby="settings-dialog-title"
    @close="restoreFocus('settings')"
  >
    <div class="modal-box gv-modal-box">
      <div class="dialoghead">
        <div>
          <span class="eyebrow">PREFERÊNCIAS</span>
          <h2 id="settings-dialog-title">Configurações</h2>
        </div>
        <button
          class="btn btn-square btn-ghost"
          aria-label="Fechar configurações"
          autofocus
          @click="settingsDialog?.close()"
        >
          <AppIcon name="close" />
        </button>
      </div>
      <section class="settings">
        <div v-if="settingsError" class="alert alert-error" role="alert">
          {{ settingsError }}
        </div>
        <h3>Ferramentas</h3>
        <p class="section-help">
          GoVideo verifica essas ferramentas antes de iniciar as tarefas.
        </p>
        <div v-for="d in deps" :key="d.name" class="dep">
          <div>
            <b>{{ d.name }}</b
            ><span>{{ d.version || d.path || "Não disponível" }}</span
            ><small>{{ d.license }}</small>
          </div>
          <span class="badge badge-outline gv-status" :class="d.state">{{
            dependencyLabels[d.state] || d.state
          }}</span
          ><button
            v-if="['failed', 'invalid', 'missing'].includes(d.state)"
            class="btn btn-outline btn-sm"
            @click="api.RetryDependency(d.name)"
          >
            <AppIcon name="refresh" />Tentar novamente
          </button>
        </div>
        <h3>Atualizações</h3>
        <div v-if="update" class="update">
          <div>
            <b>GoVideo {{ update.currentVersion }}</b
            ><span v-if="!update.enabled"
              >Atualizações desativadas nesta build de desenvolvimento.</span
            ><span v-else-if="update.availableVersion"
              >Versão {{ update.availableVersion }} disponível.</span
            ><span v-else>Canal estável · GitHub Releases</span
            ><small>{{ update.repository }}</small
            ><small v-if="updateError || update.error" class="text-error">{{
              updateError || update.error
            }}</small>
          </div>
          <button
            class="btn btn-outline btn-sm"
            :disabled="!update.enabled || checkingUpdate"
            @click="checkUpdates"
          >
            <AppIcon name="refresh" />{{
              checkingUpdate
                ? "Verificando…"
                : update.availableVersion
                  ? "Baixar atualização"
                  : "Verificar atualizações"
            }}
          </button>
        </div>
        <h3>Pastas padrão</h3>
        <label
          v-for="k in [
            'videoDirectory',
            'musicDirectory',
            'conversionDirectory',
          ] as const"
          :key="k"
          >{{
            k === "videoDirectory"
              ? "Vídeos"
              : k === "musicDirectory"
                ? "Músicas"
                : "Conversões"
          }}
          <div class="path">
            <input class="input w-full" v-model="settings[k]" /><button
              class="btn btn-outline btn-sm"
              @click="folder(k)"
            >
              <AppIcon name="folder" />Selecionar
            </button>
          </div></label
        ><label
          >Tema<select class="select w-full" v-model="settings.theme">
            <option value="system">Acompanhar o sistema</option>
            <option value="light">Claro</option>
            <option value="dark">Escuro</option>
          </select></label
        >
      </section>
      <div class="dialogactions">
        <p>As pastas são criadas quando você salva.</p>
        <div>
          <button class="btn btn-outline" @click="settingsDialog?.close()">
            Cancelar</button
          ><button
            class="btn btn-primary"
            :disabled="savingSettings"
            @click="save"
          >
            <span
              v-if="savingSettings"
              class="loading loading-spinner loading-sm"
            /><AppIcon v-else name="check" />{{
              savingSettings ? "Salvando…" : "Salvar alterações"
            }}
          </button>
        </div>
      </div>
    </div>
  </dialog>
</template>

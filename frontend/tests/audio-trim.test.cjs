const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');
const vue = require('vue');

function loadEditor(api) {
  const source = fs.readFileSync(path.join(__dirname, '../src/composables/useAudioTrim.ts'), 'utf8');
  const js = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText;
  const exports = {};
  vm.runInNewContext(js, { exports, require: name => name === 'vue'
    ? { ...vue, onMounted() {}, onBeforeUnmount() {} } : api });
  return exports.useAudioTrim();
}

test('regeneration waits for discarded preview before acquiring backend render lock', async () => {
  let finishDiscard;
  let rendering = false;
  const events = [];
  const editor = loadEditor({
    DiscardAudioPreview(id) {
      events.push('discard:'+id);
      return new Promise(resolve => { finishDiscard = resolve; });
    },
    async CreateAudioPreview() {
      rendering = true;
      events.push('render');
      return { id:'new', playbackURL:'http://127.0.0.1:1234/audio-preview/new' };
    },
  });
  editor.source.value = {path:'/music.mp3',durationSeconds:60};
  editor.startSeconds.value = 5;
  await vue.nextTick();
  editor.preview.value = {id:'old',playbackURL:'http://127.0.0.1:1234/audio-preview/old'};
  const generating = editor.generatePreview();
  await Promise.resolve();
  assert.equal(editor.busy.value,'preview');
  assert.equal(rendering,false,'new render raced discard');
  assert.deepEqual(events,['discard:old']);
  finishDiscard();
  await generating;
  assert.deepEqual(events,['discard:old','render']);
  assert.equal(editor.previewURL.value,'http://127.0.0.1:1234/audio-preview/new');
  assert.equal(editor.busy.value,null);
});

test('cancel during cleanup prevents new render', async () => {
  let finishDiscard;
  let renders=0;
  const editor=loadEditor({DiscardAudioPreview(){return new Promise(resolve=>{finishDiscard=resolve})},CreateAudioPreview(){renders++;}});
  editor.source.value={path:'/music.mp3',durationSeconds:60};
  editor.startSeconds.value=5;await vue.nextTick();
  editor.preview.value={id:'old'};
  const generating=editor.generatePreview();await Promise.resolve();
  await editor.cancelPreview();finishDiscard();await generating;
  assert.equal(renders,0);assert.equal(editor.busy.value,null);
});

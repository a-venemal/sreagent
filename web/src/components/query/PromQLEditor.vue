<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, shallowRef } from 'vue'
import { EditorView, keymap, placeholder as cmPlaceholder } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { completionKeymap, autocompletion } from '@codemirror/autocomplete'
import { PromQLExtension } from '@prometheus-io/codemirror-promql'
import { oneDark } from '@codemirror/theme-one-dark'

const props = withDefaults(defineProps<{
  modelValue: string
  datasourceId?: number | null
  placeholder?: string
  disabled?: boolean
  dark?: boolean
}>(), {
  datasourceId: null,
  placeholder: 'Enter PromQL expression...',
  disabled: false,
  dark: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'execute'): void
}>()

const editorRef = ref<HTMLDivElement>()
const view = shallowRef<EditorView>()

const promQLExt = new PromQLExtension()

function createExtensions() {
  const exts = [
    history(),
    keymap.of([
      ...defaultKeymap,
      ...historyKeymap,
      ...completionKeymap,
      { key: 'Ctrl-Enter', run: () => { emit('execute'); return true } },
      { key: 'Cmd-Enter', run: () => { emit('execute'); return true } },
    ]),
    promQLExt.asExtension(
      props.datasourceId
        ? { completeStrategy: { remote: { url: '/api/v1', datasourceId: props.datasourceId } } }
        : {}
    ),
    autocompletion(),
    cmPlaceholder(props.placeholder),
    EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        emit('update:modelValue', update.state.doc.toString())
      }
    }),
    EditorView.lineWrapping,
  ]
  if (props.dark) exts.push(oneDark)
  if (props.disabled) exts.push(EditorState.readOnly.of(true))
  return exts
}

onMounted(() => {
  if (!editorRef.value) return
  const state = EditorState.create({
    doc: props.modelValue,
    extensions: createExtensions(),
  })
  view.value = new EditorView({ state, parent: editorRef.value })
})

onUnmounted(() => {
  view.value?.destroy()
})

watch(() => props.modelValue, (val) => {
  if (view.value && view.value.state.doc.toString() !== val) {
    view.value.dispatch({
      changes: { from: 0, to: view.value.state.doc.length, insert: val },
    })
  }
})

watch(() => props.datasourceId, () => {
  // Recreate editor when datasource changes to update completion context
  if (view.value) {
    const doc = view.value.state.doc.toString()
    view.value.destroy()
    if (editorRef.value) {
      const state = EditorState.create({
        doc,
        extensions: createExtensions(),
      })
      view.value = new EditorView({ state, parent: editorRef.value })
    }
  }
})
</script>

<template>
  <div ref="editorRef" class="promql-editor" :class="{ disabled }" />
</template>

<style scoped>
.promql-editor {
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  overflow: hidden;
  min-height: 42px;
}
.promql-editor:focus-within {
  border-color: #18a058;
  box-shadow: 0 0 0 2px rgba(24, 160, 88, 0.15);
}
.promql-editor.disabled {
  opacity: 0.6;
  pointer-events: none;
}
.promql-editor :deep(.cm-editor) {
  min-height: 42px;
  font-size: 13px;
}
.promql-editor :deep(.cm-content) {
  padding: 8px 12px;
}
</style>

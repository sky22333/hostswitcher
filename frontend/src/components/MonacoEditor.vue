<template>
  <div class="monaco-editor-wrapper" :style="{ height }">    
    <!-- 编辑器容器 -->
    <div
      ref="editorContainer"
      class="monaco-editor-container"
      :style="{ height }"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick, computed } from 'vue';
import { useTheme } from 'vuetify';
import * as monaco from 'monaco-editor';

// 属性定义
const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  language: {
    type: String,
    default: 'plaintext'
  },
  options: {
    type: Object,
    default: () => ({})
  },
  loading: {
    type: Boolean,
    default: false
  },
  height: {
    type: String,
    default: '400px'
  }
});

// 事件定义
const emit = defineEmits(['update:modelValue', 'change', 'ready']);

// 编辑器引用
const editorContainer = ref(null);
let editor = null;
let contentChangeDisposable = null;
let isUpdatingFromProps = false;

// 主题相关
const vuetifyTheme = useTheme();
const isDarkTheme = computed(() => vuetifyTheme.global.current.value.dark);

// 生命周期钩子
onMounted(async () => {
  await nextTick();
  setupMonacoLanguage();
  initEditor();
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResize);
});

onBeforeUnmount(() => {
  // 清理资源
  if (contentChangeDisposable) {
    contentChangeDisposable.dispose();
    contentChangeDisposable = null;
  }
  
  if (editor) {
    editor.dispose();
    editor = null;
  }
  
  window.removeEventListener('resize', handleResize);
});

// 监听属性变化
watch(() => props.modelValue, (newValue) => {
  if (editor && !isUpdatingFromProps && editor.getValue() !== newValue) {
    isUpdatingFromProps = true;
    const selection = editor.getSelection();
    editor.setValue(newValue);
    if (selection) {
      editor.setSelection(selection);
    }
    isUpdatingFromProps = false;
  }
});

watch(() => props.language, (newValue) => {
  if (editor && editor.getModel()) {
    monaco.editor.setModelLanguage(editor.getModel(), newValue);
  }
});

watch(() => props.options, (newValue) => {
  if (editor) {
    editor.updateOptions(newValue);
  }
}, { deep: true });

// 监听主题变化
watch(isDarkTheme, (newValue) => {
  if (editor) {
    applyEditorTheme();
  }
}, { immediate: true });

function setupMonacoLanguage() {
  // 简化的语言配置
  if (!window.MonacoEnvironment) {
    window.MonacoEnvironment = { Locale: 'zh-cn' };
  }
}

// 初始化编辑器
function initEditor() {
  if (!editorContainer.value) return;
  
  // 精简的默认选项 - 移除冗余配置
  const defaultOptions = {
    theme: isDarkTheme.value ? 'vs-dark' : 'vs',
    fontSize: 14,
    wordWrap: 'on',
    automaticLayout: true,
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    lineNumbers: 'on',
    tabSize: 4,
    insertSpaces: true,
    folding: true,
    renderWhitespace: 'selection',
    scrollbar: {
      verticalScrollbarSize: 8,
      horizontalScrollbarSize: 8,
    },
  };
  
  // 合并选项
  const editorOptions = {
    ...defaultOptions,
    ...props.options,
    value: props.modelValue,
    language: props.language,
  };
  
  try {
    // 创建编辑器
    editor = monaco.editor.create(editorContainer.value, editorOptions);
    
    // 监听内容变化
    contentChangeDisposable = editor.onDidChangeModelContent(() => {
      if (!isUpdatingFromProps) {
        const value = editor.getValue();
        emit('update:modelValue', value);
        emit('change', value);
      }
    });
    
    // 设置编辑器大小
    handleResize();
    
    // 发出准备完成事件
    emit('ready', editor);
    
    // 添加hosts文件语法高亮
    setupHostsLanguage();
    
    // 应用主题
    applyEditorTheme();
    
  } catch (error) {
    console.error('Failed to initialize Monaco Editor:', error);
  }
}

// 简化的主题应用方法
function applyEditorTheme() {
  if (!editor) return;
  
  const theme = isDarkTheme.value ? 'vs-dark' : 'vs';
  monaco.editor.setTheme(theme);
  
  // 延迟布局确保主题应用完成
  setTimeout(() => {
    if (editor) {
      editor.layout();
    }
  }, 50);
}

// 设置hosts文件语法高亮
function setupHostsLanguage() {
  if (!monaco.languages.getLanguages().find(lang => lang.id === 'hosts')) {
    // 注册hosts语言
    monaco.languages.register({ id: 'hosts' });
    
    // 设置语言配置
    monaco.languages.setLanguageConfiguration('hosts', {
      comments: {
        lineComment: '#'
      },
      brackets: [],
      autoClosingPairs: [],
      surroundingPairs: [],
    });
    
    // 设置语法高亮规则
    monaco.languages.setMonarchTokensProvider('hosts', {
      tokenizer: {
        root: [
          [/#.*$/, 'comment'],
          [/^\s*$/, 'whitespace'],
          [/^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/, 'number.ipv4'],
          [/^([0-9a-fA-F:]+)/, 'number.ipv6'],
          [/\S+/, 'string.hostname'],
        ]
      }
    });
  }
  
  // 如果当前语言是properties或其他，切换到hosts
  if (editor && editor.getModel()) {
    const currentLang = monaco.editor.getModel(editor.getModel().uri).getLanguageId();
    if (currentLang !== 'hosts') {
      monaco.editor.setModelLanguage(editor.getModel(), 'hosts');
    }
  }
}

// 处理窗口大小变化
function handleResize() {
  if (editor) {
    // 延迟调用layout，确保容器大小已更新
    setTimeout(() => {
      editor.layout();
    }, 100);
  }
}

// 暴露编辑器实例和方法
defineExpose({
  getEditor: () => editor,
  focus: () => editor?.focus(),
  setValue: (value) => editor?.setValue(value),
  getValue: () => editor?.getValue(),
  getSelection: () => editor?.getSelection(),
  setSelection: (selection) => editor?.setSelection(selection),
  layout: () => editor?.layout(),
});
</script>

<style scoped>
.monaco-editor-wrapper {
  position: relative;
  width: 100%;
  overflow: hidden;
}

.monaco-editor-container {
  width: 100%;
  height: 100%;
}

/* 加载状态样式 */
.editor-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  font-size: 14px;
}
</style>

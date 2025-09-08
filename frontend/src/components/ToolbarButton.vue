<template>
  <v-tooltip :location="tooltipLocation">
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        :icon="icon"
        :variant="variant"
        :size="size"
        :color="color"
        :loading="loading"
        :disabled="disabled"
        :class="buttonClass"
        @click="$emit('click')"
      >
        <v-icon>{{ icon }}</v-icon>
      </v-btn>
    </template>
    <span>{{ tooltip }}</span>
  </v-tooltip>
</template>

<script setup>
defineProps({
  icon: {
    type: String,
    required: true
  },
  tooltip: {
    type: String,
    required: true
  },
  variant: {
    type: String,
    default: 'text'
  },
  size: {
    type: String,
    default: 'small'
  },
  color: {
    type: String,
    default: undefined
  },
  loading: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  },
  buttonClass: {
    type: String,
    default: 'icon-btn'
  },
  tooltipLocation: {
    type: String,
    default: 'bottom'
  }
})

defineEmits(['click'])
</script>

<style scoped>
/* 使用全局样式，无需重复定义 */
</style>
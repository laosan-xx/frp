<template>
  <BaseDialog
    v-model="visible"
    :title="title"
    :width="isMobile ? '92vw' : '400px'"
    :close-on-click-modal="false"
    :append-to-body="true"
    :is-mobile="isMobile"
  >
    <p class="confirm-message">{{ message }}</p>
    <template #footer>
      <div class="dialog-footer">
        <ActionButton variant="outline" @click="handleCancel">
          {{ cancelText }}
        </ActionButton>
        <ActionButton
          :danger="danger"
          :loading="loading"
          @click="handleConfirm"
        >
          {{ confirmText }}
        </ActionButton>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useBreakpoints } from '@vueuse/core'
import BaseDialog from './BaseDialog.vue'
import ActionButton from '@shared/components/ActionButton.vue'

const breakpoints = useBreakpoints({ mobile: 0, desktop: 768 })

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    message: string
    confirmText?: string
    cancelText?: string
    danger?: boolean
    loading?: boolean
    isMobile?: boolean
  }>(),
  {
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    danger: false,
    loading: false,
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

// 外部未显式传入 isMobile 时，自动按视口宽度判定，避免移动端固定 400px 超宽
const isMobile = computed(
  () => props.isMobile ?? breakpoints.smaller('desktop').value,
)

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  visible.value = false
  emit('cancel')
}
</script>

<style scoped lang="scss">
.confirm-message {
  margin: 0;
  font-size: $font-size-md;
  color: $color-text-secondary;
  line-height: 1.6;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: $spacing-md;
}
</style>

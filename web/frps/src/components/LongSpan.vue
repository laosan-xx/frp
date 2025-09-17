<template>
  <el-tooltip :content="content" placement="top">
    <span class="long-span" v-show="content.length > length"
      >{{ content.slice(0, length) }}...</span
    >
  </el-tooltip>
  <span class="long-span" v-show="content.length < 30">{{ content }}</span>
</template>

<script setup lang="ts">
defineProps<{
  content: string
  length: number
}>()
</script>

<style scoped>
.long-span {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 600;
}
</style>

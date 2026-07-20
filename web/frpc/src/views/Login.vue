<template>
  <div class="login-wrapper">
    <div class="login-card">
      <div class="login-brand">
        <span class="login-logo">frp</span>
        <span class="login-title">Client</span>
      </div>

      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        label-position="top"
        class="login-form"
      >
        <el-form-item label="Username" prop="username">
          <el-input
            v-model="form.username"
            placeholder="Enter username"
            clearable
            size="large"
          />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input
            v-model="form.password"
            placeholder="Enter password"
            show-password
            clearable
            size="large"
          />
        </el-form-item>
        <el-form-item label="Captcha" prop="captchaAns">
          <div class="captcha-row">
            <el-input
              v-model="form.captchaAns"
              placeholder="Enter captcha"
              clearable
              size="large"
              class="captcha-input"
              @keyup.enter="onSubmit"
            />
            <div
              class="captcha-box"
              v-html="captchaSvg"
              @click="refreshCaptcha"
              title="Click to refresh"
            ></div>
          </div>
        </el-form-item>
        <el-button
          type="primary"
          :loading="loading"
          class="submit-btn"
          @click="onSubmit"
          size="large"
        >
          Login
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import CryptoJS from 'crypto-js'

interface LoginForm {
  username: string
  password: string
  captchaId?: string
  captchaAns?: string
}

const router = useRouter()
const route = useRoute()
const formRef = ref()
const loading = ref(false)
const form = reactive<LoginForm>({
  username: '',
  password: '',
  captchaId: '',
  captchaAns: '',
})
const captchaSvg = ref('')

const rules = {
  username: [{ required: true, message: 'Please enter username', trigger: 'blur' }],
  password: [{ required: true, message: 'Please enter password', trigger: 'blur' }],
  captchaAns: [{ required: true, message: 'Please enter captcha', trigger: 'blur' }],
}

const refreshCaptcha = async () => {
  try {
    const resp = await fetch('/api/captcha')
    if (!resp.ok) throw new Error('captcha failed')
    const data = await resp.json()
    form.captchaId = data.id
    captchaSvg.value = data.svg
    form.captchaAns = ''
  } catch {
    captchaSvg.value = ''
    form.captchaAns = ''
  }
}

const sha256Hash = (text: string): string => {
  return CryptoJS.SHA256(text).toString()
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    try {
      loading.value = true

      const hashedPassword = sha256Hash(form.password)

      const resp = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: form.username.trim(),
          password: hashedPassword,
          captchaId: form.captchaId,
          captchaAns: form.captchaAns?.trim(),
        }),
      })

      if (!resp.ok) {
        if (resp.status === 401) {
          ElMessage.error('Invalid username, password or captcha')
        } else {
          ElMessage.error(`Login failed (${resp.status})`)
        }
        await refreshCaptcha()
        return
      }

      ElMessage.success('Login successful')
      const redirect = (route.query.redirect as string) || '/'
      router.replace(redirect)
    } catch (e) {
      console.error('Login error:', e)
      ElMessage.error('Network error, please check connection')
      await refreshCaptcha()
    } finally {
      loading.value = false
    }
  })
}

onMounted(() => {
  refreshCaptcha()
})
</script>

<style lang="scss" scoped>
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  background: $color-bg-secondary;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: $color-bg-primary;
  border: 1px solid $color-border-light;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.login-brand {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 24px;
}

.login-logo {
  font-weight: $font-weight-semibold;
  font-size: 24px;
  color: $color-text-primary;
}

.login-title {
  font-size: 16px;
  color: $color-text-muted;
}

.submit-btn {
  width: 100%;
  margin-top: 8px;
}

.captcha-row {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.captcha-input {
  flex: 1;
}

.captcha-box {
  width: 80px;
  height: 40px;
  border-radius: $radius-sm;
  overflow: hidden;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid $color-border-light;
  cursor: pointer;
  flex-shrink: 0;
}
</style>

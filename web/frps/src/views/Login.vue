<template>
  <div class="login-wrapper">
    <div class="login-card">
      <div class="login-brand">
        <span class="login-logo">frp</span>
        <span class="login-title">{{ $t('login.dashboard') }}</span>
      </div>

      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        label-position="top"
        class="login-form"
      >
        <el-form-item :label="$t('login.username')" prop="username">
          <el-input
            v-model="form.username"
            :placeholder="$t('login.usernamePlaceholder')"
            clearable
            size="large"
          />
        </el-form-item>
        <el-form-item :label="$t('login.password')" prop="password">
          <el-input
            v-model="form.password"
            :placeholder="$t('login.passwordPlaceholder')"
            show-password
            clearable
            size="large"
          />
        </el-form-item>
        <el-form-item :label="$t('login.captcha')" prop="captchaAns">
          <div class="captcha-row">
            <el-input
              v-model="form.captchaAns"
              :placeholder="$t('login.captchaPlaceholder')"
              clearable
              size="large"
              class="captcha-input"
              @keyup.enter="onSubmit"
            />
            <div
              class="captcha-box"
              v-html="captchaSvg"
              @click="refreshCaptcha"
              :title="$t('login.captchaRefresh')"
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
          {{ $t('login.submit') }}
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CryptoJS from 'crypto-js'

interface LoginForm {
  username: string
  password: string
  captchaId?: string
  captchaAns?: string
}

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
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
  username: [{ required: true, message: () => t('login.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: () => t('login.passwordRequired'), trigger: 'blur' }],
  captchaAns: [{ required: true, message: () => t('login.captchaRequired'), trigger: 'blur' }],
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
          ElMessage.error(t('login.invalidCredentials'))
        } else {
          ElMessage.error(t('login.loginFailed', { status: resp.status }))
        }
        await refreshCaptcha()
        return
      }

      ElMessage.success(t('login.loginSuccess'))
      const redirect = (route.query.redirect as string) || '/'
      router.replace(redirect)
    } catch (e) {
      console.error('Login error:', e)
      ElMessage.error(t('login.networkError'))
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

<style scoped>
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  background: var(--content-bg, #f9f9f9);
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: var(--header-bg, #ffffff);
  border: 1px solid var(--header-border, #e4e7ed);
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
  font-weight: 700;
  font-size: 24px;
  color: var(--text-primary, #303133);
}

.login-title {
  font-size: 16px;
  color: var(--text-muted, #909399);
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
  border-radius: 6px;
  overflow: hidden;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--header-border, #e4e7ed);
  cursor: pointer;
  flex-shrink: 0;
}
</style>

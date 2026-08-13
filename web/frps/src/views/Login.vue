<template>
  <div class="login-wrapper">
    <aside class="login-aside">
      <div class="login-aside-inner">
        <h1 class="login-aside-title">{{ $t('login.welcome') }}</h1>
        <p class="login-aside-desc">{{ $t('login.welcomeDesc') }}</p>
        <ul class="login-aside-features">
          <li>
            <span class="login-aside-dot"></span>{{ $t('login.featureRealtime') }}
          </li>
          <li>
            <span class="login-aside-dot"></span>{{ $t('login.featureProxy') }}
          </li>
          <li>
            <span class="login-aside-dot"></span>{{ $t('login.featureSecure') }}
          </li>
        </ul>
      </div>
      <div class="login-aside-deco login-aside-deco-1"></div>
      <div class="login-aside-deco login-aside-deco-2"></div>
    </aside>

    <div class="login-card">
      <div class="login-brand">
        <span class="login-brand-logo">frp</span>
        <span>
          <span class="login-logo">frp</span>
          <span class="login-title">{{ $t('login.dashboard') }}</span>
        </span>
      </div>

      <div class="login-welcome">
        <div class="login-welcome-title">{{ $t('login.welcome') }}</div>
        <div class="login-welcome-desc">{{ $t('login.welcomeDesc') }}</div>
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
          >
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item :label="$t('login.password')" prop="password">
          <el-input
            v-model="form.password"
            :placeholder="$t('login.passwordPlaceholder')"
            show-password
            clearable
            size="large"
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
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

      <div class="login-footer">{{ $t('login.footer') }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'

defineOptions({ name: 'LoginView' })
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CryptoJS from 'crypto-js'
import { User, Lock } from '@element-plus/icons-vue'

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
  align-items: stretch;
  background: var(--content-bg, #f9f9f9);
}

/* Left brand showcase (desktop only) */
.login-aside {
  position: relative;
  flex: 1 1 0;
  display: flex;
  align-items: center;
  padding: 48px;
  overflow: hidden;
  color: #fff;
  background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
}

.login-aside-inner {
  position: relative;
  z-index: 2;
  max-width: 420px;
}

.login-aside-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 48px;
}

.login-aside-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 52px;
  height: 52px;
  border-radius: 14px;
  font-weight: 800;
  font-size: 22px;
  color: #3b82f6;
  background: #fff;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
}

.login-aside-name {
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.login-aside-title {
  font-size: 44px;
  font-weight: 800;
  line-height: 1.25;
  margin: 0 0 16px;
}

.login-aside-desc {
  font-size: 15px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.85);
  margin: 0 0 32px;
}

.login-aside-features {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.login-aside-features li {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.92);
}

.login-aside-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 0 0 4px rgba(255, 255, 255, 0.25);
  flex-shrink: 0;
}

.login-aside-deco {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.12);
  z-index: 1;
}

.login-aside-deco-1 {
  width: 280px;
  height: 280px;
  top: -80px;
  right: -60px;
}

.login-aside-deco-2 {
  width: 180px;
  height: 180px;
  bottom: -50px;
  left: -40px;
  background: rgba(255, 255, 255, 0.08);
}

/* Right login card */
.login-card {
  flex: 1 1 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  width: 100%;
  padding: 48px;
  background: var(--header-bg, #ffffff);
  border-left: 1px solid var(--header-border, #e4e7ed);
}

.login-card > * {
  width: 100%;
  max-width: 380px;
}

.login-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 28px;
}

/* Welcome block only shows on mobile (desktop uses the aside title) */
.login-welcome {
  display: none;
}

.login-brand-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  font-weight: 800;
  font-size: 16px;
  color: #fff;
  background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
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

/* Inputs follow the active theme on desktop too */
.login-card :deep(.el-input__wrapper) {
  background: var(--color-bg-input, #fff);
  box-shadow: 0 0 0 1px var(--header-border, #e4e7ed) inset;
}

.login-card :deep(.el-input__inner) {
  color: var(--text-primary, #303133);
}

.captcha-box {
  background: var(--color-bg-input, #fff);
}

.login-footer {
  display: none;
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
  position: relative;
  width: 86px;
  height: 40px;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
  border: 1px solid var(--header-border, #e4e7ed);
  cursor: pointer;
  flex-shrink: 0;
  transition:
    box-shadow 0.2s ease,
    border-color 0.2s ease,
    transform 0.1s ease;
}

.captcha-box:hover {
  border-color: #3b82f6;
  box-shadow: 0 2px 10px rgba(59, 130, 246, 0.18);
}

.captcha-box:active {
  transform: scale(0.97);
}

.captcha-box :deep(svg) {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  display: block;
}

/* Mobile: light, card-less native login page (no floating card) */
@media (max-width: 767px) {
  .login-wrapper {
    display: block;
    padding: 0;
    min-height: 100vh;
    min-height: 100dvh;
    position: relative;
    overflow-x: hidden;
    background: var(--content-bg, #f9f9f9);
  }

  /* Hide the desktop brand showcase on mobile */
  .login-aside {
    display: none;
  }

  /* Restore the card to its light, borderless mobile form */
  .login-card {
    flex: none;
    max-width: none;
    margin: 0;
    padding: 56px 24px 32px;
    border: none;
    border-radius: 0;
    border-left: none;
    box-shadow: none;
    background: transparent;
  }

  /* Soft brand glow at the top — stays light, no dark overlay */
  .login-wrapper::before {
    content: '';
    position: absolute;
    top: -120px;
    left: 50%;
    transform: translateX(-50%);
    width: 360px;
    height: 320px;
    background: radial-gradient(
      circle,
      rgba(59, 130, 246, 0.16),
      rgba(6, 182, 212, 0.08) 45%,
      transparent 70%
    );
    pointer-events: none;
    z-index: 0;
  }

  /* Login content blends into the page: no border, no shadow, no radius */
  .login-card {
    position: relative;
    z-index: 1;
    width: auto;
    max-width: none;
    margin: 0;
    border: none;
    border-radius: 0;
    box-shadow: none;
    padding: 56px 24px 32px;
    background: transparent;
  }

  /* Card-internal brand shows on mobile as the top logo block */
  .login-brand {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 14px;
    margin-bottom: 36px;
  }

  .login-brand-logo {
    width: 48px;
    height: 48px;
    border-radius: 14px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 800;
    font-size: 20px;
    color: #fff;
    background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
    box-shadow: 0 6px 16px rgba(59, 130, 246, 0.28);
  }

  .login-logo {
    font-size: 24px;
    background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .login-title {
    display: block;
    font-size: 13px;
    color: var(--text-muted, #909399);
    margin-top: 2px;
  }

  /* Welcome heading block */
  .login-welcome {
    display: block;
    margin: -16px 0 24px;
  }

  .login-welcome-title {
    font-size: 22px;
    font-weight: 700;
    color: var(--text-primary, #303133);
    letter-spacing: 0.3px;
  }

  .login-welcome-desc {
    margin-top: 6px;
    font-size: 13px;
    color: var(--text-muted, #909399);
  }

  /* Gradient divider above the form */
  .login-card > .el-form::before {
    content: '';
    display: block;
    width: 44px;
    height: 3px;
    border-radius: 3px;
    margin-bottom: 24px;
    background: linear-gradient(90deg, #3b82f6, #06b6d4);
  }

  /* Keep inputs clean and full-width, following the active theme */
  .login-card :deep(.el-input__wrapper) {
    background: var(--color-bg-input, #fff);
    box-shadow: 0 0 0 1px var(--header-border, #e4e7ed) inset !important;
  }

  .login-card :deep(.el-input__wrapper:hover) {
    box-shadow: 0 0 0 1px #c0c4cc inset !important;
  }

  .login-card :deep(.el-input__wrapper.is-focus) {
    box-shadow: 0 0 0 1px #3b82f6 inset !important;
  }

  .login-card :deep(.el-input__prefix) {
    color: var(--text-muted, #909399);
  }

  .login-card :deep(.el-input__inner) {
    color: var(--text-primary, #303133);
  }

  .submit-btn {
    margin-top: 20px;
    height: 48px;
    font-size: 16px;
    font-weight: 600;
    color: #fff;
    background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
    border: none;
    box-shadow: 0 8px 18px rgba(59, 130, 246, 0.25);
  }

  .submit-btn:hover {
    box-shadow: 0 10px 22px rgba(6, 182, 212, 0.32);
  }

  /* Brand footer pinned at the bottom */
  .login-footer {
    display: block;
    position: fixed;
    left: 0;
    right: 0;
    bottom: 16px;
    text-align: center;
    font-size: 12px;
    color: var(--text-muted, #909399);
    opacity: 0.8;
  }
}
</style>

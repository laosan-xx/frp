<template>
  <div class="auth-wrapper">
    <div class="glass-card">
      <div class="brand">
        <div class="logo">frp</div>
        <div class="title">Dashboard 登录</div>
      </div>

      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        label-position="top"
        class="login-form"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            clearable
            size="large"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            placeholder="请输入密码"
            show-password
            clearable
            size="large"
          />
        </el-form-item>
        <el-form-item label="验证码" prop="captchaAns">
          <div class="captcha-row">
            <el-input
              v-model="form.captchaAns"
              placeholder="请输入右侧验证码"
              clearable
              size="large"
              class="captcha-input"
              @keyup.enter="onSubmit"
            />
            <div
              class="captcha-box"
              v-html="captchaSvg"
              @click="refreshCaptcha"
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
          登录
        </el-button>
      </el-form>

      <div class="footer">© {{ new Date().getFullYear() }} frp Dashboard</div>
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
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captchaAns: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

const refreshCaptcha = async () => {
  try {
    const resp = await fetch('/api/captcha')
    if (!resp.ok) throw new Error('captcha failed')
    const data = await resp.json()
    form.captchaId = data.id
    captchaSvg.value = data.svg
    // 清空用户输入的验证码
    form.captchaAns = ''
  } catch {
    captchaSvg.value = ''
    // 即使请求失败也清空验证码输入
    form.captchaAns = ''
  }
}

// SHA256 加密函数 - 使用 crypto-js 包
const sha256Hash = (text: string): string => {
  return CryptoJS.SHA256(text).toString()
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    try {
      loading.value = true

      // 对密码进行 SHA256 加密
      const hashedPassword = sha256Hash(form.password)

      const resp = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: form.username.trim(),
          password: hashedPassword, // 发送加密后的密码
          captchaId: form.captchaId,
          captchaAns: form.captchaAns?.trim(), // 去除前后空格
        }),
      })

      if (!resp.ok) {
        if (resp.status === 401) {
          ElMessage.error('用户名、密码或验证码错误')
        } else if (resp.status === 400) {
          ElMessage.error('请求参数错误')
        } else {
          ElMessage.error(`登录失败 (${resp.status})`)
        }

        // 刷新验证码
        await refreshCaptcha()
        return
      }

      // 登录成功
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/'
      router.replace(redirect)
    } catch (e) {
      console.error('登录异常:', e)
      ElMessage.error('网络错误，请检查连接')
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
.auth-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  background:
    radial-gradient(
      1200px 600px at 10% 10%,
      rgba(102, 126, 234, 0.25),
      transparent
    ),
    radial-gradient(
      1200px 600px at 90% 90%,
      rgba(118, 75, 162, 0.25),
      transparent
    ),
    linear-gradient(135deg, #f7fafc 0%, #edf2f7 100%);
}

html.dark .auth-wrapper {
  background:
    radial-gradient(
      1200px 600px at 10% 10%,
      rgba(45, 55, 72, 0.35),
      transparent
    ),
    radial-gradient(
      1200px 600px at 90% 90%,
      rgba(26, 32, 44, 0.35),
      transparent
    ),
    linear-gradient(135deg, #1a202c 0%, #2d3748 100%);
}

.glass-card {
  width: 100%;
  max-width: 420px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(12px);
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  padding: 28px;
  margin-bottom: 8vh;
}

html.dark .glass-card {
  background: rgba(45, 55, 72, 0.8);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
}

.brand {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 20px;
}

.logo {
  font-weight: 800;
  font-size: 28px;
  color: #5a67d8;
}

html.dark .logo {
  color: #63b3ed;
}

.title {
  font-size: 18px;
  color: #4a5568;
}

html.dark .title {
  color: #e2e8f0;
}

.submit-btn {
  width: 100%;
  margin-top: 4px;
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
  border: 1px solid #e2e8f0;
}
</style>

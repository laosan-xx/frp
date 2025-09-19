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
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'

interface LoginForm {
  username: string
  password: string
}

const router = useRouter()
const route = useRoute()
const formRef = ref()
const loading = ref(false)
const form = reactive<LoginForm>({ username: '', password: '' })

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    try {
      loading.value = true
      // 调用后端登录接口，示例使用 fetch，请按需替换 URL
      const resp = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      })
      if (!resp.ok) {
        throw new Error('登录失败')
      }
      // 这里假设后端设置了 cookie 或返回 token；项目无需复杂状态
      const redirect = (route.query.redirect as string) || '/'
      router.replace(redirect)
    } catch (e) {
      ElMessage.error('用户名或密码错误')
    } finally {
      loading.value = false
    }
  })
}
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

.footer {
  margin-top: 14px;
  text-align: center;
  font-size: 12px;
  color: #a0aec0;
}

html.dark .footer {
  color: #cbd5e0;
}
</style>

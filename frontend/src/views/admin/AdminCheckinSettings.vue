<template>
  <AppLayout>
    <div class="mx-auto max-w-3xl space-y-6">
      <div
        v-if="toast"
        class="fixed top-4 left-1/2 z-50 -translate-x-1/2 rounded-lg px-4 py-2 text-sm text-white shadow-lg"
        :class="toastError ? 'bg-red-500' : 'bg-green-500'"
      >
        {{ toast }}
      </div>

      <!-- settings -->
      <div class="card">
        <div class="p-6">
          <h2 class="mb-6 text-xl font-semibold text-gray-800 dark:text-gray-100">签到设置</h2>

          <div v-if="form" class="space-y-5">
            <label class="flex items-center justify-between">
              <span class="text-sm text-gray-700 dark:text-gray-300">开启签到功能</span>
              <input v-model="form.enabled" type="checkbox" class="h-5 w-5" />
            </label>

            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-700 dark:text-gray-300">每次基础额度</span>
              <input
                v-model.number="form.base_amount"
                type="number"
                step="0.01"
                min="0"
                class="w-40 rounded-lg border px-3 py-2 dark:border-gray-600 dark:bg-gray-700"
              />
            </div>

            <label class="flex items-center justify-between">
              <span class="text-sm text-gray-700 dark:text-gray-300">开启连续签到加成</span>
              <input v-model="form.consecutive_bonus" type="checkbox" class="h-5 w-5" />
            </label>

            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-700 dark:text-gray-300">每连续一天额外加成</span>
              <input
                v-model.number="form.bonus_per_day"
                type="number"
                step="0.01"
                min="0"
                :disabled="!form.consecutive_bonus"
                class="w-40 rounded-lg border px-3 py-2 disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700"
              />
            </div>

            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-700 dark:text-gray-300">最多加成天数</span>
              <input
                v-model.number="form.max_bonus_days"
                type="number"
                step="1"
                min="0"
                :disabled="!form.consecutive_bonus"
                class="w-40 rounded-lg border px-3 py-2 disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700"
              />
            </div>

            <div class="pt-2">
              <button
                class="rounded-lg bg-primary-500 px-6 py-2 font-semibold text-white hover:bg-primary-600 disabled:opacity-50"
                :disabled="saving"
                @click="save"
              >
                {{ saving ? '保存中...' : '保存设置' }}
              </button>
            </div>
          </div>
          <div v-else class="py-6 text-sm text-gray-400">加载中...</div>
        </div>
      </div>

      <!-- logs -->
      <div class="card">
        <div class="p-6">
          <h3 class="mb-4 text-base font-semibold text-gray-700 dark:text-gray-200">签到记录</h3>
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b text-left text-gray-500 dark:border-gray-700">
                <th class="py-2">用户ID</th>
                <th class="py-2">日期</th>
                <th class="py-2">连续天数</th>
                <th class="py-2 text-right">额度</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="logs.length === 0">
                <td colspan="4" class="py-6 text-center text-gray-400">暂无记录</td>
              </tr>
              <tr
                v-for="item in logs"
                :key="item.id"
                class="border-b last:border-0 dark:border-gray-700"
              >
                <td class="py-2">{{ item.user_id }}</td>
                <td class="py-2">{{ formatDate(item.checkin_date) }}</td>
                <td class="py-2">{{ item.consecutive_days }}</td>
                <td class="py-2 text-right text-green-500">+{{ formatAmount(item.amount) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import {
  adminGetCheckinSettings,
  adminUpdateCheckinSettings,
  adminGetCheckinLogs,
  type CheckinSettings,
  type CheckinLogItem,
} from '@/api/checkin'

const form = ref<CheckinSettings | null>(null)
const logs = ref<CheckinLogItem[]>([])
const saving = ref(false)
const toast = ref('')
const toastError = ref(false)

function showToast(msg: string, isError = false): void {
  toast.value = msg
  toastError.value = isError
  window.setTimeout(() => {
    toast.value = ''
  }, 3000)
}

function formatAmount(v: number): string {
  return v.toFixed(4).replace(/\.?0+$/, '') || '0'
}

function formatDate(v: string): string {
  return v.length >= 10 ? v.slice(0, 10) : v
}

async function load(): Promise<void> {
  try {
    form.value = await adminGetCheckinSettings()
  } catch (err: any) {
    showToast(err?.response?.data?.message || '加载设置失败', true)
  }
  try {
    const res = await adminGetCheckinLogs({ page: 1, page_size: 20 })
    logs.value = res.items
  } catch {
    // ignore
  }
}

async function save(): Promise<void> {
  if (!form.value || saving.value) return
  saving.value = true
  try {
    await adminUpdateCheckinSettings(form.value)
    showToast('保存成功')
  } catch (err: any) {
    showToast(err?.response?.data?.message || '保存失败', true)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

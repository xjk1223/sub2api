<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- toast -->
      <div
        v-if="toast"
        class="fixed top-4 left-1/2 z-50 -translate-x-1/2 rounded-lg px-4 py-2 text-sm text-white shadow-lg"
        :class="toastError ? 'bg-red-500' : 'bg-green-500'"
      >
        {{ toast }}
      </div>

      <!-- check-in card -->
      <div class="card">
        <div class="p-6">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-100">每日签到</h2>
            <span
              v-if="status && status.consecutive_days > 1"
              class="rounded-full bg-orange-100 px-3 py-1 text-sm font-medium text-orange-600"
            >
              连续 {{ status.consecutive_days }} 天
            </span>
          </div>

          <div v-if="!status?.enabled" class="py-8 text-center text-gray-400">
            签到功能暂未开启
          </div>

          <template v-else>
            <div class="my-6 flex items-center justify-center">
              <div class="text-center">
                <div class="text-4xl font-bold text-primary-600 dark:text-primary-400">
                  +{{ formatAmount(status?.today_amount ?? 0) }}
                </div>
                <div class="mt-1 text-sm text-gray-500">
                  {{ status?.checked_in_today ? '今日已获得额度' : '签到可获得额度' }}
                </div>
              </div>
            </div>

            <button
              class="w-full rounded-xl py-3 font-semibold text-white transition-all"
              :class="
                status?.checked_in_today
                  ? 'cursor-not-allowed bg-gray-300 dark:bg-gray-600'
                  : 'bg-primary-500 hover:bg-primary-600 active:scale-95'
              "
              :disabled="status?.checked_in_today || loading"
              @click="handleCheckin"
            >
              <span v-if="loading">签到中...</span>
              <span v-else-if="status?.checked_in_today">今日已签到</span>
              <span v-else>立即签到</span>
            </button>
          </template>
        </div>
      </div>

      <!-- history card -->
      <div class="card">
        <div class="p-6">
          <h3 class="mb-4 text-base font-semibold text-gray-700 dark:text-gray-200">签到记录</h3>
          <div v-if="history.length === 0" class="py-6 text-center text-sm text-gray-400">
            暂无签到记录
          </div>
          <ul v-else class="divide-y divide-gray-100 dark:divide-gray-700">
            <li
              v-for="item in history"
              :key="item.id"
              class="flex items-center justify-between py-3"
            >
              <div>
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ formatDate(item.checkin_date) }}
                </span>
                <span v-if="item.consecutive_days > 1" class="ml-2 text-xs text-orange-500">
                  连续 {{ item.consecutive_days }} 天
                </span>
              </div>
              <span class="text-sm font-semibold text-green-500">
                +{{ formatAmount(item.amount) }}
              </span>
            </li>
          </ul>
          <div v-if="total > history.length" class="mt-4 text-center">
            <button class="text-sm text-primary-500 hover:underline" @click="loadMore">
              加载更多
            </button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import {
  getCheckinStatus,
  doCheckin,
  getCheckinHistory,
  type CheckinStatus,
  type CheckinLogItem,
} from '@/api/checkin'

const loading = ref(false)
const status = ref<CheckinStatus | null>(null)
const history = ref<CheckinLogItem[]>([])
const total = ref(0)
const page = ref(1)
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

async function loadStatus(): Promise<void> {
  try {
    status.value = await getCheckinStatus()
  } catch {
    // ignore
  }
}

async function loadHistory(): Promise<void> {
  try {
    const res = await getCheckinHistory(1, 10)
    history.value = res.items
    total.value = res.total
    page.value = 1
  } catch {
    // ignore
  }
}

async function loadMore(): Promise<void> {
  page.value += 1
  const res = await getCheckinHistory(page.value, 10)
  history.value.push(...res.items)
}

async function handleCheckin(): Promise<void> {
  if (loading.value || status.value?.checked_in_today) return
  loading.value = true
  try {
    const res = await doCheckin()
    showToast(
      `签到成功！获得 ${formatAmount(res.amount)} 额度，连续 ${res.consecutive_days} 天，当前余额 ${formatAmount(res.new_balance)}`,
    )
    await Promise.all([loadStatus(), loadHistory()])
  } catch (err: any) {
    showToast(err?.response?.data?.message || '签到失败，请稍后重试', true)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadStatus()
  void loadHistory()
})
</script>

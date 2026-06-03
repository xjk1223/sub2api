/**
 * Daily check-in API endpoints.
 */
import { apiClient } from './client'

export interface CheckinStatus {
  enabled: boolean
  checked_in_today: boolean
  consecutive_days: number
  today_amount: number
  last_checkin_date?: string
}

export interface CheckinResult {
  amount: number
  consecutive_days: number
  new_balance: number
}

export interface CheckinLogItem {
  id: number
  user_id?: number
  amount: number
  consecutive_days: number
  checkin_date: string
  created_at: string
}

export interface CheckinSettings {
  enabled: boolean
  base_amount: number
  consecutive_bonus: boolean
  bonus_per_day: number
  max_bonus_days: number
}

export interface CheckinPage {
  items: CheckinLogItem[]
  total: number
  page: number
  page_size: number
  pages: number
}

// ---- user ----
export async function getCheckinStatus(): Promise<CheckinStatus> {
  const { data } = await apiClient.get<CheckinStatus>('/user/checkin/status')
  return data
}

export async function doCheckin(): Promise<CheckinResult> {
  const { data } = await apiClient.post<CheckinResult>('/user/checkin')
  return data
}

export async function getCheckinHistory(page = 1, pageSize = 10): Promise<CheckinPage> {
  const { data } = await apiClient.get<CheckinPage>('/user/checkin/history', {
    params: { page, page_size: pageSize },
  })
  return data
}

// ---- admin ----
export async function adminGetCheckinSettings(): Promise<CheckinSettings> {
  const { data } = await apiClient.get<CheckinSettings>('/admin/checkin/settings')
  return data
}

export async function adminUpdateCheckinSettings(payload: CheckinSettings): Promise<CheckinSettings> {
  const { data } = await apiClient.put<CheckinSettings>('/admin/checkin/settings', payload)
  return data
}

export async function adminGetCheckinLogs(params: {
  user_id?: number
  page?: number
  page_size?: number
}): Promise<CheckinPage> {
  const { data } = await apiClient.get<CheckinPage>('/admin/checkin/logs', { params })
  return data
}

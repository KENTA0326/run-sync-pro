<template>
  <div class="space-y-6">
    <div class="rounded-lg bg-white p-6 shadow-md">
      <h1 class="mb-2 text-xl font-bold text-gray-800">走行解析</h1>
      <p class="mb-4 text-sm text-gray-600">
        月別の走行距離・平均ペースを並列処理で集計しています。ログが増えても高速に結果を返します。
      </p>

      <div v-if="loading" class="py-8 text-center text-gray-500">
        解析中…
      </div>
      <div v-else-if="error" class="rounded bg-red-50 p-3 text-sm text-red-700">
        {{ error }}
      </div>
      <div v-else class="space-y-6">
        <!-- サマリーカード -->
        <div class="grid gap-4 sm:grid-cols-3">
          <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
            <p class="text-xs font-medium uppercase tracking-wide text-gray-500">全期間 総走行距離</p>
            <p class="mt-1 text-2xl font-bold text-gray-800">
              {{ analysis?.total_distance?.toFixed(1) ?? '0' }} <span class="text-base font-normal text-gray-600">km</span>
            </p>
          </div>
          <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
            <p class="text-xs font-medium uppercase tracking-wide text-gray-500">全期間 走行回数</p>
            <p class="mt-1 text-2xl font-bold text-gray-800">
              {{ analysis?.total_run_count ?? 0 }} <span class="text-base font-normal text-gray-600">回</span>
            </p>
          </div>
          <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
            <p class="text-xs font-medium uppercase tracking-wide text-gray-500">全期間 平均ペース</p>
            <p class="mt-1 text-2xl font-bold text-gray-800">
              {{ formatPace(overallAvgPaceSecPerKm) }}
            </p>
          </div>
        </div>

        <!-- VDOT推移グラフ -->
        <div v-if="analysis?.monthly_reports?.length" class="rounded-lg border border-gray-200 bg-gray-50/50 p-4">
          <h2 class="mb-3 text-lg font-semibold text-gray-800">VDOT推移</h2>
          <p class="mb-4 text-sm text-gray-600">
            月ごとの走行から算出したVDOT（走力指標）の推移です。最高VDOTはその月のベスト走、平均VDOTは月内の平均的な走力の目安です。
          </p>
          <ClientOnly>
            <VdotTrendChart v-if="vdotChartReports.length" :monthly-reports="vdotChartReports" />
            <p v-else class="py-4 text-center text-sm text-gray-500">VDOTを算出できる走行がまだありません（距離・タイムが有効なログが必要です）</p>
          </ClientOnly>
        </div>

        <!-- VDOT推移テーブル（グラフと同じデータ） -->
        <div v-if="analysis?.monthly_reports?.length && vdotChartReports.length" class="rounded-lg border border-gray-200 bg-white p-4">
          <h2 class="mb-3 text-lg font-semibold text-gray-800">VDOT推移データ</h2>
          <div class="overflow-x-auto rounded-lg border border-gray-200">
            <table class="min-w-full divide-y divide-gray-200 text-left text-sm">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-4 py-2 font-medium text-gray-700">月</th>
                  <th class="px-4 py-2 font-medium text-gray-700">距離 (km)</th>
                  <th class="px-4 py-2 font-medium text-gray-700">走行回数</th>
                  <th class="px-4 py-2 font-medium text-gray-700">平均ペース</th>
                  <th class="px-4 py-2 font-medium text-gray-700">平均VDOT</th>
                  <th class="px-4 py-2 font-medium text-gray-700">最高VDOT</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 bg-white">
                <tr
                  v-for="m in vdotChartReports"
                  :key="m.year_month"
                  class="hover:bg-gray-50"
                >
                  <td class="px-4 py-2 font-medium text-gray-800">{{ formatYearMonth(m.year_month) }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ m.total_distance.toFixed(1) }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ m.run_count }} 回</td>
                  <td class="px-4 py-2 text-gray-700">{{ formatPace(m.avg_pace_sec_per_km) }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ formatVdot(m.avg_vdot) }}</td>
                  <td class="px-4 py-2 font-medium text-gray-800">{{ formatVdot(m.max_vdot) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 月別レポート（全データ） -->
        <div>
          <h2 class="mb-3 text-lg font-semibold text-gray-800">月別レポート</h2>
          <div v-if="!analysis?.monthly_reports?.length" class="rounded border border-gray-200 bg-gray-50 p-4 text-center text-gray-500">
            走行ログがまだありません。走行ログを登録するとここに表示されます。
          </div>
          <div v-else class="overflow-x-auto rounded-lg border border-gray-200">
            <table class="min-w-full divide-y divide-gray-200 text-left text-sm">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-4 py-2 font-medium text-gray-700">月</th>
                  <th class="px-4 py-2 font-medium text-gray-700">距離 (km)</th>
                  <th class="px-4 py-2 font-medium text-gray-700">走行回数</th>
                  <th class="px-4 py-2 font-medium text-gray-700">平均ペース</th>
                  <th class="px-4 py-2 font-medium text-gray-700">平均VDOT</th>
                  <th class="px-4 py-2 font-medium text-gray-700">最高VDOT</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 bg-white">
                <tr
                  v-for="m in analysis.monthly_reports"
                  :key="m.year_month"
                  class="hover:bg-gray-50"
                >
                  <td class="px-4 py-2 font-medium text-gray-800">{{ formatYearMonth(m.year_month) }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ m.total_distance.toFixed(1) }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ m.run_count }} 回</td>
                  <td class="px-4 py-2 text-gray-700">{{ formatPace(m.avg_pace_sec_per_km) }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ formatVdot(m.avg_vdot) }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ formatVdot(m.max_vdot) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AnalysisResponse } from '~/types/api'

definePageMeta({
  middleware: 'auth',
})

const api = useApi()
const analysis = ref<AnalysisResponse | null>(null)
const loading = ref(true)
const error = ref('')

function formatPace(secPerKm: number): string {
  if (!secPerKm || Number.isNaN(secPerKm)) return '-'
  const sec = Math.round(secPerKm)
  const min = Math.floor(sec / 60)
  const s = sec % 60
  return `${min}:${s.toString().padStart(2, '0')}/km`
}

function formatYearMonth(ym: string): string {
  if (!ym || ym.length < 7) return ym
  const [y, m] = ym.split('-')
  const months = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
  const mi = parseInt(m, 10) - 1
  return `${y}年${months[mi] ?? m}`
}

function formatVdot(v: number): string {
  if (v == null || Number.isNaN(v) || v <= 0) return '-'
  return v.toFixed(1)
}

/** VDOTが有効な月だけ（グラフ・VDOTテーブル用） */
const vdotChartReports = computed(() => {
  const list = analysis.value?.monthly_reports ?? []
  return list.filter(m => m.max_vdot > 0 || m.avg_vdot > 0)
})

const overallAvgPaceSecPerKm = computed(() => {
  const a = analysis.value
  if (!a?.total_distance || !a?.total_duration || a.total_distance <= 0) return 0
  return a.total_duration / a.total_distance
})

async function fetchAnalysis() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<AnalysisResponse>('/auth/analysis')
    analysis.value = data
  } catch (err) {
    error.value = api.getErrorMessage(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchAnalysis()
})
</script>

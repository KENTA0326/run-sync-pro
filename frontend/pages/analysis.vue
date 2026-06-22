<template>
  <div class="space-y-6">
    <div class="rounded-lg bg-white p-6 shadow-md">
      <h1 class="mb-2 text-xl font-bold text-gray-800">ダッシュボード</h1>
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
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
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
          <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
            <p class="text-xs font-medium uppercase tracking-wide text-gray-500">推定完走タイム（フル）</p>
            <p class="mt-1 text-2xl font-bold text-gray-800">
              {{ formatDuration(fullMarathonEstimateLatestRaceSeconds) }}
            </p>
            <p class="mt-1 text-xs text-gray-500">
              <span v-if="latestRaceLog">
                直近レース（{{ formatDate(latestRaceLog.training_date) }} / VDOT {{ formatVdot(latestRaceVdot) }}）から算出
              </span>
              <span v-else>レース種別(kind=4)のログがあると算出されます</span>
            </p>
          </div>
        </div>

        <details class="rounded-lg border border-gray-200 bg-white p-4">
          <summary class="cursor-pointer list-none text-sm font-semibold text-gray-800">
            計算式を表示
          </summary>
          <div class="mt-3 space-y-1 rounded bg-gray-50 p-3 font-mono text-xs text-gray-700">
            <p>VDOT = (-4.60 + 0.182258*V + 0.000104*V^2) / (0.8 + 0.1894393*e^(-0.012778*T) + 0.2989558*e^(-0.1932605*T))</p>
            <p>V = distance_meters / (time_seconds / 60)</p>
            <p>PredictRaceTimeSeconds(vdot, 42195) は 2 分探索で VDOT が一致する秒数を求める</p>
          </div>
          <ul class="mt-3 list-disc space-y-1 pl-5 text-xs text-gray-600">
            <li><span class="font-mono">VDOT</span>: 走力指標（値が高いほど速い）</li>
            <li><span class="font-mono">V</span>: 速度（m/min）</li>
            <li><span class="font-mono">T</span>: タイム（分）</li>
            <li><span class="font-mono">distance_meters</span>: 距離（メートル）</li>
            <li><span class="font-mono">time_seconds</span>: タイム（秒）</li>
            <li><span class="font-mono">42195</span>: フルマラソン距離（m）</li>
            <li><span class="font-mono">e</span>: 自然対数の底（ネイピア数）</li>
            <li>
              分子・分母の係数（例: <span class="font-mono">0.000104</span>, <span class="font-mono">0.1894393</span>）は
              Jack Daniels の Running Formula で使われる経験式係数です
            </li>
          </ul>
          <div class="mt-3 rounded bg-blue-50 p-3 text-xs text-blue-900">
            <p class="font-semibold">係数がどうやって現れるか</p>
            <p class="mt-1">
              これらの係数は、レース実測データに対して「VDOTと速度・時間の関係」を最もよく表すように
              回帰フィットして得られた値です（理論的に手計算で導出する定数ではなく、データ近似で決まる定数）。
            </p>
            <p class="mt-1">
              そのため係数を変えるとモデル自体が別物になり、同じ走行データでもVDOTや推定タイムが変わります。
            </p>
          </div>
          <p class="mt-2 text-xs text-gray-500">
            参考: Jack Daniels' Running Formula（VDOT）, Riegel Formula（距離換算予測）
          </p>
          <p class="mt-2 text-xs text-gray-500">※ 上記は `backend/internal/domainservice/vdot.go` と同じ式・同じ探索方針です。</p>
        </details>

        <!-- VDOT推移グラフ -->
        <div v-if="analysis?.monthly_reports?.length" class="rounded-lg border border-gray-200 bg-gray-50/50 p-4">
          <h2 class="mb-3 text-lg font-semibold text-gray-800">VDOT推移</h2>
          <VdotScopeToggle v-model="vdotTrendScope" class="mb-3" />
          <p class="mb-4 text-sm text-gray-600">
            月ごとの走行から算出したVDOT（走力指標）の推移です。最高VDOTはその月のベスト走、平均VDOTは月内の平均的な走力の目安です。
            グラフの点をクリックすると、その月のデータを強調表示できます。
          </p>
          <p v-if="selectedVdotMonth" class="mb-3 text-sm text-blue-700">
            選択中: {{ formatYearMonth(selectedVdotMonth) }}
            <button
              type="button"
              class="ml-2 text-xs text-blue-600 underline hover:text-blue-800"
              @click="selectedVdotMonth = ''"
            >
              解除
            </button>
          </p>
          <ClientOnly>
            <VdotTrendChart
              v-if="displayedVdotReports.length"
              :monthly-reports="displayedVdotReports"
              @select-month="selectedVdotMonth = $event"
            />
            <p v-else class="py-4 text-center text-sm text-gray-500">VDOTを算出できる走行がまだありません（距離・タイムが有効なログが必要です）</p>
          </ClientOnly>
        </div>

        <!-- VDOT推移テーブル（グラフと同じデータ） -->
        <div v-if="analysis?.monthly_reports?.length && displayedVdotReports.length" class="rounded-lg border border-gray-200 bg-white p-4">
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
                  v-for="m in displayedVdotReports"
                  :key="m.year_month"
                  class="hover:bg-gray-50"
                  :class="selectedVdotMonth === m.year_month ? 'bg-blue-50 ring-1 ring-inset ring-blue-200' : ''"
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
import { apiPath } from '~/composables/apiPaths'
import type { AnalysisResponse, PaginatedTrainingLogsResponse, TrainingLog } from '~/types/api'

definePageMeta({
  middleware: 'auth',
})

const api = useApi()
const analysis = ref<AnalysisResponse | null>(null)
const recentRaceLogs = ref<TrainingLog[]>([])
const loading = ref(true)
const error = ref('')
const vdotTrendScope = ref<'all' | 'recent3'>('all')
const selectedVdotMonth = ref('')

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

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return dateStr.slice(0, 10)
}

function formatVdot(v: number): string {
  if (v == null || Number.isNaN(v) || v <= 0) return '-'
  return v.toFixed(1)
}

function formatDuration(totalSeconds: number): string {
  if (!totalSeconds || Number.isNaN(totalSeconds) || totalSeconds <= 0) return '-'
  const sec = Math.round(totalSeconds)
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

/** VDOTが有効な月だけ（グラフ・VDOTテーブル用） */
const vdotReportsAll = computed(() => {
  const list = analysis.value?.monthly_reports ?? []
  return list.filter(m => m.max_vdot > 0 || m.avg_vdot > 0)
})

function parseYearMonth(ym: string): number {
  const [y, m] = ym.split('-').map(v => parseInt(v, 10))
  if (!y || !m) return Number.NaN
  return y * 12 + (m - 1)
}

const vdotReportsRecent3Months = computed(() => {
  const list = vdotReportsAll.value
  if (!list.length) return []
  const latest = parseYearMonth(list[list.length - 1].year_month)
  if (Number.isNaN(latest)) return list.slice(-3)
  const minKey = latest - 2
  return list.filter(m => {
    const key = parseYearMonth(m.year_month)
    return !Number.isNaN(key) && key >= minKey && key <= latest
  })
})

const displayedVdotReports = computed(() => {
  return vdotTrendScope.value === 'recent3' ? vdotReportsRecent3Months.value : vdotReportsAll.value
})

const overallAvgPaceSecPerKm = computed(() => {
  const a = analysis.value
  if (!a?.total_distance || !a?.total_duration || a.total_distance <= 0) return 0
  return a.total_duration / a.total_distance
})

function calculateVDOTRaw(distanceMeters: number, timeSeconds: number): number {
  if (distanceMeters <= 0 || timeSeconds <= 0) return 0
  const timeMin = timeSeconds / 60.0
  const velocity = distanceMeters / timeMin
  const num = -4.60 + 0.182258 * velocity + 0.000104 * velocity * velocity
  const denom = 0.8 + 0.1894393 * Math.exp(-0.012778 * timeMin) + 0.2989558 * Math.exp(-0.1932605 * timeMin)
  if (denom <= 0) return 0
  const vdot = num / denom
  return vdot > 0 ? vdot : 0
}

function predictRaceTimeSeconds(vdot: number, distanceMeters: number): number {
  if (vdot <= 0 || distanceMeters <= 0) return 0

  let lo = 60
  let hi = distanceMeters * 3
  if (hi < 600) hi = 600
  if (distanceMeters >= 42195) hi = 8 * 60 * 60
  else if (distanceMeters >= 21097.5) hi = 5 * 60 * 60
  else if (distanceMeters >= 10000) hi = 3 * 60 * 60

  for (let i = 0; i < 20; i++) {
    const cur = calculateVDOTRaw(distanceMeters, hi)
    if (cur < vdot) break
    hi *= 1.5
    if (hi > 12 * 60 * 60) {
      hi = 12 * 60 * 60
      break
    }
  }

  for (let i = 0; i < 60; i++) {
    const mid = (lo + hi) / 2
    const cur = calculateVDOTRaw(distanceMeters, mid)
    if (cur <= 0) {
      hi = mid
      continue
    }
    if (cur > vdot) lo = mid
    else hi = mid
  }
  const sec = Math.round(hi)
  return sec > 0 ? sec : 0
}

const latestRaceLog = computed(() => {
  return recentRaceLogs.value.find(l => l.kind === 4) ?? null
})

const latestRaceVdot = computed(() => {
  const race = latestRaceLog.value
  if (!race || race.distance <= 0 || race.duration <= 0) return 0
  return calculateVDOTRaw(race.distance * 1000, race.duration)
})

const fullMarathonEstimateLatestRaceSeconds = computed(() => {
  if (!latestRaceVdot.value) return 0
  return predictRaceTimeSeconds(latestRaceVdot.value, 42195)
})

async function fetchAnalysis() {
  loading.value = true
  error.value = ''
  try {
    const [analysisData, logsData] = await Promise.all([
      api.get<AnalysisResponse>(apiPath.analysisMonthly),
      api.get<PaginatedTrainingLogsResponse>(apiPath.trainingLogs, { limit: 500 }),
    ])
    analysis.value = analysisData
    recentRaceLogs.value = logsData.items
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

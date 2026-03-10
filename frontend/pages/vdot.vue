<template>
  <div class="space-y-6">
    <div class="rounded-lg bg-white p-6 shadow-md">
      <h1 class="mb-2 text-xl font-bold text-gray-800">VDOT 計算</h1>
      <p class="mb-4 text-sm text-gray-600">
        レースの距離とタイムから、Jack Daniels の公式でVDOTとトレーニングペースを算出します（3km以上のレース結果推奨）。
      </p>

      <div class="mb-4 grid gap-4 sm:grid-cols-2">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">距離 (km)</label>
          <input
            v-model.number="distanceKm"
            type="number"
            step="0.1"
            min="0.1"
            placeholder="例: 5"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">タイム (分:秒)</label>
          <div class="flex gap-2">
            <input
              v-model.number="timeMin"
              type="number"
              min="0"
              placeholder="分"
              class="w-20 rounded border border-gray-300 p-2 text-gray-800"
            />
            <span class="self-center text-gray-500">:</span>
            <input
              v-model.number="timeSec"
              type="number"
              min="0"
              max="59"
              placeholder="秒"
              class="w-20 rounded border border-gray-300 p-2 text-gray-800"
            />
          </div>
        </div>
      </div>

      <fieldset class="mb-4 rounded border border-gray-200 p-3">
        <legend class="px-1 text-sm font-medium text-gray-700">リーゲル公式の減衰率（指数）</legend>
        <div class="mt-2 grid gap-2 sm:grid-cols-3">
          <label class="flex cursor-pointer items-center gap-2 rounded px-2 py-1 hover:bg-gray-50">
            <input v-model.number="riegelExponent" type="radio" :value="1.06" />
            <span class="text-sm text-gray-700">1.06（超スタミナ型）</span>
          </label>
          <label class="flex cursor-pointer items-center gap-2 rounded px-2 py-1 hover:bg-gray-50">
            <input v-model.number="riegelExponent" type="radio" :value="1.08" />
            <span class="text-sm text-gray-700">1.08（標準）</span>
          </label>
          <label class="flex cursor-pointer items-center gap-2 rounded px-2 py-1 hover:bg-gray-50">
            <input v-model.number="riegelExponent" type="radio" :value="1.12" />
            <span class="text-sm text-gray-700">1.12（初心者/スピード型）</span>
          </label>
        </div>
      </fieldset>

      <button
        type="button"
        class="w-full rounded bg-blue-500 py-2 text-white hover:bg-blue-600"
        @click="calculate"
      >
        算出
      </button>

      <p v-if="error" class="mt-3 text-sm text-red-500">{{ error }}</p>
    </div>

    <div v-if="result" class="rounded-lg bg-white p-6 shadow-md">
      <h2 class="mb-4 text-lg font-bold text-gray-800">結果</h2>
      <p class="mb-4 text-2xl font-bold text-blue-600">
        あなたのVDOTは <span class="text-3xl">{{ result.vdot }}</span> です
      </p>

      <div class="mb-6 rounded bg-gray-50 p-4">
        <p class="mb-2 text-sm font-semibold text-gray-800">
          予想タイム（リーゲル換算: 指数 {{ result.riegel_exponent.toFixed(2) }}）
        </p>
        <table class="w-full border-collapse text-sm">
          <thead>
            <tr class="border-b border-gray-200">
              <th class="p-2 text-left font-medium text-gray-700">距離</th>
              <th class="p-2 text-left font-medium text-gray-700">予想タイム</th>
            </tr>
          </thead>
          <tbody class="text-gray-800">
            <tr class="border-b border-gray-100">
              <td class="p-2">フルマラソン</td>
              <td class="p-2 font-mono">{{ formatDuration(result.riegel_predictions.full_seconds) }}</td>
            </tr>
            <tr class="border-b border-gray-100">
              <td class="p-2">ハーフ</td>
              <td class="p-2 font-mono">{{ formatDuration(result.riegel_predictions.half_seconds) }}</td>
            </tr>
            <tr class="border-b border-gray-100">
              <td class="p-2">10km</td>
              <td class="p-2 font-mono">{{ formatDuration(result.riegel_predictions.ten_k_seconds) }}</td>
            </tr>
            <tr class="border-b border-gray-100">
              <td class="p-2">5km</td>
              <td class="p-2 font-mono">{{ formatDuration(result.riegel_predictions.five_k_seconds) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <table class="w-full border-collapse text-sm">
        <thead>
          <tr class="border-b border-gray-200 bg-gray-50">
            <th class="p-2 text-left font-medium text-gray-700">強度</th>
            <th class="p-2 text-left font-medium text-gray-700">推奨ペース (/km)</th>
          </tr>
        </thead>
        <tbody class="text-gray-800">
          <tr class="border-b border-gray-100">
            <td class="p-2">
              <button
                type="button"
                class="w-full rounded px-2 py-1.5 text-left font-medium text-gray-800 transition-colors hover:bg-gray-100"
                @click="selectIntensity('E')"
              >
                ジョグ (Easy)
              </button>
            </td>
            <td class="p-2 font-mono">{{ formatPace(result.paces.easy_min_sec_per_km) }} ～ {{ formatPace(result.paces.easy_max_sec_per_km) }}</td>
          </tr>
          <tr class="border-b border-gray-100">
            <td class="p-2">
              <button
                type="button"
                class="w-full rounded px-2 py-1.5 text-left font-medium text-gray-800 transition-colors hover:bg-gray-100"
                @click="selectIntensity('M')"
              >
                マラソン (M)
              </button>
            </td>
            <td class="p-2 font-mono">{{ formatPace(result.paces.marathon_sec_per_km) }}</td>
          </tr>
          <tr class="border-b border-gray-100">
            <td class="p-2">
              <button
                type="button"
                class="w-full rounded px-2 py-1.5 text-left font-medium text-gray-800 transition-colors hover:bg-gray-100"
                @click="selectIntensity('T')"
              >
                閾値走 (Threshold)
              </button>
            </td>
            <td class="p-2 font-mono">{{ formatPace(result.paces.threshold_sec_per_km) }}</td>
          </tr>
          <tr class="border-b border-gray-100">
            <td class="p-2">
              <button
                type="button"
                class="w-full rounded px-2 py-1.5 text-left font-medium text-gray-800 transition-colors hover:bg-gray-100"
                @click="selectIntensity('I')"
              >
                インターバル (Interval)
              </button>
            </td>
            <td class="p-2 font-mono">{{ formatPace(result.paces.interval_sec_per_km) }}</td>
          </tr>
          <tr class="border-b border-gray-100">
            <td class="p-2">
              <button
                type="button"
                class="w-full rounded px-2 py-1.5 text-left font-medium text-gray-800 transition-colors hover:bg-gray-100"
                @click="selectIntensity('R')"
              >
                レペティション (R)
              </button>
            </td>
            <td class="p-2 font-mono">{{ formatPace(result.paces.repetition_sec_per_km) }}</td>
          </tr>
        </tbody>
      </table>

      <div ref="intensitySectionRef" class="mt-6 rounded-lg bg-white p-6 shadow-md">
        <h3 class="mb-2 text-lg font-bold text-gray-800">強度の解説</h3>
        <p v-if="!hasClickedIntensity" class="mb-4 text-sm text-gray-600">
          上の表の「強度」をクリックすると、ここに説明が表示されます。
        </p>

        <div v-if="selectedIntensity" class="space-y-4">
          <div class="rounded bg-gray-50 p-4">
            <p class="text-base font-bold text-gray-800">{{ selectedIntensity.label }}</p>
            <p class="mt-1 text-sm text-gray-600">{{ selectedIntensity.title }}</p>
          </div>

          <div v-if="selectedKey === 'M'">
            <p class="mb-2 text-sm font-semibold text-gray-800">フルマラソンのスプリット（1kmごと）</p>
            <p class="mb-3 text-sm text-gray-600">
              マラソン（M）ペース（{{ formatPace(result?.paces.marathon_sec_per_km ?? 0) }} /km）で走った場合の通過タイムです。
              10kmごとにページ切り替えできます。
            </p>

            <div class="rounded border border-gray-200 bg-white">
              <div class="flex items-center justify-between gap-2 border-b border-gray-200 px-3 py-2">
                <button
                  type="button"
                  class="rounded px-3 py-1 text-sm text-gray-700 hover:bg-gray-100 disabled:opacity-40"
                  :disabled="!splits || splits.page <= 1 || splitsLoading"
                  @click="fetchFullSplits((splits?.page ?? 2) - 1)"
                >
                  前へ
                </button>
                <p class="text-sm text-gray-700">
                  {{ splits ? `ページ ${splits.page} / ${splits.total_pages}` : '—' }}
                </p>
                <button
                  type="button"
                  class="rounded px-3 py-1 text-sm text-gray-700 hover:bg-gray-100 disabled:opacity-40"
                  :disabled="!splits || splits.page >= splits.total_pages || splitsLoading"
                  @click="fetchFullSplits((splits?.page ?? 0) + 1)"
                >
                  次へ
                </button>
              </div>

              <div v-if="splitsError" class="px-3 py-2 text-sm text-red-600">
                {{ splitsError }}
              </div>
              <div v-else-if="splitsLoading" class="px-3 py-2 text-sm text-gray-600">
                読み込み中...
              </div>
              <table v-else-if="splits" class="w-full border-collapse text-sm">
                <thead>
                  <tr class="border-b border-gray-200 bg-gray-50">
                    <th class="p-2 text-left font-medium text-gray-700">距離</th>
                    <th class="p-2 text-left font-medium text-gray-700">通過タイム</th>
                  </tr>
                </thead>
                <tbody class="text-gray-800">
                  <tr v-for="row in splits.rows" :key="`${row.label}-${row.km}`" class="border-b border-gray-100">
                    <td class="p-2">{{ row.label }}</td>
                    <td class="p-2 font-mono">{{ formatDuration(row.cumulative_seconds) }}</td>
                  </tr>
                </tbody>
              </table>

              <div v-else class="px-3 py-2 text-sm text-gray-600">
                まずは「マラソン（M）」をクリックしてください。
              </div>
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-semibold text-gray-800">意味・目的</p>
            <ul class="list-disc space-y-1 pl-5 text-sm text-gray-700">
              <li v-for="(p, idx) in selectedIntensity.purpose" :key="idx">{{ p }}</li>
            </ul>
          </div>

          <div>
            <p class="mb-2 text-sm font-semibold text-gray-800">推奨練習量の目安</p>
            <ul class="list-disc space-y-1 pl-5 text-sm text-gray-700">
              <li v-for="(g, idx) in selectedIntensity.guideline" :key="idx">{{ g }}</li>
            </ul>
          </div>

          <div>
            <p class="mb-2 text-sm font-semibold text-gray-800">メニュー例</p>
            <ul class="list-disc space-y-1 pl-5 text-sm text-gray-700">
              <li v-for="(e, idx) in selectedIntensity.examples" :key="idx">{{ e }}</li>
            </ul>
          </div>
        </div>

        <div v-else class="rounded bg-gray-50 p-4 text-sm text-gray-600">
          まずは「強度」をクリックしてください。
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {
  FullMarathonSplitsRequest,
  FullMarathonSplitsResponse,
  VDOTCalculateRequest,
  VDOTCalculateResponse,
} from '~/types/api'
import { TRAINING_INTENSITY_MAP, type TrainingIntensityKey } from '~/data/training-intensities'

definePageMeta({
  middleware: 'auth',
})

const api = useApi()
const distanceKm = ref<number>(5)
const timeMin = ref<number>(20)
const timeSec = ref<number>(0)
const riegelExponent = ref<number>(1.08)
const error = ref('')
const result = ref<VDOTCalculateResponse | null>(null)
const selectedKey = ref<TrainingIntensityKey | null>(null)
const intensitySectionRef = ref<HTMLElement | null>(null)
const hasClickedIntensity = ref(false)
const splits = ref<FullMarathonSplitsResponse | null>(null)
const splitsLoading = ref(false)
const splitsError = ref('')

const selectedIntensity = computed(() => {
  if (!selectedKey.value) return null
  return TRAINING_INTENSITY_MAP[selectedKey.value]
})

function formatPace(secPerKm: number): string {
  if (!secPerKm || !Number.isFinite(secPerKm)) return '-'
  const m = Math.floor(secPerKm / 60)
  const s = Math.round(secPerKm % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

function formatDuration(totalSeconds: number): string {
  if (!totalSeconds || !Number.isFinite(totalSeconds) || totalSeconds <= 0) return '-'
  const sec = Math.round(totalSeconds)
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  return `${m}:${s.toString().padStart(2, '0')}`
}

function selectIntensity(key: TrainingIntensityKey) {
  hasClickedIntensity.value = true
  selectedKey.value = key
  if (key === 'M') {
    void fetchFullSplits(1)
  }
  // レイアウトのヘッダー分を考慮しつつスクロール
  if (import.meta.client) {
    requestAnimationFrame(() => {
      intensitySectionRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    })
  }
}

async function fetchFullSplits(page: number) {
  if (!result.value) return
  splitsError.value = ''
  splitsLoading.value = true
  try {
    const body: FullMarathonSplitsRequest = {
      pace_sec_per_km: result.value.paces.marathon_sec_per_km,
      page,
    }
    const data = await api.post<FullMarathonSplitsResponse>('/splits/fullmarathon', body)
    splits.value = data
  } catch (err) {
    splitsError.value = api.getErrorMessage(err)
  } finally {
    splitsLoading.value = false
  }
}

async function calculate() {
  error.value = ''
  result.value = null
  splits.value = null
  splitsError.value = ''
  const dist = distanceKm.value
  const min = timeMin.value ?? 0
  const sec = timeSec.value ?? 0
  if (!dist || dist <= 0) {
    error.value = '距離を入力してください'
    return
  }
  const timeSeconds = min * 60 + sec
  if (timeSeconds <= 0) {
    error.value = 'タイムを入力してください'
    return
  }
  try {
    const body: VDOTCalculateRequest = {
      distance_meters: Math.round(dist * 1000),
      time_seconds: timeSeconds,
      riegel_exponent: riegelExponent.value,
    }
    const data = await api.post<VDOTCalculateResponse>('/vdot/calculate', body)
    result.value = data
  } catch (err) {
    error.value = api.getErrorMessage(err)
  }
}
</script>

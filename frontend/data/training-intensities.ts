export type TrainingIntensityKey = 'E' | 'M' | 'T' | 'I' | 'R'

export type TrainingIntensityDetail = {
  key: TrainingIntensityKey
  label: string
  title: string
  purpose: string[]
  guideline: string[]
  examples: string[]
}

export const TRAINING_INTENSITIES: TrainingIntensityDetail[] = [
  {
    key: 'E',
    label: 'ジョグ (Easy)',
    title: 'Easy（ジョグ）',
    purpose: [
      '回復と土台作り（有酸素能力の底上げ）',
      'フォームを崩さずに走行距離を積む',
      '疲労を溜めすぎずに日々の練習を継続する',
    ],
    guideline: [
      '会話できる強度（息が上がりすぎない）',
      '時間目安: 30〜90分（ロングは90〜150分でも可）',
      'きつければ「ゆっくり長く」が正解。ペースより継続を優先',
    ],
    examples: [
      '45分イージー + 流し100m×4（余裕がある日）',
      '60〜90分のLSD（とにかく楽に）',
      '前日のポイント練習の翌日は30〜45分の回復ジョグ',
    ],
  },
  {
    key: 'M',
    label: 'マラソン (M)',
    title: 'M（マラソンペース）',
    purpose: [
      'マラソン本番に近いペースでの「持続力」を高める',
      '補給・フォーム・リズムを実戦レベルで練習する',
      'レース後半の粘り（脚づくり）',
    ],
    guideline: [
      '時間目安: 30〜120分（初心者は短めから）',
      '週1回のポイント練習の一部として入れるのが安全',
      '疲労が強い週はやらない（故障予防）',
    ],
    examples: [
      '20分E + 40分M + 10分E',
      'M 8〜16km（余裕が残る範囲で）',
      'ロング走の後半だけMに上げる（例: 20kmE + 6kmM）',
    ],
  },
  {
    key: 'T',
    label: '閾値走 (Threshold)',
    title: 'T（閾値走/テンポ走）',
    purpose: [
      '乳酸が溜まり始める手前の強度を鍛え、長く速く走れる体を作る',
      '「きついけど制御できる」感覚の習得（レースペースの芯）',
    ],
    guideline: [
      '時間目安: 20〜40分の「連続」または分割（クルーズインターバル）',
      '分割する場合は休憩を短く（例: 1分ジョグ）',
      '終盤にフォームが崩れるほど上げすぎない',
    ],
    examples: [
      'T 20分 1本（慣れてきたら30分）',
      'T 10分×3（つなぎ1分ジョグ）',
      'T 2km×4（つなぎ60〜90秒）',
    ],
  },
  {
    key: 'I',
    label: 'インターバル (Interval)',
    title: 'I（インターバル/VO₂max）',
    purpose: [
      '最大酸素摂取量（VO₂max）付近の刺激で心肺能力を引き上げる',
      '高い強度でもフォームを保つ練習',
    ],
    guideline: [
      '1本あたり: 2〜5分（400m〜1200mが目安）',
      'レスト（つなぎ）は「同じくらいの時間」か少し短めの軽いジョグ',
      '合計の速い時間: 12〜20分程度（やりすぎ注意）',
    ],
    examples: [
      'I 1000m×5（つなぎ2〜3分ジョグ）',
      'I 800m×6（つなぎ2分ジョグ）',
      'I 3分×6（つなぎ2分ジョグ）',
    ],
  },
  {
    key: 'R',
    label: 'レペティション (R)',
    title: 'R（レペティション/スピード）',
    purpose: [
      'スピードとフォーム（神経系）を鍛える',
      '効率よく速く走る「動きづくり」',
    ],
    guideline: [
      '1本あたり: 200〜400m（30〜90秒程度）',
      'レストは長め（完全に呼吸が整うまで）',
      '合計距離: 1.5〜3km程度（やりすぎない）',
    ],
    examples: [
      'R 200m×8（完全回復）',
      'R 400m×6（つなぎ200mジョグ or 2〜3分休憩）',
      '坂ダッシュ 10〜15秒×8（下りは完全回復）',
    ],
  },
]

export const TRAINING_INTENSITY_MAP: Record<TrainingIntensityKey, TrainingIntensityDetail> =
  TRAINING_INTENSITIES.reduce(
    (acc, cur) => {
      acc[cur.key] = cur
      return acc
    },
    {} as Record<TrainingIntensityKey, TrainingIntensityDetail>
  )


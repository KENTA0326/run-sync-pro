<template>
  <div class="min-h-screen font-sans overflow-x-hidden transition-colors duration-700" :class="pageColor">
    
    <main class="relative h-screen flex items-center justify-center">
      <div class="w-full max-w-5xl px-10">
        <transition name="slide-fade" mode="out-in">
          <div :key="currentSlide" class="text-center">
            <h2 class="text-4xl md:text-7xl font-black leading-tight tracking-tight" :class="textColor">
              {{ slides[currentSlide].pre }}
              <span class="inline-block text-sky-500 transform scale-125 mx-4">
                {{ slides[currentSlide].keyword }}
              </span>
              {{ slides[currentSlide].post }}
            </h2>
          </div>
        </transition>
      </div>

      <div v-if="currentSlide < 4" class="absolute bottom-10 left-1/2 -translate-x-1/2 flex flex-col items-center">
        <span class="text-[10px] font-black tracking-[0.2em] uppercase mb-2" :class="textColor">Scroll</span>
        <div class="w-[1px] h-16 bg-current animate-scroll-line" :class="textColor"></div>
      </div>
    </main>

    <section v-if="currentSlide === 4" class="fixed inset-0 bg-sky-500 text-white flex flex-col items-center justify-center z-10 animate-fade-in">
        <p class="text-lg font-bold tracking-widest mb-4 opacity-80">走るって、最高だ。</p>
        <h3 class="text-5xl md:text-8xl font-black mb-12 italic tracking-tighter">I LOVE RUNNING.</h3>
        <h4 class="text-3xl md:text-5xl font-extrabold mb-10">北海道マラソン 2026</h4>
        <div class="text-7xl md:text-[10rem] font-black leading-none tracking-tighter">08.30</div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';

// スライドデータ（ご提示の区切りに合わせました）
const slides = [
  /* ページ1のグループ */
  { pre: 'この', keyword: '達成感', post: 'がやめられない。' },
  { pre: '去年の自分よりも', keyword: '速く', post: '。' }, 
  /* ページ2のグループ */
  { pre: '初出場も', keyword: '大歓迎', post: '！' },
  { pre: '今年は友達も', keyword: '連れて', post: '。' },
  /* ページ3（最終ページ） */
  { pre: '遊びにいく', keyword: 'ノリ', post: 'で来てよ。' }
];

const currentSlide = ref(0);
let slideTimer: any = null;

// スライド状況に合わせて背景色や文字色を計算
const pageColor = computed(() => {
  if (currentSlide.value >= 4) return 'bg-sky-500'; // 最後のページ
  return 'bg-white'; // それ以外
});

const textColor = computed(() => {
  if (currentSlide.value >= 4) return 'text-white';
  return 'text-blue-950';
});

onMounted(() => {
  slideTimer = setInterval(() => {
    // スライドをループさせる（4秒ごとに切り替え）
    currentSlide.value = (currentSlide.value + 1) % slides.length;
  }, 4000);
});

onUnmounted(() => {
  if (slideTimer) clearInterval(slideTimer);
});
</script>

<style scoped>
/* スライド切り替え（文字の動き） */
.slide-fade-enter-active, .slide-fade-leave-active {
  transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}
.slide-fade-enter-from { opacity: 0; transform: translateY(40px); }
.slide-fade-leave-to { opacity: 0; transform: translateY(-40px); }

/* 最終ページのフェードイン */
.animate-fade-in {
  animation: fadeIn 1s ease-out forwards;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* スクロール線のアニメーション */
@keyframes scroll-line {
  0% { transform: scaleY(0); transform-origin: top; }
  50% { transform: scaleY(1); transform-origin: top; }
  51% { transform: scaleY(1); transform-origin: bottom; }
  100% { transform: scaleY(0); transform-origin: bottom; }
}
.animate-scroll-line {
  animation: scroll-line 2s infinite;
}
</style>
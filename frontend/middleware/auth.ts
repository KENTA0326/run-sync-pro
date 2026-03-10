export default defineNuxtRouteMiddleware(() => {
  const auth = useAuthStore()

  // localStorage を見る必要があるので、クライアント側で判定する
  if (import.meta.client && !auth.tokenOrStorage) {
    return navigateTo('/login')
  }
})


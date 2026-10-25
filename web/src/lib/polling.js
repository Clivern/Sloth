import { onBeforeUnmount, onMounted } from 'vue'

// Runs fn every ms while the tab is visible, and once right away when the tab comes back into view.
export function usePolling(fn, ms = 15000) {
  let timer = null

  function start() {
    stop()
    timer = setInterval(fn, ms)
  }

  function stop() {
    clearInterval(timer)
    timer = null
  }

  function onVisibility() {
    if (document.hidden) {
      stop()
      return
    }

    fn()
    start()
  }

  onMounted(() => {
    start()
    document.addEventListener('visibilitychange', onVisibility)
  })

  onBeforeUnmount(() => {
    stop()
    document.removeEventListener('visibilitychange', onVisibility)
  })
}

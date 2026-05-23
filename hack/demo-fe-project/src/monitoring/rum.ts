import armsRumModule from '@arms/rum-browser'

type RumEnv = 'prod' | 'gray' | 'pre' | 'daily' | 'local'

let initialized = false

const rumClient =
  (armsRumModule as unknown as { default?: typeof armsRumModule }).default || armsRumModule

function parseBooleanFlag(flag: string | undefined, fallback = false) {
  if (flag == null) {
    return fallback
  }
  return flag === 'true'
}

export function initRum() {
  const enabled = parseBooleanFlag(import.meta.env.VITE_ARMS_ENABLED, false)
  const endpoint = import.meta.env.VITE_ARMS_ENDPOINT

  if (!enabled) {
    return
  }

  if (!endpoint) {
    console.warn('[ARMS RUM] VITE_ARMS_ENDPOINT is missing, skip initialization.')
    return
  }

  if (initialized) {
    return
  }

  if (typeof rumClient.init !== 'function') {
    console.warn('[ARMS RUM] Invalid SDK instance: init() is not available.')
    return
  }

  rumClient.init({
    endpoint,
    env: (import.meta.env.VITE_ARMS_ENV || 'prod') as RumEnv,
    spaMode: 'history',
    collectors: {
      perf: true,
      webVitals: true,
      api: true,
      staticResource: true,
      jsError: true,
      consoleError: true,
      action: true,
    },
    tracing: false,
  })

  initialized = true
}

export const simulateRumErrors = {
  syncError() {
    setTimeout(() => {
      throw new Error('RUM_SYNC_ERROR: simulated sync runtime exception')
    }, 0)
  },
  unhandledRejection() {
    Promise.reject(new Error('RUM_UNHANDLED_REJECTION: simulated promise rejection'))
  },
  consoleError() {
    console.error('RUM_CONSOLE_ERROR: simulated console error')
  },
  apiFailure() {
    fetch('/__arms_mock_api_fail__', { method: 'POST' }).catch((error) => {
      console.error('RUM_API_FAILURE: simulated api failure', error)
    })
  },
  resourceFailure() {
    const image = new Image()
    image.src = `/__arms_mock_image_fail__-${Date.now()}.png`
  },
}

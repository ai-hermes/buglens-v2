import { readFileSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { rumVitePlugin, type IUserOptions } from '@arms/rum-vite-plugin'

function parseBooleanFlag(flag: string | undefined, fallback = false) {
  if (flag == null) {
    return fallback
  }
  return flag === 'true'
}

const packageJson = JSON.parse(
  readFileSync(fileURLToPath(new URL('./package.json', import.meta.url)), 'utf8'),
) as { version?: string }

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  const enableAutoUpload = parseBooleanFlag(env.ARMS_RUM_SOURCEMAP_AUTO_UPLOAD, false)
  const pid = env.ARMS_RUM_PID
  const region = env.ARMS_RUM_REGION || 'cn-hangzhou'
  const accessKeyId = env.ARMS_RUM_ACCESS_KEY_ID
  const accessKeySecret = env.ARMS_RUM_ACCESS_KEY_SECRET

  const baseRumOptions: IUserOptions = {
    version: packageJson.version || '1.0.0',
    silent: !enableAutoUpload,
  }

  const hasAutoUploadOptions = Boolean(pid && accessKeyId && accessKeySecret)
  const rumOptions: IUserOptions =
    enableAutoUpload && hasAutoUploadOptions
      ? {
          ...baseRumOptions,
          pid,
          region,
          accessKeyId,
          accessKeySecret,
          clearSourceMap: true,
        }
      : baseRumOptions

  if (enableAutoUpload && !hasAutoUploadOptions) {
    console.warn(
      '[ARMS RUM] ARMS_RUM_SOURCEMAP_AUTO_UPLOAD=true but missing required envs. Fallback to manual upload mode.',
    )
  }

  return {
    plugins: [react(), rumVitePlugin(rumOptions)],
    define: {
      global: 'globalThis',
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    css: {
      preprocessorOptions: {
        less: {
          javascriptEnabled: true,
        },
      },
    },
    build: {
      sourcemap: true,
    },
  }
})

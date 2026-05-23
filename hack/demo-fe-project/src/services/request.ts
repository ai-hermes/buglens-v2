import axios, { AxiosError } from 'axios'

type ApiErrorPayload = {
  message?: string
}

export const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000,
})

request.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiErrorPayload>) => {
    const message = error.response?.data?.message ?? error.message ?? 'Request failed'
    return Promise.reject(new Error(message))
  },
)

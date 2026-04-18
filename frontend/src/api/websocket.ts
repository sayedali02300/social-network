import { API_BASE_URL } from "@/api/api"

export const buildWebSocketURL = (path: string = '/ws') => {
  const url = new URL(API_BASE_URL)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = path
  return url.toString()
}
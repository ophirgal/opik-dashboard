import { useQuery } from '@tanstack/react-query'
import api from '../api/client'
import type { Metrics } from '../types/metrics'

export function useMetrics() {
  return useQuery({
    queryKey: ['metrics'],
    queryFn: () => api.get<Metrics>('/metrics').then((res) => res.data),
  })
}

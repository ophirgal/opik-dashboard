import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { DashboardPage } from './DashboardPage'
import api from '../api/client'
import type { Metrics } from '../types/metrics'

jest.mock('../api/client')
const mockedApi = api as jest.Mocked<typeof api>

const fixture: Metrics = {
  repo: 'comet-ml/opik',
  window_days: 90,
  generated_at: '2026-06-22T12:00:00Z',
  avg_time_to_address_pr_hours: 36,
  addressed_pr_count: 120,
  disengaged_pr_count: 3,
  disengaged_multiplier: 2,
  open_bug_issue_count: 7,
  avg_time_to_close_bug_hours: 60,
  closed_bug_count: 45,
}

function renderWithClient(ui: React.ReactElement) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>)
}

test('renders the four metrics from the API', async () => {
  mockedApi.get.mockResolvedValue({ data: fixture } as never)

  renderWithClient(<DashboardPage />)

  await waitFor(() => expect(screen.getByText('Disengaged PRs')).toBeInTheDocument())

  // Metric titles
  expect(screen.getByText('Avg time to address PR')).toBeInTheDocument()
  expect(screen.getByText('Open bug issues')).toBeInTheDocument()
  expect(screen.getByText('Avg time to close bug')).toBeInTheDocument()

  // Formatted values
  expect(screen.getByText('36.0 h')).toBeInTheDocument() // address time
  expect(screen.getByText('2.5 d')).toBeInTheDocument() // bug close time (60h)
  expect(screen.getByText('3')).toBeInTheDocument() // disengaged count
  expect(screen.getByText('7')).toBeInTheDocument() // open bugs

  // Header reflects the repo
  expect(screen.getByText(/comet-ml\/opik/)).toBeInTheDocument()
})

test('shows an error state when the API fails', async () => {
  mockedApi.get.mockRejectedValue(new Error('Bad Gateway'))

  renderWithClient(<DashboardPage />)

  await waitFor(() =>
    expect(screen.getByText(/Failed to load metrics/)).toBeInTheDocument(),
  )
})

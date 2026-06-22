import axios from 'axios'
import { GitPullRequest, AlertTriangle, Bug, Timer } from 'lucide-react'
import { useMetrics } from '../hooks/useMetrics'
import { MetricCard } from '../components/MetricCard'

function formatHours(hours: number): string {
  if (!hours || hours <= 0) return '—'
  if (hours >= 48) return `${(hours / 24).toFixed(1)} d`
  return `${hours.toFixed(1)} h`
}

// Prefer the API's `{ "error": ... }` body over Axios's generic status message.
function errorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const apiError = (error.response?.data as { error?: string } | undefined)?.error
    if (apiError) return apiError
  }
  return (error as Error)?.message ?? 'unknown error'
}

export function DashboardPage() {
  const { data, isLoading, isError, error } = useMetrics()

  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-6xl px-6 py-10">
        <header className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Opik Repo Health</h1>
          <p className="mt-1 text-muted-foreground">
            {data
              ? `${data.repo} · last ${data.window_days} days`
              : 'Repository metrics from GitHub'}
          </p>
        </header>

        {isLoading && <p className="text-muted-foreground">Loading metrics…</p>}

        {isError && (
          <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
            Failed to load metrics: {errorMessage(error)}
          </div>
        )}

        {data && (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <MetricCard
              title="Avg time to address PR"
              value={formatHours(data.avg_time_to_address_pr_hours)}
              subtitle={`${data.addressed_pr_count} PRs addressed`}
              icon={Timer}
            />
            <MetricCard
              title="Disengaged PRs"
              value={String(data.disengaged_pr_count)}
              subtitle={`open > ${data.disengaged_multiplier}× the average`}
              icon={GitPullRequest}
            />
            <MetricCard
              title="Open bug issues"
              value={String(data.open_bug_issue_count)}
              subtitle="issues prefixed [BUG]"
              icon={Bug}
            />
            <MetricCard
              title="Avg time to close bug"
              value={formatHours(data.avg_time_to_close_bug_hours)}
              subtitle={`${data.closed_bug_count} bugs closed`}
              icon={AlertTriangle}
            />
          </div>
        )}
      </div>
    </div>
  )
}

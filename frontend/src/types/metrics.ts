export interface Metrics {
  repo: string
  window_days: number
  generated_at: string

  avg_time_to_address_pr_hours: number
  addressed_pr_count: number
  disengaged_pr_count: number
  disengaged_multiplier: number

  open_bug_issue_count: number
  avg_time_to_close_bug_hours: number
  closed_bug_count: number
}

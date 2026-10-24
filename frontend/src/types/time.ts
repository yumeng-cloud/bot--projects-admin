export type TimeEntry = {
  id: string
  workspace_id: string
  issue_id: string
  member_id: string
  hours: number
  note: string
  billable: boolean
  rate?: number | null
  worked_at: string
  created_at?: string
  updated_at?: string
  project_id?: string
  issue_identifier?: string
  member_name?: string
}

export type BillingRate = {
  id: string
  workspace_id: string
  member_id?: string | null
  hourly_rate: number
  currency: string
  created_at?: string
  updated_at?: string
}

export type BillingSummaryRow = {
  group_key: string
  group_label: string
  hours: number
  billable_hours: number
  amount: number
  currency: string
  entries: number
}

export type BillingSummary = {
  from: string
  to: string
  group_by: string
  rows: BillingSummaryRow[]
  total_hours: number
  total_amount: number
  currency: string
}

export type TimeEntryCreate = {
  issue_id: string
  hours: number
  note?: string
  billable?: boolean
  rate?: number
  worked_at?: string
  member_id?: string
}

export type TimeEntryPatch = {
  hours?: number
  note?: string
  billable?: boolean
  rate?: number | null
  worked_at?: string
}

export type BillingRateUpsert = {
  member_id?: string | null
  hourly_rate: number
  currency?: string
}

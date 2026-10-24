export type WikiSpace = {
  id: string
  workspace_id: string
  name: string
  slug: string
  description: string
  created_at?: string
  updated_at?: string
}

export type WikiPage = {
  id: string
  space_id: string
  workspace_id: string
  parent_id?: string | null
  title: string
  slug: string
  body_md: string
  project_id?: string | null
  issue_id?: string | null
  created_by?: string | null
  updated_by?: string | null
  created_at?: string
  updated_at?: string
}

export type WikiTreeNode = {
  id: string
  title: string
  slug: string
  parent_id?: string | null
  children?: WikiTreeNode[]
}

import { useParams } from 'react-router-dom'
import { BoardPage } from './board'
import { IssueListPage } from './list'
import { CyclesPage } from './cycles'
import { TimelinePage } from './timeline'

export function BoardRoute() {
  const { wsId = '', projectId = '' } = useParams()
  return <BoardPage workspaceId={wsId} projectId={projectId} />
}

export function ListRoute() {
  const { wsId = '', projectId = '' } = useParams()
  return <IssueListPage workspaceId={wsId} projectId={projectId} />
}

export function CyclesRoute() {
  const { wsId = '', projectId = '' } = useParams()
  return <CyclesPage workspaceId={wsId} projectId={projectId} />
}

export function TimelineRoute() {
  const { wsId = '', projectId = '' } = useParams()
  return <TimelinePage workspaceId={wsId} projectId={projectId} />
}

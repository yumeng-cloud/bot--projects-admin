import { useParams } from 'react-router-dom'
import TimelinePage from './TimelinePage'

export default function TimelineRoute() {
  const { wsId = '', projectId = '' } = useParams()
  return <TimelinePage workspaceId={wsId} projectId={projectId} />
}

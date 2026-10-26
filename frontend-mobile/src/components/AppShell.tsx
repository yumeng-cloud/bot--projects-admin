import { Outlet } from 'react-router-dom'
import BottomNav from './BottomNav'

export default function AppShell() {
  return (
    <div className="m-app">
      <div className="m-app__body">
        <Outlet />
      </div>
      <BottomNav />
    </div>
  )
}

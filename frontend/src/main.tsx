import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider, App as AntApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import App from './App'
import { getLinearTheme, resolveThemeMode } from './theme'
import './index.css'
import './styles/page.css'
import './utils/time'

// Default product UI is light (Linear). Dark only if user toggled and saved.
const themeMode = resolveThemeMode('light')
document.documentElement.setAttribute('data-theme', themeMode)
document.documentElement.style.colorScheme = themeMode

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider locale={zhCN} theme={getLinearTheme(themeMode)}>
      <AntApp>
        <App />
      </AntApp>
    </ConfigProvider>
  </React.StrictMode>,
)

import type { ThemeConfig } from 'antd'
import { theme as antdTheme } from 'antd'

/** Linear-inspired brand indigo — single accent, not Ant default blue */
export const BRAND_PRIMARY = '#5E6AD2'

const sharedToken: ThemeConfig['token'] = {
  colorPrimary: BRAND_PRIMARY,
  borderRadius: 8,
  fontSize: 13,
  controlHeight: 28,
  fontFamily: 'var(--acb-font-sans)',
  lineWidth: 1,
  /* near-flat — edge via border, not elevation */
  boxShadow: 'none',
  boxShadowSecondary: '0 1px 2px rgba(0, 0, 0, 0.03)',
}

const sharedComponents: ThemeConfig['components'] = {
  Layout: {
    headerHeight: 48,
    headerPadding: '0 20px',
  },
  Menu: {
    itemHeight: 36,
    itemMarginInline: 8,
    itemBorderRadius: 6,
    iconSize: 15,
    itemPaddingInline: 12,
  },
  Card: {
    paddingLG: 20,
    paddingMD: 16,
    paddingSM: 12,
  },
  Table: {
    cellPaddingBlock: 8,
    cellPaddingInline: 12,
    headerBorderRadius: 8,
  },
  Button: {
    fontWeight: 450,
    paddingInline: 12,
    controlHeight: 28,
    borderRadius: 8,
    primaryShadow: 'none',
    defaultShadow: 'none',
  },
  Input: {
    controlHeight: 28,
    paddingBlock: 4,
  },
  Select: {
    controlHeight: 28,
  },
  Tag: {
    borderRadiusSM: 6,
    defaultBg: 'rgba(0, 0, 0, 0.04)',
  },
  /** Login + drawers: compact forms, thin chrome */
  Form: {
    itemMarginBottom: 14,
    verticalLabelPadding: '0 0 4px',
    labelFontSize: 12,
  },
  Tabs: {
    horizontalItemPadding: '8px 0',
    titleFontSize: 13,
    cardHeight: 36,
  },
  Drawer: {
    paddingLG: 16,
    paddingXS: 12,
  },
  Modal: {
    titleFontSize: 15,
    borderRadiusLG: 8,
  },
  Typography: {
    titleMarginBottom: '0.35em',
    titleMarginTop: '0.6em',
  },
  Divider: {
    marginLG: 16,
  },
}

/** Light: cool-gray canvas, sider chrome #F2F4F8, white work surface */
export const linearLightTheme: ThemeConfig = {
  algorithm: antdTheme.defaultAlgorithm,
  token: {
    ...sharedToken,
    colorBgLayout: '#F7F8FA',
    colorBgContainer: '#FFFFFF',
    colorBgElevated: '#FCFCFD',
    colorBorder: '#E6E8EE',
    colorBorderSecondary: '#EEF0F4',
    colorText: '#2A3140',
    colorTextSecondary: '#4A5263',
    colorTextTertiary: '#8A93A6',
    colorFillSecondary: '#F2F4F8',
    colorFillTertiary: '#F7F8FA',
    colorPrimaryBg: 'rgba(94, 106, 210, 0.10)',
    colorPrimaryBgHover: 'rgba(94, 106, 210, 0.16)',
    colorPrimaryBorder: 'rgba(94, 106, 210, 0.28)',
    colorPrimaryBorderHover: 'rgba(94, 106, 210, 0.42)',
  },
  components: {
    ...sharedComponents,
    Layout: {
      ...sharedComponents?.Layout,
      headerBg: '#FFFFFF',
      bodyBg: '#F7F8FA',
      siderBg: '#F2F4F8',
    },
    Menu: {
      ...sharedComponents?.Menu,
      itemBg: 'transparent',
      subMenuItemBg: 'transparent',
      /* soft indigo wash — not Ant solid blue block */
      itemSelectedBg: 'rgba(94, 106, 210, 0.12)',
      itemSelectedColor: BRAND_PRIMARY,
      itemHoverBg: 'rgba(42, 49, 64, 0.04)',
      itemActiveBg: 'rgba(94, 106, 210, 0.14)',
    },
    Button: {
      ...sharedComponents?.Button,
      /* keep default solid for rare CTAs; soft primary via CSS overrides */
      colorPrimary: BRAND_PRIMARY,
    },
    Tag: {
      ...sharedComponents?.Tag,
      defaultColor: '#4A5263',
    },
  },
}

/** Dark: shell toggles via localStorage; ConfigProvider reads at boot */
export const linearDarkTheme: ThemeConfig = {
  algorithm: antdTheme.darkAlgorithm,
  token: {
    ...sharedToken,
    colorBgLayout: '#0f1115',
    colorBgContainer: '#16181d',
    colorBgElevated: '#1c1f26',
    colorBorder: '#2a2e38',
    colorBorderSecondary: '#23262e',
    colorText: '#f0f1f4',
    colorTextSecondary: '#b0b5c3',
    colorTextTertiary: '#858b9a',
    colorPrimaryBg: 'rgba(94, 106, 210, 0.18)',
    colorPrimaryBgHover: 'rgba(94, 106, 210, 0.26)',
    colorPrimaryBorder: 'rgba(94, 106, 210, 0.35)',
  },
  components: {
    ...sharedComponents,
    Layout: {
      ...sharedComponents?.Layout,
      headerBg: '#16181d',
      bodyBg: '#0f1115',
      siderBg: '#16181d',
    },
    Menu: {
      ...sharedComponents?.Menu,
      itemBg: 'transparent',
      subMenuItemBg: 'transparent',
      itemSelectedBg: 'rgba(94, 106, 210, 0.18)',
      itemSelectedColor: '#a8b0f0',
      itemHoverBg: 'rgba(255, 255, 255, 0.04)',
      itemActiveBg: 'rgba(94, 106, 210, 0.24)',
    },
  },
}

/** localStorage key for shell-driven theme mode (light | dark) */
export const THEME_MODE_STORAGE_KEY = 'ai-collab-board:theme-mode-v2'

export type ThemeMode = 'light' | 'dark'

/** Product default is always light; dark only when user explicitly saved it. */
export function resolveThemeMode(fallback: ThemeMode = 'light'): ThemeMode {
  try {
    const v = localStorage.getItem(THEME_MODE_STORAGE_KEY)
    if (v === 'dark') return 'dark'
    if (v === 'light') return 'light'
  } catch {
    /* SSR / private mode */
  }
  return fallback === 'dark' ? 'light' : fallback
}

export function getLinearTheme(mode: ThemeMode = 'light'): ThemeConfig {
  return mode === 'dark' ? linearDarkTheme : linearLightTheme
}

/** Default for ConfigProvider — light */
export const linearTheme: ThemeConfig = linearLightTheme

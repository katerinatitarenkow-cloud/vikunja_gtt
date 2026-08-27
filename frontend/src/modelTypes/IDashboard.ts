export const DASHBOARD_LAYOUT_VERSION = 1

export type DashboardWidgetType =
| 'activities'
| 'calendar'
| 'activityStream'
| 'recordList'
| 'metric'

export type DashboardWidgetWidth =
| 'normal'
| 'wide'
| 'full'

export interface IDashboardWidget {
id: string
type: DashboardWidgetType
title: string
width: DashboardWidgetWidth
settings: Record<string, unknown>
}

export interface IDashboardTab {
id: string
title: string
widgets: IDashboardWidget[]
}

export interface IDashboardLayout {
version: number
activeTabId: string
locked: boolean
tabs: IDashboardTab[]
}
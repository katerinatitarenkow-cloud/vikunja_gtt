import {onBeforeUnmount, ref, watch} from 'vue'
import {nanoid} from 'nanoid'

import {useAuthStore} from '@/stores/auth'
import UserSettingsModel from '@/models/userSettings'

import {
DASHBOARD_LAYOUT_VERSION,
type DashboardWidgetType,
type DashboardWidgetWidth,
type IDashboardLayout,
type IDashboardTab,
type IDashboardWidget,
} from '@/modelTypes/IDashboard'

export const DASHBOARD_WIDGET_TITLES: Record<DashboardWidgetType, string> = {
activities: 'Мои дела',
calendar: 'Календарь',
activityStream: 'Последняя активность',
recordList: 'Список',
metric: 'Показатели',
}

export const DASHBOARD_WIDGET_DESCRIPTIONS: Record<DashboardWidgetType, string> = {
activities: 'Ближайшие задачи и действия текущего пользователя.',
calendar: 'Личные события Google Calendar и рабочие события Vikunja в одном календаре.',
activityStream: 'Хронологическая лента последних изменений и действий.',
recordList: 'Универсальный список данных с фильтрами и сортировкой.',
metric: 'Компактные рабочие показатели.',
}

function cloneLayout(layout: IDashboardLayout): IDashboardLayout {
return JSON.parse(JSON.stringify(layout)) as IDashboardLayout
}

export function createDashboardWidget(
type: DashboardWidgetType,
title = DASHBOARD_WIDGET_TITLES[type],
width?: DashboardWidgetWidth,
): IDashboardWidget {
const defaultWidth: DashboardWidgetWidth =
type === 'activities'
? 'wide'
: type === 'metric'
? 'normal'
: 'normal'

return {
id: `widget-${nanoid(10)}`,
type,
title,
width: width ?? defaultWidth,
settings: {},
}
}

export function createDashboardTab(
title = 'Обзор',
widgets: IDashboardWidget[] = [],
): IDashboardTab {
return {
id: `tab-${nanoid(10)}`,
title,
widgets,
}
}

export function createDefaultDashboardLayout(): IDashboardLayout {
const tab = createDashboardTab('Обзор', [
createDashboardWidget('activities', 'Мои дела', 'wide'),
createDashboardWidget('calendar', 'Календарь'),
createDashboardWidget('activityStream', 'Последняя активность'),
createDashboardWidget('recordList', 'Мои задачи'),
createDashboardWidget('metric', 'Показатели'),
])

return {
version: DASHBOARD_LAYOUT_VERSION,
activeTabId: tab.id,
locked: false,
tabs: [tab],
}
}

function isDashboardLayout(value: unknown): value is IDashboardLayout {
if (
value === null ||
typeof value !== 'object'
) {
return false
}

const candidate = value as Partial<IDashboardLayout>

if (
candidate.version !== DASHBOARD_LAYOUT_VERSION ||
typeof candidate.activeTabId !== 'string' ||
typeof candidate.locked !== 'boolean' ||
!Array.isArray(candidate.tabs) ||
candidate.tabs.length === 0
) {
return false
}

return candidate.tabs.every(tab =>
typeof tab?.id === 'string' &&
typeof tab?.title === 'string' &&
Array.isArray(tab?.widgets),
)
}

function normalizeDashboardLayout(value: unknown): IDashboardLayout {
if (!isDashboardLayout(value)) {
return createDefaultDashboardLayout()
}

const layout = cloneLayout(value)

if (!layout.tabs.some(tab => tab.id === layout.activeTabId)) {
layout.activeTabId = layout.tabs[0].id
}

return layout
}

export function useDashboardLayout() {
const authStore = useAuthStore()

const layout = ref<IDashboardLayout>(
normalizeDashboardLayout(
authStore.settings.frontendSettings.dashboard,
),
)

const saving = ref(false)

let saveTimer: ReturnType<typeof setTimeout> | undefined

async function saveLayout() {
saveTimer = undefined
saving.value = true

try {
const settings = new UserSettingsModel({
...authStore.settings,
frontendSettings: {
...authStore.settings.frontendSettings,
dashboard: cloneLayout(layout.value),
},
})

await authStore.saveUserSettings({
settings,
showMessage: false,
})
} finally {
saving.value = false
}
}

function scheduleSave() {
if (saveTimer !== undefined) {
clearTimeout(saveTimer)
}

saveTimer = setTimeout(() => {
void saveLayout()
}, 500)
}

function resetLayout() {
layout.value = createDefaultDashboardLayout()
}

watch(
layout,
scheduleSave,
{deep: true},
)

onBeforeUnmount(() => {
if (saveTimer === undefined) {
return
}

clearTimeout(saveTimer)
void saveLayout()
})

return {
layout,
saving,
resetLayout,
saveLayout,
}
}
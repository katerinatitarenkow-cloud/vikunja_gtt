<template>
<section class="dashboard">
<header class="dashboard-header">
<div>
<h1>Обзор</h1>
<p class="dashboard-header__subtitle">
Персональная рабочая панель
</p>
</div>

<div class="dashboard-header__actions">
<span
v-if="saving"
class="dashboard-saving"
>
Сохранение…
</span>

<XButton
v-if="editing"
variant="secondary"
@click="toggleLayoutLock"
>
{{ layout.locked ? 'Разблокировать расположение' : 'Заблокировать расположение' }}
</XButton>

<XButton
v-if="editing"
variant="secondary"
@click="resetDashboard"
>
Сбросить
</XButton>

<XButton
variant="secondary"
@click="toggleEditing"
>
{{ editing ? 'Готово' : 'Настроить dashboard' }}
</XButton>
</div>
</header>

<DashboardTabs
:tabs="layout.tabs"
:active-tab-id="layout.activeTabId"
:editing="editing"
@select="selectTab"
@add="addTab"
@delete="deleteTab"
@rename="renameTab"
/>

<div
v-if="editing"
class="dashboard-editor"
>
<DashboardWidgetPicker @add="addWidget" />

<div class="dashboard-editor__status">
<span v-if="layout.locked">
Расположение заблокировано
</span>

<span v-else>
Перетаскивайте виджеты за маркер ⋮⋮
</span>
</div>
</div>

<div
v-if="widgetSettingsDraft !== null"
class="dashboard-widget-settings"
>
<div class="dashboard-widget-settings__header">
<strong>Настройки виджета</strong>

<button
type="button"
class="dashboard-widget-settings__close"
aria-label="Закрыть настройки"
@click="widgetSettingsDraft = null"
>
×
</button>
</div>

<div class="dashboard-widget-settings__fields">
<label>
<span>Название</span>

<input
v-model="widgetSettingsDraft.title"
type="text"
class="dashboard-widget-settings__input"
>
</label>

<label>
<span>Ширина</span>

<select
v-model="widgetSettingsDraft.width"
class="dashboard-widget-settings__input"
>
<option
v-for="option in widthOptions"
:key="option.value"
:value="option.value"
>
{{ option.label }}
</option>
</select>
</label>

<XButton @click="saveWidgetSettings">
Сохранить
</XButton>
</div>
</div>

<DashboardGrid
v-if="activeTab !== null"
:widgets="activeWidgets"
:editing="editing"
:locked="layout.locked"
@update:widgets="updateWidgets"
@remove="removeWidget"
@settings="openWidgetSettings"
/>

<div
v-if="activeWidgets.length === 0"
class="dashboard-empty"
>
<strong>На этой вкладке пока нет виджетов</strong>

<p v-if="editing">
Добавьте нужный виджет через панель настройки выше.
</p>

<p v-else>
Откройте настройку dashboard, чтобы добавить рабочие блоки.
</p>
</div>
</section>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'

import XButton from '@/components/input/Button.vue'

import DashboardTabs from './DashboardTabs.vue'
import DashboardGrid from './DashboardGrid.vue'
import DashboardWidgetPicker from './DashboardWidgetPicker.vue'

import {
createDashboardTab,
createDashboardWidget,
useDashboardLayout,
} from '@/composables/useDashboardLayout'

import type {
DashboardWidgetType,
DashboardWidgetWidth,
IDashboardTab,
IDashboardWidget,
} from '@/modelTypes/IDashboard'

interface WidgetSettingsDraft {
id: string
title: string
width: DashboardWidgetWidth
}

const {
layout,
saving,
resetLayout,
} = useDashboardLayout()

const editing = ref(false)
const widgetSettingsDraft = ref<WidgetSettingsDraft | null>(null)

const widthOptions: Array<{
value: DashboardWidgetWidth
label: string
}> = [
{value: 'normal', label: 'Обычный'},
{value: 'wide', label: 'Широкий'},
{value: 'full', label: 'На всю ширину'},
]

const activeTab = computed<IDashboardTab | null>(() => {
return (
layout.value.tabs.find(
tab => tab.id === layout.value.activeTabId,
) ??
layout.value.tabs[0] ??
null
)
})

const activeWidgets = computed<IDashboardWidget[]>(() => {
return activeTab.value?.widgets ?? []
})

function toggleEditing() {
editing.value = !editing.value

if (!editing.value) {
widgetSettingsDraft.value = null
}
}

function toggleLayoutLock() {
layout.value.locked = !layout.value.locked
}

function selectTab(tabId: string) {
if (!layout.value.tabs.some(tab => tab.id === tabId)) {
return
}

layout.value.activeTabId = tabId
widgetSettingsDraft.value = null
}

function addTab() {
const tab = createDashboardTab(
`Вкладка ${layout.value.tabs.length + 1}`,
)

layout.value.tabs.push(tab)
layout.value.activeTabId = tab.id
widgetSettingsDraft.value = null
}

function deleteTab(tabId: string) {
if (layout.value.tabs.length <= 1) {
return
}

const index = layout.value.tabs.findIndex(
tab => tab.id === tabId,
)

if (index < 0) {
return
}

const tab = layout.value.tabs[index]

if (
tab.widgets.length > 0 &&
!window.confirm(
`Удалить вкладку «${tab.title}» и все размещённые на ней виджеты?`,
)
) {
return
}

layout.value.tabs.splice(index, 1)

if (layout.value.activeTabId === tabId) {
const fallbackIndex = Math.min(
index,
layout.value.tabs.length - 1,
)

layout.value.activeTabId =
layout.value.tabs[fallbackIndex].id
}

widgetSettingsDraft.value = null
}

function renameTab(tabId: string, title: string) {
const tab = layout.value.tabs.find(
item => item.id === tabId,
)

if (!tab) {
return
}

const trimmed = title.trim()

if (trimmed === '') {
return
}

tab.title = trimmed
}

function addWidget(type: DashboardWidgetType) {
if (!activeTab.value) {
return
}

activeTab.value.widgets.push(
createDashboardWidget(type),
)
}

function removeWidget(widgetId: string) {
if (!activeTab.value) {
return
}

activeTab.value.widgets =
activeTab.value.widgets.filter(
widget => widget.id !== widgetId,
)

if (
widgetSettingsDraft.value?.id === widgetId
) {
widgetSettingsDraft.value = null
}
}

function updateWidgets(widgets: IDashboardWidget[]) {
if (!activeTab.value) {
return
}

activeTab.value.widgets = widgets
}

function openWidgetSettings(widgetId: string) {
const widget = activeTab.value?.widgets.find(
item => item.id === widgetId,
)

if (!widget) {
return
}

widgetSettingsDraft.value = {
id: widget.id,
title: widget.title,
width: widget.width,
}
}

function saveWidgetSettings() {
const draft = widgetSettingsDraft.value

if (!draft || !activeTab.value) {
return
}

const widget = activeTab.value.widgets.find(
item => item.id === draft.id,
)

if (!widget) {
widgetSettingsDraft.value = null
return
}

const title = draft.title.trim()

if (title !== '') {
widget.title = title
}

widget.width = draft.width

widgetSettingsDraft.value = null
}

function resetDashboard() {
if (
!window.confirm(
'Вернуть стандартное расположение dashboard? Пользовательская настройка будет удалена.',
)
) {
return
}

resetLayout()
widgetSettingsDraft.value = null
}
</script>

<style scoped lang="scss">
.dashboard {
display: flex;
flex-direction: column;
gap: 1rem;
}

.dashboard-header {
display: flex;
align-items: flex-start;
justify-content: space-between;
gap: 1rem;
flex-wrap: wrap;

h1 {
margin: 0;
font-size: clamp(1.65rem, 3vw, 2.15rem);
font-weight: 750;
letter-spacing: -.035em;
color: var(--text-strong);
}
}

.dashboard-header__subtitle {
margin: .2rem 0 0;
color: var(--grey-600);
font-size: .9rem;
}

.dashboard-header__actions {
display: flex;
align-items: center;
justify-content: flex-end;
gap: .5rem;
flex-wrap: wrap;
}

.dashboard-saving {
color: var(--grey-600);
font-size: .82rem;
}

.dashboard-editor {
display: flex;
align-items: center;
justify-content: space-between;
gap: 1rem;
flex-wrap: wrap;
padding: .7rem .8rem;
border: 1px solid var(--brand-border);
border-radius: 7px;
background: var(--brand-surface-soft);
}

.dashboard-editor__status {
color: var(--grey-600);
font-size: .82rem;
}

.dashboard-widget-settings {
padding: .85rem;
border: 1px solid var(--brand-border);
border-radius: 7px;
background: var(--white);
}

.dashboard-widget-settings__header {
display: flex;
align-items: center;
justify-content: space-between;
gap: 1rem;
margin-block-end: .75rem;
color: var(--text-strong);
}

.dashboard-widget-settings__close {
display: inline-flex;
align-items: center;
justify-content: center;
inline-size: 28px;
block-size: 28px;
padding: 0;
border: 0;
border-radius: 5px;
background: transparent;
color: var(--grey-600);
cursor: pointer;
font-size: 1.2rem;

&:hover {
background: var(--brand-surface-soft);
color: var(--text-strong);
}
}

.dashboard-widget-settings__fields {
display: flex;
align-items: flex-end;
gap: .75rem;
flex-wrap: wrap;

label {
display: flex;
flex-direction: column;
gap: .3rem;
color: var(--grey-600);
font-size: .8rem;
}
}

.dashboard-widget-settings__input {
min-inline-size: 180px;
block-size: 38px;
padding-inline: .65rem;
border: 1px solid var(--brand-border);
border-radius: 6px;
background: var(--white);
color: var(--text);
}

.dashboard-empty {
display: flex;
flex-direction: column;
align-items: center;
justify-content: center;
min-block-size: 220px;
padding: 2rem;
border: 1px dashed var(--brand-border);
border-radius: 8px;
text-align: center;
color: var(--grey-600);

strong {
color: var(--text-strong);
}

p {
margin: .4rem 0 0;
}
}
</style>
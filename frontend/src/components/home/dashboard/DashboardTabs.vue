<template>
<div class="dashboard-tabs">
<div
v-for="tab in tabs"
:key="tab.id"
class="dashboard-tabs__item"
>
<input
v-if="editing && tab.id === activeTabId"
class="dashboard-tabs__input"
:value="tab.title"
aria-label="Название вкладки"
@change="renameTab(tab.id, $event)"
>

<button
v-else
type="button"
class="dashboard-tabs__tab"
:class="{'is-active': tab.id === activeTabId}"
@click="$emit('select', tab.id)"
>
{{ tab.title }}
</button>

<button
v-if="editing && tabs.length > 1"
type="button"
class="dashboard-tabs__delete"
aria-label="Удалить вкладку"
@click="$emit('delete', tab.id)"
>
×
</button>
</div>

<button
v-if="editing"
type="button"
class="dashboard-tabs__add"
aria-label="Добавить вкладку"
@click="$emit('add')"
>
+
</button>
</div>
</template>

<script setup lang="ts">
import type {IDashboardTab} from '@/modelTypes/IDashboard'

defineProps<{
tabs: IDashboardTab[]
activeTabId: string
editing: boolean
}>()

const emit = defineEmits<{
select: [tabId: string]
add: []
delete: [tabId: string]
rename: [tabId: string, title: string]
}>()

function renameTab(tabId: string, event: Event) {
const input = event.target as HTMLInputElement
emit('rename', tabId, input.value)
}
</script>

<style lang="scss" scoped>
.dashboard-tabs {
display: flex;
align-items: center;
gap: .2rem;
min-inline-size: 0;
border-block-end: 1px solid var(--brand-border);
overflow-x: auto;
}

.dashboard-tabs__item {
display: flex;
align-items: center;
flex: 0 0 auto;
}

.dashboard-tabs__tab,
.dashboard-tabs__add {
position: relative;
padding: .65rem .85rem;
border: 0;
background: transparent;
color: var(--grey-600);
font-weight: 600;
cursor: pointer;

&:hover {
color: var(--text-strong);
}

&.is-active {
color: var(--text-strong);

&::after {
content: '';
position: absolute;
inset-inline: .65rem;
inset-block-end: -1px;
block-size: 2px;
background: var(--primary);
}
}
}

.dashboard-tabs__input {
inline-size: 160px;
margin: .25rem .25rem .25rem 0;
padding: .38rem .55rem;
border: 1px solid var(--brand-border);
border-radius: 5px;
background: var(--white);
color: var(--text);
}

.dashboard-tabs__delete {
inline-size: 26px;
block-size: 26px;
padding: 0;
border: 0;
border-radius: 4px;
background: transparent;
color: var(--grey-500);
cursor: pointer;

&:hover {
background: var(--brand-surface-soft);
color: var(--danger);
}
}

.dashboard-tabs__add {
font-size: 1.15rem;
}
</style>
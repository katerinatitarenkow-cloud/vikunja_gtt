<template>
<div
ref="grid"
class="dashboard-grid"
>
<DashboardWidget
v-for="widget in widgets"
:key="widget.id"
:data-widget-id="widget.id"
:widget="widget"
:editing="editing"
:locked="locked"
@remove="$emit('remove', widget.id)"
@settings="$emit('settings', widget.id)"
>
<div class="dashboard-widget-placeholder">
<p>{{ widgetDescription(widget.type) }}</p>
<span>
Данные будут подключены на следующем этапе.
</span>
</div>
</DashboardWidget>
</div>
</template>

<script setup lang="ts">
import {
nextTick,
onBeforeUnmount,
onMounted,
ref,
watch,
} from 'vue'
import Sortable from 'sortablejs'

import DashboardWidget from './DashboardWidget.vue'

import type {
DashboardWidgetType,
IDashboardWidget,
} from '@/modelTypes/IDashboard'

import {
DASHBOARD_WIDGET_DESCRIPTIONS,
} from '@/composables/useDashboardLayout'

const props = defineProps<{
widgets: IDashboardWidget[]
editing: boolean
locked: boolean
}>()

const emit = defineEmits<{
'update:widgets': [widgets: IDashboardWidget[]]
remove: [widgetId: string]
settings: [widgetId: string]
}>()

const grid = ref<HTMLElement | null>(null)

let sortable: Sortable | null = null

function createSortable() {
if (!grid.value || sortable !== null) {
return
}

sortable = Sortable.create(grid.value, {
animation: 150,
handle: '.dashboard-widget__drag-handle',
disabled: !props.editing || props.locked,
onEnd(event) {
const oldIndex = event.oldIndex
const newIndex = event.newIndex

if (
oldIndex === undefined ||
newIndex === undefined ||
oldIndex === newIndex
) {
return
}

const reordered = [...props.widgets]
const [moved] = reordered.splice(oldIndex, 1)

if (!moved) {
return
}

reordered.splice(newIndex, 0, moved)

emit('update:widgets', reordered)
},
})
}

watch(
() => [props.editing, props.locked] as const,
([editing, locked]) => {
sortable?.option(
'disabled',
!editing || locked,
)
},
)

onMounted(async () => {
await nextTick()
createSortable()
})

onBeforeUnmount(() => {
sortable?.destroy()
sortable = null
})

function widgetDescription(type: DashboardWidgetType): string {
return DASHBOARD_WIDGET_DESCRIPTIONS[type]
}
</script>

<style lang="scss" scoped>
.dashboard-grid {
display: grid;
grid-template-columns: repeat(12, minmax(0, 1fr));
align-items: stretch;
gap: 1rem;
min-block-size: 180px;
}

.dashboard-widget-placeholder {
display: flex;
flex-direction: column;
justify-content: center;
min-block-size: 110px;
color: var(--grey-600);

p {
margin: 0 0 .35rem;
color: var(--text);
font-weight: 600;
}

span {
font-size: .85rem;
}
}
</style>
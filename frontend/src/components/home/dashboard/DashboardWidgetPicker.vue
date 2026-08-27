<template>
<div class="dashboard-widget-picker">
<select
v-model="selectedType"
class="dashboard-widget-picker__select"
>
<option
v-for="option in options"
:key="option.value"
:value="option.value"
>
{{ option.label }}
</option>
</select>

<XButton
variant="secondary"
icon="plus"
@click="$emit('add', selectedType)"
>
Добавить виджет
</XButton>
</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'

import XButton from '@/components/input/Button.vue'

import type {DashboardWidgetType} from '@/modelTypes/IDashboard'
import {DASHBOARD_WIDGET_TITLES} from '@/composables/useDashboardLayout'

defineEmits<{
add: [type: DashboardWidgetType]
}>()

const selectedType = ref<DashboardWidgetType>('activities')

const options = Object.entries(DASHBOARD_WIDGET_TITLES).map(
([value, label]) => ({
value: value as DashboardWidgetType,
label,
}),
)
</script>

<style lang="scss" scoped>
.dashboard-widget-picker {
display: flex;
align-items: center;
gap: .6rem;
flex-wrap: wrap;
}

.dashboard-widget-picker__select {
min-inline-size: 190px;
block-size: 38px;
padding-inline: .65rem 2rem;
border: 1px solid var(--brand-border);
border-radius: 6px;
background: var(--white);
color: var(--text);
}
</style>
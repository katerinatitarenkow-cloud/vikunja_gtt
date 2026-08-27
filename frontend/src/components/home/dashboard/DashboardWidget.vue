<template>
<article
class="dashboard-widget"
:class="`dashboard-widget--${widget.width}`"
>
<header class="dashboard-widget__header">
<div class="dashboard-widget__heading">
<button
v-if="editing && !locked"
type="button"
class="dashboard-widget__drag-handle"
aria-label="Переместить виджет"
>
⋮⋮
</button>

<h2>{{ widget.title }}</h2>
</div>

<div
v-if="editing"
class="dashboard-widget__actions"
>
<button
type="button"
class="dashboard-widget__action"
aria-label="Настройки виджета"
@click="$emit('settings')"
>
⚙
</button>

<button
type="button"
class="dashboard-widget__action dashboard-widget__action--danger"
aria-label="Удалить виджет"
@click="$emit('remove')"
>
×
</button>
</div>
</header>

<div class="dashboard-widget__content">
<slot />
</div>
</article>
</template>

<script setup lang="ts">
import type {IDashboardWidget} from '@/modelTypes/IDashboard'

defineProps<{
widget: IDashboardWidget
editing: boolean
locked: boolean
}>()

defineEmits<{
remove: []
settings: []
}>()
</script>

<style lang="scss" scoped>
.dashboard-widget {
display: flex;
flex-direction: column;
min-inline-size: 0;
min-block-size: 180px;
border: 1px solid var(--brand-border);
border-radius: 8px;
background: var(--white);
box-shadow: var(--shadow-xs);
overflow: hidden;

&--normal {
grid-column: span 4;
}

&--wide {
grid-column: span 8;
}

&--full {
grid-column: span 12;
}

@media screen and (max-width: $widescreen) {
&--normal {
grid-column: span 6;
}

&--wide,
&--full {
grid-column: span 12;
}
}

@media screen and (max-width: $tablet) {
&--normal,
&--wide,
&--full {
grid-column: span 12;
}
}
}

.dashboard-widget__header {
display: flex;
align-items: center;
justify-content: space-between;
gap: .75rem;
min-block-size: 46px;
padding: .65rem .8rem;
border-block-end: 1px solid var(--brand-border);
}

.dashboard-widget__heading {
display: flex;
align-items: center;
gap: .55rem;
min-inline-size: 0;

h2 {
margin: 0;
font-size: .95rem;
font-weight: 700;
color: var(--text-strong);
overflow: hidden;
text-overflow: ellipsis;
white-space: nowrap;
}
}

.dashboard-widget__drag-handle,
.dashboard-widget__action {
display: inline-flex;
align-items: center;
justify-content: center;
inline-size: 30px;
block-size: 30px;
padding: 0;
border: 0;
border-radius: 5px;
background: transparent;
color: var(--grey-600);
cursor: pointer;

&:hover {
background: var(--brand-surface-soft);
color: var(--text-strong);
}
}

.dashboard-widget__drag-handle {
cursor: grab;
font-size: 1rem;

&:active {
cursor: grabbing;
}
}

.dashboard-widget__actions {
display: flex;
align-items: center;
gap: .15rem;
}

.dashboard-widget__action--danger:hover {
color: var(--danger);
}

.dashboard-widget__content {
flex: 1;
min-block-size: 0;
padding: .85rem;
}
</style>
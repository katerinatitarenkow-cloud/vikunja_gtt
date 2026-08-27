<template>
<ul
class="project-grid"
:class="{ 'show-even-number-of-projects': showEvenNumberOfProjects }"
>
<li
v-for="(item, index) in filteredProjects"
:key="`project_${item.id}_${index}`"
class="project-grid-item"
:class="{
'is-selection-mode': selectionMode,
'is-selected': selectionMode && isSelected(item.id),
'is-not-selectable': selectionMode && !isSelectable(item.id),
}"
>
<button
v-if="selectionMode"
type="button"
class="project-selection-overlay"
:disabled="!isSelectable(item.id)"
:aria-pressed="isSelected(item.id)"
:aria-label="isSelected(item.id) ? `Снять выбор: ${item.title}` : `Выбрать: ${item.title}`"
@click="emit('toggle-selection', item.id)"
>
<span
v-if="isSelectable(item.id)"
class="project-selection-checkbox"
:class="{'is-selected': isSelected(item.id)}"
>
<Icon
v-if="isSelected(item.id)"
icon="check"
/>
</span>
</button>

<ProjectCard
:project="item"
:due-date="props.dueDates?.[item.id] ?? null"
/>
</li>
</ul>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import type {IProject} from '@/modelTypes/IProject'

import ProjectCard from './ProjectCard.vue'

const props = withDefaults(defineProps<{
projects: readonly IProject[],
showArchived?: boolean,
itemLimit?: boolean,
showEvenNumberOfProjects?: boolean,
dueDates?: Record<number, string>,
selectionMode?: boolean,
selectedProjectIds?: number[],
selectableProjectIds?: number[],
}>(), {
showArchived: false,
itemLimit: false,
showEvenNumberOfProjects: false,
selectionMode: false,
selectedProjectIds: () => [],
selectableProjectIds: () => [],
})

const emit = defineEmits<{
'toggle-selection': [projectId: number],
}>()

const filteredProjects = computed(() => {
return props.showArchived
? props.projects
: props.projects.filter(project => !project.isArchived)
})

function isSelected(projectId: number): boolean {
return props.selectedProjectIds.includes(projectId)
}

function isSelectable(projectId: number): boolean {
return props.selectableProjectIds.includes(projectId)
}
</script>

<style lang="scss" scoped>
.project-grid {
--project-grid-item-height: 150px;
--project-grid-gap: 1rem;
margin: 0;
list-style-type: none;
display: grid;
grid-template-columns: repeat(var(--project-grid-columns), 1fr);
grid-auto-rows: var(--project-grid-item-height);
gap: var(--project-grid-gap);

@media screen and (min-width: $mobile) {
--project-grid-columns: 1;
}

@media screen and (min-width: $mobile) and (max-width: $tablet) {
--project-grid-columns: 2;
}

@media screen and (min-width: $tablet) and (max-width: $widescreen) {
--project-grid-columns: 3;
}

@media screen and (min-width: $widescreen) {
--project-grid-columns: 5;
}

&.show-even-number-of-projects {
@media screen and (min-width: $widescreen) {
.project-grid-item:nth-child(5) {
display: none;
}
}
}
}

.project-grid-item {
display: grid;
margin-block-start: 0;
position: relative;
border-radius: $radius;

&.is-selection-mode {
:deep(.complete-toggle),
:deep(.favorite) {
opacity: 0 !important;
pointer-events: none;
}
}

&.is-not-selectable {
opacity: .65;
}
}

.project-selection-overlay {
position: absolute;
inset: 0;
z-index: 10;
padding: 0;
border: 3px solid transparent;
border-radius: $radius;
background: transparent;
cursor: pointer;
appearance: none;
transition:
border-color $transition,
box-shadow $transition;

&:hover:not(:disabled) {
border-color: color-mix(in srgb, var(--primary) 45%, transparent);
}

&:focus-visible {
outline: none;
box-shadow: var(--focus-ring);
}

&:disabled {
cursor: not-allowed;
}
}

.project-grid-item.is-selected .project-selection-overlay {
border-color: var(--primary);
box-shadow: 0 0 0 2px color-mix(in srgb, var(--primary) 18%, transparent);
}

.project-selection-checkbox {
position: absolute;
inset-block-start: .65rem;
inset-inline-start: .65rem;
display: flex;
align-items: center;
justify-content: center;
inline-size: 1.8rem;
block-size: 1.8rem;
border: 2px solid var(--grey-400);
border-radius: 6px;
background: var(--white);
color: var(--white);
box-shadow: var(--shadow-sm);
font-size: .85rem;

&.is-selected {
background: var(--primary);
border-color: var(--primary);
}
}
</style>
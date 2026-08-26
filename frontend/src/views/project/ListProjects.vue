<template>
<div
v-cy="'projects-list'"
class="content loader-container"
:class="{'is-loading': loading}"
>
<div class="project-tabs">
<BaseButton
class="project-tab"
:class="{'is-active': activeSection === 'active'}"
@click="activeSection = 'active'"
>
Проекты
</BaseButton>

<BaseButton
class="project-tab"
:class="{'is-active': activeSection === 'completed'}"
@click="activeSection = 'completed'"
>
Выполненные проекты
</BaseButton>
</div>

<header class="project-header">
<FancyCheckbox
v-model="showArchived"
v-cy="'show-archived-check'"
>
{{ $t('project.showArchived') }}
</FancyCheckbox>

<div class="action-buttons">
<XButton
:to="{name: 'filters.create'}"
icon="filter"
>
{{ $t('filters.create.title') }}
</XButton>

<XButton
v-if="accessStore.can(ACCESS_PERMISSION.PROJECTS_MANAGE)"
v-cy="'new-project'"
:to="{name: 'project.create'}"
icon="plus"
>
{{ $t('project.create.header') }}
</XButton>
</div>
</header>

<div
v-if="activeSection === 'completed' && projects.length === 0"
class="empty-completed"
>
Выполненных проектов пока нет.
</div>

<ProjectCardGrid
v-else
:projects="projects"
:show-archived="showArchived"
:due-dates="dueDates"
/>
</div>
</template>

<script setup lang="ts">
import {computed, onActivated, onBeforeUnmount, onMounted, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import ProjectCardGrid from '@/components/project/partials/ProjectCardGrid.vue'

import {useTitle} from '@/composables/useTitle'
import {useStorage} from '@vueuse/core'

import {useProjectStore} from '@/stores/projects'
import {useAccessStore} from '@/stores/access'
import {ACCESS_PERMISSION} from '@/modelTypes/IAccessControl'
import TaskService from '@/services/task'

const {t} = useI18n()
const projectStore = useProjectStore()
const accessStore = useAccessStore()

useTitle(() => t('project.title'))

const showArchived = useStorage('showArchived', false)
const activeSection = ref<'active' | 'completed'>('active')

const dueDates = ref<Record<number, string>>({})

async function loadDueDates() {
const service = new TaskService()
const nearest: Record<number, string> = {}

const now = Date.now()
const warningLimit = now + 5 * 24 * 60 * 60 * 1000

const params = {
sort_by: ['id'],
order_by: ['asc'],
filter: 'done = false',
per_page: 50,
}

try {
let page = 1

while (true) {
const tasks = await service.getAll({}, params, page)

for (const task of tasks) {
if (task.projectId <= 0 || task.done) {
continue
}

const candidateDates = [
task.dueDate,
task.endDate,
]

let nearestTaskDeadline: number | null = null

for (const candidate of candidateDates) {
if (!candidate) {
continue
}

const timestamp = new Date(candidate).getTime()

if (!Number.isFinite(timestamp)) {
continue
}

// Дата дальше чем через 5 дней — пока не предупреждаем.
if (timestamp > warningLimit) {
continue
}

if (
nearestTaskDeadline === null ||
timestamp < nearestTaskDeadline
) {
nearestTaskDeadline = timestamp
}
}

if (nearestTaskDeadline === null) {
continue
}

const current = nearest[task.projectId]

if (
!current ||
nearestTaskDeadline < new Date(current).getTime()
) {
nearest[task.projectId] =
new Date(nearestTaskDeadline).toISOString()
}
}

if (
tasks.length === 0 ||
!service.totalPages ||
page >= service.totalPages
) {
break
}

page++
}

dueDates.value = nearest
} catch (error) {
dueDates.value = {}
console.error('Could not load project due-date warnings', error)
}
}

function refreshDueDates() {
void loadDueDates()
}

onMounted(() => {
refreshDueDates()
window.addEventListener('focus', refreshDueDates)
})

onActivated(refreshDueDates)

onBeforeUnmount(() => {
window.removeEventListener('focus', refreshDueDates)
})



const loading = computed(() => projectStore.isLoading)

const projects = computed(() => {
const visibleProjects = showArchived.value
? projectStore.projectsArray
: projectStore.projectsArray.filter(({isArchived}) => !isArchived)

if (activeSection.value === 'completed') {
return visibleProjects.filter(({isCompleted}) => isCompleted)
}

return visibleProjects.filter(({isCompleted}) => !isCompleted)
})
</script>

<style lang="scss" scoped>
.project-tabs {
display: flex;
align-items: center;
gap: .35rem;
margin-block-end: 1.25rem;
border-block-end: 1px solid var(--brand-border);
}

.project-tab {
position: relative;
padding: .7rem 1rem;
color: var(--text);
font-weight: 600;
transition:
color $transition,
background $transition;

&:hover {
color: var(--brand-forest);
}

&.is-active {
color: var(--brand-forest);

&::after {
content: '';
position: absolute;
inset-inline: .75rem;
inset-block-end: -1px;
block-size: 3px;
border-radius: 3px 3px 0 0;
background: var(--brand-forest);
}
}
}

.empty-completed {
padding: 3rem 1rem;
text-align: center;
color: var(--grey-500);
font-size: .95rem;
}

.project-header {
display: flex;
justify-content: space-between;
align-items: center;
gap: 1rem;
margin-block-end: 1rem;

@media screen and (max-width: $tablet) {
flex-direction: column;
}
}

.action-buttons {
display: flex;
justify-content: space-between;
gap: 1rem;

@media screen and (max-width: $tablet) {
inline-size: 100%;
flex-direction: column;
align-items: stretch;
}
}

.project:not(:first-child) {
margin-block-start: 1rem;
}

.project-title {
display: flex;
align-items: center;
}

.is-archived {
font-size: 0.75rem;
border: 1px solid var(--grey-500);
color: $grey !important;
padding: 2px 4px;
border-radius: 3px;
font-family: $vikunja-font;
background: var(--white-translucent);
margin-inline-start: .5rem;
}
</style>











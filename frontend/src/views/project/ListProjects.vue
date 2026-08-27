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
v-if="
accessStore.can(ACCESS_PERMISSION.PROJECTS_MANAGE) &&
selectableProjects.length > 0
"
variant="secondary"
@click="bulkEditMode ? cancelBulkEdit() : startBulkEdit()"
>
{{ bulkEditMode ? 'Закончить редактирование' : 'Массовое редактирование' }}
</XButton>

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
v-if="bulkEditMode"
class="bulk-toolbar"
>
<div class="bulk-toolbar-info">
<strong>
Выбрано: {{ selectedProjectIds.length }}
</strong>

<span class="bulk-toolbar-total">
из {{ selectableProjects.length }}
</span>
</div>

<div class="bulk-toolbar-actions">
<XButton
variant="tertiary"
@click="toggleSelectAll"
>
{{ allSelectableSelected ? 'Снять выделение' : 'Выбрать все' }}
</XButton>

<XButton
danger
:disabled="selectedProjectIds.length === 0"
@click="showBulkDeleteModal = true"
>
Удалить выбранные
</XButton>

<XButton
variant="secondary"
@click="cancelBulkEdit"
>
Отмена
</XButton>
</div>
</div>

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
:selection-mode="bulkEditMode"
:selected-project-ids="selectedProjectIds"
:selectable-project-ids="selectableProjectIds"
@toggle-selection="toggleProjectSelection"
/>

<Modal
v-if="showBulkDeleteModal"
@close="showBulkDeleteModal = false"
@submit="deleteSelectedProjects"
>
<template #header>
<span>Удалить выбранные проекты?</span>
</template>

<template #text>
<p>
Выбрано проектов: <strong>{{ selectedProjectIds.length }}</strong>.
</p>

<p>
Если среди выбранных есть родительский проект,
его вложенные проекты и задачи также будут удалены.
</p>

<p class="has-text-weight-bold">
Это действие нельзя отменить.
</p>

<p v-if="bulkDeleting">
Удаление...
</p>
</template>
</Modal>
</div>
</template>

<script setup lang="ts">
import {
computed,
onActivated,
onBeforeUnmount,
onMounted,
ref,
watch,
} from 'vue'
import {useI18n} from 'vue-i18n'
import {useStorage} from '@vueuse/core'

import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import ProjectCardGrid from '@/components/project/partials/ProjectCardGrid.vue'

import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'
import {useAccessStore} from '@/stores/access'
import {ACCESS_PERMISSION} from '@/modelTypes/IAccessControl'
import {PERMISSIONS} from '@/constants/permissions'
import type {IProject} from '@/modelTypes/IProject'
import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import ProjectModel from '@/models/project'
import {success, error as showError, getErrorText} from '@/message'

const {t} = useI18n()
const projectStore = useProjectStore()
const accessStore = useAccessStore()

useTitle(() => t('project.title'))

const showArchived = useStorage('showArchived', false)
const activeSection = ref<'active' | 'completed'>('active')

const dueDates = ref<Record<number, string>>({})

const bulkEditMode = ref(false)
const selectedProjectIds = ref<number[]>([])
const showBulkDeleteModal = ref(false)
const bulkDeleting = ref(false)

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
const tasks = await service.getAll(
new TaskModel(),
params,
page,
)

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

const projects = computed<IProject[]>(() => {
const allProjects = projectStore.projectsArray.map(
project => project as unknown as IProject,
)

const visibleProjects = showArchived.value
? allProjects
: allProjects.filter(({isArchived}) => !isArchived)

if (activeSection.value === 'completed') {
return visibleProjects.filter(({isCompleted}) => isCompleted)
}

return visibleProjects.filter(({isCompleted}) => !isCompleted)
})

const selectableProjects = computed(() => {
return projects.value.filter(project => {
if (project.id <= 0) {
return false
}

return (
project.maxPermission === null ||
project.maxPermission === PERMISSIONS.ADMIN
)
})
})

const selectableProjectIds = computed(() =>
selectableProjects.value.map(project => project.id),
)

const allSelectableSelected = computed(() => {
if (selectableProjectIds.value.length === 0) {
return false
}

return selectableProjectIds.value.every(projectId =>
selectedProjectIds.value.includes(projectId),
)
})

const selectedRootProjectIds = computed<number[]>(() => {
const selected = new Set(selectedProjectIds.value)

return projects.value
.filter(project => selected.has(project.id))
.filter(project => {
let parentProjectId = project.parentProjectId

while (parentProjectId > 0) {
if (selected.has(parentProjectId)) {
return false
}

const parentProject = projectStore.projects[parentProjectId]

if (!parentProject) {
break
}

parentProjectId = parentProject.parentProjectId
}

return true
})
.map(project => project.id)
})

function startBulkEdit() {
bulkEditMode.value = true
selectedProjectIds.value = []
}

function cancelBulkEdit() {
bulkEditMode.value = false
selectedProjectIds.value = []
showBulkDeleteModal.value = false
}

function toggleProjectSelection(projectId: number) {
if (!selectableProjectIds.value.includes(projectId)) {
return
}

if (selectedProjectIds.value.includes(projectId)) {
selectedProjectIds.value =
selectedProjectIds.value.filter(id => id !== projectId)

return
}

selectedProjectIds.value = [
...selectedProjectIds.value,
projectId,
]
}

function toggleSelectAll() {
if (allSelectableSelected.value) {
selectedProjectIds.value = []
return
}

selectedProjectIds.value = [...selectableProjectIds.value]
}

async function deleteSelectedProjects() {
if (
bulkDeleting.value ||
selectedRootProjectIds.value.length === 0
) {
return
}

bulkDeleting.value = true

const failedProjectIds: number[] = []
const failures: string[] = []
let deletedCount = 0

try {
const projectIdsToDelete = [...selectedRootProjectIds.value]

for (const projectId of projectIdsToDelete) {
const project = projectStore.projects[projectId]
const projectTitle = project?.title ?? `#${projectId}`

try {
await projectStore.deleteProject(
new ProjectModel({id: projectId}),
)

deletedCount++
} catch (error) {
failedProjectIds.push(projectId)

const message = getErrorText(error)

failures.push(
`«${projectTitle}»: ${message}`,
)
}
}

await loadDueDates()

if (failures.length === 0) {
selectedProjectIds.value = []
showBulkDeleteModal.value = false
bulkEditMode.value = false

success({
message: `Удалено проектов: ${deletedCount}.`,
})

return
}

selectedProjectIds.value = failedProjectIds
showBulkDeleteModal.value = false
bulkEditMode.value = true

if (deletedCount > 0) {
success({
message: `Удалено проектов: ${deletedCount}.`,
})
}

showError({
message:
`Не удалось удалить ${failures.length} проект(а):\n` +
failures.join('\n'),
})
} finally {
bulkDeleting.value = false
}
}

watch(
[activeSection, showArchived],
() => {
selectedProjectIds.value = []
showBulkDeleteModal.value = false
},
)

watch(
projects,
visibleProjects => {
const visibleIds = new Set(
visibleProjects.map(project => project.id),
)

selectedProjectIds.value =
selectedProjectIds.value.filter(projectId =>
visibleIds.has(projectId),
)
},
)
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

.bulk-toolbar {
display: flex;
align-items: center;
justify-content: space-between;
gap: 1rem;
margin-block-end: 1rem;
padding: .8rem 1rem;
border: 1px solid var(--brand-border);
border-radius: $radius;
background: var(--brand-surface-soft);
box-shadow: var(--shadow-xs);

@media screen and (max-width: $tablet) {
align-items: stretch;
flex-direction: column;
}
}

.bulk-toolbar-info {
display: flex;
align-items: baseline;
gap: .4rem;
}

.bulk-toolbar-total {
color: var(--grey-500);
font-size: .85rem;
}

.bulk-toolbar-actions {
display: flex;
align-items: center;
gap: .6rem;

@media screen and (max-width: $tablet) {
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
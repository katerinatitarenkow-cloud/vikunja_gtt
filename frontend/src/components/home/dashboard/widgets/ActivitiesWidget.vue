<template>
<div class="activities-widget">
<div class="activities-widget__toolbar">
<span>
Ближайшие задачи
</span>

<button
type="button"
class="activities-widget__refresh"
:disabled="loading"
@click="loadTasks"
>
{{ loading ? 'Загрузка…' : 'Обновить' }}
</button>
</div>

<div
v-if="loading && tasks.length === 0"
class="activities-widget__state"
>
Загрузка задач…
</div>

<div
v-else-if="loadError"
class="activities-widget__state activities-widget__state--error"
>
<p>Не удалось загрузить задачи.</p>

<button
type="button"
class="activities-widget__retry"
@click="loadTasks"
>
Повторить
</button>
</div>

<div
v-else-if="groups.length === 0"
class="activities-widget__state"
>
Нет ближайших задач, назначенных вам.
</div>

<div
v-else
class="activities-widget__groups"
>
<section
v-for="group in groups"
:key="group.key"
class="activities-group"
>
<header class="activities-group__header">
<span>{{ group.title }}</span>

<span class="activities-group__count">
{{ group.tasks.length }}
</span>
</header>

<ul class="activities-group__list">
<li
v-for="task in group.tasks"
:key="task.id"
class="activities-task"
>
<RouterLink
:to="{
name: 'task.detail',
params: {id: task.id},
}"
class="activities-task__link"
>
<div class="activities-task__time">
{{ formatTaskDate(task) }}
</div>

<div class="activities-task__main">
<strong class="activities-task__title">
{{ task.title }}
</strong>

<div class="activities-task__meta">
<span v-if="projectTitle(task)">
{{ projectTitle(task) }}
</span>

<span v-if="assigneeNames(task)">
{{ assigneeNames(task) }}
</span>
</div>
</div>
</RouterLink>
</li>
</ul>
</section>
</div>
</div>
</template>

<script setup lang="ts">
import {
computed,
onMounted,
ref,
} from 'vue'

import TaskService from '@/services/task'
import TaskModel from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'

import {useAuthStore} from '@/stores/auth'
import {useProjectStore} from '@/stores/projects'

interface ActivityGroup {
key: 'overdue' | 'today' | 'tomorrow' | 'upcoming'
title: string
tasks: ITask[]
}

const authStore = useAuthStore()
const projectStore = useProjectStore()

const tasks = ref<ITask[]>([])
const loading = ref(false)
const loadError = ref(false)

function startOfDay(date: Date): Date {
const result = new Date(date)
result.setHours(0, 0, 0, 0)

return result
}

function addDays(date: Date, days: number): Date {
const result = new Date(date)
result.setDate(result.getDate() + days)

return result
}

function taskDate(task: ITask): Date | null {
const value =
task.dueDate ??
task.startDate ??
task.endDate

if (!value) {
return null
}

const date = new Date(value)

return Number.isNaN(date.getTime())
? null
: date
}

function isTaskMine(task: ITask, userId: number): boolean {
const assignedDirectly = task.assignees.some(
assignee => assignee.id === userId,
)

if (assignedDirectly) {
return true
}

return (
task.assignees.length === 0 &&
task.createdBy?.id === userId
)
}

async function loadTasks() {
const userId = authStore.info?.id

if (!userId) {
tasks.value = []
return
}

loading.value = true
loadError.value = false

try {
const service = new TaskService()

const params = {
sort_by: ['due_date', 'start_date', 'end_date', 'id'],
order_by: ['asc', 'asc', 'asc', 'desc'],
filter: 'done = false',
filter_include_nulls: true,
s: '',
per_page: 100,
}

const loaded: ITask[] = []
let page = 1

while (true) {
const pageTasks = await service.getAll(
new TaskModel(),
{...params},
page,
)

loaded.push(...pageTasks)

if (
pageTasks.length === 0 ||
service.totalPages <= page
) {
break
}

page++

if (page > 100) {
break
}
}

tasks.value = loaded
.filter(task => !task.done)
.filter(task => isTaskMine(task, userId))
.filter(task => taskDate(task) !== null)
.sort((first, second) => {
const firstDate = taskDate(first)?.getTime() ?? 0
const secondDate = taskDate(second)?.getTime() ?? 0

return firstDate - secondDate
})
} catch {
loadError.value = true
} finally {
loading.value = false
}
}

const groups = computed<ActivityGroup[]>(() => {
const today = startOfDay(new Date())
const tomorrow = addDays(today, 1)
const dayAfterTomorrow = addDays(today, 2)
const upcomingEnd = addDays(today, 8)

const result: ActivityGroup[] = [
{
key: 'overdue',
title: 'Просрочено',
tasks: [],
},
{
key: 'today',
title: 'Сегодня',
tasks: [],
},
{
key: 'tomorrow',
title: 'Завтра',
tasks: [],
},
{
key: 'upcoming',
title: 'Ближайшие дни',
tasks: [],
},
]

for (const task of tasks.value) {
const date = taskDate(task)

if (!date) {
continue
}

const timestamp = date.getTime()

if (timestamp < today.getTime()) {
result[0].tasks.push(task)
continue
}

if (timestamp < tomorrow.getTime()) {
result[1].tasks.push(task)
continue
}

if (timestamp < dayAfterTomorrow.getTime()) {
result[2].tasks.push(task)
continue
}

if (timestamp < upcomingEnd.getTime()) {
result[3].tasks.push(task)
}
}

return result.filter(group => group.tasks.length > 0)
})

function formatTaskDate(task: ITask): string {
const date = taskDate(task)

if (!date) {
return ''
}

return new Intl.DateTimeFormat('ru-RU', {
day: '2-digit',
month: '2-digit',
hour: '2-digit',
minute: '2-digit',
}).format(date)
}

function projectTitle(task: ITask): string {
return projectStore.projects[task.projectId]?.title ?? ''
}

function assigneeNames(task: ITask): string {
return task.assignees
.map(assignee =>
assignee.name ||
assignee.username,
)
.filter(Boolean)
.join(', ')
}

onMounted(() => {
void loadTasks()
})
</script>

<style scoped lang="scss">
.activities-widget {
display: flex;
flex-direction: column;
gap: .7rem;
}

.activities-widget__toolbar {
display: flex;
align-items: center;
justify-content: space-between;
gap: .75rem;
color: var(--grey-600);
font-size: .78rem;
}

.activities-widget__refresh,
.activities-widget__retry {
padding: .2rem .4rem;
border: 0;
border-radius: 4px;
background: transparent;
color: var(--primary);
cursor: pointer;

&:hover:not(:disabled) {
background: var(--brand-surface-soft);
}

&:disabled {
opacity: .55;
cursor: default;
}
}

.activities-widget__state {
display: flex;
flex-direction: column;
align-items: center;
justify-content: center;
min-block-size: 120px;
color: var(--grey-600);
text-align: center;

p {
margin: 0 0 .4rem;
}
}

.activities-widget__state--error {
color: var(--danger);
}

.activities-widget__groups {
display: flex;
flex-direction: column;
gap: 1rem;
}

.activities-group__header {
display: flex;
align-items: center;
gap: .45rem;
margin-block-end: .3rem;
color: var(--text-strong);
font-size: .78rem;
font-weight: 750;
text-transform: uppercase;
letter-spacing: .025em;
}

.activities-group__count {
display: inline-flex;
align-items: center;
justify-content: center;
min-inline-size: 20px;
block-size: 20px;
padding-inline: .35rem;
border-radius: 999px;
background: var(--brand-surface-soft);
color: var(--grey-600);
font-size: .7rem;
}

.activities-group:first-child .activities-group__header {
color: var(--danger);
}

.activities-group__list {
margin: 0;
padding: 0;
list-style: none;
}

.activities-task {
margin: 0 !important;
border-block-end: 1px solid var(--brand-border);

&:last-child {
border-block-end: 0;
}
}

.activities-task__link {
display: grid;
grid-template-columns: 92px minmax(0, 1fr);
gap: .75rem;
align-items: center;
padding: .55rem .25rem;
color: inherit;
text-decoration: none;

&:hover {
background: var(--brand-surface-soft);
}
}

.activities-task__time {
color: var(--grey-600);
font-size: .73rem;
font-variant-numeric: tabular-nums;
}

.activities-task__main {
min-inline-size: 0;
}

.activities-task__title {
display: block;
color: var(--text-strong);
font-size: .86rem;
font-weight: 650;
overflow: hidden;
text-overflow: ellipsis;
white-space: nowrap;
}

.activities-task__meta {
display: flex;
gap: .6rem;
margin-block-start: .18rem;
color: var(--grey-600);
font-size: .7rem;
overflow: hidden;

span {
overflow: hidden;
text-overflow: ellipsis;
white-space: nowrap;
}
}

@media screen and (max-width: $tablet) {
.activities-task__link {
grid-template-columns: 1fr;
gap: .15rem;
}
}
</style>
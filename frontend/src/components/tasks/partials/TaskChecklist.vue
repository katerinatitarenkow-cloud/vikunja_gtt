<template>
	<section class="structured-checklist">
		<div class="checklist-heading">
			<div>
				<h2 class="task-section-title checklist-title">
					<span class="icon is-grey"><Icon icon="tasks" /></span>
					{{ $t('task.structuredChecklist.title') }}
				</h2>
				<p
					v-if="state.total > 0"
					class="checklist-progress-text"
				>
					{{ $t('task.structuredChecklist.progress', {completed: state.completed, total: state.total}) }}
				</p>
			</div>
			<div
				v-if="state.total > 0"
				class="checklist-progress"
				aria-hidden="true"
			>
				<span :style="{width: progressPercent + '%'}" />
			</div>
		</div>

		<div
			v-if="loading"
			class="checklist-loading"
		>
			{{ $t('misc.loading') }}
		</div>

		<div
			v-else-if="state.items.length > 0"
			class="checklist-items"
		>
			<div
				v-for="item in state.items"
				:key="item.id"
				v-tooltip="item.done ? completionTooltip(item) : ''"
				class="checklist-item"
				:class="{'is-done': item.done}"
			>
				<FancyCheckbox
					:model-value="item.done"
					:disabled="!canWrite || savingIds.has(item.id)"
					class="checklist-checkbox"
					@update:modelValue="done => toggleItem(item, done)"
				/>
				<span class="checklist-item-title">{{ item.title }}</span>
				<BaseButton
					v-if="canWrite"
					class="checklist-delete"
					:aria-label="$t('task.structuredChecklist.delete')"
					:disabled="savingIds.has(item.id)"
					@click="removeItem(item)"
				>
					<Icon icon="trash-alt" />
				</BaseButton>
			</div>
		</div>

		<p
			v-else-if="!loading"
			class="checklist-empty"
		>
			{{ $t('task.structuredChecklist.empty') }}
		</p>

		<form
			v-if="canWrite"
			class="checklist-add"
			@submit.prevent="addItem"
		>
			<div class="control is-expanded">
				<input
					v-model="newTitle"
					class="input"
					type="text"
					maxlength="1000"
					:placeholder="$t('task.structuredChecklist.placeholder')"
					:disabled="creating"
				>
			</div>
			<BaseButton
				class="button is-primary"
				:disabled="creating || newTitle.trim() === ''"
				type="submit"
			>
				<Icon icon="plus" />
				<span>{{ $t('task.structuredChecklist.add') }}</span>
			</BaseButton>
		</form>
	</section>
</template>

<script setup lang="ts">
import {computed, reactive, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import TaskChecklistService from '@/services/taskChecklist'
import type {ITaskChecklistItem, ITaskChecklistState} from '@/modelTypes/ITaskChecklist'
import {formatDateLong} from '@/helpers/time/formatDate'
import {error} from '@/message'

const props = defineProps<{
	taskId: number
	canWrite: boolean
}>()

const emit = defineEmits<{
	'state-changed': [state: ITaskChecklistState]
}>()

const {t} = useI18n({useScope: 'global'})
const service = new TaskChecklistService()
const state = reactive<ITaskChecklistState>({
	items: [],
	total: 0,
	completed: 0,
	task_done: false,
	task_done_at: null,
})
const loading = ref(true)
const creating = ref(false)
const savingIds = reactive(new Set<number>())
const newTitle = ref('')

const progressPercent = computed(() => state.total === 0 ? 0 : Math.round((state.completed / state.total) * 100))

function applyState(next: ITaskChecklistState) {
	state.items = next.items ?? []
	state.total = next.total ?? state.items.length
	state.completed = next.completed ?? state.items.filter(item => item.done).length
	state.task_done = Boolean(next.task_done)
	state.task_done_at = next.task_done_at ?? null
	emit('state-changed', {...next, items: [...state.items]})
}

async function load() {
	loading.value = true
	state.items = []
	state.total = 0
	state.completed = 0
	try {
		applyState(await service.getAll(props.taskId))
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

async function addItem() {
	const title = newTitle.value.trim()
	if (!title || creating.value) return
	creating.value = true
	try {
		applyState(await service.create(props.taskId, title))
		newTitle.value = ''
	} catch (e) {
		error(e)
	} finally {
		creating.value = false
	}
}

async function toggleItem(item: ITaskChecklistItem, done: boolean) {
	if (!props.canWrite || savingIds.has(item.id)) return
	savingIds.add(item.id)
	try {
		applyState(await service.update(props.taskId, item.id, item.title, done))
	} catch (e) {
		error(e)
	} finally {
		savingIds.delete(item.id)
	}
}

async function removeItem(item: ITaskChecklistItem) {
	if (!props.canWrite || savingIds.has(item.id)) return
	savingIds.add(item.id)
	try {
		applyState(await service.delete(props.taskId, item.id))
	} catch (e) {
		error(e)
	} finally {
		savingIds.delete(item.id)
	}
}

function completionTooltip(item: ITaskChecklistItem): string {
	if (!item.done) return ''
	const userName = item.completed_by?.name || item.completed_by?.username || t('task.structuredChecklist.unknownUser')
	const date = item.completed_at ? formatDateLong(item.completed_at) : t('task.structuredChecklist.unknownTime')
	return t('task.structuredChecklist.completedBy', {user: userName, date})
}

watch(() => props.taskId, load, {immediate: true})
</script>

<style scoped lang="scss">
.structured-checklist {
	margin-block: 1.25rem 1.75rem;
	padding: 1.15rem;
	border: 1px solid var(--grey-200);
	border-radius: 14px;
	background: var(--white);
}

.checklist-heading {
	display: flex;
	align-items: flex-end;
	justify-content: space-between;
	gap: 1rem;
	margin-block-end: .85rem;
}

.checklist-title {
	margin-block-end: .2rem !important;
}

.checklist-progress-text,
.checklist-empty,
.checklist-loading {
	color: var(--grey-600);
	font-size: .875rem;
}

.checklist-progress {
	inline-size: min(190px, 35%);
	block-size: 7px;
	overflow: hidden;
	border-radius: 999px;
	background: var(--grey-200);

	span {
		display: block;
		block-size: 100%;
		border-radius: inherit;
		background: var(--primary);
		transition: width .2s ease;
	}
}

.checklist-items {
	display: grid;
	gap: .35rem;
}

.checklist-item {
	display: grid;
	grid-template-columns: auto minmax(0, 1fr) auto;
	align-items: center;
	gap: .7rem;
	min-block-size: 42px;
	padding: .48rem .55rem;
	border-radius: 10px;
	transition: background .15s ease;

	&:hover {
		background: var(--grey-100);
	}

	&.is-done .checklist-item-title {
		color: var(--grey-500);
		text-decoration: line-through;
	}
}

.checklist-item-title {
	min-inline-size: 0;
	overflow-wrap: anywhere;
}

.checklist-delete {
	padding: .3rem;
	color: var(--grey-500);
	opacity: 0;
	transition: opacity .15s ease, color .15s ease;

	.checklist-item:hover &,
	&:focus-visible {
		opacity: 1;
	}

	&:hover {
		color: var(--danger);
	}
}

.checklist-add {
	display: flex;
	gap: .65rem;
	margin-block-start: .9rem;

	.control {
		flex: 1;
	}
}

@media (max-width: 600px) {
	.checklist-heading {
		align-items: flex-start;
		flex-direction: column;
	}

	.checklist-progress {
		inline-size: 100%;
	}


	.checklist-add {
		align-items: stretch;
		flex-direction: column;
	}
}
</style>

<template>
	<div class="edit-assignees-wrapper">
		<Multiselect
			v-model="assignees"
			class="edit-assignees"
			:class="{'has-assignees': assignees.length > 0}"
			:loading="projectUserService.loading"
			:placeholder="$t('task.assignee.placeholder')"
			:multiple="true"
			:search-results="foundUsers"
			:show-empty="true"
			label="name"
			:select-placeholder="$t('task.assignee.selectPlaceholder')"
			:autocomplete-enabled="false"
			@search="findUser"
			@select="addAssignee"
			@focus="preloadUsers"
		>
			<template #items="{items}">
				<AssigneeList
					:assignees="items"
					:disabled="disabled"
					can-remove
					@remove="removeAssignee"
				/>
			</template>
			<template #searchResult="{option: user}">
				<User
					:avatar-size="24"
					:show-username="true"
					:user="user"
				/>
			</template>
		</Multiselect>

		<div class="edit-assignees__groups mt-2">
			<div class="edit-assignees__groups-label is-size-7 has-text-grey mb-1">
				<Icon icon="users" />
				{{ $t('workGroups.assignedGroups') }}
			</div>
			<Multiselect
				v-model="assignedGroups"
				class="edit-assignees edit-assignees--groups"
				:loading="groupsLoading"
				:placeholder="$t('workGroups.assignPlaceholder')"
				:multiple="true"
				:search-results="foundGroups"
				:show-empty="true"
				label="name"
				:select-placeholder="$t('workGroups.assignSelect')"
				:autocomplete-enabled="false"
				:disabled="disabled"
				@search="findGroup"
				@select="addGroupAssignee"
				@remove="removeGroupAssignee"
				@focus="preloadGroups"
			>
				<template #searchResult="{option: group}">
					<div>
						<strong>{{ group.name }}</strong>
						<div class="is-size-7 has-text-grey">
							{{ $t('workGroups.membersCount', {count: group.member_count}) }}
							<span v-if="group.leader"> · {{ $t('workGroups.leader') }}: {{ group.leader.name || group.leader.username }}</span>
						</div>
					</div>
				</template>
			</Multiselect>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive, watch, nextTick} from 'vue'
import {useI18n} from 'vue-i18n'

import User from '@/components/misc/User.vue'
import Multiselect from '@/components/input/Multiselect.vue'

import {includesById} from '@/helpers/utils'
import ProjectUserService from '@/services/projectUsers'
import WorkGroupService from '@/services/workGroups'
import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import {success, error} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useTaskStore} from '@/stores/tasks'

import type {IUser} from '@/modelTypes/IUser'
import type {IWorkGroup} from '@/modelTypes/IWorkGroup'
import {getDisplayName} from '@/models/user'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'

const props = withDefaults(defineProps<{
	modelValue: IUser[] | undefined,
	taskId: number,
	projectId: number,
	disabled?: boolean,
}>(), {
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: IUser[] | undefined],
}>()

const authStore = useAuthStore()
const taskStore = useTaskStore()
const {t} = useI18n({useScope: 'global'})

const projectUserService = shallowReactive(new ProjectUserService())
const workGroupService = new WorkGroupService()
const taskService = new TaskService()
const foundUsers = ref<IUser[]>([])
const assignees = ref<IUser[]>([])
const assignedGroups = ref<IWorkGroup[]>([])
const foundGroups = ref<IWorkGroup[]>([])
const groupsLoading = ref(false)
let isAdding = false
let hasPreloaded = false
let groupsPreloaded = false

function preloadUsers() {
	if (hasPreloaded) return
	hasPreloaded = true
	findUser()
}

function preloadGroups() {
	if (groupsPreloaded) return
	groupsPreloaded = true
	findGroup()
}

watch(
	() => props.modelValue,
	(value) => {
		assignees.value = value ?? []
	},
	{
		immediate: true,
		deep: true,
	},
)

watch(
	() => props.taskId,
	async () => {
		groupsPreloaded = false
		await loadAssignedGroups()
	},
	{immediate: true},
)

async function reloadAssignees() {
	const loaded = await taskService.get(new TaskModel({id: props.taskId}))
	assignees.value = loaded.assignees ?? []
	emit('update:modelValue', assignees.value)
}

async function loadAssignedGroups() {
	if (!props.taskId) return
	groupsLoading.value = true
	try {
		assignedGroups.value = await workGroupService.getTaskGroups(props.taskId)
	} catch (e) {
		error(e)
	} finally {
		groupsLoading.value = false
	}
}

async function addAssignee(user: IUser) {
	if (isAdding) {
		return
	}

	try {
		nextTick(() => isAdding = true)

		await taskStore.addAssignee({user: user, taskId: props.taskId})
		emit('update:modelValue', assignees.value)
		success({message: t('task.assignee.assignSuccess')})
	} finally {
		nextTick(() => isAdding = false)
	}
}

async function removeAssignee(user: IUser) {
	try {
		await taskStore.removeAssignee({user: user, taskId: props.taskId})
		const idx = assignees.value.findIndex(a => a.id === user.id)
		if (idx !== -1) {
			assignees.value.splice(idx, 1)
		}
		success({message: t('task.assignee.unassignSuccess')})
	} catch (e) {
		await reloadAssignees()
		error(e)
	}
}

async function findUser(query = '') {
	const response = await projectUserService.getAll({projectId: props.projectId}, {s: query}) as IUser[]

	const currentUserId = authStore.info?.id

	foundUsers.value = response
		.filter(({id}) => !includesById(assignees.value, id))
		.map(u => {
			u.name = getDisplayName(u)
			return u
		})
		.sort((a, b) => {
			if (a.id === currentUserId) return -1
			if (b.id === currentUserId) return 1
			return a.name.localeCompare(b.name)
		})
}

async function findGroup(query = '') {
	groupsLoading.value = true
	try {
		const groups = await workGroupService.getAll(query)
		const selected = new Set(assignedGroups.value.map(group => group.id))
		foundGroups.value = groups.filter(group => !selected.has(group.id))
	} catch (e) {
		error(e)
	} finally {
		groupsLoading.value = false
	}
}

async function addGroupAssignee(group: IWorkGroup) {
	groupsLoading.value = true
	try {
		const result = await workGroupService.assignTaskGroup(props.taskId, group.id)
		await Promise.all([loadAssignedGroups(), reloadAssignees()])
		const skipped = result.skipped_users?.length ?? 0
		const message = skipped > 0
			? `${t('workGroups.assignSuccess')} ${t('workGroups.skippedUsers', {count: skipped})}`
			: t('workGroups.assignSuccess')
		success({message})
	} catch (e) {
		await loadAssignedGroups()
		error(e)
	} finally {
		groupsLoading.value = false
	}
}

async function removeGroupAssignee(group: IWorkGroup) {
	groupsLoading.value = true
	try {
		await workGroupService.unassignTaskGroup(props.taskId, group.id)
		await Promise.all([loadAssignedGroups(), reloadAssignees()])
		success({message: t('workGroups.unassignSuccess')})
	} catch (e) {
		await loadAssignedGroups()
		error(e)
	} finally {
		groupsLoading.value = false
	}
}
</script>

<style lang="scss">
.edit-assignees.has-assignees.multiselect .input {
	padding-inline-start: 0;
}

.edit-assignees__groups-label {
	display: flex;
	align-items: center;
	gap: .35rem;
}

.edit-assignees--groups.multiselect .tag {
	font-weight: 600;
}
</style>

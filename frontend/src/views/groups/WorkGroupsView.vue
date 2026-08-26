<template>
	<div class="work-groups-view">
		<Card>
			<div class="work-groups-view__header">
				<div>
					<h2 class="title is-4 mb-1">{{ $t('workGroups.title') }}</h2>
					<p class="has-text-grey">{{ $t('workGroups.description') }}</p>
				</div>
				<XButton v-if="isAdmin" variant="primary" @click="openCreate">
					{{ $t('workGroups.add') }}
				</XButton>
			</div>
		</Card>

		<Card class="mt-4" :loading="loading">
			<div class="work-groups-view__toolbar">
				<FormInput
					v-model="search"
					type="search"
					:placeholder="$t('workGroups.search')"
					@keyup.enter="loadGroups"
				/>
				<XButton variant="secondary" @click="loadGroups">{{ $t('workGroups.searchButton') }}</XButton>
			</div>

			<p v-if="groups.length === 0 && !loading" class="has-text-grey mt-4">
				{{ $t('workGroups.empty') }}
			</p>

			<div v-else class="work-groups-view__grid mt-4">
				<div v-for="group in groups" :key="group.id" class="work-groups-view__card">
					<div class="work-groups-view__card-head">
						<div>
							<strong>{{ group.name }}</strong>
							<div v-if="group.leader" class="is-size-7 has-text-grey mt-1">
								{{ $t('workGroups.leader') }}: {{ displayName(group.leader) }}
							</div>
						</div>
						<XButton v-if="isAdmin" variant="secondary" @click="openEdit(group)">
							{{ $t('workGroups.editButton') }}
						</XButton>
					</div>
					<p v-if="group.description" class="has-text-grey mt-2">{{ group.description }}</p>
					<div class="mt-3">
						<span class="tag mr-1 mb-1">{{ $t('workGroups.membersCount', {count: group.member_count}) }}</span>
						<span class="tag mr-1 mb-1">{{ $t('workGroups.tasksCount', {count: group.task_count}) }}</span>
					</div>
					<div class="work-groups-view__members mt-3">
						<span v-for="member in group.members" :key="member.id" class="tag mr-1 mb-1">
							{{ displayName(member) }}
							<span v-if="member.id === group.leader_user_id" class="ml-1">в…</span>
						</span>
					</div>
				</div>
			</div>
		</Card>

		<Modal v-if="showModal" variant="hint-modal" @close="closeModal">
			<Card class="has-no-shadow" :title="editingId ? $t('workGroups.editTitle') : $t('workGroups.createTitle')">
				<FormField :label="$t('workGroups.name')">
					<FormInput v-model="form.name" type="text" />
				</FormField>
				<FormField :label="$t('workGroups.groupDescription')">
					<textarea v-model="form.description" class="textarea" rows="3" />
				</FormField>

				<FormField :label="$t('workGroups.leader')">
					<select v-model.number="form.leaderUserId" class="input">
						<option :value="0">{{ $t('workGroups.noLeader') }}</option>
						<option v-for="person in users" :key="person.id" :value="person.id">
							{{ person.name || person.username }} (@{{ person.username }})
						</option>
					</select>
				</FormField>

				<div class="mt-4">
					<strong>{{ $t('workGroups.members') }}</strong>
					<p class="help mb-2">{{ $t('workGroups.membersHelp') }}</p>
					<FormInput v-model="userSearch" type="search" :placeholder="$t('workGroups.searchUsers')" />
					<div class="work-groups-view__user-list mt-2">
						<label v-for="person in filteredUsers" :key="person.id" class="work-groups-view__check-row">
							<input
								type="checkbox"
								:checked="form.memberIds.includes(person.id)"
								@change="toggleMember(person.id)"
							>
							<span>
								{{ person.name || person.username }}
								<span class="has-text-grey is-size-7">@{{ person.username }}</span>
							</span>
						</label>
					</div>
				</div>

				<div class="work-groups-view__modal-actions mt-5">
					<XButton variant="primary" :loading="saving" @click="save">{{ $t('misc.save') }}</XButton>
					<XButton v-if="editingId" variant="secondary" :loading="deleting" @click="remove">
						{{ $t('misc.delete') }}
					</XButton>
					<XButton variant="secondary" @click="closeModal">{{ $t('misc.cancel') }}</XButton>
				</div>
			</Card>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'
import FormInput from '@/components/input/FormInput.vue'
import FormField from '@/components/input/FormField.vue'
import {error, success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import AccessControlService from '@/services/accessControl'
import WorkGroupService from '@/services/workGroups'
import type {IWorkGroup} from '@/modelTypes/IWorkGroup'
import type {IAccessUser} from '@/modelTypes/IAccessControl'
import type {IUser} from '@/modelTypes/IUser'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('workGroups.title'))

const authStore = useAuthStore()
const service = new WorkGroupService()
const accessService = new AccessControlService()
const groups = ref<IWorkGroup[]>([])
const users = ref<IAccessUser[]>([])
const search = ref('')
const userSearch = ref('')
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const showModal = ref(false)
const editingId = ref<number | null>(null)
const isAdmin = computed(() => Boolean(authStore.info?.isAdmin))
const form = reactive({name: '', description: '', leaderUserId: 0, memberIds: [] as number[]})

const filteredUsers = computed(() => {
	const q = userSearch.value.trim().toLowerCase()
	if (!q) return users.value
	return users.value.filter(person => [person.name, person.username, person.email].some(value => value?.toLowerCase().includes(q)))
})

function displayName(person: IUser): string {
	return person.name || person.username
}

async function loadGroups() {
	loading.value = true
	try {
		groups.value = await service.getAll(search.value)
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

async function loadUsers() {
	if (!isAdmin.value) return
	try {
		users.value = await accessService.getUsers()
	} catch (e) {
		error(e)
	}
}

function resetForm() {
	Object.assign(form, {name: '', description: '', leaderUserId: 0, memberIds: []})
	userSearch.value = ''
}

function openCreate() {
	resetForm()
	editingId.value = null
	showModal.value = true
}

function openEdit(group: IWorkGroup) {
	resetForm()
	editingId.value = group.id
	Object.assign(form, {
		name: group.name,
		description: group.description,
		leaderUserId: group.leader_user_id || 0,
		memberIds: group.members.map(member => member.id),
	})
	showModal.value = true
}

function closeModal() {
	showModal.value = false
	editingId.value = null
	resetForm()
}

function toggleMember(userId: number) {
	form.memberIds = form.memberIds.includes(userId)
		? form.memberIds.filter(id => id !== userId)
		: [...form.memberIds, userId]
	if (form.leaderUserId === userId && !form.memberIds.includes(userId)) {
		form.leaderUserId = 0
	}
}

async function save() {
	if (!form.name.trim()) {
		error({message: 'Введите название группы'})
		return
	}
	saving.value = true
	try {
		const payload = {
			name: form.name.trim(),
			description: form.description.trim(),
			leader_user_id: form.leaderUserId,
			member_ids: form.memberIds,
		}
		if (editingId.value) {
			await service.update(editingId.value, payload)
			success({message: t('workGroups.saved')})
		} else {
			await service.create(payload)
			success({message: t('workGroups.created')})
		}
		await loadGroups()
		closeModal()
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}

async function remove() {
	if (!editingId.value || !window.confirm(t('workGroups.deleteConfirm', {name: form.name}))) return
	deleting.value = true
	try {
		await service.delete(editingId.value)
		success({message: t('workGroups.deleted')})
		await loadGroups()
		closeModal()
	} catch (e) {
		error(e)
	} finally {
		deleting.value = false
	}
}

onMounted(async () => {
	await Promise.all([loadGroups(), loadUsers()])
})
</script>

<style scoped>
.work-groups-view__header,
.work-groups-view__toolbar,
.work-groups-view__card-head,
.work-groups-view__modal-actions {
	display: flex;
	align-items: center;
	gap: .75rem;
}
.work-groups-view__header,
.work-groups-view__card-head { justify-content: space-between; }
.work-groups-view__toolbar { flex-wrap: wrap; }
.work-groups-view__toolbar > :first-child { flex: 1 1 260px; }
.work-groups-view__grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1rem; }
.work-groups-view__card { border: 1px solid var(--grey-200); border-radius: 8px; padding: 1rem; }
.work-groups-view__user-list { max-height: 300px; overflow-y: auto; border: 1px solid var(--grey-200); border-radius: 6px; padding: .5rem; }
.work-groups-view__check-row { display: flex; gap: .6rem; align-items: flex-start; padding: .4rem; cursor: pointer; }
.work-groups-view__check-row input { margin-top: .25rem; }
.work-groups-view__modal-actions { justify-content: flex-end; flex-wrap: wrap; }
@media (max-width: 800px) {
	.work-groups-view__header { align-items: flex-start; flex-direction: column; }
}
</style>


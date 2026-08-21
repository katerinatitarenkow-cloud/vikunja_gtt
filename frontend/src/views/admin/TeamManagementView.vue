<template>
	<div class="team-management">
		<Card>
			<div class="team-management__header">
				<div>
					<h2 class="title is-4 mb-1">{{ $t('admin.team.title') }}</h2>
					<p class="has-text-grey">{{ $t('admin.team.description') }}</p>
				</div>
				<div class="team-management__tabs">
					<XButton :variant="tab === 'users' ? 'primary' : 'secondary'" @click="tab = 'users'">
						{{ $t('admin.team.usersTab') }}
					</XButton>
					<XButton :variant="tab === 'groups' ? 'primary' : 'secondary'" @click="tab = 'groups'">
						{{ $t('admin.team.groupsTab') }}
					</XButton>
				</div>
			</div>
		</Card>

		<Card v-if="tab === 'users'" class="mt-4" :loading="loadingUsers">
			<div class="team-management__toolbar">
				<FormInput
					v-model="search"
					type="search"
					:placeholder="$t('admin.team.searchUsers')"
					@keyup.enter="loadUsers"
				/>
				<XButton variant="secondary" @click="loadUsers">{{ $t('admin.team.searchButton') }}</XButton>
				<XButton variant="primary" @click="openCreateUser">{{ $t('admin.team.addUser') }}</XButton>
			</div>

			<div class="table-container mt-4">
				<table class="table is-striped is-hoverable is-fullwidth">
					<thead>
						<tr>
							<th>{{ $t('admin.team.person') }}</th>
							<th>{{ $t('user.auth.email') }}</th>
							<th>{{ $t('admin.team.phone') }}</th>
							<th>{{ $t('admin.team.groups') }}</th>
							<th>{{ $t('admin.users.status') }}</th>
							<th />
						</tr>
					</thead>
					<tbody>
						<tr v-for="person in users" :key="person.id">
							<td>
								<strong>{{ person.name || person.username }}</strong>
								<div class="is-size-7 has-text-grey">@{{ person.username }}</div>
							</td>
							<td>{{ person.email }}</td>
							<td>{{ person.phone || '—' }}</td>
							<td>
								<span v-for="group in person.groups" :key="group.id" class="tag mr-1 mb-1">
									{{ group.name }}
								</span>
							</td>
							<td>{{ statusLabel(person.status) }}</td>
							<td class="has-text-right">
								<XButton variant="secondary" @click="openUserCard(person)">{{ $t('admin.team.openCard') }}</XButton>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</Card>

		<Card v-else class="mt-4" :loading="loadingGroups">
			<div class="team-management__toolbar team-management__toolbar--end">
				<XButton variant="primary" @click="openCreateGroup">{{ $t('admin.team.addGroup') }}</XButton>
			</div>
			<div class="team-management__group-grid mt-4">
				<div v-for="group in groups" :key="group.id" class="team-management__group-card">
					<div class="team-management__group-title">
						<div>
							<strong>{{ group.name }}</strong>
							<span v-if="group.system_key" class="tag ml-2">{{ $t('admin.team.systemGroup') }}</span>
						</div>
						<span class="has-text-grey is-size-7">{{ $t('admin.team.memberCount', {count: group.member_count}) }}</span>
					</div>
					<p class="has-text-grey mt-2">{{ group.description || '—' }}</p>
					<div class="mt-3">
						<span v-for="permission in group.permissions" :key="permission" class="tag mr-1 mb-1">
							{{ permissionLabel(permission) }}
						</span>
					</div>
					<div class="team-management__group-actions mt-3">
						<XButton variant="secondary" @click="openEditGroup(group)">{{ $t('admin.team.editButton') }}</XButton>
						<XButton v-if="!group.system_key" variant="secondary" @click="removeGroup(group)">{{ $t('misc.delete') }}</XButton>
					</div>
				</div>
			</div>
		</Card>

		<Modal
			v-if="showUserModal"
			:class="{'team-management__create-user-modal': creatingUser}"
			variant="hint-modal"
			@close="closeUserModal"
		>
			<Card class="has-no-shadow" :title="userModalTitle">
				<template v-if="creatingUser">
					<FormField
						id="new-user-username"
						:label="$t('admin.team.loginLabel')"
						:error="createAttempted && newUser.username.trim() === '' ? $t('admin.team.usernameRequired') : null"
					>
						<FormInput
							id="new-user-username"
							v-model="newUser.username"
							v-focus
							name="username"
							type="text"
							autocomplete="username"
							required
						/>
					</FormField>

					<div class="field">
						<label class="label" for="password">{{ $t('user.auth.password') }}</label>
						<Password
							v-model="newUser.password"
							autocomplete="new-password"
							:validate-initially="createAttempted"
							@submit="saveUser"
						/>
					</div>
				</template>

				<template v-else>
					<FormField :label="$t('admin.team.fullName')">
						<FormInput v-model="userForm.name" type="text" />
					</FormField>
					<FormField :label="$t('user.auth.email')">
						<FormInput v-model="userForm.email" type="email" autocomplete="email" />
					</FormField>
					<FormField :label="$t('admin.team.phone')">
						<FormInput v-model="userForm.phone" type="tel" />
					</FormField>
					<FormField :label="$t('admin.team.notes')">
						<textarea v-model="userForm.notes" class="textarea" rows="5" />
					</FormField>

					<FormCheckbox v-model="userForm.isAdmin" :label="$t('admin.team.instanceAdmin')" />

					<FormField :label="$t('admin.users.status')">
						<select v-model.number="userForm.status" class="input">
							<option :value="0">{{ $t('admin.users.statusActive') }}</option>
							<option :value="1">{{ $t('admin.users.statusEmailConfirmation') }}</option>
							<option :value="2">{{ $t('admin.users.statusDisabled') }}</option>
							<option :value="3">{{ $t('admin.users.statusLocked') }}</option>
						</select>
					</FormField>

					<div class="mt-4">
						<strong>{{ $t('admin.team.groupMembership') }}</strong>
						<p class="help mb-2">{{ $t('admin.team.groupMembershipHelp') }}</p>
						<label v-for="group in assignableGroups" :key="group.id" class="team-management__check-row">
							<input
								type="checkbox"
								:checked="userForm.groupIds.includes(group.id)"
								@change="toggleUserGroup(group.id)"
							>
							<span>{{ group.name }}</span>
						</label>
					</div>
				</template>

				<div class="team-management__modal-actions mt-4">
					<XButton variant="secondary" @click="closeUserModal">{{ $t('misc.cancel') }}</XButton>
					<XButton
						variant="primary"
						:loading="savingUser"
						:disabled="creatingUser && !canCreateUser"
						@click="saveUser"
					>
						{{ creatingUser ? $t('user.auth.createAccount') : $t('misc.save') }}
					</XButton>
				</div>
			</Card>
		</Modal>

		<Modal v-if="showGroupModal" variant="hint-modal" @close="closeGroupModal">
			<Card class="has-no-shadow" :title="groupModalTitle">
				<FormField :label="$t('admin.team.groupName')">
					<FormInput v-model="groupForm.name" type="text" :disabled="editingSystemGroup" />
				</FormField>
				<FormField :label="$t('admin.team.groupDescription')">
					<textarea v-model="groupForm.description" class="textarea" rows="3" />
				</FormField>

				<div class="mt-4">
					<strong>{{ $t('admin.team.permissionsTitle') }}</strong>
					<p v-if="editingSystemAdminGroup" class="help mb-2">{{ $t('admin.team.adminPermissionsLocked') }}</p>
					<label v-for="permission in permissions" :key="permission.key" class="team-management__check-row">
						<input
							type="checkbox"
							:checked="groupForm.permissions.includes(permission.key)"
							:disabled="editingSystemAdminGroup"
							@change="togglePermission(permission.key)"
						>
						<span>{{ permissionLabel(permission.key) }}</span>
					</label>
				</div>

				<div class="team-management__modal-actions mt-5">
					<XButton variant="primary" :loading="savingGroup" @click="saveGroup">{{ $t('misc.save') }}</XButton>
					<XButton variant="secondary" @click="closeGroupModal">{{ $t('misc.cancel') }}</XButton>
				</div>
			</Card>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'
import {error, success} from '@/message'
import Password from '@/components/input/Password.vue'
import AccessControlService from '@/services/accessControl'
import type {IAccessGroup, IAccessPermissionDefinition, IAccessUser} from '@/modelTypes/IAccessControl'
import {useAccessStore} from '@/stores/access'
import {useAuthStore} from '@/stores/auth'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('admin.team.title'))

const service = new AccessControlService()
const accessStore = useAccessStore()
const authStore = useAuthStore()
const tab = ref<'users' | 'groups'>('users')
const users = ref<IAccessUser[]>([])
const groups = ref<IAccessGroup[]>([])
const permissions = ref<IAccessPermissionDefinition[]>([])
const search = ref('')
const loadingUsers = ref(false)
const loadingGroups = ref(false)
const savingUser = ref(false)
const savingGroup = ref(false)
const showUserModal = ref(false)
const showGroupModal = ref(false)
const creatingUser = ref(false)
const editingUserId = ref<number | null>(null)
const editingGroupId = ref<number | null>(null)

const userForm = reactive({
	username: '', name: '', email: '', phone: '', notes: '', password: '',
	isAdmin: false, status: 0, groupIds: [] as number[],
})
const newUser = reactive({username: '', password: ''})
const createAttempted = ref(false)
const groupForm = reactive({name: '', description: '', permissions: [] as string[], systemKey: ''})

const assignableGroups = computed(() => groups.value.filter(group => group.system_key !== 'admin'))
const editingSystemGroup = computed(() => Boolean(groupForm.systemKey))
const editingSystemAdminGroup = computed(() => groupForm.systemKey === 'admin')
const canCreateUser = computed(() => newUser.username.trim() !== '' && newUser.password !== '')
const userModalTitle = computed(() => creatingUser.value ? t('admin.team.createUserTitle') : t('admin.team.userCardTitle'))
const groupModalTitle = computed(() => editingGroupId.value ? t('admin.team.editGroupTitle') : t('admin.team.createGroupTitle'))

function permissionLabel(permission: string): string {
	const labels: Record<string, string> = {
		'projects.view': t('admin.team.permissions.projectsView'),
		'projects.manage': t('admin.team.permissions.projectsManage'),
		'tasks.view': t('admin.team.permissions.tasksView'),
		'tasks.manage': t('admin.team.permissions.tasksManage'),
		'labels.view': t('admin.team.permissions.labelsView'),
		'labels.manage': t('admin.team.permissions.labelsManage'),
		'teams.view': t('admin.team.permissions.teamsView'),
		'teams.manage': t('admin.team.permissions.teamsManage'),
		'kanban.use': t('admin.team.permissions.kanbanUse'),
		'time_tracking.use': t('admin.team.permissions.timeTracking'),
		'wialon.view': t('admin.team.permissions.wialonView'),
	}
	return labels[permission] ?? permission
}

function statusLabel(status: number): string {
	return [t('admin.users.statusActive'), t('admin.users.statusEmailConfirmation'), t('admin.users.statusDisabled'), t('admin.users.statusLocked')][status] ?? String(status)
}

async function loadUsers() {
	loadingUsers.value = true
	try { users.value = await service.getUsers(search.value) } catch (e) { error(e) } finally { loadingUsers.value = false }
}

async function loadGroups() {
	loadingGroups.value = true
	try {
		const [loadedGroups, loadedPermissions] = await Promise.all([service.getGroups(), service.getPermissions()])
		groups.value = loadedGroups
		permissions.value = loadedPermissions
	} catch (e) { error(e) } finally { loadingGroups.value = false }
}

function resetUserForm() {
	Object.assign(userForm, {username: '', name: '', email: '', phone: '', notes: '', password: '', isAdmin: false, status: 0, groupIds: []})
}

function openCreateUser() {
	resetUserForm()
	Object.assign(newUser, {username: '', password: ''})
	createAttempted.value = false
	creatingUser.value = true
	editingUserId.value = null
	showUserModal.value = true
}

function openUserCard(person: IAccessUser) {
	resetUserForm(); creatingUser.value = false; editingUserId.value = person.id
	Object.assign(userForm, {
		username: person.username, name: person.name, email: person.email, phone: person.phone,
		notes: person.notes, isAdmin: person.is_admin, status: person.status,
		groupIds: person.groups.filter(group => group.system_key !== 'admin').map(group => group.id),
	})
	showUserModal.value = true
}

function closeUserModal() {
	showUserModal.value = false
	editingUserId.value = null
	Object.assign(newUser, {username: '', password: ''})
	createAttempted.value = false
	resetUserForm()
}
function toggleUserGroup(id: number) {
	userForm.groupIds = userForm.groupIds.includes(id) ? userForm.groupIds.filter(groupId => groupId !== id) : [...userForm.groupIds, id]
}

async function saveUser() {
	if (creatingUser.value) {
		createAttempted.value = true
		if (!canCreateUser.value) return
	}
	savingUser.value = true
	try {
		if (creatingUser.value) {
			await service.createUser({
				username: newUser.username.trim(),
				password: newUser.password,
			})
			success({message: t('admin.team.userCreated')})
		} else if (editingUserId.value) {
			await service.updateUser(editingUserId.value, {
				name: userForm.name.trim(), email: userForm.email.trim(), phone: userForm.phone.trim(), notes: userForm.notes,
				is_admin: userForm.isAdmin, status: userForm.status, group_ids: userForm.groupIds,
			})
			success({message: t('admin.team.userSaved')})
			if (editingUserId.value === authStore.info?.id) await authStore.refreshUserInfo()
		}
		await Promise.all([loadUsers(), loadGroups()])
		await accessStore.refresh()
		closeUserModal()
	} catch (e) { error(e) } finally { savingUser.value = false }
}

function resetGroupForm() { Object.assign(groupForm, {name: '', description: '', permissions: [], systemKey: ''}) }
function openCreateGroup() { resetGroupForm(); editingGroupId.value = null; showGroupModal.value = true }
function openEditGroup(group: IAccessGroup) {
	resetGroupForm(); editingGroupId.value = group.id
	Object.assign(groupForm, {name: group.name, description: group.description, permissions: [...group.permissions], systemKey: group.system_key ?? ''})
	showGroupModal.value = true
}
function closeGroupModal() { showGroupModal.value = false; editingGroupId.value = null; resetGroupForm() }
function togglePermission(key: string) {
	if (editingSystemAdminGroup.value) return
	groupForm.permissions = groupForm.permissions.includes(key) ? groupForm.permissions.filter(item => item !== key) : [...groupForm.permissions, key]
}

async function saveGroup() {
	savingGroup.value = true
	try {
		if (editingGroupId.value) {
			await service.updateGroup(editingGroupId.value, {name: groupForm.name.trim(), description: groupForm.description, permissions: groupForm.permissions})
			success({message: t('admin.team.groupSaved')})
		} else {
			await service.createGroup({name: groupForm.name.trim(), description: groupForm.description, permissions: groupForm.permissions})
			success({message: t('admin.team.groupCreated')})
		}
		await Promise.all([loadGroups(), loadUsers()])
		await accessStore.refresh()
		closeGroupModal()
	} catch (e) { error(e) } finally { savingGroup.value = false }
}

async function removeGroup(group: IAccessGroup) {
	if (!window.confirm(t('admin.team.deleteGroupConfirm', {name: group.name}))) return
	try {
		await service.deleteGroup(group.id)
		success({message: t('admin.team.groupDeleted')})
		await Promise.all([loadGroups(), loadUsers()])
	} catch (e) { error(e) }
}

onMounted(async () => { await loadGroups(); await loadUsers() })
</script>

<style scoped>
.team-management__header,
.team-management__toolbar,
.team-management__group-title,
.team-management__group-actions,
.team-management__modal-actions {
	display: flex;
	align-items: center;
	gap: .75rem;
}
.team-management__header,
.team-management__group-title { justify-content: space-between; }
.team-management__tabs { display: flex; gap: .5rem; flex-wrap: wrap; }
.team-management__toolbar { flex-wrap: wrap; }
.team-management__toolbar > :first-child { flex: 1 1 260px; }
.team-management__toolbar--end { justify-content: flex-end; }
.team-management__group-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1rem; }
.team-management__group-card { border: 1px solid var(--grey-200); border-radius: 8px; padding: 1rem; }
.team-management__check-row { display: flex; align-items: flex-start; gap: .6rem; padding: .35rem 0; cursor: pointer; }
.team-management__check-row input { margin-top: .25rem; }
.team-management__modal-actions { justify-content: flex-end; flex-wrap: wrap; }
:global(.team-management__create-user-modal .modal-content) { max-inline-size: 440px; }
@media (max-width: 800px) {
	.team-management__header { align-items: flex-start; flex-direction: column; }
}
</style>

<template>
	<div>
		<h3 class="has-text-weight-bold share-heading">
			{{ $t('project.share.userTeam.shared', {type: shareTypeNames}) }}
		</h3>
		<div v-if="userIsAdmin">
			<div class="field has-addons">
				<p
					class="control is-expanded"
					:class="{ 'is-loading': searchService.loading }"
				>
					<Multiselect
						v-model="sharable"
						:loading="searchService.loading"
						:placeholder="$t('misc.searchPlaceholder')"
						:aria-label="$t('project.share.userTeam.search', {type: shareTypeName})"
						:search-results="found"
						:label="searchLabel"
						@search="find"
					>
						<template #searchResult="{option: result}">
							<User
								v-if="shareType === 'user'"
								:avatar-size="24"
								:show-username="true"
								:user="asUser(result)"
							/>
							<span 
								v-else
								class="search-result"
							>
								{{ getSharableName(result) }}
							</span>
						</template>
					</Multiselect>
				</p>
				<p class="control">
					<XButton @click="add()">
						{{ $t('project.share.share') }}
					</XButton>
				</p>
			</div>
		</div>

		<div
			v-if="sharables.length > 0"
			class="has-horizontal-overflow mbe-4"
		>
			<table class="table has-actions is-striped is-hoverable is-fullwidth">
				<tbody>
					<tr
						v-for="s in sharables"
						:key="s.id"
					>
						<template v-if="shareType === 'user'">
							<td>{{ getSharableDisplayName(s) }}</td>
							<td>
								<template v-if="s.id === userInfo?.id">
									<b class="is-success">{{ $t('project.share.userTeam.you') }}</b>
								</template>
							</td>
						</template>
						<template v-if="shareType === 'team'">
							<td>
								<RouterLink
									:to="{
										name: 'teams.edit',
										params: { id: s.id },
									}"
								>
									{{ s.name }}
								</RouterLink>
							</td>
						</template>
						<td class="type">
							<template v-if="s.permission === PERMISSIONS.ADMIN">
								<span class="icon is-small">
									<Icon icon="lock" />
								</span>
								{{ $t('project.share.permission.admin') }}
							</template>
							<template v-else-if="s.permission === PERMISSIONS.READ_WRITE">
								<span class="icon is-small">
									<Icon icon="pen" />
								</span>
								{{ $t('project.share.permission.readWrite') }}
							</template>
							<template v-else>
								<span class="icon is-small">
									<Icon icon="users" />
								</span>
								{{ $t('project.share.permission.read') }}
							</template>
						</td>
						<td
							v-if="userIsAdmin"
							class="actions"
						>
							<div class="select">
								<select
									v-model="selectedPermission[s.id]"
									class="mie-2"
									:aria-label="$t('project.share.userTeam.permissionFor', {sharable: getSharableDisplayName(s)})"
									@change="toggleType(s)"
								>
									<option
										:selected="s.permission === PERMISSIONS.READ"
										:value="PERMISSIONS.READ"
									>
										{{ $t('project.share.permission.read') }}
									</option>
									<option
										:selected="s.permission === PERMISSIONS.READ_WRITE"
										:value="PERMISSIONS.READ_WRITE"
									>
										{{ $t('project.share.permission.readWrite') }}
									</option>
									<option
										:selected="s.permission === PERMISSIONS.ADMIN"
										:value="PERMISSIONS.ADMIN"
									>
										{{ $t('project.share.permission.admin') }}
									</option>
								</select>
							</div>
							<XButton
								danger
								icon="trash-alt"
								:aria-label="$t('project.share.userTeam.remove', {type: shareTypeName})"
								@click="
									() => {
										sharable = s
										showDeleteModal = true
									}
								"
							/>
						</td>
					</tr>
				</tbody>
			</table>
		</div>

		<Nothing v-else>
			{{ $t('project.share.userTeam.notShared', {type: shareTypeNames}) }}
		</Nothing>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteSharable()"
		>
			<template #header>
				<span>{{
					$t('project.share.userTeam.removeHeader', {type: shareTypeName, sharable: sharableName})
				}}</span>
			</template>
			<template #text>
				<p>{{ $t('project.share.userTeam.removeText', {type: shareTypeName, sharable: sharableName}) }}</p>
			</template>
		</Modal>
	</div>
</template>


<script setup lang="ts">
import {ref, reactive, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import UserProjectService from '@/services/userProject'
import UserProjectModel from '@/models/userProject'
import type {IUserProject} from '@/modelTypes/IUserProject'

import UserService from '@/services/user'
import UserModel, {getDisplayName} from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

import TeamProjectService from '@/services/teamProject'
import TeamProjectModel from '@/models/teamProject'
import type {ITeamProject} from '@/modelTypes/ITeamProject'

import TeamService from '@/services/team'
import TeamModel from '@/models/team'
import type {ITeam} from '@/modelTypes/ITeam'

import {PERMISSIONS, type Permission} from '@/constants/permissions'
import Multiselect from '@/components/input/Multiselect.vue'
import Nothing from '@/components/misc/Nothing.vue'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import User from '@/components/misc/User.vue'

type Sharable = (
(IUser & {permission?: Permission}) |
ITeam
) & Record<string, unknown>

const props = withDefaults(defineProps<{
type?: 'project',
shareType: 'user' | 'team',
id: number,
userIsAdmin?: boolean
}>(), {
type: 'project',
userIsAdmin: false,
})

defineOptions({name: 'UserTeamShare'})

const {t} = useI18n({useScope: 'global'})

let stuffService: UserProjectService | TeamProjectService
let stuffModel: IUserProject | ITeamProject
let searchService: UserService | TeamService

const sharable = ref<Sharable>(
new UserModel() as unknown as Sharable,
)

const searchLabel = ref('')
const selectedPermission = ref<Record<number, Permission>>({})
const sharables = ref<Sharable[]>([])
const found = ref<Sharable[]>([])
const showDeleteModal = ref(false)

const authStore = useAuthStore()
const configStore = useConfigStore()

const userInfo = computed(() => authStore.info)

function createShareTypeNameComputed(count: number) {
return computed(() => {
if (props.shareType === 'user') {
return t('project.share.userTeam.typeUser', count)
}

return t('project.share.userTeam.typeTeam', count)
})
}

const shareTypeNames = createShareTypeNameComputed(2)
const shareTypeName = createShareTypeNameComputed(1)

const sharableName = computed(() => {
if (props.type === 'project') {
return t('project.list.title')
}

return ''
})

function asUser(value: unknown): IUser {
return value as IUser
}

function getSharableName(value: unknown): string {
if (
typeof value === 'object' &&
value !== null &&
'name' in value
) {
return String(
(value as {name?: unknown}).name ?? '',
)
}

return String(value ?? '')
}

function getSharableDisplayName(value: Sharable): string {
if (props.shareType === 'user') {
return getDisplayName(value as IUser)
}

return value.name
}

if (props.shareType === 'user') {
searchService = shallowReactive(new UserService())
sharable.value = new UserModel() as unknown as Sharable
searchLabel.value = 'username'

stuffService = shallowReactive(
new UserProjectService(),
)

stuffModel = reactive(
new UserProjectModel({
projectId: props.id,
}),
)
} else {
searchService = shallowReactive(new TeamService())
sharable.value = new TeamModel() as unknown as Sharable
searchLabel.value = 'name'

stuffService = shallowReactive(
new TeamProjectService(),
)

stuffModel = reactive(
new TeamProjectModel({
projectId: props.id,
}),
)
}

load()

async function load() {
if (props.shareType === 'user') {
const service =
stuffService as UserProjectService

const model =
stuffModel as IUserProject

sharables.value =
await service.getAll(model) as unknown as Sharable[]
} else {
const service =
stuffService as TeamProjectService

const model =
stuffModel as ITeamProject

sharables.value =
await service.getAll(model) as unknown as Sharable[]
}

for (const item of sharables.value) {
selectedPermission.value[item.id] =
item.permission ?? PERMISSIONS.READ
}
}

async function deleteSharable() {
let idx = -1

if (props.shareType === 'user') {
const service =
stuffService as UserProjectService

const model =
stuffModel as IUserProject

const selected =
sharable.value as IUser

model.username = selected.username

await service.delete(model)

idx = sharables.value.findIndex(item =>
'username' in item &&
item.username === model.username,
)
} else {
const service =
stuffService as TeamProjectService

const model =
stuffModel as ITeamProject

model.teamId = sharable.value.id

await service.delete(model)

idx = sharables.value.findIndex(
item => item.id === model.teamId,
)
}

showDeleteModal.value = false

if (idx !== -1) {
sharables.value.splice(idx, 1)
}

success({
message: t(
'project.share.userTeam.removeSuccess',
{
type: shareTypeName.value,
sharable: sharableName.value,
},
),
})
}

async function add(admin = false) {
const permission = admin
? PERMISSIONS.ADMIN
: PERMISSIONS.READ

if (props.shareType === 'user') {
const service =
stuffService as UserProjectService

const model =
stuffModel as IUserProject

const selected =
sharable.value as IUser

model.permission = permission
model.username = selected.username

await service.create(model)
} else {
const service =
stuffService as TeamProjectService

const model =
stuffModel as ITeamProject

model.permission = permission
model.teamId = sharable.value.id

await service.create(model)
}

success({
message: t(
'project.share.userTeam.addedSuccess',
{type: shareTypeName.value},
),
})

await load()
}

async function toggleType(item: Sharable) {
let permission =
selectedPermission.value[item.id]

if (
permission !== PERMISSIONS.ADMIN &&
permission !== PERMISSIONS.READ &&
permission !== PERMISSIONS.READ_WRITE
) {
permission = PERMISSIONS.READ
selectedPermission.value[item.id] =
PERMISSIONS.READ
}

if (props.shareType === 'user') {
const service =
stuffService as UserProjectService

const model =
stuffModel as IUserProject

const user =
item as IUser

model.permission = permission
model.username = user.username

const result = await service.update(model)

for (const entry of sharables.value) {
if (
'username' in entry &&
entry.username === model.username
) {
entry.permission =
result.permission
}
}
} else {
const service =
stuffService as TeamProjectService

const model =
stuffModel as ITeamProject

model.permission = permission
model.teamId = item.id

const result = await service.update(model)

for (const entry of sharables.value) {
if (entry.id === model.teamId) {
entry.permission =
result.permission
}
}
}

success({
message: t(
'project.share.userTeam.updatedSuccess',
{type: shareTypeName.value},
),
})
}

const currentUserId = computed(
() => authStore.info?.id ?? 0,
)

async function find(query: string) {
if (query === '') {
found.value = []
return
}

let results: Sharable[]

if (props.shareType === 'user') {
const service =
searchService as UserService

results =
await service.getAll(
new UserModel(),
{s: query},
) as unknown as Sharable[]
} else {
const service =
searchService as TeamService

results =
await service.getAll(
new TeamModel(),
{
s: query,
includePublic:
configStore.publicTeamsEnabled,
},
) as unknown as Sharable[]
}

found.value = results.filter(item => {
if (
props.shareType === 'user' &&
item.id === currentUserId.value
) {
return false
}

return !sharables.value.some(
existing =>
existing.id === item.id,
)
})
}
</script>

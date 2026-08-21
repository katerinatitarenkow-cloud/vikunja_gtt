<template>
	<ProjectWrapper
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id="0"
	>
		<div
			v-if="profile"
			class="crm-client"
		>
			<header class="crm-client__hero">
				<div>
					<div class="crm-client__eyebrow">{{ $t('clientProfile.cardEyebrow') }}</div>
					<h2>{{ profile.display_name || $t('clientProfile.unnamed') }}</h2>
					<p>{{ $t('clientProfile.subtitle') }}</p>
					<div class="crm-client__summary">
						<span>{{ statusLabel(profile.status) }}</span>
						<span>{{ selectedResponsible?.name || $t('clientProfile.noResponsible') }}</span>
						<span>{{ $t('clientProfile.contactsCount', {count: profile.contact_persons.length}) }}</span>
					</div>
				</div>
				<XButton
					v-if="canWrite"
					:loading="saving"
					icon="save"
					@click="save"
				>
					{{ $t('clientProfile.save') }}
				</XButton>
			</header>

			<section class="crm-card crm-card--overview">
				<div class="crm-section-heading">
					<div>
						<h3>{{ $t('clientProfile.mainInfo') }}</h3>
						<p>{{ $t('clientProfile.mainInfoHint') }}</p>
					</div>
					<div class="crm-added-date">
						<span>{{ $t('clientProfile.addedAt') }}</span>
						<strong>{{ formatDate(profile.added_at) }}</strong>
					</div>
				</div>

				<div class="crm-form-grid crm-form-grid--3">
					<label class="crm-field">
						<span>{{ $t('clientProfile.type') }}</span>
						<select v-model="profile.client_type" :disabled="!canWrite" class="select input">
							<option value="person">{{ $t('clientProfile.typePerson') }}</option>
							<option value="fop">{{ $t('clientProfile.typeFop') }}</option>
							<option value="company">{{ $t('clientProfile.typeCompany') }}</option>
						</select>
					</label>
					<label class="crm-field crm-field--wide">
						<span>{{ profile.client_type === 'company' ? $t('clientProfile.companyName') : $t('clientProfile.personName') }}</span>
						<input v-model="profile.display_name" :disabled="!canWrite" class="input" type="text">
					</label>
					<label v-if="profile.client_type === 'company'" class="crm-field">
						<span>{{ $t('clientProfile.contactName') }}</span>
						<input v-model="profile.contact_name" :disabled="!canWrite" class="input" type="text">
					</label>
				</div>

				<div class="crm-field crm-field--status">
					<span>{{ $t('clientProfile.status') }}</span>
					<div class="crm-status">
						<button
							v-for="status in statuses"
							:key="status.value"
							type="button"
							:disabled="!canWrite"
							:class="['crm-status__item', `crm-status__item--${status.value}`, {'is-active': profile.status === status.value}]"
							@click="profile.status = status.value"
						>
							{{ status.label }}
						</button>
					</div>
				</div>

				<div class="crm-form-grid crm-form-grid--2">
					<div class="crm-field">
						<span>{{ $t('clientProfile.responsible') }}</span>
						<Multiselect
							v-model="selectedResponsible"
							:disabled="!canWrite"
							:loading="responsibleUsersLoading"
							:placeholder="$t('clientProfile.responsiblePlaceholder')"
							:search-results="foundUsers"
							label="name"
							:show-empty="true"
							:autocomplete-enabled="false"
							@search="findUsers"
							@focus="findUsers('')"
						>
							<template #searchResult="{option: person}">
								<User :user="person" :avatar-size="26" :show-username="true" />
							</template>
						</Multiselect>
					</div>
					<label class="crm-field">
						<span>{{ $t('clientProfile.source') }}</span>
						<select v-model="profile.source" :disabled="!canWrite" class="select input">
							<option value="">—</option>
							<option value="site">{{ $t('clientProfile.sourceSite') }}</option>
							<option value="referral">{{ $t('clientProfile.sourceReferral') }}</option>
							<option value="advertising">{{ $t('clientProfile.sourceAdvertising') }}</option>
							<option value="cold_contact">{{ $t('clientProfile.sourceCold') }}</option>
						</select>
					</label>
				</div>
			</section>

			<section class="crm-card crm-card--proposal">
				<div class="crm-section-heading">
					<div>
						<h3>{{ $t('clientProfile.proposal') }}</h3>
						<p>{{ $t('clientProfile.proposalHint') }}</p>
					</div>
					<XButton v-if="canWrite" variant="secondary" icon="cloud-upload-alt" :loading="uploadingProposal" @click="triggerProposalUpload">
						{{ profile.commercial_proposal ? $t('clientProfile.proposalReplace') : $t('clientProfile.proposalUpload') }}
					</XButton>
					<input ref="proposalInput" class="is-hidden" type="file" accept="application/pdf,.pdf" @change="onProposalSelected">
				</div>

				<div v-if="profile.commercial_proposal" class="crm-proposal">
					<div class="crm-proposal__badge">PDF</div>
					<div class="crm-proposal__meta">
						<strong>{{ profile.commercial_proposal.name }}</strong>
						<span>{{ formatFileSize(profile.commercial_proposal.size) }} · {{ formatDate(profile.commercial_proposal.created) }}</span>
					</div>
					<div class="crm-proposal__actions">
						<XButton variant="secondary" icon="download" :loading="downloadingProposal" @click="downloadProposal">{{ $t('clientProfile.proposalDownload') }}</XButton>
						<XButton v-if="canWrite" variant="secondary" icon="trash-alt" :loading="deletingProposal" @click="deleteProposal">{{ $t('clientProfile.proposalDelete') }}</XButton>
					</div>
				</div>
				<div v-else class="crm-empty-state">
					<strong>{{ $t('clientProfile.proposalEmpty') }}</strong>
					<span>{{ $t('clientProfile.proposalEmptyHint') }}</span>
				</div>
			</section>

			<section class="crm-card">
				<div class="crm-section-heading">
					<div>
						<h3>{{ $t('clientProfile.contacts') }}</h3>
						<p>{{ $t('clientProfile.contactsHint') }}</p>
					</div>
				</div>
				<div class="crm-form-grid crm-form-grid--3">
					<CrmInput v-model="profile.phone" :disabled="!canWrite" :label="$t('clientProfile.phone')" type="tel" />
					<CrmInput v-model="profile.phone_secondary" :disabled="!canWrite" :label="$t('clientProfile.phoneSecondary')" type="tel" />
					<CrmInput v-model="profile.email" :disabled="!canWrite" :label="$t('clientProfile.email')" type="email" />
					<CrmInput v-model="profile.email_secondary" :disabled="!canWrite" :label="$t('clientProfile.emailSecondary')" type="email" />
					<CrmInput v-model="profile.telegram" :disabled="!canWrite" label="Telegram" />
					<CrmInput v-model="profile.viber" :disabled="!canWrite" label="Viber" />
					<CrmInput v-model="profile.whatsapp" :disabled="!canWrite" label="WhatsApp" />
					<CrmInput v-model="profile.website" :disabled="!canWrite" :label="$t('clientProfile.website')" type="url" />
					<label class="crm-field">
						<span>{{ $t('clientProfile.preferredContact') }}</span>
						<select v-model="profile.preferred_contact_method" :disabled="!canWrite" class="select input">
							<option value="">—</option>
							<option value="phone">{{ $t('clientProfile.phone') }}</option>
							<option value="email">Email</option>
							<option value="telegram">Telegram</option>
							<option value="viber">Viber</option>
							<option value="whatsapp">WhatsApp</option>
						</select>
					</label>
					<label class="crm-field">
						<span>{{ $t('clientProfile.preferredLanguage') }}</span>
						<select v-model="profile.preferred_language" :disabled="!canWrite" class="select input">
							<option value="">—</option>
							<option value="uk">Українська</option>
							<option value="ru">Русский</option>
							<option value="en">English</option>
							<option value="other">{{ $t('clientProfile.other') }}</option>
						</select>
					</label>
				</div>
			</section>

			<section class="crm-card">
				<div class="crm-section-heading">
					<div>
						<h3>{{ $t('clientProfile.addresses') }}</h3>
						<p>{{ $t('clientProfile.addressesHint') }}</p>
					</div>
				</div>
				<div class="crm-addresses">
					<details v-for="address in profile.addresses" :key="address.type" class="crm-address" :open="address.type === 'object'">
						<summary>
							<span>{{ addressLabel(address.type) }}</span>
							<small>{{ shortAddress(address) }}</small>
						</summary>
						<div class="crm-address__body">
							<div class="crm-form-grid crm-form-grid--4">
								<CrmInput v-model="address.country" :disabled="!canWrite" :label="$t('clientProfile.country')" @blur="scheduleAddressLookup(address)" />
								<CrmInput v-model="address.region" :disabled="!canWrite" :label="$t('clientProfile.region')" @blur="scheduleAddressLookup(address)" />
								<CrmInput v-model="address.city" :disabled="!canWrite" :label="$t('clientProfile.city')" @blur="scheduleAddressLookup(address)" />
								<CrmInput v-model="address.street" :disabled="!canWrite" :label="$t('clientProfile.street')" @blur="scheduleAddressLookup(address)" />
								<CrmInput v-model="address.house" :disabled="!canWrite" :label="$t('clientProfile.house')" @blur="scheduleAddressLookup(address)" />
								<CrmInput v-model="address.office" :disabled="!canWrite" :label="$t('clientProfile.office')" />
								<CrmInput v-model="address.postal_code" :disabled="!canWrite" :label="$t('clientProfile.postalCode')" />
							</div>
							<div class="crm-address__actions">
								<XButton variant="secondary" icon="map-marker-alt" :loading="geocodingType === address.type" @click="showAddressOnMap(address)">
									{{ $t('clientProfile.showMap') }}
								</XButton>
								<span v-if="address.latitude && address.longitude" class="crm-coordinates">{{ address.latitude.toFixed(5) }}, {{ address.longitude.toFixed(5) }}</span>
							</div>
							<iframe
								v-if="openMaps.has(address.type) && address.latitude && address.longitude"
								class="crm-map"
								:src="mapUrl(address)"
								loading="lazy"
								referrerpolicy="no-referrer"
							/>
						</div>
					</details>
				</div>
			</section>

			<section v-if="profile.client_type !== 'person'" class="crm-card">
				<div class="crm-section-heading">
					<div>
						<h3>{{ $t('clientProfile.legalDetails') }}</h3>
						<p>{{ $t('clientProfile.legalDetailsHint') }}</p>
					</div>
				</div>
				<div class="crm-form-grid crm-form-grid--3">
					<CrmInput v-model="profile.tax_id" :disabled="!canWrite" :label="$t('clientProfile.taxId')" />
					<CrmInput v-model="profile.legal_name" :disabled="!canWrite" :label="$t('clientProfile.legalName')" />
					<CrmInput v-model="profile.director_name" :disabled="!canWrite" :label="$t('clientProfile.directorName')" />
					<CrmInput v-model="profile.ownership_form" :disabled="!canWrite" :label="$t('clientProfile.ownershipForm')" />
					<CrmInput v-model="profile.industry" :disabled="!canWrite" :label="$t('clientProfile.industry')" />
					<CrmInput v-model.number="profile.employee_count" :disabled="!canWrite" :label="$t('clientProfile.employeeCount')" type="number" />
					<CrmInput v-model="profile.iban" :disabled="!canWrite" label="IBAN" />
					<CrmInput v-model="profile.bank" :disabled="!canWrite" :label="$t('clientProfile.bank')" />
					<CrmInput v-model="profile.mfo" :disabled="!canWrite" :label="$t('clientProfile.mfo')" />
					<CrmInput v-model="profile.vat_number" :disabled="!canWrite" :label="$t('clientProfile.vatNumber')" />
					<CrmInput v-model="profile.tax_system" :disabled="!canWrite" :label="$t('clientProfile.taxSystem')" />
					<label class="crm-field crm-field--full">
						<span>{{ $t('clientProfile.requisites') }}</span>
						<textarea v-model="profile.requisites" :disabled="!canWrite" class="textarea" rows="4" />
					</label>
				</div>
			</section>

			<section class="crm-card">
				<div class="crm-section-heading">
					<div>
						<h3>{{ $t('clientProfile.contactPersons') }}</h3>
						<p>{{ $t('clientProfile.contactPersonsHint') }}</p>
					</div>
					<XButton v-if="canWrite" variant="secondary" icon="plus" @click="addContactPerson">{{ $t('clientProfile.contactPersonAdd') }}</XButton>
				</div>

				<div v-if="profile.contact_persons.length" class="crm-contact-persons">
					<article v-for="(person, index) in profile.contact_persons" :key="person.id || `new-${index}`" class="crm-contact-person">
						<header class="crm-contact-person__header">
							<div>
								<span class="crm-contact-person__number">{{ String(index + 1).padStart(2, '0') }}</span>
								<div>
									<strong>{{ person.full_name || $t('clientProfile.contactPersonNew') }}</strong>
									<small>{{ decisionRoleLabel(person.decision_role) }}</small>
								</div>
							</div>
							<XButton v-if="canWrite" variant="secondary" icon="trash-alt" @click="removeContactPerson(index)">{{ $t('clientProfile.remove') }}</XButton>
						</header>

						<div class="crm-form-grid crm-form-grid--3">
							<CrmInput v-model="person.full_name" :disabled="!canWrite" :label="$t('clientProfile.contactPersonName')" />
							<CrmInput v-model="person.position" :disabled="!canWrite" :label="$t('clientProfile.position')" />
							<CrmInput v-model="person.department" :disabled="!canWrite" :label="$t('clientProfile.department')" />
							<CrmInput v-model="person.phone" :disabled="!canWrite" :label="$t('clientProfile.phone')" type="tel" />
							<CrmInput v-model="person.email" :disabled="!canWrite" :label="$t('clientProfile.email')" type="email" />
							<CrmInput v-model="person.birthday" :disabled="!canWrite" :label="$t('clientProfile.birthday')" type="date" />
							<CrmInput v-model="person.telegram" :disabled="!canWrite" label="Telegram" />
							<CrmInput v-model="person.viber" :disabled="!canWrite" label="Viber" />
							<CrmInput v-model="person.whatsapp" :disabled="!canWrite" label="WhatsApp" />
							<label class="crm-field">
								<span>{{ $t('clientProfile.preferredContact') }}</span>
								<select v-model="person.preferred_contact_method" :disabled="!canWrite" class="select input">
									<option value="">—</option>
									<option value="phone">{{ $t('clientProfile.phone') }}</option>
									<option value="email">Email</option>
									<option value="telegram">Telegram</option>
									<option value="viber">Viber</option>
									<option value="whatsapp">WhatsApp</option>
								</select>
							</label>
							<label class="crm-field">
								<span>{{ $t('clientProfile.decisionRole') }}</span>
								<select v-model="person.decision_role" :disabled="!canWrite" class="select input">
									<option v-for="role in decisionRoles" :key="role.value" :value="role.value">{{ role.label }}</option>
								</select>
							</label>
							<label class="crm-field crm-field--full">
								<span>{{ $t('clientProfile.notes') }}</span>
								<textarea v-model="person.notes" :disabled="!canWrite" class="textarea" rows="3" />
							</label>
						</div>
					</article>
				</div>
				<div v-else class="crm-empty-state">
					<strong>{{ $t('clientProfile.contactPersonsEmpty') }}</strong>
					<span>{{ $t('clientProfile.contactPersonsEmptyHint') }}</span>
				</div>
			</section>

			<section class="crm-card crm-card--custom-fields">
				<div class="crm-section-heading">
					<div>
						<h3>{{ $t('clientProfile.customFields') }}</h3>
						<p>{{ $t('clientProfile.customFieldsHint') }}</p>
					</div>
					<XButton v-if="canWrite" variant="secondary" icon="plus" @click="addCustomField">{{ $t('clientProfile.customFieldAdd') }}</XButton>
				</div>

				<div v-if="customFields.length" class="crm-custom-fields">
					<div v-for="(field, index) in customFields" :key="field.id || `new-${index}`" class="crm-custom-field">
						<label class="crm-field">
							<span>{{ $t('clientProfile.customFieldName') }}</span>
							<input v-model="field.name" :disabled="!canWrite" class="input" type="text" :placeholder="$t('clientProfile.customFieldNamePlaceholder')">
						</label>
						<label class="crm-field crm-custom-field__value">
							<span>{{ $t('clientProfile.customFieldValue') }}</span>
							<input v-model="field.value" :disabled="!canWrite" class="input" type="text" :placeholder="$t('clientProfile.customFieldValuePlaceholder')">
						</label>
						<div v-if="canWrite" class="crm-custom-field__actions">
							<XButton variant="secondary" icon="save" :loading="savingCustomFieldIndex === index" @click="saveCustomField(field, index)">{{ $t('misc.save') }}</XButton>
							<XButton variant="secondary" icon="trash-alt" :loading="deletingCustomFieldIndex === index" @click="deleteCustomField(field, index)">{{ $t('misc.delete') }}</XButton>
						</div>
					</div>
				</div>
				<div v-else class="crm-empty-state">
					<strong>{{ $t('clientProfile.customFieldsEmpty') }}</strong>
					<span>{{ $t('clientProfile.customFieldsEmptyHint') }}</span>
				</div>
			</section>

			<div v-if="canWrite" class="crm-savebar">
				<div>
					<strong>{{ $t('clientProfile.saveReminder') }}</strong>
					<span>{{ $t('clientProfile.saveReminderHint') }}</span>
				</div>
				<XButton :loading="saving" icon="save" @click="save">{{ $t('clientProfile.save') }}</XButton>
			</div>
		</div>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, defineComponent, h, onBeforeUnmount, onMounted, ref, shallowReactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'
import XButton from '@/components/input/Button.vue'

import ProjectService from '@/services/project'
import ClientProfileService from '@/services/clientProfile'
import ClientCustomFieldService from '@/services/clientCustomField'
import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'
import {useAccessStore} from '@/stores/access'
import {useAuthStore} from '@/stores/auth'
import {saveProjectToHistory} from '@/modules/projectHistory'
import {error, success, translatedError} from '@/message'
import {PERMISSIONS} from '@/constants/permissions'
import {ACCESS_PERMISSION} from '@/modelTypes/IAccessControl'
import {getDisplayName} from '@/models/user'

import type {ClientAddressType, ClientDecisionRole, ClientStatus, IClientAddress, IClientContactPerson, IClientProfile} from '@/modelTypes/IClientProfile'
import type {IUser} from '@/modelTypes/IUser'
import type {IClientCustomField} from '@/modelTypes/IClientCustomField'

const props = defineProps<{projectId: number}>()
const {t, locale} = useI18n()
const projectStore = useProjectStore()
const baseStore = useBaseStore()
const projectService = shallowReactive(new ProjectService())
const clientService = new ClientProfileService()
const customFieldService = new ClientCustomFieldService()
const accessStore = useAccessStore()
const authStore = useAuthStore()

const profile = ref<IClientProfile | null>(null)
const saving = ref(false)
const selectedResponsible = ref<IUser | null>(null)
const foundUsers = ref<IUser[]>([])
const responsibleUsersLoading = ref(false)
const geocodingType = ref<ClientAddressType | null>(null)
const openMaps = ref(new Set<ClientAddressType>())
const geocodeTimers = new Map<ClientAddressType, ReturnType<typeof setTimeout>>()
const proposalInput = ref<HTMLInputElement | null>(null)
const uploadingProposal = ref(false)
const deletingProposal = ref(false)
const downloadingProposal = ref(false)
const customFields = ref<IClientCustomField[]>([])
const savingCustomFieldIndex = ref<number | null>(null)
const deletingCustomFieldIndex = ref<number | null>(null)

const currentProject = computed(() => projectStore.projects[props.projectId])
const isLoadingProject = computed(() => projectService.loading || profile.value === null)
const canWrite = computed(() => Boolean(currentProject.value)
	&& !authStore.isLinkShareAuth
	&& accessStore.can(ACCESS_PERMISSION.PROJECTS_MANAGE)
	&& !currentProject.value.isArchived
	&& (currentProject.value.maxPermission === null || currentProject.value.maxPermission >= PERMISSIONS.READ_WRITE))

const statuses = computed<Array<{value: ClientStatus, label: string}>>(() => [
	{value: 'potential', label: t('clientProfile.statusPotential')},
	{value: 'active', label: t('clientProfile.statusActive')},
	{value: 'inactive', label: t('clientProfile.statusInactive')},
	{value: 'vip', label: 'VIP'},
])

const decisionRoles = computed<Array<{value: ClientDecisionRole, label: string}>>(() => [
	{value: '', label: '—'},
	{value: 'leader', label: t('clientProfile.roleLeader')},
	{value: 'decision_maker', label: t('clientProfile.roleDecisionMaker')},
	{value: 'technical', label: t('clientProfile.roleTechnical')},
	{value: 'procurement', label: t('clientProfile.roleProcurement')},
	{value: 'accountant', label: t('clientProfile.roleAccountant')},
	{value: 'user', label: t('clientProfile.roleUser')},
	{value: 'other', label: t('clientProfile.other')},
])

const addressTypes: ClientAddressType[] = ['legal', 'actual', 'postal', 'delivery', 'object']

const CrmInput = defineComponent({
	name: 'CrmInput',
	props: {
		modelValue: {type: [String, Number], default: ''},
		label: {type: String, required: true},
		disabled: Boolean,
		type: {type: String, default: 'text'},
	},
	emits: ['update:modelValue', 'blur'],
	setup(inputProps, {emit}) {
		return () => h('label', {class: 'crm-field'}, [
			h('span', inputProps.label),
			h('input', {
				class: 'input',
				type: inputProps.type,
				value: inputProps.modelValue,
				disabled: inputProps.disabled,
				onInput: (event: Event) => emit('update:modelValue', inputProps.type === 'number' ? Number((event.target as HTMLInputElement).value) : (event.target as HTMLInputElement).value),
				onBlur: () => emit('blur'),
			}),
		])
	},
})

function emptyAddress(type: ClientAddressType): IClientAddress {
	return {
		id: 0, project_id: props.projectId, type, country: '', region: '', city: '',
		street: '', house: '', office: '', postal_code: '', latitude: 0, longitude: 0,
	}
}

function normalizeProfile(data: IClientProfile): IClientProfile {
	data.addresses = addressTypes.map(type => data.addresses?.find(address => address.type === type) ?? emptyAddress(type))
	data.contact_persons = (data.contact_persons ?? []).map((person, index) => ({...person, position_index: index}))
	data.commercial_proposal ??= null
	return data
}

function emptyContactPerson(index: number): IClientContactPerson {
	return {
		id: 0, project_id: props.projectId, full_name: '', position: '', department: '', phone: '', email: '',
		telegram: '', viber: '', whatsapp: '', birthday: '', preferred_contact_method: '', decision_role: '',
		notes: '', position_index: index,
	}
}

async function load() {
	try {
		const [project, data] = await Promise.all([
			projectService.get({id: props.projectId}),
			clientService.get(props.projectId),
		])
		projectStore.setProject(project)
		baseStore.handleSetCurrentProject({project, currentProjectViewId: 0})
		profile.value = normalizeProfile(data)
		selectedResponsible.value = data.responsible
			? {...data.responsible, name: getDisplayName(data.responsible)}
			: null
		saveProjectToHistory({id: props.projectId})
	} catch (e) {
		error(e)
		return
	}

	try {
		customFields.value = await customFieldService.getAll(props.projectId)
	} catch {
		customFields.value = []
		error(translatedError('clientProfile.customFieldsLoadFailed'))
	}
}

watch(selectedResponsible, user => {
	if (profile.value) profile.value.responsible_user_id = user?.id ?? 0
})
watch(() => props.projectId, load)
onMounted(load)
onBeforeUnmount(() => {
	for (const timer of geocodeTimers.values()) clearTimeout(timer)
	geocodeTimers.clear()
})

async function save() {
	if (!profile.value) return
	saving.value = true
	try {
		profile.value.responsible_user_id = selectedResponsible.value?.id ?? 0
		profile.value = normalizeProfile(await clientService.save(props.projectId, profile.value))
		selectedResponsible.value = profile.value.responsible
			? {...profile.value.responsible, name: getDisplayName(profile.value.responsible)}
			: null

		// Saving the client name also updates the project title on the backend.
		const project = await projectService.get({id: props.projectId})
		projectStore.setProject(project)
		baseStore.handleSetCurrentProject({project, currentProjectViewId: 0})
		success({message: t('clientProfile.saved')})
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}

function addCustomField() {
	if (!canWrite.value) return
	customFields.value.push({
		id: 0,
		project_id: props.projectId,
		name: '',
		value: '',
		position: customFields.value.length,
		created: '',
		updated: '',
	})
}

async function saveCustomField(field: IClientCustomField, index: number) {
	if (!canWrite.value) return
	if (!field.name.trim()) {
		error(new Error(t('clientProfile.customFieldNameRequired')))
		return
	}
	savingCustomFieldIndex.value = index
	try {
		const saved = field.id
			? await customFieldService.update(props.projectId, field.id, field.name, field.value)
			: await customFieldService.create(props.projectId, field.name, field.value)
		customFields.value[index] = saved
		success({message: t('clientProfile.customFieldSaved')})
	} catch (e) {
		error(e)
	} finally {
		savingCustomFieldIndex.value = null
	}
}

async function deleteCustomField(field: IClientCustomField, index: number) {
	if (!canWrite.value) return
	if (field.id === 0) {
		customFields.value.splice(index, 1)
		return
	}
	if (!window.confirm(t('clientProfile.customFieldDeleteConfirm', {name: field.name}))) return
	deletingCustomFieldIndex.value = index
	try {
		await customFieldService.delete(props.projectId, field.id)
		customFields.value.splice(index, 1)
		success({message: t('clientProfile.customFieldDeleted')})
	} catch (e) {
		error(e)
	} finally {
		deletingCustomFieldIndex.value = null
	}
}

function statusLabel(value: ClientStatus) {
	return statuses.value.find(status => status.value === value)?.label ?? value
}

function decisionRoleLabel(value: ClientDecisionRole) {
	return decisionRoles.value.find(role => role.value === value)?.label ?? t('clientProfile.other')
}

function addContactPerson() {
	if (!profile.value || !canWrite.value) return
	profile.value.contact_persons.push(emptyContactPerson(profile.value.contact_persons.length))
}

function removeContactPerson(index: number) {
	if (!profile.value || !canWrite.value) return
	profile.value.contact_persons.splice(index, 1)
	profile.value.contact_persons.forEach((person, personIndex) => { person.position_index = personIndex })
}

function triggerProposalUpload() {
	proposalInput.value?.click()
}

async function onProposalSelected(event: Event) {
	const input = event.target as HTMLInputElement
	const file = input.files?.[0]
	input.value = ''
	if (!file || !profile.value) return
	if (file.type && file.type !== 'application/pdf' && !file.name.toLowerCase().endsWith('.pdf')) {
		error(new Error(t('clientProfile.proposalPdfOnly')))
		return
	}
	uploadingProposal.value = true
	try {
		profile.value = normalizeProfile(await clientService.uploadProposal(props.projectId, file))
		success({message: t('clientProfile.proposalUploaded')})
	} catch (e) {
		error(e)
	} finally {
		uploadingProposal.value = false
	}
}

async function downloadProposal() {
	if (!profile.value?.commercial_proposal) return
	downloadingProposal.value = true
	try {
		await clientService.downloadProposal(props.projectId, profile.value.commercial_proposal.name)
	} catch (e) {
		error(e)
	} finally {
		downloadingProposal.value = false
	}
}

async function deleteProposal() {
	if (!profile.value?.commercial_proposal || !canWrite.value) return
	if (!window.confirm(t('clientProfile.proposalDeleteConfirm'))) return
	deletingProposal.value = true
	try {
		profile.value = normalizeProfile(await clientService.deleteProposal(props.projectId))
		success({message: t('clientProfile.proposalDeleted')})
	} catch (e) {
		error(e)
	} finally {
		deletingProposal.value = false
	}
}

function formatFileSize(bytes: number) {
	if (!bytes) return '0 B'
	const units = ['B', 'KB', 'MB', 'GB']
	const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
	return `${(bytes / 1024 ** index).toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

async function findUsers(query = '') {
	responsibleUsersLoading.value = true
	try {
		const users = await clientService.searchProjectUsers(props.projectId, query)
		foundUsers.value = users.map(user => ({...user, name: getDisplayName(user)}))
	} catch (e) {
		error(e)
	} finally {
		responsibleUsersLoading.value = false
	}
}

function addressLabel(type: ClientAddressType) {
	const keys: Record<ClientAddressType, string> = {
		legal: 'clientProfile.addressLegal',
		actual: 'clientProfile.addressActual',
		postal: 'clientProfile.addressPostal',
		delivery: 'clientProfile.addressDelivery',
		object: 'clientProfile.addressObject',
	}
	return t(keys[type])
}

function shortAddress(address: IClientAddress) {
	return [address.city, address.street, address.house].filter(Boolean).join(', ') || t('clientProfile.notFilled')
}

function addressQuery(address: IClientAddress) {
	return [address.country, address.region, address.city, address.street, address.house].filter(Boolean).join(', ')
}

function scheduleAddressLookup(address: IClientAddress) {
	if (address.postal_code || !address.country || !address.city || !address.street || !address.house) return
	const previous = geocodeTimers.get(address.type)
	if (previous) clearTimeout(previous)
	geocodeTimers.set(address.type, setTimeout(() => void lookupAddress(address, false), 900))
}

async function lookupAddress(address: IClientAddress, openMap: boolean) {
	const query = addressQuery(address)
	if (query.length < 3) return
	geocodingType.value = address.type
	try {
		const result = await clientService.geocode(props.projectId, query)
		address.latitude = result.latitude
		address.longitude = result.longitude
		if (!address.postal_code && result.postal_code) address.postal_code = result.postal_code
		if (openMap) {
			const next = new Set(openMaps.value)
			next.add(address.type)
			openMaps.value = next
		}
	} catch (e) {
		if (openMap) error(e)
	} finally {
		geocodingType.value = null
	}
}

async function showAddressOnMap(address: IClientAddress) {
	await lookupAddress(address, true)
}

function mapUrl(address: IClientAddress) {
	const delta = 0.008
	const minLon = address.longitude - delta
	const maxLon = address.longitude + delta
	const minLat = address.latitude - delta
	const maxLat = address.latitude + delta
	return `https://www.openstreetmap.org/export/embed.html?bbox=${minLon}%2C${minLat}%2C${maxLon}%2C${maxLat}&layer=mapnik&marker=${address.latitude}%2C${address.longitude}`
}

function formatDate(value: string | Date) {
	return value
		? new Intl.DateTimeFormat(locale.value, {dateStyle: 'medium'}).format(new Date(value))
		: '—'
}
</script>

<style lang="scss" scoped>
.crm-client { max-inline-size: 1480px; margin: 0 auto; padding-block-end: 5rem; color: var(--brand-text, #21352f); }
.crm-client__hero { display:flex; align-items:flex-start; justify-content:space-between; gap:2rem; margin: .4rem 0 1.2rem; padding: 1.35rem 1.5rem; background: linear-gradient(115deg, var(--brand-forest, #153f34), #285f4f); color:#fff; border-radius:18px; box-shadow:0 12px 30px rgba(21,63,52,.16); }
.crm-client__hero h2 { margin:.15rem 0 .2rem; color:#fff; font-size:1.6rem; }
.crm-client__hero p { margin:0; color:rgba(255,255,255,.72); }
.crm-client__eyebrow { text-transform:uppercase; letter-spacing:.11em; font-size:.69rem; font-weight:800; color:var(--brand-accent, #d8ff80); }
.crm-card { background:rgba(255,255,255,.96); border:1px solid var(--brand-border, #dfe8df); border-radius:17px; padding:1.3rem; margin-block-end:1rem; box-shadow:0 7px 22px rgba(31,91,73,.055); }
.crm-section-heading { display:flex; align-items:flex-start; justify-content:space-between; gap:1rem; margin-block-end:1.05rem; }
.crm-section-heading h3 { font-size:1.03rem; margin:0 0 .15rem; color:var(--brand-forest, #153f34); }
.crm-section-heading p { margin:0; font-size:.8rem; color:var(--brand-text-muted, #61726c); }
.crm-added-date { display:flex; flex-direction:column; align-items:flex-end; font-size:.76rem; color:var(--brand-text-muted,#61726c); }
.crm-added-date strong { color:var(--brand-forest,#153f34); font-size:.86rem; }
.crm-form-grid { display:grid; gap:.9rem 1rem; align-items:end; }
.crm-form-grid--2 { grid-template-columns:repeat(2,minmax(0,1fr)); }
.crm-form-grid--3 { grid-template-columns:repeat(3,minmax(0,1fr)); }
.crm-form-grid--4 { grid-template-columns:repeat(4,minmax(0,1fr)); }
:deep(.crm-field) { display:flex; flex-direction:column; gap:.34rem; min-inline-size:0; }
:deep(.crm-field > span) { color:var(--brand-text-muted,#61726c); font-weight:700; font-size:.74rem; }
:deep(.crm-field .input), :deep(.crm-field .textarea) { background:#fbfdfb; border-color:var(--brand-border,#dfe8df); box-shadow:none; border-radius:10px; }
:deep(.crm-field .input:focus), :deep(.crm-field .textarea:focus) { border-color:var(--brand-green,#1f5b49); box-shadow:0 0 0 3px rgba(31,91,73,.09); }
.crm-field--wide { grid-column:span 2; }
.crm-field--full { grid-column:1/-1; }
.crm-field--status { margin:1.05rem 0; }
.crm-field--status > span { display:block; margin-block-end:.35rem; }
.crm-status { display:grid; grid-template-columns:repeat(4,1fr); border-radius:11px; overflow:hidden; border:1px solid var(--brand-border,#dfe8df); }
.crm-status__item { border:0; padding:.62rem .8rem; font-weight:750; font-size:.78rem; cursor:pointer; color:#52645e; opacity:.66; transition:.15s ease; }
.crm-status__item:disabled { cursor:default; }
.crm-status__item--potential { background:#fff4c7; }
.crm-status__item--active { background:#d9f4e4; }
.crm-status__item--inactive { background:#edf0ee; }
.crm-status__item--vip { background:#eadcfb; }
.crm-status__item.is-active { opacity:1; color:#173d32; box-shadow:inset 0 -3px 0 rgba(15,62,48,.45); transform:translateY(-1px); }
.crm-addresses { display:flex; flex-direction:column; gap:.65rem; }
.crm-address { border:1px solid var(--brand-border,#dfe8df); border-radius:12px; overflow:hidden; background:#fbfdfb; }
.crm-address summary { cursor:pointer; display:flex; align-items:center; gap:1rem; padding:.8rem 1rem; font-weight:800; color:var(--brand-forest,#153f34); list-style:none; }
.crm-address summary::-webkit-details-marker { display:none; }
.crm-address summary::before { content:'›'; font-size:1.3rem; transition:transform .15s; }
.crm-address[open] summary::before { transform:rotate(90deg); }
.crm-address summary small { margin-inline-start:auto; font-weight:500; color:var(--brand-text-muted,#61726c); }
.crm-address__body { border-block-start:1px solid var(--brand-border,#dfe8df); padding:1rem; background:#fff; }
.crm-address__actions { margin-block-start:.8rem; display:flex; align-items:center; gap:.8rem; }
.crm-coordinates { font-size:.75rem; color:var(--brand-text-muted,#61726c); }
.crm-map { inline-size:100%; block-size:300px; border:0; border-radius:12px; margin-block-start:.8rem; }
.crm-card--custom-fields { margin-block-end:1rem; }
.crm-custom-fields { display:flex; flex-direction:column; gap:.7rem; }
.crm-custom-field { display:grid; grid-template-columns:minmax(180px,.8fr) minmax(240px,1.5fr) auto; gap:.75rem; align-items:end; padding:.8rem; border:1px solid var(--brand-border,#dfe8df); border-radius:12px; background:#fbfdfb; }
.crm-custom-field__actions { display:flex; gap:.45rem; align-items:flex-end; }
.crm-custom-field__value { min-inline-size:0; }
.crm-savebar { position:sticky; z-index:5; inset-block-end:1rem; display:flex; align-items:center; justify-content:space-between; gap:1rem; padding:.8rem 1rem; border:1px solid rgba(31,91,73,.18); border-radius:14px; background:rgba(255,255,255,.94); backdrop-filter:blur(12px); box-shadow:0 10px 28px rgba(21,63,52,.14); }
.crm-savebar div { display:flex; flex-direction:column; }
.crm-savebar span { font-size:.72rem; color:var(--brand-text-muted,#61726c); }
.crm-client__summary { display:flex; flex-wrap:wrap; gap:.45rem; margin-block-start:.75rem; }
.crm-client__summary span { display:inline-flex; align-items:center; min-block-size:1.7rem; padding:.25rem .6rem; border:1px solid rgba(216,255,128,.24); border-radius:999px; background:rgba(255,255,255,.08); color:rgba(255,255,255,.9); font-size:.72rem; font-weight:700; }
.crm-card--proposal { overflow:hidden; }
.crm-proposal { display:flex; align-items:center; gap:1rem; padding:1rem; border:1px solid var(--brand-border,#dfe8df); border-radius:13px; background:#fbfdfb; }
.crm-proposal__badge { display:grid; place-items:center; inline-size:50px; block-size:56px; flex:0 0 50px; border-radius:10px; background:#e5484d; color:#fff; font-weight:900; font-size:.78rem; letter-spacing:.08em; box-shadow:0 6px 16px rgba(229,72,77,.16); }
.crm-proposal__meta { display:flex; flex:1; min-inline-size:0; flex-direction:column; }
.crm-proposal__meta strong { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; color:var(--brand-forest,#153f34); }
.crm-proposal__meta span { margin-block-start:.15rem; color:var(--brand-text-muted,#61726c); font-size:.75rem; }
.crm-proposal__actions { display:flex; gap:.5rem; flex-wrap:wrap; }
.crm-empty-state { display:flex; flex-direction:column; align-items:flex-start; gap:.2rem; padding:1.1rem; border:1px dashed #cddbd1; border-radius:12px; background:#fafcfa; color:var(--brand-text-muted,#61726c); }
.crm-empty-state strong { color:var(--brand-forest,#153f34); }
.crm-empty-state span { font-size:.78rem; }
.crm-contact-persons { display:flex; flex-direction:column; gap:.85rem; }
.crm-contact-person { border:1px solid var(--brand-border,#dfe8df); border-radius:14px; padding:1rem; background:linear-gradient(180deg,#fff,#fbfdfb); }
.crm-contact-person__header { display:flex; align-items:center; justify-content:space-between; gap:1rem; margin-block-end:1rem; padding-block-end:.75rem; border-block-end:1px solid #edf2ed; }
.crm-contact-person__header > div { display:flex; align-items:center; gap:.75rem; min-inline-size:0; }
.crm-contact-person__header > div > div { display:flex; min-inline-size:0; flex-direction:column; }
.crm-contact-person__header strong { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; color:var(--brand-forest,#153f34); }
.crm-contact-person__header small { color:var(--brand-text-muted,#61726c); }
.crm-contact-person__number { display:grid; place-items:center; inline-size:34px; block-size:34px; border-radius:10px; background:#eaf2eb; color:var(--brand-green,#1f5b49); font-size:.7rem; font-weight:900; }
.is-hidden { display:none !important; }
@media (max-width:1024px) { .crm-form-grid--4,.crm-form-grid--3 { grid-template-columns:repeat(2,minmax(0,1fr)); } .crm-field--wide{grid-column:span 1;} }
@media (max-width:700px) { .crm-client__hero,.crm-section-heading,.crm-savebar { align-items:stretch; flex-direction:column; } .crm-form-grid--2,.crm-form-grid--3,.crm-form-grid--4 { grid-template-columns:1fr; } .crm-status { grid-template-columns:1fr 1fr; } .crm-added-date{align-items:flex-start;} .crm-proposal{align-items:flex-start;flex-direction:column;} .crm-proposal__actions{inline-size:100%;} .crm-contact-person__header{align-items:flex-start;flex-direction:column;} .crm-custom-field{grid-template-columns:1fr;} .crm-custom-field__actions{flex-wrap:wrap;} }
</style>

<template>
	<ProjectWrapper
		:is-loading-project="loading"
		:project-id="projectId"
		:view-id="0"
	>
		<div class="client-history">
			<header class="history-hero">
				<div>
					<div class="history-hero__eyebrow">{{ $t('clientHistory.eyebrow') }}</div>
					<h2>{{ $t('clientHistory.title') }}</h2>
					<p>{{ $t('clientHistory.subtitle') }}</p>
				</div>
				<XButton v-if="canWrite" icon="plus" @click="showComposer = !showComposer">
					{{ $t('clientHistory.add') }}
				</XButton>
			</header>

			<section v-if="showComposer && canWrite" class="history-composer">
				<div class="history-section-heading">
					<div>
						<h3>{{ $t('clientHistory.newActivity') }}</h3>
						<p>{{ $t('clientHistory.newActivityHint') }}</p>
					</div>
				</div>

				<div class="history-form-grid history-form-grid--3">
					<label class="history-field">
						<span>{{ $t('clientHistory.type') }}</span>
						<select v-model="draft.event_type" class="select input">
							<option v-for="type in manualTypes" :key="type.value" :value="type.value">{{ type.label }}</option>
						</select>
					</label>
					<label class="history-field">
						<span>{{ $t('clientHistory.when') }}</span>
						<input v-model="draftOccurredLocal" class="input" type="datetime-local">
					</label>
					<label class="history-field">
						<span>{{ $t('clientHistory.contactPerson') }}</span>
						<select v-model.number="draft.metadata.contact_person_id" class="select input">
							<option :value="0">—</option>
							<option v-for="person in profile?.contact_persons ?? []" :key="person.id" :value="person.id">{{ person.full_name }}</option>
						</select>
					</label>

					<label v-if="showsDirection" class="history-field">
						<span>{{ $t('clientHistory.direction') }}</span>
						<select v-model="draft.metadata.direction" class="select input">
							<option value="">—</option>
							<option value="incoming">{{ $t('clientHistory.incoming') }}</option>
							<option value="outgoing">{{ $t('clientHistory.outgoing') }}</option>
						</select>
					</label>
					<label v-if="showsChannel" class="history-field">
						<span>{{ $t('clientHistory.channel') }}</span>
						<select v-model="draft.metadata.channel" class="select input">
							<option value="">—</option>
							<option value="telegram">Telegram</option>
							<option value="viber">Viber</option>
							<option value="whatsapp">WhatsApp</option>
							<option value="email">Email</option>
							<option value="phone">{{ $t('clientHistory.phoneChannel') }}</option>
							<option value="other">{{ $t('clientHistory.other') }}</option>
						</select>
					</label>
					<label v-if="showsDuration" class="history-field">
						<span>{{ $t('clientHistory.duration') }}</span>
						<input v-model.number="draft.metadata.duration_minutes" class="input" type="number" min="0" step="1">
					</label>
					<label class="history-field history-field--full">
						<span>{{ $t('clientHistory.subject') }}</span>
						<input v-model="draft.title" class="input" type="text" :placeholder="$t('clientHistory.subjectPlaceholder')">
					</label>
					<label class="history-field history-field--full">
						<span>{{ $t('clientHistory.description') }}</span>
						<textarea v-model="draft.description" class="textarea" rows="4" :placeholder="$t('clientHistory.descriptionPlaceholder')" />
					</label>
					<label class="history-field history-field--full">
						<span>{{ $t('clientHistory.result') }}</span>
						<input v-model="draft.metadata.result" class="input" type="text" :placeholder="$t('clientHistory.resultPlaceholder')">
					</label>
				</div>
				<div class="history-composer__actions">
					<XButton variant="secondary" @click="cancelComposer">{{ $t('misc.cancel') }}</XButton>
					<XButton :loading="creating" icon="save" @click="createActivity">{{ $t('clientHistory.saveActivity') }}</XButton>
				</div>
			</section>

			<section class="history-toolbar">
				<label>
					<span>{{ $t('clientHistory.filter') }}</span>
					<select v-model="filterType" class="select input" @change="reloadHistory">
						<option value="">{{ $t('clientHistory.allEvents') }}</option>
						<option value="call">{{ $t('clientHistory.calls') }}</option>
						<option value="message">{{ $t('clientHistory.messages') }}</option>
						<option value="meeting">{{ $t('clientHistory.meetings') }}</option>
						<option value="manual_note">{{ $t('clientHistory.notes') }}</option>
						<option value="task_created">{{ $t('clientHistory.tasksCreated') }}</option>
						<option value="task_completed">{{ $t('clientHistory.tasksCompleted') }}</option>
						<option value="comment_created">{{ $t('clientHistory.comments') }}</option>
					</select>
				</label>
				<div class="history-toolbar__count">{{ $t('clientHistory.eventsCount', {count: total}) }}</div>
			</section>

			<div v-if="!events.length && !loadingHistory" class="history-empty">
				<div class="history-empty__icon"><Icon icon="history" /></div>
				<strong>{{ $t('clientHistory.empty') }}</strong>
				<span>{{ $t('clientHistory.emptyHint') }}</span>
			</div>

			<div v-else class="history-days">
				<section v-for="group in groupedEvents" :key="group.key" class="history-day">
					<div class="history-day__label">{{ group.label }}</div>
					<div class="history-timeline">
						<article v-for="event in group.events" :key="event.id" class="history-event">
							<div :class="['history-event__icon', `history-event__icon--${eventCategory(event.event_type)}`]">
								<Icon :icon="eventIcon(event.event_type)" />
							</div>
							<div class="history-event__body">
								<div class="history-event__topline">
									<div>
										<strong>{{ eventTitle(event) }}</strong>
										<span>{{ actorName(event) }} · {{ formatTime(event.occurred_at) }}</span>
									</div>
									<span v-if="event.system_generated" class="history-event__system">{{ $t('clientHistory.automatic') }}</span>
								</div>

								<p v-if="event.description" class="history-event__description">{{ plainText(event.description) }}</p>

								<div v-if="eventMetadataRows(event).length" class="history-event__meta">
									<span v-for="row in eventMetadataRows(event)" :key="row">{{ row }}</span>
								</div>

								<div class="history-event__actions">
									<RouterLink v-if="event.entity_type === 'task' && event.entity_id" class="history-link" :to="{name: 'task.detail', params: {id: event.entity_id}}">
										{{ $t('clientHistory.openTask') }}
									</RouterLink>
									<RouterLink v-else-if="event.entity_type === 'commercial_proposal' || event.entity_type === 'custom_field'" class="history-link" :to="{name: 'project.client', params: {projectId}}">
										{{ $t('clientHistory.openClient') }}
									</RouterLink>
									<button v-if="canWrite && !event.system_generated" type="button" class="history-delete" @click="deleteActivity(event)">{{ $t('misc.delete') }}</button>
								</div>
							</div>
						</article>
					</div>
				</section>
			</div>

			<div v-if="events.length < total" class="history-load-more">
				<XButton variant="secondary" :loading="loadingMore" @click="loadMore">{{ $t('clientHistory.loadMore') }}</XButton>
			</div>
		</div>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, onMounted, reactive, ref, shallowReactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import XButton from '@/components/input/Button.vue'
import Icon from '@/components/misc/Icon'

import ProjectService from '@/services/project'
import ClientProfileService from '@/services/clientProfile'
import ClientActivityService from '@/services/clientActivity'
import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'
import {useAccessStore} from '@/stores/access'
import {useAuthStore} from '@/stores/auth'
import {saveProjectToHistory} from '@/modules/projectHistory'
import {error, success} from '@/message'
import {PERMISSIONS} from '@/constants/permissions'
import {ACCESS_PERMISSION} from '@/modelTypes/IAccessControl'
import {getDisplayName} from '@/models/user'

import type {IClientProfile} from '@/modelTypes/IClientProfile'
import type {ClientActivityType, IClientActivityCreate, IClientActivityEvent} from '@/modelTypes/IClientActivity'

const props = defineProps<{projectId: number}>()
const {t, locale} = useI18n()
const projectStore = useProjectStore()
const baseStore = useBaseStore()
const accessStore = useAccessStore()
const authStore = useAuthStore()
const projectService = shallowReactive(new ProjectService())
const clientService = new ClientProfileService()
const activityService = new ClientActivityService()

const profile = ref<IClientProfile | null>(null)
const events = ref<IClientActivityEvent[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 50
const filterType = ref('')
const loadingHistory = ref(false)
const loadingMore = ref(false)
const creating = ref(false)
const showComposer = ref(false)
const draftOccurredLocal = ref(toLocalDateTimeInput(new Date()))

const draft = reactive<IClientActivityCreate>(emptyDraft())

const currentProject = computed(() => projectStore.projects[props.projectId])
const loading = computed(() => projectService.loading || profile.value === null || loadingHistory.value && events.value.length === 0)
const canWrite = computed(() => Boolean(currentProject.value)
	&& !authStore.isLinkShareAuth
	&& accessStore.can(ACCESS_PERMISSION.PROJECTS_MANAGE)
	&& !currentProject.value.isArchived
	&& (currentProject.value.maxPermission === null || currentProject.value.maxPermission >= PERMISSIONS.READ_WRITE))

const manualTypes = computed(() => [
	{value: 'call', label: t('clientHistory.calls')},
	{value: 'message', label: t('clientHistory.messages')},
	{value: 'meeting', label: t('clientHistory.meetings')},
	{value: 'manual_note', label: t('clientHistory.notes')},
	{value: 'document_sent', label: t('clientHistory.documentSent')},
	{value: 'commercial_proposal_sent', label: t('clientHistory.proposalSent')},
	{value: 'invoice_sent', label: t('clientHistory.invoiceSent')},
] as const)

const showsDirection = computed(() => draft.event_type === 'call' || draft.event_type === 'message')
const showsChannel = computed(() => draft.event_type === 'message')
const showsDuration = computed(() => draft.event_type === 'call' || draft.event_type === 'meeting')

const groupedEvents = computed(() => {
	const groups: Array<{key: string, label: string, events: IClientActivityEvent[]}> = []
	const map = new Map<string, {key: string, label: string, events: IClientActivityEvent[]}>()
	for (const event of events.value) {
		const date = new Date(event.occurred_at)
		const key = `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
		let group = map.get(key)
		if (!group) {
			group = {key, label: formatDay(date), events: []}
			map.set(key, group)
			groups.push(group)
		}
		group.events.push(event)
	}
	return groups
})

function emptyDraft(): IClientActivityCreate {
	return {
		event_type: 'call',
		occurred_at: new Date().toISOString(),
		title: '',
		description: '',
		metadata: {direction: '', channel: '', duration_minutes: 0, result: '', contact_person_id: 0},
	}
}

function resetDraft() {
	Object.assign(draft, emptyDraft())
	draftOccurredLocal.value = toLocalDateTimeInput(new Date())
}

function cancelComposer() {
	showComposer.value = false
	resetDraft()
}

async function loadBase() {
	try {
		const [project, client] = await Promise.all([
			projectService.get({id: props.projectId}),
			clientService.get(props.projectId),
		])
		projectStore.setProject(project)
		baseStore.handleSetCurrentProject({project, currentProjectViewId: 0})
		profile.value = client
		saveProjectToHistory({id: props.projectId})
		await reloadHistory()
	} catch (e) {
		error(e)
	}
}

async function reloadHistory() {
	page.value = 1
	loadingHistory.value = true
	try {
		const result = await activityService.getAll(props.projectId, page.value, perPage, filterType.value)
		events.value = result.items
		total.value = result.total
	} catch (e) {
		error(e)
	} finally {
		loadingHistory.value = false
	}
}

async function loadMore() {
	loadingMore.value = true
	try {
		const nextPage = page.value + 1
		const result = await activityService.getAll(props.projectId, nextPage, perPage, filterType.value)
		events.value.push(...result.items)
		page.value = nextPage
		total.value = result.total
	} catch (e) {
		error(e)
	} finally {
		loadingMore.value = false
	}
}

async function createActivity() {
	creating.value = true
	try {
		draft.occurred_at = new Date(draftOccurredLocal.value).toISOString()
		await activityService.create(props.projectId, draft)
		success({message: t('clientHistory.saved')})
		showComposer.value = false
		resetDraft()
		await reloadHistory()
	} catch (e) {
		error(e)
	} finally {
		creating.value = false
	}
}

async function deleteActivity(event: IClientActivityEvent) {
	if (!window.confirm(t('clientHistory.deleteConfirm'))) return
	try {
		await activityService.delete(props.projectId, event.id)
		events.value = events.value.filter(item => item.id !== event.id)
		total.value = Math.max(0, total.value - 1)
		success({message: t('clientHistory.deleted')})
	} catch (e) {
		error(e)
	}
}

function eventTitle(event: IClientActivityEvent) {
	if (!event.system_generated && event.title) return event.title
	const meta = event.metadata ?? {}
	switch (event.event_type) {
		case 'call': return t('clientHistory.calls')
		case 'message': return t('clientHistory.messages')
		case 'meeting': return t('clientHistory.meetings')
		case 'manual_note': return t('clientHistory.note')
		case 'document_sent': return t('clientHistory.documentSent')
		case 'commercial_proposal_sent': return t('clientHistory.proposalSent')
		case 'invoice_sent': return t('clientHistory.invoiceSent')
		case 'task_created': return t('clientHistory.eventTaskCreated', {title: meta.task_title || `#${event.entity_id}`})
		case 'task_completed': return t('clientHistory.eventTaskCompleted', {title: meta.task_title || `#${event.entity_id}`})
		case 'task_reopened': return t('clientHistory.eventTaskReopened', {title: meta.task_title || `#${event.entity_id}`})
		case 'comment_created': return t('clientHistory.eventCommentCreated', {title: meta.task_title || `#${event.entity_id}`})
		case 'status_changed': return t('clientHistory.eventStatusChanged')
		case 'responsible_changed': return t('clientHistory.eventResponsibleChanged')
		case 'commercial_proposal_uploaded': return t('clientHistory.eventProposalUploaded')
		case 'commercial_proposal_replaced': return t('clientHistory.eventProposalReplaced')
		case 'commercial_proposal_deleted': return t('clientHistory.eventProposalDeleted')
		case 'custom_field_created': return t('clientHistory.eventCustomFieldCreated', {name: meta.field_name || `#${event.entity_id}`})
		case 'custom_field_updated': return t('clientHistory.eventCustomFieldUpdated', {name: meta.field_name || `#${event.entity_id}`})
		case 'custom_field_deleted': return t('clientHistory.eventCustomFieldDeleted', {name: meta.field_name || `#${event.entity_id}`})
		default: return event.event_type
	}
}

function eventMetadataRows(event: IClientActivityEvent) {
	const meta = event.metadata
	if (!meta) return []
	const rows: string[] = []
	if (event.event_type === 'status_changed') {
		rows.push(`${clientStatusLabel(meta.old_value)} → ${clientStatusLabel(meta.new_value)}`)
	}
	if (event.event_type === 'responsible_changed') {
		rows.push(`${meta.old_value || t('clientHistory.notAssigned')} → ${meta.new_value || t('clientHistory.notAssigned')}`)
	}
	if (event.event_type === 'custom_field_created' && meta.new_value !== undefined) {
		rows.push(`${meta.field_name || t('clientHistory.customField')}: ${meta.new_value || '—'}`)
	}
	if (event.event_type === 'custom_field_updated') {
		if (meta.old_field_name && meta.new_field_name && meta.old_field_name !== meta.new_field_name) {
			rows.push(`${t('clientHistory.fieldName')}: ${meta.old_field_name} → ${meta.new_field_name}`)
		}
		if (meta.old_value !== undefined || meta.new_value !== undefined) {
			rows.push(`${t('clientHistory.value')}: ${meta.old_value || '—'} → ${meta.new_value || '—'}`)
		}
	}
	if (event.event_type === 'custom_field_deleted') {
		rows.push(`${meta.field_name || t('clientHistory.customField')}: ${meta.old_value || '—'}`)
	}
	if (meta.contact_person_name) rows.push(`${t('clientHistory.contactPerson')}: ${meta.contact_person_name}`)
	if (meta.direction) rows.push(meta.direction === 'incoming' ? t('clientHistory.incoming') : t('clientHistory.outgoing'))
	if (meta.channel) rows.push(`${t('clientHistory.channel')}: ${channelLabel(meta.channel)}`)
	if (meta.duration_minutes) rows.push(`${t('clientHistory.duration')}: ${meta.duration_minutes} ${t('clientHistory.minutes')}`)
	if (meta.result) rows.push(`${t('clientHistory.result')}: ${meta.result}`)
	if (meta.file_name) rows.push(meta.file_name)
	return rows
}

function clientStatusLabel(value?: string) {
	switch (value) {
		case 'potential': return t('clientProfile.statusPotential')
		case 'active': return t('clientProfile.statusActive')
		case 'inactive': return t('clientProfile.statusInactive')
		case 'vip': return 'VIP'
		default: return value || '—'
	}
}

function channelLabel(channel: string) {
	if (channel === 'phone') return t('clientHistory.phoneChannel')
	if (channel === 'other') return t('clientHistory.other')
	return channel.charAt(0).toUpperCase() + channel.slice(1)
}

function eventCategory(type: ClientActivityType) {
	if (type.startsWith('task_')) return 'task'
	if (type.includes('proposal') || type.includes('document') || type.includes('invoice')) return 'document'
	if (type.includes('status') || type.includes('responsible') || type.startsWith('custom_field_')) return 'change'
	if (type === 'comment_created' || type === 'message') return 'message'
	if (type === 'call') return 'call'
	if (type === 'meeting') return 'meeting'
	return 'note'
}

function eventIcon(type: ClientActivityType) {
	const category = eventCategory(type)
	if (category === 'task') return type === 'task_completed' ? 'check' : 'tasks'
	if (category === 'document') return 'file'
	if (category === 'change') return 'history'
	if (category === 'message') return 'comments'
	if (category === 'meeting') return 'calendar'
	if (category === 'call') return 'clock'
	return 'pen'
}

function actorName(event: IClientActivityEvent) {
	return event.actor ? getDisplayName(event.actor) : t('clientHistory.system')
}

function plainText(value: string) {
	if (!value.includes('<')) return value
	const doc = new DOMParser().parseFromString(`<div>${value}</div>`, 'text/html')
	return doc.body.textContent?.trim() || ''
}

function formatTime(value: string | Date) {
	return new Intl.DateTimeFormat(locale.value, {hour: '2-digit', minute: '2-digit'}).format(new Date(value))
}

function formatDay(value: Date) {
	const today = new Date()
	const yesterday = new Date(today)
	yesterday.setDate(today.getDate() - 1)
	if (sameDay(value, today)) return t('clientHistory.today')
	if (sameDay(value, yesterday)) return t('clientHistory.yesterday')
	return new Intl.DateTimeFormat(locale.value, {day: '2-digit', month: 'long', year: 'numeric'}).format(value)
}

function sameDay(a: Date, b: Date) {
	return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

function toLocalDateTimeInput(value: Date) {
	const pad = (n: number) => String(n).padStart(2, '0')
	return `${value.getFullYear()}-${pad(value.getMonth() + 1)}-${pad(value.getDate())}T${pad(value.getHours())}:${pad(value.getMinutes())}`
}

watch(() => props.projectId, loadBase)
onMounted(loadBase)
</script>

<style lang="scss" scoped>
.client-history { max-inline-size: 1180px; margin:0 auto; padding-block-end:5rem; color:var(--brand-text,#21352f); }
.history-hero { display:flex; align-items:flex-start; justify-content:space-between; gap:1.5rem; padding:1.35rem 1.5rem; margin:.4rem 0 1rem; border-radius:18px; background:linear-gradient(115deg,var(--brand-forest,#153f34),#285f4f); color:#fff; box-shadow:0 12px 30px rgba(21,63,52,.16); }
.history-hero h2 { margin:.12rem 0 .2rem; color:#fff; font-size:1.55rem; }
.history-hero p { margin:0; color:rgba(255,255,255,.72); }
.history-hero__eyebrow { text-transform:uppercase; letter-spacing:.11em; color:var(--brand-accent,#d8ff80); font-size:.68rem; font-weight:800; }
.history-composer,.history-toolbar { background:rgba(255,255,255,.97); border:1px solid var(--brand-border,#dfe8df); border-radius:16px; box-shadow:0 7px 22px rgba(31,91,73,.055); }
.history-composer { padding:1.2rem; margin-block-end:1rem; }
.history-section-heading { margin-block-end:1rem; }
.history-section-heading h3 { margin:0 0 .15rem; color:var(--brand-forest,#153f34); font-size:1rem; }
.history-section-heading p { margin:0; color:var(--brand-text-muted,#61726c); font-size:.8rem; }
.history-form-grid { display:grid; gap:.85rem 1rem; }
.history-form-grid--3 { grid-template-columns:repeat(3,minmax(0,1fr)); }
.history-field { display:flex; flex-direction:column; gap:.32rem; min-inline-size:0; }
.history-field > span,.history-toolbar label > span { color:var(--brand-text-muted,#61726c); font-size:.73rem; font-weight:750; }
.history-field--full { grid-column:1/-1; }
.history-field .input,.history-field .textarea,.history-toolbar .input { border-color:var(--brand-border,#dfe8df); border-radius:10px; box-shadow:none; background:#fbfdfb; }
.history-composer__actions { display:flex; justify-content:flex-end; gap:.55rem; margin-block-start:1rem; }
.history-toolbar { display:flex; align-items:end; justify-content:space-between; gap:1rem; padding:.8rem 1rem; margin-block-end:1rem; }
.history-toolbar label { display:flex; flex-direction:column; gap:.25rem; min-inline-size:220px; }
.history-toolbar__count { padding-block-end:.5rem; color:var(--brand-text-muted,#61726c); font-size:.78rem; font-weight:700; }
.history-days { display:flex; flex-direction:column; gap:1rem; }
.history-day__label { position:sticky; inset-block-start:.5rem; z-index:2; inline-size:max-content; margin:0 0 .55rem 3.6rem; padding:.3rem .65rem; border:1px solid var(--brand-border,#dfe8df); border-radius:999px; background:rgba(244,247,242,.94); backdrop-filter:blur(8px); color:var(--brand-forest,#153f34); font-size:.72rem; font-weight:800; }
.history-timeline { position:relative; display:flex; flex-direction:column; gap:.65rem; }
.history-timeline::before { content:''; position:absolute; inset-block:0; inset-inline-start:1.55rem; inline-size:2px; background:#dbe6dc; }
.history-event { position:relative; display:grid; grid-template-columns:3.2rem minmax(0,1fr); gap:.7rem; }
.history-event__icon { z-index:1; display:grid; place-items:center; inline-size:2.35rem; block-size:2.35rem; margin:.75rem auto 0; border:4px solid var(--brand-page,#f4f7f2); border-radius:50%; background:#eaf2eb; color:var(--brand-green,#1f5b49); }
.history-event__icon--call { background:#e2f1ff; color:#2874a6; }
.history-event__icon--message { background:#ece7ff; color:#6651a6; }
.history-event__icon--meeting { background:#fff1d6; color:#9b6b12; }
.history-event__icon--task { background:#dff3e6; color:#247044; }
.history-event__icon--document { background:#ffe5e5; color:#aa4545; }
.history-event__icon--change { background:#eef0ee; color:#56645f; }
.history-event__body { padding:.85rem 1rem; border:1px solid var(--brand-border,#dfe8df); border-radius:14px; background:rgba(255,255,255,.97); box-shadow:0 4px 16px rgba(31,91,73,.045); }
.history-event__topline { display:flex; align-items:flex-start; justify-content:space-between; gap:1rem; }
.history-event__topline > div { display:flex; flex-direction:column; }
.history-event__topline strong { color:var(--brand-forest,#153f34); }
.history-event__topline span { color:var(--brand-text-muted,#61726c); font-size:.72rem; }
.history-event__system { padding:.18rem .45rem; border-radius:999px; background:#eef4ed; white-space:nowrap; font-size:.65rem !important; font-weight:800; }
.history-event__description { margin:.65rem 0 0; white-space:pre-wrap; line-height:1.5; }
.history-event__meta { display:flex; flex-wrap:wrap; gap:.35rem; margin-block-start:.65rem; }
.history-event__meta span { padding:.25rem .5rem; border-radius:8px; background:#f2f6f2; color:#53665f; font-size:.7rem; }
.history-event__actions { display:flex; align-items:center; gap:.8rem; margin-block-start:.7rem; }
.history-link,.history-delete { border:0; background:none; padding:0; color:var(--brand-green,#1f5b49); font-size:.73rem; font-weight:800; cursor:pointer; }
.history-delete { color:#ad4b4b; }
.history-empty { display:flex; flex-direction:column; align-items:center; text-align:center; gap:.25rem; padding:3rem 1rem; border:1px dashed #cbd9ce; border-radius:16px; background:rgba(255,255,255,.7); color:var(--brand-text-muted,#61726c); }
.history-empty strong { color:var(--brand-forest,#153f34); }
.history-empty__icon { display:grid; place-items:center; inline-size:46px; block-size:46px; margin-block-end:.35rem; border-radius:14px; background:#eaf2eb; color:var(--brand-green,#1f5b49); font-size:1.1rem; }
.history-load-more { display:flex; justify-content:center; margin-block-start:1.25rem; }
@media (max-width:900px) { .history-form-grid--3{grid-template-columns:repeat(2,minmax(0,1fr));} }
@media (max-width:650px) { .history-hero,.history-toolbar{align-items:stretch;flex-direction:column;} .history-form-grid--3{grid-template-columns:1fr;} .history-field--full{grid-column:auto;} .history-toolbar label{min-inline-size:0;} .history-day__label{margin-inline-start:0;} .history-event{grid-template-columns:2.7rem minmax(0,1fr);} .history-timeline::before{inset-inline-start:1.3rem;} .history-event__icon{inline-size:2rem;block-size:2rem;} .history-event__topline{flex-direction:column;gap:.4rem;} }
</style>

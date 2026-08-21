<template>
	<div class="mailbox-page">
		<header class="mailbox-header">
			<div>
				<span class="mailbox-eyebrow">{{ $t('mailbox.eyebrow') }}</span>
				<h1>{{ $t('mailbox.title') }}</h1>
				<p>{{ $t('mailbox.subtitle') }}</p>
			</div>
			<XButton icon="plus" @click="startCompose()">{{ $t('mailbox.compose') }}</XButton>
		</header>

		<section v-if="composing" class="mailbox-compose">
			<div class="mailbox-compose__heading">
				<div>
					<h2>{{ draft.reply_to_id ? $t('mailbox.reply') : $t('mailbox.newMessage') }}</h2>
					<p>{{ $t('mailbox.composeHint') }}</p>
				</div>
				<button class="mailbox-icon-button" type="button" @click="cancelCompose">×</button>
			</div>

			<div class="mailbox-compose__grid">
				<label class="mailbox-field mailbox-field--recipient">
					<span>{{ $t('mailbox.recipient') }}</span>
					<div v-if="selectedRecipient" class="mailbox-selected-user">
						<div>
							<strong>{{ displayUser(selectedRecipient) }}</strong>
							<small>@{{ selectedRecipient.username }}</small>
						</div>
						<button type="button" @click="clearRecipient">{{ $t('mailbox.change') }}</button>
					</div>
					<div v-else class="mailbox-recipient-search">
						<input
							v-model="recipientQuery"
							class="input"
							type="search"
							:placeholder="$t('mailbox.recipientSearch')"
							@input="scheduleRecipientSearch"
							@focus="searchRecipients"
						>
						<div v-if="recipientResults.length" class="mailbox-recipient-results">
							<button
								v-for="person in recipientResults"
								:key="person.id"
								type="button"
								@click="selectRecipient(person)"
							>
								<strong>{{ displayUser(person) }}</strong>
								<span>@{{ person.username }}</span>
							</button>
						</div>
					</div>
				</label>

				<label class="mailbox-field">
					<span>{{ $t('mailbox.subject') }}</span>
					<input v-model="draft.subject" class="input" maxlength="500" type="text">
				</label>

				<label class="mailbox-field mailbox-field--full">
					<span>{{ $t('mailbox.message') }}</span>
					<textarea v-model="draft.body" class="textarea" maxlength="50000" rows="8" />
				</label>
			</div>

			<div class="mailbox-compose__actions">
				<XButton variant="secondary" @click="cancelCompose">{{ $t('misc.cancel') }}</XButton>
				<XButton :disabled="!canSend" :loading="sending" @click="sendMessage">{{ $t('mailbox.send') }}</XButton>
			</div>
		</section>

		<section class="mailbox-shell">
			<div class="mailbox-list-panel">
				<div class="mailbox-toolbar">
					<div class="mailbox-tabs">
						<button :class="{'is-active': folder === 'inbox'}" type="button" @click="setFolder('inbox')">
							{{ $t('mailbox.inbox') }}
							<span v-if="unreadCount > 0" class="mailbox-count">{{ unreadCount }}</span>
						</button>
						<button :class="{'is-active': folder === 'sent'}" type="button" @click="setFolder('sent')">{{ $t('mailbox.sent') }}</button>
					</div>
					<input
						v-model="search"
						class="input mailbox-search"
						type="search"
						:placeholder="$t('mailbox.search')"
						@input="scheduleSearch"
					>
				</div>

				<div v-if="loadingList" class="mailbox-loading">{{ $t('misc.loading') }}</div>
				<div v-else-if="messages.length === 0" class="mailbox-empty">
					<strong>{{ folder === 'inbox' ? $t('mailbox.emptyInbox') : $t('mailbox.emptySent') }}</strong>
					<span>{{ $t('mailbox.emptyHint') }}</span>
				</div>
				<div v-else class="mailbox-list">
					<button
						v-for="message in messages"
						:key="message.id"
						type="button"
						:class="['mailbox-row', {'is-selected': selected?.id === message.id, 'is-unread': folder === 'inbox' && !isMessageRead(message)}]"
						@click="openMessage(message)"
					>
						<div class="mailbox-row__top">
							<strong>{{ folder === 'inbox' ? displayUser(message.sender) : displayUser(message.recipient) }}</strong>
							<time>{{ formatCompactDate(message.created) }}</time>
						</div>
						<div class="mailbox-row__subject">{{ message.subject }}</div>
						<div class="mailbox-row__preview">{{ preview(message.body) }}</div>
					</button>
				</div>

				<div v-if="totalPages > 1" class="mailbox-pager">
					<button :disabled="page <= 1" type="button" @click="changePage(page - 1)">‹</button>
					<span>{{ page }} / {{ totalPages }}</span>
					<button :disabled="page >= totalPages" type="button" @click="changePage(page + 1)">›</button>
				</div>
			</div>

			<div class="mailbox-detail-panel">
				<div v-if="selected" class="mailbox-detail">
					<header>
						<div>
							<span class="mailbox-detail__party">
								{{ folder === 'inbox' ? $t('mailbox.from') : $t('mailbox.to') }}:
								<strong>{{ folder === 'inbox' ? displayUser(selected.sender) : displayUser(selected.recipient) }}</strong>
							</span>
							<h2>{{ selected.subject }}</h2>
							<time>{{ formatDate(selected.created) }}</time>
						</div>
						<div class="mailbox-detail__actions">
							<XButton v-if="folder === 'inbox'" variant="secondary" @click="replyToSelected">{{ $t('mailbox.reply') }}</XButton>
							<XButton v-if="folder === 'inbox'" variant="secondary" @click="toggleReadSelected">
								{{ isMessageRead(selected) ? $t('mailbox.markUnread') : $t('mailbox.markRead') }}
							</XButton>
							<XButton variant="secondary" danger @click="deleteSelected">{{ $t('mailbox.delete') }}</XButton>
						</div>
					</header>
					<div class="mailbox-detail__body">{{ selected.body }}</div>
				</div>
				<div v-else class="mailbox-detail-empty">
					<Icon icon="envelope" />
					<strong>{{ $t('mailbox.selectMessage') }}</strong>
					<span>{{ $t('mailbox.selectMessageHint') }}</span>
				</div>
			</div>
		</section>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, onMounted, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import XButton from '@/components/input/Button.vue'
import Icon from '@/components/misc/Icon'
import MailboxService from '@/services/mailbox'
import {error, success} from '@/message'
import type {IMailboxDraft, IMailboxMessage, IMailboxUser} from '@/modelTypes/IMailbox'

const {t, locale} = useI18n()
const service = new MailboxService()

const folder = ref<'inbox' | 'sent'>('inbox')
const messages = ref<IMailboxMessage[]>([])
const selected = ref<IMailboxMessage | null>(null)
const loadingList = ref(false)
const page = ref(1)
const totalPages = ref(1)
const search = ref('')
const unreadCount = ref(0)
const composing = ref(false)
const sending = ref(false)
const selectedRecipient = ref<IMailboxUser | null>(null)
const recipientQuery = ref('')
const recipientResults = ref<IMailboxUser[]>([])
let searchTimer: ReturnType<typeof setTimeout> | undefined
let recipientTimer: ReturnType<typeof setTimeout> | undefined

const draft = reactive<IMailboxDraft>({recipient_id: 0, reply_to_id: 0, subject: '', body: ''})
const canSend = computed(() => draft.recipient_id > 0 && draft.body.trim().length > 0)

function displayUser(user: IMailboxUser | null | undefined) {
	if (!user) return '—'
	return user.name?.trim() || user.username
}

function preview(value: string) {
	const clean = value.replace(/\s+/g, ' ').trim()
	return clean.length > 110 ? `${clean.slice(0, 110)}…` : clean
}

function isMessageRead(message: IMailboxMessage) {
	if (!message.read_at) return false
	const date = new Date(message.read_at)
	return Number.isFinite(date.getTime()) && date.getUTCFullYear() > 1
}

function formatDate(value: string | Date) {
	return new Intl.DateTimeFormat(locale.value, {dateStyle: 'medium', timeStyle: 'short'}).format(new Date(value))
}

function formatCompactDate(value: string | Date) {
	const date = new Date(value)
	const today = new Date()
	if (date.toDateString() === today.toDateString()) {
		return new Intl.DateTimeFormat(locale.value, {hour: '2-digit', minute: '2-digit'}).format(date)
	}
	return new Intl.DateTimeFormat(locale.value, {day: '2-digit', month: '2-digit'}).format(date)
}

async function loadMessages() {
	loadingList.value = true
	try {
		const result = await service.list(folder.value, page.value, search.value.trim())
		messages.value = result.items
		totalPages.value = Math.max(result.total_pages, 1)
		if (selected.value && !messages.value.some(message => message.id === selected.value?.id)) selected.value = null
	} catch (e) {
		error(e)
	} finally {
		loadingList.value = false
	}
}

async function loadUnreadCount() {
	try {
		unreadCount.value = await service.unreadCount()
	} catch {
		unreadCount.value = 0
	}
}

function setFolder(value: 'inbox' | 'sent') {
	if (folder.value === value) return
	folder.value = value
	page.value = 1
	selected.value = null
	void loadMessages()
}

function changePage(value: number) {
	page.value = value
	selected.value = null
	void loadMessages()
}

function scheduleSearch() {
	if (searchTimer) clearTimeout(searchTimer)
	searchTimer = setTimeout(() => {
		page.value = 1
		void loadMessages()
	}, 350)
}

async function openMessage(message: IMailboxMessage) {
	try {
		selected.value = await service.get(message.id)
		if (folder.value === 'inbox' && !isMessageRead(selected.value)) {
			selected.value = await service.setRead(selected.value.id, true)
			const row = messages.value.find(item => item.id === selected.value?.id)
			if (row) row.read_at = selected.value.read_at
			await loadUnreadCount()
		}
	} catch (e) {
		error(e)
	}
}

function resetDraft() {
	draft.recipient_id = 0
	draft.reply_to_id = 0
	draft.subject = ''
	draft.body = ''
	selectedRecipient.value = null
	recipientQuery.value = ''
	recipientResults.value = []
}

function startCompose(recipient?: IMailboxUser, subject = '', replyToID = 0) {
	resetDraft()
	composing.value = true
	if (recipient) selectRecipient(recipient)
	draft.subject = subject
	draft.reply_to_id = replyToID
	if (!recipient) void searchRecipients()
}

function cancelCompose() {
	composing.value = false
	resetDraft()
}

function selectRecipient(user: IMailboxUser) {
	selectedRecipient.value = user
	draft.recipient_id = user.id
	recipientResults.value = []
	recipientQuery.value = ''
}

function clearRecipient() {
	selectedRecipient.value = null
	draft.recipient_id = 0
	void searchRecipients()
}

function scheduleRecipientSearch() {
	if (recipientTimer) clearTimeout(recipientTimer)
	recipientTimer = setTimeout(() => void searchRecipients(), 250)
}

async function searchRecipients() {
	try {
		recipientResults.value = await service.recipients(recipientQuery.value.trim())
	} catch (e) {
		error(e)
	}
}

async function sendMessage() {
	if (!canSend.value) return
	sending.value = true
	try {
		await service.send({...draft})
		success({message: t('mailbox.sentSuccess')})
		cancelCompose()
		if (folder.value === 'sent') await loadMessages()
	} catch (e) {
		error(e)
	} finally {
		sending.value = false
	}
}

function replyToSelected() {
	if (!selected.value) return
	const subject = /^re:/i.test(selected.value.subject) ? selected.value.subject : `Re: ${selected.value.subject}`
	startCompose(selected.value.sender, subject, selected.value.id)
}

async function toggleReadSelected() {
	if (!selected.value || folder.value !== 'inbox') return
	try {
		selected.value = await service.setRead(selected.value.id, !isMessageRead(selected.value))
		const row = messages.value.find(item => item.id === selected.value?.id)
		if (row) row.read_at = selected.value.read_at
		await loadUnreadCount()
	} catch (e) {
		error(e)
	}
}

async function deleteSelected() {
	if (!selected.value || !window.confirm(t('mailbox.deleteConfirm'))) return
	try {
		await service.delete(selected.value.id)
		selected.value = null
		await Promise.all([loadMessages(), loadUnreadCount()])
		success({message: t('mailbox.deletedSuccess')})
	} catch (e) {
		error(e)
	}
}

onMounted(() => {
	void Promise.all([loadMessages(), loadUnreadCount()])
})

onBeforeUnmount(() => {
	if (searchTimer) clearTimeout(searchTimer)
	if (recipientTimer) clearTimeout(recipientTimer)
})
</script>

<style scoped lang="scss">
.mailbox-page { max-inline-size: 1500px; margin: 0 auto; padding: 1rem 1.25rem 3rem; }
.mailbox-header { display:flex; align-items:flex-start; justify-content:space-between; gap:1rem; margin-block-end:1rem; padding:1.3rem 1.4rem; border-radius:18px; background:linear-gradient(115deg,var(--brand-forest,#153f34),#285f4f); color:#fff; }
.mailbox-header h1 { margin:.1rem 0 .2rem; color:#fff; font-size:1.65rem; }
.mailbox-header p { margin:0; color:rgba(255,255,255,.72); }
.mailbox-eyebrow { color:var(--brand-accent,#d8ff80); font-size:.7rem; font-weight:800; text-transform:uppercase; letter-spacing:.11em; }
.mailbox-compose,.mailbox-shell { border:1px solid var(--brand-border,#dfe8df); border-radius:16px; background:#fff; box-shadow:0 8px 26px rgba(31,91,73,.06); }
.mailbox-compose { margin-block-end:1rem; padding:1.1rem; }
.mailbox-compose__heading,.mailbox-detail header { display:flex; align-items:flex-start; justify-content:space-between; gap:1rem; }
.mailbox-compose__heading h2,.mailbox-detail h2 { margin:0; color:var(--brand-forest,#153f34); }
.mailbox-compose__heading p { margin:.15rem 0 0; color:#697a74; font-size:.8rem; }
.mailbox-icon-button { border:0; background:transparent; font-size:1.5rem; cursor:pointer; color:#60716b; }
.mailbox-compose__grid { display:grid; grid-template-columns:1fr 1fr; gap:.85rem 1rem; margin-block-start:1rem; }
.mailbox-field { display:flex; flex-direction:column; gap:.35rem; }
.mailbox-field > span { color:#61726c; font-size:.74rem; font-weight:700; }
.mailbox-field--full { grid-column:1/-1; }
.mailbox-selected-user { display:flex; align-items:center; justify-content:space-between; min-block-size:42px; padding:.45rem .65rem; border:1px solid var(--brand-border,#dfe8df); border-radius:9px; background:#fbfdfb; }
.mailbox-selected-user div { display:flex; flex-direction:column; }
.mailbox-selected-user small { color:#788881; }
.mailbox-selected-user button { border:0; background:transparent; color:var(--brand-green,#1f5b49); cursor:pointer; }
.mailbox-recipient-search { position:relative; }
.mailbox-recipient-results { position:absolute; z-index:20; inset:calc(100% + .25rem) 0 auto; max-block-size:240px; overflow:auto; border:1px solid var(--brand-border,#dfe8df); border-radius:10px; background:#fff; box-shadow:0 12px 28px rgba(21,63,52,.14); }
.mailbox-recipient-results button { display:flex; inline-size:100%; align-items:center; justify-content:space-between; gap:1rem; padding:.65rem .75rem; border:0; border-block-end:1px solid #edf2ed; background:#fff; cursor:pointer; text-align:start; }
.mailbox-recipient-results button:hover { background:#f3f8f4; }
.mailbox-recipient-results span { color:#75867f; font-size:.75rem; }
.mailbox-compose__actions { display:flex; justify-content:flex-end; gap:.55rem; margin-block-start:1rem; }
.mailbox-shell { display:grid; grid-template-columns:minmax(320px,38%) minmax(0,1fr); min-block-size:620px; overflow:hidden; }
.mailbox-list-panel { border-inline-end:1px solid var(--brand-border,#dfe8df); background:#fbfdfb; }
.mailbox-toolbar { padding:.85rem; border-block-end:1px solid var(--brand-border,#dfe8df); }
.mailbox-tabs { display:flex; gap:.35rem; margin-block-end:.65rem; }
.mailbox-tabs button { border:1px solid transparent; border-radius:9px; padding:.48rem .7rem; background:transparent; cursor:pointer; color:#52645e; font-weight:750; }
.mailbox-tabs button.is-active { border-color:#d5e3d8; background:#eaf2eb; color:var(--brand-forest,#153f34); }
.mailbox-count { display:inline-grid; place-items:center; min-inline-size:19px; min-block-size:19px; margin-inline-start:.3rem; padding:0 .3rem; border-radius:999px; background:var(--brand-forest,#153f34); color:#fff; font-size:.67rem; }
.mailbox-search { inline-size:100%; }
.mailbox-list { display:flex; flex-direction:column; }
.mailbox-row { inline-size:100%; padding:.85rem .9rem; border:0; border-block-end:1px solid #edf2ed; background:#fff; cursor:pointer; text-align:start; }
.mailbox-row:hover,.mailbox-row.is-selected { background:#f0f6f1; }
.mailbox-row.is-unread { border-inline-start:4px solid var(--brand-green,#1f5b49); background:#f8fcf8; }
.mailbox-row.is-unread strong,.mailbox-row.is-unread .mailbox-row__subject { font-weight:850; }
.mailbox-row__top { display:flex; justify-content:space-between; gap:.75rem; color:#263a33; font-size:.8rem; }
.mailbox-row__top time { color:#83918c; font-size:.7rem; white-space:nowrap; }
.mailbox-row__subject { margin:.28rem 0 .12rem; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; color:#354a43; font-weight:650; }
.mailbox-row__preview { display:-webkit-box; overflow:hidden; color:#75867f; font-size:.75rem; -webkit-box-orient:vertical; -webkit-line-clamp:2; }
.mailbox-loading,.mailbox-empty { padding:2rem 1rem; text-align:center; color:#75867f; }
.mailbox-empty { display:flex; flex-direction:column; gap:.2rem; }
.mailbox-empty strong { color:#40564e; }
.mailbox-pager { display:flex; align-items:center; justify-content:center; gap:.7rem; padding:.7rem; }
.mailbox-pager button { inline-size:32px; block-size:32px; border:1px solid var(--brand-border,#dfe8df); border-radius:8px; background:#fff; cursor:pointer; }
.mailbox-pager button:disabled { opacity:.4; cursor:not-allowed; }
.mailbox-detail-panel { min-inline-size:0; background:#fff; }
.mailbox-detail { padding:1.3rem; }
.mailbox-detail header { padding-block-end:1rem; border-block-end:1px solid #edf2ed; }
.mailbox-detail__party { color:#61726c; font-size:.78rem; }
.mailbox-detail__party strong { margin-inline-start:.3rem; color:#284039; }
.mailbox-detail h2 { margin:.35rem 0 .2rem; font-size:1.35rem; }
.mailbox-detail time { color:#84928d; font-size:.75rem; }
.mailbox-detail__actions { display:flex; flex-wrap:wrap; gap:.45rem; justify-content:flex-end; }
.mailbox-detail__body { padding-block:1.25rem; white-space:pre-wrap; overflow-wrap:anywhere; color:#2c4039; line-height:1.65; }
.mailbox-detail-empty { display:flex; min-block-size:620px; align-items:center; justify-content:center; flex-direction:column; gap:.35rem; color:#8a9893; }
.mailbox-detail-empty :deep(svg) { font-size:2.5rem; color:#b8c8c0; }
.mailbox-detail-empty strong { color:#53665f; }
.mailbox-detail-empty span { font-size:.78rem; }
@media (max-width:900px) { .mailbox-shell{grid-template-columns:1fr;} .mailbox-list-panel{border-inline-end:0;border-block-end:1px solid var(--brand-border,#dfe8df);} .mailbox-detail-empty{min-block-size:220px;} }
@media (max-width:650px) { .mailbox-page{padding:.65rem;} .mailbox-header,.mailbox-compose__heading,.mailbox-detail header{flex-direction:column;} .mailbox-compose__grid{grid-template-columns:1fr;} .mailbox-field--full{grid-column:auto;} .mailbox-detail__actions{justify-content:flex-start;} }
</style>

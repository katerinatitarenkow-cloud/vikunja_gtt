<template>
	<div class="notifications">
		<slot
			name="trigger"
			toggle-open="() => showNotifications = !showNotifications"
			:has-unread-notifications="unreadNotifications > 0"
		>
			<BaseButton
				class="trigger-button"
				:aria-expanded="showNotifications"
				@click.stop="showNotifications = !showNotifications"
			>
				<span class="is-sr-only">{{ $t('notification.title') }}</span>
				<span
v-if="unreadNotifications > 0"
class="unread-indicator"
>
{{
unreadNotifications > 99
? '99+'
: unreadNotifications
}}
</span>
				<Icon icon="bell" />
			</BaseButton>
		</slot>

		<CustomTransition name="fade">
			<div
				v-if="showNotifications"
				ref="popup"
				class="notifications-list"
			>
				<div class="head">
					<span>{{ $t('notification.title') }}</span>
					<BaseButton
						v-tooltip="$t('notification.subscribeFeed')"
						class="feed-link"
						:to="{name: 'user.settings.feeds'}"
						@click="showNotifications = false"
					>
						<span class="is-sr-only">{{ $t('notification.subscribeFeed') }}</span>
						<Icon icon="rss" />
					</BaseButton>
				</div>
				<div
					v-for="(n, index) in notifications"
					:key="n.id"
					class="single-notification"
					:class="{'is-clickable': notificationHasRoute(n)}"
					@click="() => notificationHasRoute(n) && to(n, index)()"
				>
					<div
						class="read-indicator"
						:class="{'read': n.readAt !== null}"
					/>
					<User
						v-if="n.notification.doer"
						:user="n.notification.doer"
						:show-username="false"
						:avatar-size="16"
					/>
					<div class="detail">
						<div>
							<span
								v-if="n.notification.doer"
								class="has-text-weight-bold mie-1"
							>
								{{ getDisplayName(n.notification.doer) }}
							</span>
							{{ n.toText(userInfo) }}
						</div>
						<span
							v-tooltip="formatDateLong(n.created)"
							class="created"
						>
							{{ formatDisplayDate(n.created) }}
						</span>
					</div>
				</div>
				<XButton
					v-if="notifications.length > 0 && unreadNotifications > 0"
					variant="tertiary"
					class="mbs-2 is-fullwidth" 
					@click="markAllRead"
				>
					{{ $t('notification.markAllRead') }}
				</XButton>
				<p
					v-if="notifications.length === 0"
					class="nothing"
				>
					{{ $t('notification.none') }}<br>
					<span class="explainer">
						{{ $t('notification.explainer') }}
					</span>
				</p>
			</div>
		</CustomTransition>
	</div>
</template>

<script lang="ts" setup>
import {computed, onMounted, onUnmounted, ref, watch} from 'vue'
import {useRouter, isNavigationFailure, NavigationFailureType, type RouteLocationRaw} from 'vue-router'

import NotificationService from '@/services/notification'
import NotificationModel from '@/models/notification'
import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import User from '@/components/misc/User.vue'
import {NOTIFICATION_NAMES as names, type INotification} from '@/modelTypes/INotification'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/models/user'
import {useAuthStore} from '@/stores/auth'
import {useMailboxStore} from '@/stores/mailbox'
import {useWebSocket} from '@/composables/useWebSocket'
import XButton from '@/components/input/Button.vue'
import {success} from '@/message'
import {useI18n} from 'vue-i18n'

const {subscribe, connected: wsConnected} = useWebSocket()

const authStore = useAuthStore()
const mailboxStore = useMailboxStore()
const router = useRouter()
const {t} = useI18n()

const allNotifications = ref<NotificationModel[]>([])
const showNotifications = ref(false)
const popup = ref<HTMLElement | null>(null)

const unreadNotifications = computed(() => {
	return notifications.value.filter(n => n.readAt === null).length
})
const notifications = computed(() => {
	return allNotifications.value ? allNotifications.value.filter(n => n.name !== '') : []
})
const userInfo = computed(() => authStore.info)

let unsubscribeWs: (() => void) | null = null
let pollInterval: ReturnType<typeof setInterval> | null = null

const POLL_INTERVAL = 10000

onMounted(async () => {
	// Initial load via REST - wrapped in try/catch so the rest of setup
	// (click handler, WS subscription, polling) still runs if this fails
	try {
		await loadNotifications()
	} catch (e) {
		console.warn('Failed to load initial notifications:', e)
	}

	document.addEventListener('click', hidePopup)

	// Subscribe to real-time notifications
	unsubscribeWs = subscribe('notification.created', (msg) => {
		if (msg.event === 'notification.created' && msg.data) {
			const notification = new NotificationModel(msg.data as Partial<INotification>)
			// Avoid duplicates if the same notification was already loaded via REST
			const exists = allNotifications.value.some(n => n.id === notification.id)
			if (!exists) {
				allNotifications.value = [notification, ...allNotifications.value]
			}
		}
	})

	// Fallback polling when WebSocket is not available
	startPollingFallback()
})

// Reload notifications when WebSocket disconnects to catch any events
// that may have been missed during the disconnect window
watch(wsConnected, (isConnected, wasConnected) => {
	if (wasConnected && !isConnected) {
		loadNotifications().catch(e => console.warn('Failed to reload notifications after WS disconnect:', e))
	}
})

onUnmounted(() => {
	document.removeEventListener('click', hidePopup)
	unsubscribeWs?.()
	stopPollingFallback()
})

function startPollingFallback() {
	pollInterval = setInterval(async () => {
		if (!wsConnected.value && document.visibilityState === 'visible') {
			await loadNotifications()
		}
	}, POLL_INTERVAL)
}

function stopPollingFallback() {
	if (pollInterval) {
		clearInterval(pollInterval)
		pollInterval = null
	}
}

async function loadNotifications() {
	const notificationService = new NotificationService()
	allNotifications.value = await notificationService.getAll() as NotificationModel[]

// [MAILBOX] refresh unread after notifications load
await mailboxStore.refreshUnread()
}

function hidePopup(e: MouseEvent) {
	if (showNotifications.value && popup.value !== null) {
		closeWhenClickedOutside(e, popup.value, () => showNotifications.value = false)
	}
}

function getNotificationRoute(n: INotification): RouteLocationRaw | null {
const payload = n.notification as unknown as {
task?: {
id?: number
}
project?: {
id?: number
}
team?: {
id?: number
}
message_id?: number
messageId?: number
}

// Письмо -> конкретное письмо в Почте
if (n.name === names.MAILBOX_MESSAGE_RECEIVED) {
const messageID = Number(
payload.message_id ??
payload.messageId ??
0,
)

if (messageID > 0) {
return {
name: 'mailbox',
query: {
message: String(messageID),
},
}
}

return {
name: 'mailbox',
}
}

// Типы событий, у которых правильная страница
// определяется не просто наличием task/project.
switch (n.name) {
case names.PROJECT_CREATED:
if (payload.project?.id) {
return {
name: 'task.index',
params: {
projectId: payload.project.id,
},
}
}
break

case names.CLIENT_RESPONSIBLE_ASSIGNED:
if (payload.project?.id) {
return {
name: 'project.client',
params: {
projectId: payload.project.id,
},
}
}
break

case names.TEAM_MEMBER_ADDED:
if (payload.team?.id) {
return {
name: 'teams.edit',
params: {
id: payload.team.id,
},
}
}
break

// Удалённую задачу открыть уже нельзя.
case names.TASK_DELETED:
return null
}

// Любое событие конкретной задачи:
// назначение, комментарий, упоминание,
// напоминание о сроке и будущие task-уведомления.
if (payload.task?.id) {
return {
name: 'task.detail',
params: {
id: payload.task.id,
},
}
}

// Запасной маршрут для событий проекта.
if (payload.project?.id) {
return {
name: 'task.index',
params: {
projectId: payload.project.id,
},
}
}

// Запасной маршрут для событий группы.
if (payload.team?.id) {
return {
name: 'teams.edit',
params: {
id: payload.team.id,
},
}
}

return null
}

function notificationHasRoute(n: INotification): boolean {
	return getNotificationRoute(n) !== null
}

function to(n: INotification, index: number) {
	return async () => {
		const route = getNotificationRoute(n)
		if (route === null) return
		
		const failure = await router.push(route)
		if (isNavigationFailure(failure, NavigationFailureType.duplicated)) {
			router.go(0)
		}

		n.read = true
		if (allNotifications.value[index]) {
			const notificationService = new NotificationService()
			Object.assign(allNotifications.value[index], await notificationService.update(n))
		}

		showNotifications.value = false
	}
}

async function markAllRead() {
	const notificationService = new NotificationService()
	await notificationService.markAllRead()
	success({message: t('notification.markAllReadSuccess')})
	
	notifications.value.forEach(n => n.readAt = new Date())
}
</script>

<style lang="scss" scoped>
.notifications {
	display: flex;

	.trigger-button {
		inline-size: 100%;
		position: relative;
	}

	.unread-indicator {
position: absolute;
inset-block-start: .25rem;
inset-inline-end: -.15rem;
display: inline-flex;
align-items: center;
justify-content: center;
min-inline-size: 1.35rem;
block-size: 1.35rem;
padding-inline: .3rem;
background: #ef4444;
color: #fff;
border-radius: 999px;
border: 2px solid var(--white);
font-size: .62rem;
font-weight: 800;
line-height: 1;
z-index: 2;
}

	.notifications-list {
		position: absolute;
		inset-inline-end: 1rem;
		inset-block-start: calc(100% + 1rem);
		max-block-size: 400px;
		overflow-y: auto;

		background: var(--white);
		inline-size: 350px;
		max-inline-size: calc(100vw - 2rem);
		padding: .75rem .25rem;
		border-radius: $radius;
		box-shadow: var(--shadow-sm);
		font-size: .85rem;

		@media screen and (max-width: $tablet) {
			max-block-size: calc(100vh - 1rem - #{$navbar-height});
		}

		.head {
			font-family: $vikunja-font;
			font-size: 1rem;
			padding: .5rem;
			display: flex;
			align-items: center;
			justify-content: space-between;

			.feed-link {
				color: var(--grey-500);
				transition: color $transition;

				&:hover,
				&:focus {
					color: var(--primary);
				}
			}
		}

		.single-notification {
			display: flex;
			align-items: center;
			padding: 0.25rem 0;

			transition: background-color $transition;

			&.is-clickable {
				cursor: pointer;
			}

			&:hover {
				background: var(--grey-100);
				border-radius: $radius;
			}

			.read-indicator {
				inline-size: .35rem;
				block-size: .35rem;
				background: var(--primary);
				border-radius: 100%;
				margin: 0 .5rem;
				flex-shrink: 0;

				&.read {
					background: transparent;
				}
			}

			.user {
				display: inline-flex;
				align-items: center;
				inline-size: auto;
				margin: 0 .5rem;

				span {
					font-family: $family-sans-serif;
				}

				.avatar {
					block-size: 16px;
				}

				img {
					margin-inline-end: 0;
				}
			}

			.created {
				color: var(--grey-400);
			}

			&:last-child {
				margin-block-end: .25rem;
			}

			a {
				color: var(--grey-800);
			}
		}

		.nothing {
			text-align: center;
			padding: 1rem 0;
			color: var(--grey-500);

			.explainer {
				font-size: .75rem;
			}
		}
	}
}
</style>

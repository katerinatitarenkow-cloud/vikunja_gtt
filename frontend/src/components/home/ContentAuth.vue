<template>
	<div class="content-auth">
		<BaseButton
			v-show="menuActive"
			:aria-label="$t('navigation.closeSidebar')"
			class="menu-hide-button d-print-none"
			@click="baseStore.setMenuActive(false)"
		>
			<Icon icon="times" />
		</BaseButton>
		<div
			class="app-container"
			:class="{'has-background': background || blurHash}"
			:style="{'background-image': blurHash && `url(${blurHash})`}"
		>
			<div
				:class="{'is-visible': background}"
				class="app-container-background background-fade-in d-print-none"
				:style="{
					'background-image': background && `url(${background})`,
					'filter': backgroundBrightness && `brightness(${backgroundBrightness}%)`
				}"
			/>
			<Navigation class="d-print-none" />
			<main
				id="main-content"
				tabindex="-1"
				class="app-content"
				:class="[
					{ 'is-menu-enabled': menuActive },
					$route.name,
				]"
				:style="{'--sidebar-width': sidebarWidth}"
			>
				<BaseButton
					v-show="menuActive"
					:aria-label="$t('navigation.closeSidebar')"
					class="mobile-overlay d-print-none"
					@click="baseStore.setMenuActive(false)"
				/>

				<QuickActions />

				<RouterView
					v-slot="{ Component }"
					:route="routeWithModal"
				>
					<keep-alive :include="['project.view']">
						<component :is="Component" />
					</keep-alive>
				</RouterView>

				<Modal
					:enabled="typeof currentModal !== 'undefined'"
					variant="scrolling"
					class="task-detail-view-modal"
					:aria-label="$t('task.detail.title')"
					@close="closeModal()"
				>
					<component
						:is="currentModal"
						@close="closeModal()"
					/>
				</Modal>

				<BaseButton
					v-shortcut="'Shift+Slash'"
					class="keyboard-shortcuts-button d-print-none"
					@click="showKeyboardShortcuts()"
				>
					<span class="is-sr-only">{{ $t('keyboardShortcuts.title') }}</span>
					<Icon icon="keyboard" />
				</BaseButton>
			</main>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {watch, computed, onBeforeUnmount} from 'vue'
import {useRoute, useRouter} from 'vue-router'

import Navigation from '@/components/home/Navigation.vue'
import QuickActions from '@/components/quick-actions/QuickActions.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import {useBaseStore} from '@/stores/base'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'

import {useRouteWithModal} from '@/composables/useRouteWithModal'
import {useRenewTokenOnFocus} from '@/composables/useRenewTokenOnFocus'
import {useSidebarResize} from '@/composables/useSidebarResize'
import {useWebSocket} from '@/composables/useWebSocket'
import {useAuthStore} from '@/stores/auth'
import {useAccessStore} from '@/stores/access'
import {ACCESS_PERMISSION} from '@/modelTypes/IAccessControl'

const authStore = useAuthStore()
const backgroundBrightness = computed(() =>
	authStore.settings?.frontendSettings?.backgroundBrightness,
)

const {sidebarWidth} = useSidebarResize()

const {routeWithModal, currentModal, closeModal} = useRouteWithModal()

const baseStore = useBaseStore()
const background = computed(() => baseStore.background)
const blurHash = computed(() => baseStore.blurHash)
const menuActive = computed(() => baseStore.menuActive)

function showKeyboardShortcuts() {
	baseStore.setKeyboardShortcutsActive(true)
}

const route = useRoute()
const router = useRouter()

// FIXME: this is really error prone
// Reset the current project highlight in menu if the current route is not project related.
watch(() => route.name as string, (routeName) => {
	if (
		routeName &&
		(
			[
				'home',
				'teams.index',
				'teams.edit',
				'tasks.range',
				'labels.index',
				'migrate.start',
				'migrate.wunderlist',
				'projects.index',
			].includes(routeName) ||
			routeName.startsWith('user.settings')
		)
	) {
		baseStore.handleSetCurrentProject({project: null})
	}
})

// TODO: Reset the title if the page component does not set one itself

useRenewTokenOnFocus()

const {connect} = useWebSocket()
connect()

const accessStore = useAccessStore()
const labelStore = useLabelStore()
if (accessStore.can(ACCESS_PERMISSION.LABELS_VIEW)) {
	labelStore.loadAllLabels()
}

const projectStore = useProjectStore()
if (accessStore.can(ACCESS_PERMISSION.PROJECTS_VIEW)) {
	projectStore.loadAllProjects()
}

// Listen for task creation from the quick-entry window
const taskUpdateChannel = new BroadcastChannel('vikunja-task-updates')
taskUpdateChannel.onmessage = (event) => {
	if (event.data?.type === 'task-created-open' && event.data?.taskId) {
		router.push({name: 'task.detail', params: {id: event.data.taskId}})
	}
}

onBeforeUnmount(() => {
	taskUpdateChannel.close()
})
</script>

<style lang="scss" scoped>
.menu-hide-button {
	position: fixed;
	inset-block-start: .65rem;
	inset-inline-end: .65rem;
	z-index: 31;
	inline-size: 2.75rem;
	block-size: 2.75rem;
	display: flex;
	justify-content: center;
	align-items: center;
	font-size: 1.35rem;
	color: var(--brand-text-muted);
	line-height: 1;
	border-radius: 12px;
	background: var(--white);
	box-shadow: 0 8px 24px rgba(13, 48, 40, .12);
	transition: all $transition;

	@media screen and (min-width: $tablet) { display: none; }
	&:hover, &:focus { color: var(--brand-forest); }
}

.app-container {
	min-block-size: calc(100vh - 65px);
	background: var(--site-background);

	@media screen and (max-width: $tablet) {
		padding-block-start: $navbar-height;
	}
}

.app-content {
	--sidebar-width: #{$navbar-width};
	display: flow-root;
	z-index: 10;
	position: relative;
	padding: 1rem .75rem 1.5rem;
	transition: margin-inline-start $transition-duration;

	@media screen and (max-width: $tablet) {
		margin-inline-start: 0;
		margin-inline-end: 0;
		min-block-size: calc(100vh - 4rem);
	}

	@media screen and (min-width: $tablet) {
		padding: $navbar-height + 1.25rem 1.5rem 2rem;
	}

	&.is-menu-enabled {
		@media screen and (min-width: $tablet) {
			margin-inline-start: var(--sidebar-width);
		}
	}

	> .loader-container {
		min-block-size: calc(100vh - #{$navbar-height + 1.5rem + 1rem});
	}

	.card { background: var(--white); }
}

.mobile-overlay {
	display: none;
	position: fixed;
	inset: 0;
	block-size: 100vh;
	inline-size: 100vw;
	background: rgba(13, 48, 40, .45);
	backdrop-filter: blur(2px);
	z-index: 5;
	opacity: 0;
	transition: all $transition;

	@media screen and (max-width: $tablet) {
		display: block;
		opacity: 1;
	}
}

.keyboard-shortcuts-button {
	position: fixed;
	inset-block-end: 1rem;
	inset-inline-end: 1rem;
	z-index: 4500;
	color: var(--brand-text-muted);
	background: var(--white);
	border: 1px solid var(--brand-border);
	border-radius: 10px;
	padding: .45rem .55rem;
	box-shadow: 0 5px 18px rgba(31, 91, 73, .08);
	transition: all $transition;

	&:hover { color: var(--brand-forest); transform: translateY(-1px); }
	@media screen and (max-width: $tablet) { display: none; }
}

.content-auth { position: relative; z-index: 1; }
</style>

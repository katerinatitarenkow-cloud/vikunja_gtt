<template>
	<div
		class="loader-container"
		:class="{
			'is-loading': isLoadingProject,
			'is-archived': currentProject?.isArchived,
		}"
	>
		<h1 class="project-title-print">
			{{ getProjectTitle(currentProject) }}
		</h1>

		<div
			ref="switchViewContainerRef"
			class="switch-view-container d-print-none"
			:class="{'is-justify-content-flex-end': totalTabs === 1}"
		>
			<!-- Dropdown mode when buttons overflow -->
			<Dropdown
				v-if="isOverflowing && totalTabs > 1"
				class="switch-view-dropdown"
			>
				<template #trigger="{ toggleOpen, open }">
					<BaseButton
						class="switch-view switch-view-dropdown-trigger"
						:aria-expanded="open"
						@click="toggleOpen"
					>
						{{ activeViewTitle }}
						<Icon
							icon="chevron-down"
							class="dropdown-icon"
						/>
					</BaseButton>
				</template>
				<template #default="{ close }">
					<div @click="close">
						<DropdownItem
							v-if="hasClientTab"
							:to="{name: 'project.client', params: {projectId: props.projectId}}"
							:class="{'is-active': isClientView}"
						>
							{{ $t('clientProfile.tab') }}
						</DropdownItem>
						<DropdownItem
							v-for="view in views"
							:key="view.id"
							:to="getViewRoute(view)"
							:class="{'is-active': view.id === viewId}"
						>
							{{ getViewTitle(view) }}
						</DropdownItem>
						<DropdownItem
							v-if="hasClientTab"
							:to="{name: 'project.history', params: {projectId: props.projectId}}"
							:class="{'is-active': isHistoryView}"
						>
							{{ $t('clientHistory.tab') }}
						</DropdownItem>
					</div>
				</template>
			</Dropdown>

			<!-- Inline buttons, hidden when overflowing but kept in DOM for width measurement -->
			<div
				v-if="totalTabs > 1"
				ref="switchViewRef"
				class="switch-view"
				:class="{'switch-view--hidden': isOverflowing || !overflowChecked}"
				:aria-hidden="isOverflowing || undefined"
			>
				<BaseButton
					v-if="hasClientTab"
					class="switch-view-button"
					:class="{'is-active': isClientView}"
					:to="{name: 'project.client', params: {projectId: props.projectId}}"
					:tabindex="isOverflowing ? -1 : undefined"
				>
					{{ $t('clientProfile.tab') }}
				</BaseButton>
				<BaseButton
					v-for="view in views"
					:key="view.id"
					class="switch-view-button"
					:class="{'is-active': view.id === viewId}"
					:to="getViewRoute(view)"
					:tabindex="isOverflowing ? -1 : undefined"
				>
					{{ getViewTitle(view) }}
				</BaseButton>
				<BaseButton
					v-if="hasClientTab"
					class="switch-view-button"
					:class="{'is-active': isHistoryView}"
					:to="{name: 'project.history', params: {projectId: props.projectId}}"
					:tabindex="isOverflowing ? -1 : undefined"
				>
					{{ $t('clientHistory.tab') }}
				</BaseButton>
			</div>
			<slot name="header" />
		</div>
		<CustomTransition name="fade">
			<Message
				v-if="currentProject?.isArchived"
				variant="warning"
				class="mbe-4"
			>
				{{ $t('project.archivedMessage') }}
			</Message>
		</CustomTransition>

		<slot v-if="!isLoadingProject" />
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, nextTick, onMounted} from 'vue'
import {useResizeObserver} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import {useRoute} from 'vue-router'

import BaseButton from '@/components/base/BaseButton.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import Icon from '@/components/misc/Icon'
import Message from '@/components/misc/Message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {getProjectTitle} from '@/helpers/getProjectTitle'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import {useViewFiltersStore} from '@/stores/viewFilters'

import type {IProject} from '@/modelTypes/IProject'
import type {IProjectView} from '@/modelTypes/IProjectView'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: IProject['id'],
	viewId: IProjectView['id'],
}>()

const {t} = useI18n()
const route = useRoute()

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const authStore = useAuthStore()
const viewFiltersStore = useViewFiltersStore()

const switchViewContainerRef = ref<HTMLElement>()
const switchViewRef = ref<HTMLElement>()
const isOverflowing = ref(false)
const overflowChecked = ref(false)

function checkOverflow() {
	if (!switchViewRef.value || !switchViewContainerRef.value) {
		return
	}
	const buttonsWidth = switchViewRef.value.scrollWidth
	const containerWidth = switchViewContainerRef.value.clientWidth
	isOverflowing.value = buttonsWidth > containerWidth
	overflowChecked.value = true
}

onMounted(() => {
	checkOverflow()
})

useResizeObserver(switchViewContainerRef, () => {
	requestAnimationFrame(() => checkOverflow())
})

const currentProject = computed<IProject>(() => {
	return baseStore.currentProject || {
		id: 0,
		title: '',
		isArchived: false,
		maxPermission: null,
	}
})
useTitle(() => currentProject.value?.id ? getProjectTitle(currentProject.value) : '')

const views = computed(() => projectStore.projects[props.projectId]?.views ?? [])
const hasClientTab = computed(() => props.projectId > 0 && !authStore.isLinkShareAuth)
const totalTabs = computed(() => views.value.length + (hasClientTab.value ? 2 : 0))
const isClientView = computed(() => hasClientTab.value && route.name === 'project.client')
const isHistoryView = computed(() => hasClientTab.value && route.name === 'project.history')

const activeViewTitle = computed(() => {
	if (isClientView.value) return t('clientProfile.tab')
	if (isHistoryView.value) return t('clientHistory.tab')
	const activeView = views.value?.find((v: IProjectView) => v.id === props.viewId)
	return activeView ? getViewTitle(activeView) : ''
})

// Re-check overflow when views change
watch(views, () => {
	nextTick(() => checkOverflow())
})

function getViewTitle(view: IProjectView) {
	switch (view.title) {
		case 'List':
			return t('project.list.title')
		case 'Gantt':
			return t('project.gantt.title')
		case 'Table':
			return t('project.table.title')
		case 'Kanban':
			return t('project.kanban.title')
	}

	return view.title
}

function getViewRoute(view: IProjectView) {
	const storedQuery = viewFiltersStore.getViewQuery(view.id)
	return {
		name: 'project.view',
		params: {projectId: props.projectId, viewId: view.id},
		query: storedQuery,
	}
}
</script>

<style lang="scss" scoped>
.switch-view-container {
	position: relative;
	min-block-size: $switch-view-height;
	margin-block-end: 1rem;
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 1rem;

	@media screen and (max-width: $tablet) {
		justify-content: center;
		flex-direction: column;
	}
}

.switch-view {
	background: var(--white);
	display: inline-flex;
	border: 1px solid var(--brand-border);
	border-radius: 12px;
	font-size: .78rem;
	box-shadow: 0 3px 12px rgba(31, 91, 73, .055);
	padding: .28rem;
	gap: .15rem;
}

.switch-view--hidden {
	position: absolute;
	visibility: hidden;
	pointer-events: none;
	white-space: nowrap;
	inset-inline-start: 0;
	inset-inline-end: 0;
	overflow: hidden;
}

.switch-view-dropdown-trigger {
	cursor: pointer;
	display: inline-flex;
	align-items: center;
	gap: .35rem;
	font-weight: 700;
	color: #fff;
	background: var(--brand-forest);
	border-radius: 10px;
	padding: .55rem .8rem;
}

.dropdown-icon { font-size: .6rem; }

.switch-view-button {
	padding: .45rem .72rem;
	display: block;
	white-space: nowrap;
	border-radius: 9px;
	color: var(--brand-text-muted);
	font-weight: 650;
	transition: all 120ms ease;

	&:not(:last-child) { margin-inline-end: 0; }
	&:hover {
		color: var(--brand-forest);
		background: var(--brand-surface-soft);
	}
	&.is-active {
		color: #fff;
		background: var(--brand-forest);
		font-weight: 700;
		box-shadow: 0 4px 10px rgba(31, 91, 73, .16);
	}
}

.is-archived .notification.is-warning { margin-block-end: 1rem; }

.project-title-print {
	display: none;
	font-size: 1.75rem;
	text-align: center;
	margin-block-end: .5rem;
	@media print { display: block; }
}
</style>

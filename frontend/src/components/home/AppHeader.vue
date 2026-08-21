<template>
	<header
		:class="{ 'has-background': background, 'menu-active': menuActive }"
		aria-label="main navigation"
		class="navbar d-print-none"
	>
		<RouterLink
			:to="{ name: 'home' }"
			class="logo-link"
			:aria-label="$t('navigation.home')"
		>
			<img
				:src="companyLogoUrl"
				alt=""
				class="company-logo"
				height="44"
			/>
		</RouterLink>

		<MenuButton class="menu-button" />

		<div
			v-if="currentProject?.id"
			class="project-title-wrapper"
		>
			<span class="project-title">
				{{ currentProject.title === '' ? $t('misc.loading') : getProjectTitle(currentProject) }}
			</span>

			<BaseButton
				v-if="!isEditorContentEmpty(currentProject.description)"
				:to="{ name: 'project.info', params: { projectId: currentProject.id } }"
				class="project-title-button"
			>
				<span class="is-sr-only">{{ $t('project.description') }}</span>
				<Icon icon="circle-info" />
			</BaseButton>

			<ProjectSettingsDropdown
				v-if="canWriteCurrentProject && currentProject.id !== -1"
				class="project-title-dropdown"
				:project="currentProject"
			>
				<template #trigger="{ toggleOpen, open }">
					<BaseButton
						class="project-title-button"
						:aria-expanded="open"
						@click="toggleOpen"
					>
						<span class="is-sr-only">{{ $t('project.openSettingsMenu') }}</span>
						<Icon
							icon="ellipsis-h"
							class="icon"
						/>
					</BaseButton>
				</template>
			</ProjectSettingsDropdown>
		</div>

		<div
			v-else-if="pageTitle"
			class="project-title-wrapper"
		>
			<span class="project-title">{{ pageTitle }}</span>
		</div>

		<div class="navbar-end">
			<TimerBadge />
			<OpenQuickActions />
			<Notifications />
			<Dropdown>
				<template #trigger="{ toggleOpen, open }">
					<BaseButton
						class="username-dropdown-trigger"
						variant="secondary"
						:shadow="false"
						:aria-expanded="open"
						@click="toggleOpen"
					>
						<img
							:src="authStore.avatarUrl"
							alt=""
							class="avatar"
							width="40"
							height="40"
						>
						<span class="username">{{ authStore.userDisplayName }}</span>
						<span
							class="mis-1 dropdown-icon icon is-small"
							:style="{
								transform: open ? 'rotate(180deg)' : 'rotate(0)',
							}"
						>
							<Icon icon="chevron-down" />
						</span>
					</BaseButton>
				</template>

				<DropdownItem :to="{ name: 'user.settings' }">
					{{ $t('user.settings.title') }}
				</DropdownItem>
				<DropdownItem
					v-if="adminPanelEnabled && authStore.info?.isAdmin"
					:to="{ name: 'admin.overview' }"
				>
					{{ $t('admin.title') }}
				</DropdownItem>
				<DropdownItem
					v-if="imprintUrl"
					:href="imprintUrl"
				>
					{{ $t('navigation.imprint') }}
				</DropdownItem>
				<DropdownItem
					v-if="privacyPolicyUrl"
					:href="privacyPolicyUrl"
				>
					{{ $t('navigation.privacy') }}
				</DropdownItem>
				<DropdownItem @click="baseStore.setKeyboardShortcutsActive(true)">
					{{ $t('keyboardShortcuts.title') }}
				</DropdownItem>
				<DropdownItem :to="{ name: 'about' }">
					{{ $t('about.title') }}
				</DropdownItem>
				<DropdownItem @click="authStore.logout()">
					{{ $t('user.auth.logout') }}
				</DropdownItem>
			</Dropdown>
		</div>
	</header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { PERMISSIONS as Permissions } from '@/constants/permissions'
import { PRO_FEATURE } from '@/constants/proFeatures'

import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import Notifications from '@/components/notifications/Notifications.vue'
import TimerBadge from '@/components/time-tracking/TimerBadge.vue'
import companyLogoUrl from '@/assets/company-logo.png'
import BaseButton from '@/components/base/BaseButton.vue'
import MenuButton from '@/components/home/MenuButton.vue'
import OpenQuickActions from '@/components/misc/OpenQuickActions.vue'

import { getProjectTitle } from '@/helpers/getProjectTitle'
import { isEditorContentEmpty } from '@/helpers/editorContentEmpty'

import { useBaseStore } from '@/stores/base'
import { useConfigStore } from '@/stores/config'
import { useAuthStore } from '@/stores/auth'
import type { IProject } from '@/modelTypes/IProject'

const baseStore = useBaseStore()
// Create a mutable copy to satisfy type requirements (readonly deep -> mutable)
const currentProject = computed<IProject | null>(() => {
	const project = baseStore.currentProject
	return project ? { ...project } as IProject : null
})
const background = computed(() => baseStore.background)
const canWriteCurrentProject = computed(() => baseStore.currentProject?.maxPermission !== null && baseStore.currentProject?.maxPermission !== undefined && baseStore.currentProject.maxPermission > Permissions.READ)
const menuActive = computed(() => baseStore.menuActive)

// Standalone pages (no project) surface their route's title in the header.
const route = useRoute()
const { t } = useI18n()
const pageTitle = computed(() => {
	const title = route.meta.title as string | undefined
	return title ? t(title) : ''
})

const authStore = useAuthStore()

const configStore = useConfigStore()
const imprintUrl = computed(() => configStore.legal.imprintUrl)
const privacyPolicyUrl = computed(() => configStore.legal.privacyPolicyUrl)
const adminPanelEnabled = computed(() => configStore.isProFeatureEnabled(PRO_FEATURE.ADMIN_PANEL))
</script>

<style lang="scss" scoped>
$user-dropdown-width-mobile: 5rem;

.navbar {
	--navbar-button-min-width: 42px;
	--navbar-gap-width: .75rem;
	--navbar-icon-size: 1.15rem;

	position: fixed;
	inset-block-start: 0;
	inset-inline-start: 0;
	inset-inline-end: 0;
	z-index: 30;

	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: var(--navbar-gap-width);
	min-block-size: $navbar-height;
	background: color-mix(in srgb, var(--white) 94%, transparent);
	backdrop-filter: blur(16px) saturate(135%);
	border-block-end: 1px solid var(--brand-border);
	box-shadow: 0 4px 18px rgba(20, 63, 52, .035);

	@media screen and (min-width: $tablet) {
		padding-inline: .75rem 1rem;
	}

	&.menu-active {
		@media screen and (max-width: $tablet) {
			z-index: 0;
		}
	}

	:deep(.trigger-button) {
		color: var(--brand-text-muted);
		font-size: var(--navbar-icon-size);
		border-radius: 10px;
		transition: background-color 140ms ease, color 140ms ease;

		&:hover {
			background: var(--brand-surface-soft);
			color: var(--brand-forest);
		}
	}
}

.logo-link {
	display: none;

	@media screen and (min-width: $tablet) {
		align-self: center;
		display: grid;
		place-items: center;
		inline-size: 48px;
		block-size: 48px;
		margin-inline-end: .2rem;
		border-radius: 14px;
		background: linear-gradient(145deg, var(--brand-forest-strong), var(--brand-forest-deep));
		box-shadow: 0 7px 18px rgba(13, 48, 40, .18);
	}
}

.menu-button {
	margin-inline-end: auto;
	align-self: stretch;
	flex: 0 0 auto;
	color: var(--brand-forest);

	@media screen and (max-width: $tablet) {
		margin-inline-start: .5rem;
	}
}

.project-title-wrapper {
	margin-inline: auto;
	display: flex;
	align-items: center;
	min-inline-size: 0;

	@media screen and (min-width: $tablet) {
		padding-inline: var(--navbar-gap-width);
	}
}

.project-title {
	font-size: 1rem;
	font-weight: 700;
	color: var(--text-strong);
	letter-spacing: -.025em;
	text-overflow: ellipsis;
	overflow: hidden;
	white-space: nowrap;

	@media screen and (min-width: $tablet) {
		font-size: 1.25rem;
	}
}

.project-title-dropdown {
	align-self: stretch;

	.project-title-button {
		flex-grow: 1;
	}
}

.project-title-button {
	align-self: stretch;
	min-inline-size: var(--navbar-button-min-width);
	display: flex;
	place-items: center;
	justify-content: center;
	font-size: var(--navbar-icon-size);
	color: var(--brand-text-muted);
	border-radius: 10px;

	&:hover {
		color: var(--brand-forest);
		background: var(--brand-surface-soft);
	}
}

.navbar-end {
	flex: 0 0 auto;
	display: flex;
	align-items: center;
	gap: .15rem;

	>* {
		min-inline-size: var(--navbar-button-min-width);
	}
}

.username-dropdown-trigger {
	padding-inline: .5rem .7rem;
	display: inline-flex;
	align-items: center;
	min-block-size: 44px;
	font-size: .84rem;
	font-weight: 700;
	gap: .5rem;
	border-radius: 12px;
	background: var(--brand-surface-soft);
	color: var(--text-strong);

	&:hover {
		background: color-mix(in srgb, var(--brand-surface-soft) 72%, var(--brand-lime-soft));
	}
	
	:deep(.avatar) {
		margin-inline-end: 0;
	}
	
	[dir="rtl"] & {
		flex-direction: row-reverse;
	}

	@media screen and (max-width: $tablet) {
		padding-inline: .35rem;
		background: transparent;
	}
}

.username {
	font-family: $vikunja-font;

	@media screen and (max-width: $tablet) {
		display: none;
	}
}

.dropdown-icon {
	transition: transform $transition;
	color: var(--brand-text-muted);
}

.avatar {
	border-radius: 12px;
	vertical-align: middle;
	block-size: 34px;
	inline-size: 34px;
	margin-inline-end: 0;
	object-fit: cover;
	box-shadow: 0 0 0 2px var(--white), 0 0 0 3px var(--brand-border);
}

.company-logo {
	display: block;
	inline-size: 34px;
	block-size: 40px;
	object-fit: contain;
	filter: drop-shadow(0 4px 8px rgba(0, 0, 0, .14));
}
</style>

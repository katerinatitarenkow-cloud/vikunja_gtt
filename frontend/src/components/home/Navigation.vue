<template>
	<aside
		:class="{'is-active': baseStore.menuActive, 'is-resizing': isResizing}"
		class="menu-container"
		:style="{'--sidebar-width': sidebarWidth}"
	>
		<nav
			class="menu top-menu"
			:aria-label="$t('navigation.main')"
		>
			<RouterLink
				:to="{name: 'home'}"
				class="logo"
				:aria-label="$t('navigation.home')"
			>
				<img
				:src="companyLogoUrl"
				alt=""
				class="company-logo"
				height="44"
			/>
			</RouterLink>
			<menu class="menu-list other-menu-items">
				<li>
					<RouterLink
						v-shortcut="'KeyG KeyO'"
						:to="{ name: 'home'}"
					>
						<span class="menu-item-icon icon">
							<Icon icon="calendar" />
						</span>
						{{ $t('navigation.overview') }}
					</RouterLink>
				</li>
				<li v-if="canTasks">
					<RouterLink
						v-shortcut="'KeyG KeyU'"
						:to="{ name: 'tasks.range'}"
					>
						<span class="menu-item-icon icon">
							<Icon :icon="['far', 'calendar-alt']" />
						</span>
						{{ $t('navigation.upcoming') }}
					</RouterLink>
				</li>
				<li v-if="canProjects">
					<RouterLink
						v-shortcut="'KeyG KeyP'"
						:to="{ name: 'projects.index'}"
					>
						<span class="menu-item-icon icon">
							<Icon icon="layer-group" />
						</span>
						{{ $t('project.projects') }}
					</RouterLink>
				</li>
				<li>
					<RouterLink :to="{ name: 'mailbox'}">
						<span class="menu-item-icon icon">
							<Icon icon="envelope" />
						</span>
						{{ $t('mailbox.title') }}
<span
v-if="mailboxStore.unreadCount > 0"
class="mailbox-nav-badge"
>
{{
mailboxStore.unreadCount > 99
? '99+'
: mailboxStore.unreadCount
}}
</span>
					</RouterLink>
				</li>
				<li v-if="canLabels">
					<RouterLink
						v-shortcut="'KeyG KeyA'"
						:to="{ name: 'labels.index'}"
					>
						<span class="menu-item-icon icon">
							<Icon icon="tags" />
						</span>
						{{ $t('label.title') }}
					</RouterLink>
				</li>
				<li v-if="canTeams">
					<RouterLink
						v-shortcut="'KeyG KeyM'"
						:to="{ name: 'teams.index'}"
					>
						<span class="menu-item-icon icon">
							<Icon icon="users" />
						</span>
						{{ $t('team.title') }}
					</RouterLink>
				</li>
				<li v-if="isInstanceAdmin">
					<RouterLink :to="{ name: 'team-management'}">
						<span class="menu-item-icon icon">
							<Icon icon="user-edit" />
						</span>
						{{ $t('admin.team.title') }}
					</RouterLink>
				</li>
				<li v-if="canTasks">
					<RouterLink :to="{ name: 'work-groups'}">
						<span class="menu-item-icon icon">
							<Icon icon="users" />
						</span>
						{{ $t('workGroups.title') }}
					</RouterLink>
				</li>
				<li v-if="canWialon">
					<RouterLink :to="{ name: 'wialon'}">
						<span class="menu-item-icon icon">
							<Icon icon="sitemap" />
						</span>
						{{ $t('wialon.title') }}
					</RouterLink>
				</li>
				<li v-if="timeTrackingEnabled && canTimeTracking">
					<RouterLink :to="{ name: 'time-tracking'}">
						<span class="menu-item-icon icon">
							<Icon :icon="['far', 'clock']" />
						</span>
						{{ $t('timeTracking.title') }}
					</RouterLink>
				</li>
			</menu>
		</nav>

		<Loading
			v-if="canProjects && projectStore.isLoading"
			variant="small"
		/>
		<template v-else-if="canProjects">
			<nav
				v-if="favoriteProjects.length"
				class="menu"
				:aria-label="$t('project.pseudo.favorites.title')"
			>
				<ProjectsNavigation
					:model-value="favoriteProjects"
					:can-edit-order="false"
					:can-collapse="false"
				/>
			</nav>
			
			<nav
				v-if="savedFilterProjects.length"
				class="menu"
				:aria-label="$t('navigation.savedFilters')"
			>
				<ProjectsNavigation
					:model-value="savedFilterProjects"
					:can-edit-order="false"
					:can-collapse="false"
				/>
			</nav>

			<nav
				class="menu"
				:aria-label="$t('project.projects')"
			>
				<ProjectsNavigation
					:model-value="projects"
					:can-edit-order="true"
					:can-collapse="true"
				/>
			</nav>
		</template>

		<PoweredByLink
			class="mbs-auto"
			utm-medium="navigation"
		/>

		<div
			v-if="!isMobile"
			class="resize-handle"
			@mousedown="startResize"
			@touchstart="startResize"
		/>
	</aside>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import PoweredByLink from '@/components/home/PoweredByLink.vue'
import companyLogoUrl from '@/assets/company-logo.png'
import Loading from '@/components/misc/Loading.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useConfigStore} from '@/stores/config'
import {useAccessStore} from '@/stores/access'
import {useAuthStore} from '@/stores/auth'
import {useMailboxStore} from '@/stores/mailbox'
import {ACCESS_PERMISSION} from '@/modelTypes/IAccessControl'
import {PRO_FEATURE} from '@/constants/proFeatures'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'
import type {IProject} from '@/modelTypes/IProject'
import {useSidebarResize} from '@/composables/useSidebarResize'

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const configStore = useConfigStore()
const accessStore = useAccessStore()
const authStore = useAuthStore()
const mailboxStore = useMailboxStore()

void mailboxStore.refreshUnread()

const timeTrackingEnabled = computed(() => configStore.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING))
const canProjects = computed(() => accessStore.can(ACCESS_PERMISSION.PROJECTS_VIEW))
const canTasks = computed(() => accessStore.can(ACCESS_PERMISSION.TASKS_VIEW))
const canLabels = computed(() => accessStore.can(ACCESS_PERMISSION.LABELS_VIEW))
const canTeams = computed(() => accessStore.can(ACCESS_PERMISSION.TEAMS_VIEW))
const canWialon = computed(() => accessStore.can(ACCESS_PERMISSION.WIALON_VIEW))
const canTimeTracking = computed(() => accessStore.can(ACCESS_PERMISSION.TIME_TRACKING))
const isInstanceAdmin = computed(() => Boolean(authStore.info?.isAdmin))

const {sidebarWidth, isResizing, startResize, isMobile} = useSidebarResize()

// Cast readonly arrays to mutable type - the arrays are not actually mutated by the component
const projects = computed(() => projectStore.notArchivedRootProjects as IProject[])
const favoriteProjects = computed(() => projectStore.favoriteProjects as IProject[])
const savedFilterProjects = computed(() => projectStore.savedFilterProjects as IProject[])
</script>

<style lang="scss" scoped>
.logo {
	display: flex;
	align-items: center;
	padding: .75rem 1rem 1rem;
	margin: 0;

	@media screen and (min-width: $tablet) {
		display: none;
	}
}

.menu-container {
	--sidebar-width: #{$navbar-width};
	// Local token overrides make all nested project-navigation components work on the dark shell.
	--primary: var(--brand-lime);
	--primary-h: 78.4deg;
	--primary-s: 100%;
	--primary-l: 75.1%;
	--primary-hsl: var(--primary-h), var(--primary-s), var(--primary-l);
	--grey-700: #edf5f1;
	--grey-600: #d5e4de;
	--grey-500: #b1c5bd;
	--grey-400: #8fa89e;
	--grey-300: #68877b;
	--grey-200: rgba(255, 255, 255, .12);
	--grey-100: rgba(255, 255, 255, .08);
	--white: rgba(255, 255, 255, .075);

	display: flex;
	flex-direction: column;
	background:
		radial-gradient(circle at 12% 0%, rgba(216, 255, 128, .09), transparent 34%),
		linear-gradient(180deg, var(--brand-forest-strong) 0%, var(--brand-forest-deep) 100%);
	color: #edf5f1;
	padding: 1rem .75rem .75rem;
	border-inline-end: 1px solid rgba(255, 255, 255, .07);
	box-shadow: 12px 0 32px rgba(7, 38, 31, .08);
	transition: transform $transition-duration ease-in;
	position: fixed;
	inset-block-start: $navbar-height;
	inset-block-end: 0;
	inset-inline-start: 0;
	transform: translateX(-100%);
	inline-size: var(--sidebar-width);
	overflow-y: auto;
	overflow-x: hidden;

	[dir="rtl"] & {
		transform: translateX(100%);
	}

	@media screen and (max-width: $tablet) {
		inset-block-start: 0;
		inline-size: min(84vw, 320px);
		z-index: 20;
		padding-block-start: .75rem;
		box-shadow: 18px 0 45px rgba(0, 0, 0, .24);
	}

	&.is-active {
		transform: translateX(0);
		transition: transform $transition-duration ease-out;
	}

	&.is-resizing {
		transition: none;
	}

	:deep(.menu-list a),
	:deep(.menu-list .list-menu-link) {
		border-radius: 10px;
		color: #dce9e4;
		min-block-size: 42px;
		transition: background-color 140ms ease, color 140ms ease, transform 140ms ease;
	}

	:deep(.menu-list a:hover),
	:deep(.menu-list .list-menu-link:hover) {
		background: rgba(255, 255, 255, .075);
		color: #fff;
	}

	:deep(.menu-list a.router-link-exact-active) {
		background: rgba(216, 255, 128, .13);
		color: var(--brand-lime);
		font-weight: 700;
		box-shadow: inset 3px 0 0 var(--brand-lime);
	}

	:deep(.menu-list a.router-link-exact-active .icon:not(.handle)) {
		color: var(--brand-lime);
	}

	:deep(.menu-item-icon),
	:deep(.collapse-project-button svg) {
		color: #89a69a;
	}

	:deep(.color-bubble) {
		box-shadow: 0 0 0 2px rgba(255, 255, 255, .12);
	}
}

.mailbox-nav-badge {
display: inline-flex;
align-items: center;
justify-content: center;
min-inline-size: 1.35rem;
block-size: 1.35rem;
padding-inline: .32rem;
margin-inline-start: auto;
border-radius: 999px;
background: #ef4444;
color: #fff;
font-size: .68rem;
font-weight: 800;
line-height: 1;
box-shadow: 0 0 0 2px rgba(255, 255, 255, .08);
}
.resize-handle {
	position: absolute;
	inset-block-start: 0;
	inset-block-end: 0;
	inset-inline-end: 0;
	inline-size: 5px;
	cursor: ew-resize;
	background: transparent;
	transition: background-color $transition-duration ease;
	touch-action: none;

	&:hover,
	&:active {
		background-color: rgba(216, 255, 128, .38);
	}
}

.top-menu .menu-list {
	li {
		font-weight: 600;
		font-family: $vikunja-font;
	}

	.list-menu-link,
	li > a {
		padding-inline: .85rem;
		display: flex;
		gap: .15rem;

		.icon {
			padding-block-end: 0;
		}
	}
}

.menu + .menu {
	padding-block-start: .8rem;
	margin-block-start: .55rem;
	border-block-start: 1px solid rgba(255, 255, 255, .065);
}

.company-logo {
	display: block;
	inline-size: 42px;
	block-size: 42px;
	object-fit: contain;
	filter: drop-shadow(0 6px 14px rgba(0, 0, 0, .16));
}
</style>

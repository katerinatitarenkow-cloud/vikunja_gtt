import { createRouter, createWebHistory } from 'vue-router'
import type { RouteLocation } from 'vue-router'
import {saveLastVisited} from '@/helpers/saveLastVisited'
import {getProjectViewId} from '@/helpers/projectView'

import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {getNextWeekDate} from '@/helpers/time/getNextWeekDate'
import {LINK_SHARE_HASH_PREFIX} from '@/constants/linkShareHash'
import {REDIRECT_HASH_PREFIX} from '@/constants/redirectHash'
import {AUTH_ROUTE_NAMES} from '@/constants/authRouteNames'
import {PRO_FEATURE} from '@/constants/proFeatures'

import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'
import {useAccessStore} from '@/stores/access'
import {ACCESS_PERMISSION} from '@/modelTypes/IAccessControl'

import Login from '@/views/user/Login.vue'
import LinkSharingAuth from '@/views/sharing/LinkSharingAuth.vue'
import OpenIdAuth from '@/views/user/OpenIdAuth.vue'
import UpcomingTasks from '@/views/tasks/ShowTasks.vue'

import NotFoundComponent from '@/views/404.vue'

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	scrollBehavior(to, from, savedPosition) {
		// If the user is using their forward/backward keys to navigate, we want to restore the scroll view
		if (savedPosition) {
			return savedPosition
		}

		// Scroll to anchor should still work
		if (to.hash && !to.hash.startsWith(LINK_SHARE_HASH_PREFIX) && !to.hash.startsWith(REDIRECT_HASH_PREFIX)) {
			return {el: to.hash}
		}

		// Otherwise just scroll to the top
		return {
			'inset-inline-start': 0,
			'inset-block-start': 0,
		}
	},
	routes: [
		{
			path: '/',
			name: 'home',
			component: () => import('@/views/Home.vue'),
		},
		{
			path: '/:pathMatch(.*)*',
			name: 'not-found',
			component: NotFoundComponent,
		},
		// if you omit the last `*`, the `/` character in params will be encoded when resolving or pushing
		{
			path: '/:pathMatch(.*)',
			name: 'bad-not-found',
			component: NotFoundComponent,
		},
		{
			path: '/login',
			name: 'user.login',
			component: Login,
			meta: {
				title: 'user.auth.login',
			},
		},
		{
			path: '/get-password-reset',
			name: 'user.password-reset.request',
			component: () => import('@/views/user/RequestPasswordReset.vue'),
			meta: {
				title: 'user.auth.resetPassword',
			},
		},
		{
			path: '/password-reset',
			name: 'user.password-reset.reset',
			component: () => import('@/views/user/PasswordReset.vue'),
			meta: {
				title: 'user.auth.resetPassword',
			},
		},
		{
			path: '/register',
			name: 'user.register',
			redirect: {name: 'user.login'},
		},
		{
			path: '/user/settings',
			name: 'user.settings',
			component: () => import('@/views/user/Settings.vue'),
			redirect: {name: 'user.settings.general'},
			children: [
				{
					path: '/user/settings/avatar',
					name: 'user.settings.avatar',
					component: () => import('@/views/user/settings/Avatar.vue'),
				},
				{
					path: '/user/settings/caldav',
					name: 'user.settings.caldav',
					component: () => import('@/views/user/settings/Caldav.vue'),
					beforeEnter: async () => {
						const {useConfigStore} = await import('@/stores/config')
						if (!useConfigStore().caldavEnabled) {
							return {name: 'user.settings.general'}
						}
					},
				},
				{
					path: '/user/settings/data-export',
					name: 'user.settings.data-export',
					component: () => import('@/views/user/settings/DataExport.vue'),
				},
				{
					path: '/user/settings/feeds',
					name: 'user.settings.feeds',
					component: () => import('@/views/user/settings/AtomFeed.vue'),
				},
				{
					path: '/user/settings/deletion',
					name: 'user.settings.deletion',
					component: () => import('@/views/user/settings/Deletion.vue'),
				},
				{
					path: '/user/settings/email-update',
					name: 'user.settings.email-update',
					component: () => import('@/views/user/settings/EmailUpdate.vue'),
				},
				{
					path: '/user/settings/general',
					name: 'user.settings.general',
					component: () => import('@/views/user/settings/General.vue'),
				},
				{
					path: '/user/settings/password-update',
					name: 'user.settings.password-update',
					component: () => import('@/views/user/settings/PasswordUpdate.vue'),
				},
				{
					path: '/user/settings/totp',
					name: 'user.settings.totp',
					component: () => import('@/views/user/settings/TOTP.vue'),
					beforeEnter: async () => {
						const {useConfigStore} = await import('@/stores/config')
						if (!useConfigStore().totpEnabled || !useAuthStore().info?.isLocalUser) {
							return {name: 'user.settings.general'}
						}
					},
				},
				{
					path: '/user/settings/api-tokens',
					name: 'user.settings.apiTokens',
					component: () => import('@/views/user/settings/ApiTokens.vue'),
				},
				{
					path: '/user/settings/sessions',
					name: 'user.settings.sessions',
					component: () => import('@/views/user/settings/Sessions.vue'),
				},
				{
					path: '/user/settings/webhooks',
					name: 'user.settings.webhooks',
					component: () => import('@/views/user/settings/Webhooks.vue'),
				},
				{
					path: '/user/settings/team-management',
					name: 'user.settings.team-management',
					component: () => import('@/views/admin/TeamManagementView.vue'),
					meta: {requiresInstanceAdmin: true},
				},
				{
					path: '/user/settings/wialon',
					name: 'user.settings.wialon',
					component: () => import('@/views/user/settings/WialonSettings.vue'),
					meta: {requiresInstanceAdmin: true},
				},
				{
					path: '/user/settings/bots',
					name: 'user.settings.bots',
					component: () => import('@/views/user/settings/BotUsers.vue'),
				},
				{
					path: '/user/settings/migrate',
					name: 'migrate.start',
					component: () => import('@/views/migrate/Migration.vue'),
				},
				{
					path: '/migrate/csv',
					name: 'migrate.csv',
					component: () => import('@/views/migrate/MigrationCSV.vue'),
				},
				{
					path: '/migrate/:service',
					name: 'migrate.service',
					component: () => import('@/views/migrate/MigrationHandler.vue'),
					props: route => ({
						service: route.params.service as string,
						code: route.query.code as string,
					}),
				},
			],
		},
		{
			path: '/user/export/download',
			name: 'user.export.download',
			component: () => import('@/views/user/DataExportDownload.vue'),
		},
		{
			path: '/share/:share/auth',
			name: 'link-share.auth',
			// FIXME: use dynamic imports
			// component: () => import('@/views/sharing/LinkSharingAuth.vue'),
			component: LinkSharingAuth,
		},
		{
			path: '/tasks/:id',
			name: 'task.detail',
			component: () => import('@/views/tasks/TaskDetailView.vue'),
			props: route => ({ taskId: Number(route.params.id as string) }),
			meta: {requiredPermission: ACCESS_PERMISSION.TASKS_VIEW},
		},
		{
			path: '/tasks/by/upcoming',
			name: 'tasks.range',
			component: UpcomingTasks,
			props: route => ({
				dateFrom: parseDateOrString(route.query.from as string, new Date()),
				dateTo: parseDateOrString(route.query.to as string, getNextWeekDate()),
				showNulls: route.query.showNulls === 'true',
				showOverdue: route.query.showOverdue === 'true',
			}),
			meta: {requiredPermission: ACCESS_PERMISSION.TASKS_VIEW},
		},
		{
			// Redirect old list routes to the respective project routes
			// see: https://router.vuejs.org/guide/essentials/dynamic-matching.html#catch-all-404-not-found-route
			path: '/lists:pathMatch(.*)*',
			name: 'lists',
			redirect(to) {
				return {
					path: to.path.replace('/lists', '/projects'),
					query: to.query,
					hash: to.hash,
				}
			},
		},
		{
			path: '/projects',
			name: 'projects.index',
			component: () => import('@/views/project/ListProjects.vue'),
			meta: {requiredPermission: ACCESS_PERMISSION.PROJECTS_VIEW},
		},
		{
			path: '/mailbox',
			name: 'mailbox',
			component: () => import('@/views/mailbox/MailboxView.vue'),
			meta: {title: 'mailbox.title'},
		},
		{
			path: '/wialon',
			name: 'wialon',
			component: () => import('@/views/wialon/WialonView.vue'),
			meta: {
				title: 'wialon.title',
				requiredPermission: ACCESS_PERMISSION.WIALON_VIEW,
			},
		},
		{
			path: '/projects/new',
			name: 'project.create',
			component: () => import('@/views/project/NewProject.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:parentProjectId/new',
			name: 'project.createFromParent',
			component: () => import('@/views/project/NewProject.vue'),
			props: route => ({ parentProjectId: Number(route.params.parentProjectId as string) }),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:projectId/settings/edit',
			name: 'project.settings.edit',
			component: () => import('@/views/project/settings/ProjectSettingsEdit.vue'),
			props: route => ({ projectId: Number(route.params.projectId as string) }),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:projectId/settings/background',
			name: 'project.settings.background',
			component: () => import('@/views/project/settings/ProjectSettingsBackground.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:projectId/settings/duplicate',
			name: 'project.settings.duplicate',
			component: () => import('@/views/project/settings/ProjectSettingsDuplicate.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:projectId/settings/share',
			name: 'project.settings.share',
			component: () => import('@/views/project/settings/ProjectSettingsShare.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:projectId/settings/webhooks',
			name: 'project.settings.webhooks',
			component: () => import('@/views/project/settings/ProjectSettingsWebhooks.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:projectId/settings/delete',
			name: 'project.settings.delete',
			component: () => import('@/views/project/settings/ProjectSettingsDelete.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:projectId/settings/archive',
			name: 'project.settings.archive',
			component: () => import('@/views/project/settings/ProjectSettingsArchive.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
		},
		{
			path: '/projects/:projectId/settings/views',
			name: 'project.settings.views',
			component: () =>  import('@/views/project/settings/ProjectSettingsViews.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.PROJECTS_MANAGE,
			},
			props: route => ({ projectId: Number(route.params.projectId as string) }),
		},
		{
			path: '/projects/:projectId/settings/edit',
			name: 'filter.settings.edit',
			component: () => import('@/views/filters/FilterEdit.vue'),
			meta: {
				showAsModal: true,
			},
			props: route => ({ projectId: Number(route.params.projectId as string) }),
		},
		{
			path: '/projects/:projectId/settings/delete',
			name: 'filter.settings.delete',
			component: () => import('@/views/filters/FilterDelete.vue'),
			meta: {
				showAsModal: true,
			},
			props: route => ({ projectId: Number(route.params.projectId as string) }),
		},
		{
			path: '/projects/:projectId/info',
			name: 'project.info',
			component: () => import('@/views/project/ProjectInfo.vue')			,
			meta: {
				showAsModal: true,
			},
			props: route => ({ projectId: Number(route.params.projectId as string) }),
		},
		{
			path: '/projects/:projectId',
			name: 'project.index',
			meta: {requiredPermission: ACCESS_PERMISSION.PROJECTS_VIEW},
			redirect(to) {
				const projectId = parseInt(to.params.projectId as string)
				// Saved filters, Vikunja's pseudo-projects and public link shares are task
				// views, not CRM client cards. CRM contains private contact/legal data.
				if (projectId <= 0 || to.hash.startsWith(LINK_SHARE_HASH_PREFIX)) {
					return {name: 'project.view', params: {projectId, viewId: getProjectViewId(projectId) ?? 0}, hash: to.hash}
				}
				return {name: 'project.client', params: {projectId}}
			},
		},
		{
			path: '/projects/:projectId/client',
			name: 'project.client',
			component: () => import('@/views/project/ClientProfileView.vue'),
			props: route => ({projectId: parseInt(route.params.projectId as string)}),
			meta: {requiredPermission: ACCESS_PERMISSION.PROJECTS_VIEW},
		},
		{
			path: '/projects/:projectId/history',
			name: 'project.history',
			component: () => import('@/views/project/ClientHistoryView.vue'),
			props: route => ({projectId: parseInt(route.params.projectId as string)}),
			meta: {requiredPermission: ACCESS_PERMISSION.PROJECTS_VIEW},
		},
		{
			path: '/projects/:projectId/:viewId',
			name: 'project.view',
			component: () => import('@/views/project/ProjectView.vue'),
			props: route => ({ 
				projectId: parseInt(route.params.projectId as string),
				viewId: route.params.viewId ? parseInt(route.params.viewId as string): undefined,
			}),
			meta: {requiredPermission: ACCESS_PERMISSION.PROJECTS_VIEW},
		},
		{
			path: '/teams',
			name: 'teams.index',
			component: () => import('@/views/teams/ListTeams.vue'),
			meta: {requiredPermission: ACCESS_PERMISSION.TEAMS_VIEW},
		},
		{
			path: '/team-management',
			name: 'team-management',
			component: () => import('@/views/admin/TeamManagementView.vue'),
			meta: {requiresInstanceAdmin: true},
		},
		{
			path: '/groups',
			name: 'work-groups',
			component: () => import('@/views/groups/WorkGroupsView.vue'),
			meta: {requiredPermission: ACCESS_PERMISSION.TASKS_VIEW},
		},
		{
			path: '/teams/new',
			name: 'teams.create',
			component: () =>  import('@/views/teams/NewTeam.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.TEAMS_MANAGE,
			},
		},
		{
			path: '/teams/:id/edit',
			name: 'teams.edit',
			component: () => import('@/views/teams/EditTeam.vue'),
			meta: {requiredPermission: ACCESS_PERMISSION.TEAMS_MANAGE},
		},
		{
			path: '/labels',
			name: 'labels.index',
			component: () => import('@/views/labels/ListLabels.vue'),
			meta: {requiredPermission: ACCESS_PERMISSION.LABELS_VIEW},
		},
		{
			path: '/labels/new',
			name: 'labels.create',
			component: () => import('@/views/labels/NewLabel.vue'),
			meta: {
				showAsModal: true,
				requiredPermission: ACCESS_PERMISSION.LABELS_MANAGE,
			},
		},
		{
			path: '/filters/new',
			name: 'filters.create',
			component: () => import('@/views/filters/FilterNew.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/auth/openid/:provider',
			name: 'openid.auth',
			component: OpenIdAuth,
		},
		{
			path: '/oauth/authorize',
			name: 'oauth.authorize',
			component: () => import('@/views/user/OAuthAuthorize.vue'),
		},
		{
			path: '/about',
			name: 'about',
			component: () => import('@/views/About.vue'),
		},
		{
			path: '/time-tracking',
			name: 'time-tracking',
			component: () => import('@/views/time-tracking/TimeTracking.vue'),
			meta: {
				requiresTimeTracking: true,
				requiredPermission: ACCESS_PERMISSION.TIME_TRACKING,
				title: 'timeTracking.title',
			},
		},
		{
			path: '/admin',
			component: () => import('@/views/admin/AdminShell.vue'),
			meta: {
				requiresAdminPanel: true,
				adminMode: true,
			},
			children: [
				{
					path: '',
					name: 'admin.overview',
					component: () => import('@/views/admin/OverviewView.vue'),
				},
				{
					path: 'team',
					name: 'admin.team',
					component: () => import('@/views/admin/TeamManagementView.vue'),
				},
				{
					path: 'users',
					name: 'admin.users',
					component: () => import('@/views/admin/UsersView.vue'),
				},
				{
					path: 'projects',
					name: 'admin.projects',
					component: () => import('@/views/admin/ProjectsView.vue'),
				},
			],
		},
	],
})

export async function getAuthForRoute(to: RouteLocation, authStore) {
	// vue-router already decoded to.hash once, so slicing off the prefix yields the original
	// fullPath (e.g. /oauth/authorize?...) losslessly — no extra decodeURIComponent needed.
	const redirectDest = to.name === 'user.login' && to.hash.startsWith(REDIRECT_HASH_PREFIX)
		? to.hash.slice(REDIRECT_HASH_PREFIX.length)
		: ''

	if (authStore.authUser || authStore.authLinkShare) {
		// An already-signed-in browser that opens a copied /login#redirect=<oauth.authorize> URL
		// must run the OAuth flow with its existing session instead of short-circuiting to home.
		// The destination has no redirect hash, so the second guard pass just early-returns (#2654).
		if (redirectDest) {
			return redirectDest
		}
		return
	}

	// Check if password reset token is in query params
	const resetToken = to.query.userPasswordReset as string | undefined
	
	// Redirect to password reset page if we have a token stored
	if (resetToken && to.name !== 'user.password-reset.reset') {
		return {name: 'user.password-reset.reset', query: { userPasswordReset: resetToken }}
	}

	if (typeof resetToken === 'undefined' && to.name === 'user.password-reset.reset') {
		return {name: 'user.login'}
	}

	// Check if email confirmation token is in query params
	const emailConfirmToken = to.query.userEmailConfirm as string | undefined
	if (emailConfirmToken) {
		// Save token to localStorage before redirecting
		localStorage.setItem('emailConfirmToken', emailConfirmToken)
		// Redirect to login page where it will be processed
		if (to.name !== 'user.login') {
			return {name: 'user.login'}
		}
	}

	// Keep the destination in the address bar (not just per-browser localStorage) so a native
	// client's /oauth/authorize URL stays copyable into another browser. Hash, not query, so the
	// embedded OAuth params never reach access logs (#2654). Pass fullPath raw: vue-router encodes
	// the hash itself, so an extra encodeURIComponent here would be double-encoded in the URL.
	if (to.name === 'oauth.authorize') {
		return {
			name: 'user.login',
			hash: REDIRECT_HASH_PREFIX + to.fullPath,
		}
	}

	// Fold the hash destination into localStorage: it's the only bridge that survives the
	// external OIDC round-trip out of the SPA, so redirectIfSaved() works after any auth method.
	// vue-router already decoded to.hash once, so it equals the fullPath we wrote above as-is.
	if (to.hash.startsWith(REDIRECT_HASH_PREFIX)) {
		const destination = to.hash.slice(REDIRECT_HASH_PREFIX.length)
		const resolved = router.resolve(destination)
		saveLastVisited(resolved.name as string, resolved.params, resolved.query)
	}

	// Check if the route the user wants to go to is a route which needs authentication. We use this to
	// redirect the user after successful login.
	const isValidUserAppRoute = !AUTH_ROUTE_NAMES.has(to.name as string) &&
		localStorage.getItem('emailConfirmToken') === null

	if (isValidUserAppRoute) {
		saveLastVisited(to.name as string, to.params, to.query)
	}

	if (isValidUserAppRoute) {
		return {name: 'user.login'}
	}
	
	if(localStorage.getItem('emailConfirmToken') !== null && to.name !== 'user.login') {
		return {name: 'user.login', query: to.query}
	}
}

router.beforeEach(async (to, from) => {
	const authStore = useAuthStore()

	await authStore.checkAuth()

	// Load global feature permissions for every authenticated user before
	// rendering navigation. Without this, permission-controlled links could
	// briefly disappear on the first page after login.
	if (authStore.authUser) {
		await useAccessStore().ensureLoaded()
	}

	if (to.meta?.requiresAdminPanel) {
		// Await config/auth hydration so the license check doesn't race the empty default
		// on direct /admin navigation. appReady resolves without waiting on router.isReady(),
		// so awaiting it here doesn't deadlock the initial navigation.
		const baseStore = useBaseStore()
		await baseStore.appReady
		const configStore = useConfigStore()
		const featureOn = configStore.isProFeatureEnabled(PRO_FEATURE.ADMIN_PANEL)
		// isAdmin comes from /user, not the JWT; force-fetch in case checkAuth() was debounced.
		if (authStore.info?.isAdmin === undefined) {
			await authStore.refreshUserInfo()
		}
		const isAdmin = authStore.info?.isAdmin === true
		if (!featureOn || !isAdmin) {
			return {name: 'not-found'}
		}
	}

	if (to.meta?.requiresInstanceAdmin) {
		if (authStore.info?.isAdmin === undefined) await authStore.refreshUserInfo()
		if (authStore.info?.isAdmin !== true) return {name: 'not-found'}
	}

	if ((to.name === 'project.client' || to.name === 'project.history') && authStore.isLinkShareAuth) {
		return {name: 'not-found'}
	}

	if (to.meta?.requiredPermission) {
		const accessStore = useAccessStore()
		await accessStore.ensureLoaded()
		if (!accessStore.can(to.meta.requiredPermission as string)) return {name: 'not-found'}
	}

	if (to.meta?.requiresTimeTracking) {
		const baseStore = useBaseStore()
		await baseStore.appReady
		const configStore = useConfigStore()
		if (!configStore.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING)) {
			return {name: 'not-found'}
		}
	}

	if(from.hash && from.hash.startsWith(LINK_SHARE_HASH_PREFIX)) {
		to.hash = from.hash
	}

	if (to.hash.startsWith(LINK_SHARE_HASH_PREFIX) && !authStore.authLinkShare) {
		saveLastVisited(to.name as string, to.params, to.query)
		return {
			name: 'link-share.auth',
			params: {
				share: to.hash.replace(LINK_SHARE_HASH_PREFIX, ''),
			},
		}
	}

	const newRoute = await getAuthForRoute(to, authStore)
	if(newRoute) {
		// A string target (the decoded redirect destination for an authed browser) already
		// carries its own query/path and no redirect hash, so navigate to it verbatim — don't
		// re-attach to.hash or it would re-enter the redirect loop.
		if (typeof newRoute === 'string') {
			return newRoute
		}
		return {
			hash: to.hash,
			...newRoute,
		}
	}

	// to.fullPath keeps the redirect hash url-encoded while to.hash is decoded, so the endsWith
	// check below never matches and would re-append the hash forever. The hash is already on the
	// URL here, so skip the re-attach (#2654).
	if (to.hash.startsWith(REDIRECT_HASH_PREFIX)) {
		return
	}

	if(!to.fullPath.endsWith(to.hash)) {
		return to.fullPath + to.hash
	}
})

export default router

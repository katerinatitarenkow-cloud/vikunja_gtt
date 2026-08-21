export interface IAccessPermissionDefinition {
	key: string
	category: string
}

export interface IAccessGroup {
	id: number
	name: string
	description: string
	system_key?: 'admin' | 'users' | string
	permissions: string[]
	member_count: number
	created?: number
	updated?: number
}

export interface IAccessUser {
	id: number
	username: string
	name: string
	email: string
	phone: string
	notes: string
	is_admin: boolean
	status: number
	created: number
	updated: number
	groups: IAccessGroup[]
}

export interface IAccessMe {
	permissions: string[]
	groups: IAccessGroup[]
}

export const ACCESS_PERMISSION = {
	PROJECTS_VIEW: 'projects.view',
	PROJECTS_MANAGE: 'projects.manage',
	TASKS_VIEW: 'tasks.view',
	TASKS_MANAGE: 'tasks.manage',
	LABELS_VIEW: 'labels.view',
	LABELS_MANAGE: 'labels.manage',
	TEAMS_VIEW: 'teams.view',
	TEAMS_MANAGE: 'teams.manage',
	KANBAN_USE: 'kanban.use',
	TIME_TRACKING: 'time_tracking.use',
	WIALON_VIEW: 'wialon.view',
} as const

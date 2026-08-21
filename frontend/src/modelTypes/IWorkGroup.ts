import type {IUser} from '@/modelTypes/IUser'

export interface IWorkGroupSummary {
	id: number
	name: string
	leader_user_id: number
}

export interface IWorkGroup {
	id: number
	name: string
	description: string
	leader_user_id: number
	leader?: IUser | null
	members: IUser[]
	member_count: number
	task_count: number
	created?: string
	updated?: string
}

export interface IWorkGroupTaskAssignmentResult {
	group: IWorkGroup
	assigned_users: IUser[]
	skipped_users: IUser[]
}

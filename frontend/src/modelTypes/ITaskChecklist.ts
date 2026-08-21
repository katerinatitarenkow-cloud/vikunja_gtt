import type {IUser} from './IUser'

export interface ITaskChecklistItem {
	id: number
	task_id: number
	title: string
	done: boolean
	completed_by?: IUser | null
	completed_at?: string | Date | null
	position: number
	created: string | Date
	updated: string | Date
}

export interface ITaskChecklistState {
	items: ITaskChecklistItem[]
	total: number
	completed: number
	task_done: boolean
	task_done_at: string | Date | null
}

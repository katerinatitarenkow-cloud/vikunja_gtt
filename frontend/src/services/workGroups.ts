import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import type {IWorkGroup, IWorkGroupTaskAssignmentResult} from '@/modelTypes/IWorkGroup'

export interface WorkGroupPayload {
	name?: string
	description?: string
	leader_user_id?: number
	member_ids?: number[]
}

export default class WorkGroupService {
	private http = AuthenticatedHTTPFactory()

	async getAll(search = ''): Promise<IWorkGroup[]> {
		const {data} = await this.http.get<{groups: IWorkGroup[]}>(apiV2Url('work-groups'), {params: {search}})
		return data.groups
	}

	async create(payload: Required<Pick<WorkGroupPayload, 'name'>> & WorkGroupPayload): Promise<IWorkGroup> {
		const {data} = await this.http.post<IWorkGroup>(apiV2Url('admin/work-groups'), payload)
		return data
	}

	async update(id: number, payload: WorkGroupPayload): Promise<IWorkGroup> {
		const {data} = await this.http.patch<IWorkGroup>(apiV2Url(`admin/work-groups/${id}`), payload)
		return data
	}

	async delete(id: number): Promise<void> {
		await this.http.delete(apiV2Url(`admin/work-groups/${id}`))
	}

	async getTaskGroups(taskId: number): Promise<IWorkGroup[]> {
		const {data} = await this.http.get<{groups: IWorkGroup[]}>(apiV2Url(`tasks/${taskId}/group-assignees`))
		return data.groups
	}

	async assignTaskGroup(taskId: number, groupId: number): Promise<IWorkGroupTaskAssignmentResult> {
		const {data} = await this.http.post<IWorkGroupTaskAssignmentResult>(apiV2Url(`tasks/${taskId}/group-assignees`), {group_id: groupId})
		return data
	}

	async unassignTaskGroup(taskId: number, groupId: number): Promise<void> {
		await this.http.delete(apiV2Url(`tasks/${taskId}/group-assignees/${groupId}`))
	}
}

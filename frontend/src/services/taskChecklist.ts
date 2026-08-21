import {apiV2Url, AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import type {ITaskChecklistState} from '@/modelTypes/ITaskChecklist'

export default class TaskChecklistService {
	private http = AuthenticatedHTTPFactory()

	async getAll(taskId: number): Promise<ITaskChecklistState> {
		const {data} = await this.http.get<ITaskChecklistState>(apiV2Url(`tasks/${taskId}/checklist-items`))
		return data
	}

	async create(taskId: number, title: string): Promise<ITaskChecklistState> {
		const {data} = await this.http.post<ITaskChecklistState>(apiV2Url(`tasks/${taskId}/checklist-items`), {title})
		return data
	}

	async update(taskId: number, itemId: number, title: string, done: boolean): Promise<ITaskChecklistState> {
		const {data} = await this.http.put<ITaskChecklistState>(apiV2Url(`tasks/${taskId}/checklist-items/${itemId}`), {title, done})
		return data
	}

	async delete(taskId: number, itemId: number): Promise<ITaskChecklistState> {
		const {data} = await this.http.delete<ITaskChecklistState>(apiV2Url(`tasks/${taskId}/checklist-items/${itemId}`))
		return data
	}
}

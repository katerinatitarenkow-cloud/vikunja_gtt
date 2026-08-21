import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import type {IClientActivityCreate, IClientActivityEvent, IClientActivityListResponse} from '@/modelTypes/IClientActivity'

export default class ClientActivityService {
	private http = AuthenticatedHTTPFactory()

	async getAll(projectId: number, page = 1, perPage = 50, type = ''): Promise<IClientActivityListResponse> {
		const {data} = await this.http.get<IClientActivityListResponse>(
			apiV2Url(`projects/${projectId}/client/history`),
			{params: {page, per_page: perPage, type: type || undefined}},
		)
		return data
	}

	async create(projectId: number, activity: IClientActivityCreate): Promise<IClientActivityEvent> {
		const {data} = await this.http.post<IClientActivityEvent>(
			apiV2Url(`projects/${projectId}/client/history`),
			activity,
		)
		return data
	}

	async delete(projectId: number, eventId: number): Promise<void> {
		await this.http.delete(apiV2Url(`projects/${projectId}/client/history/${eventId}`))
	}
}

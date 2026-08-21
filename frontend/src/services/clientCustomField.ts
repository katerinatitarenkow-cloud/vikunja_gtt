import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import type {IClientCustomField} from '@/modelTypes/IClientCustomField'

export default class ClientCustomFieldService {
	private http = AuthenticatedHTTPFactory()

	async getAll(projectId: number): Promise<IClientCustomField[]> {
		const {data} = await this.http.get<IClientCustomField[]>(apiV2Url(`projects/${projectId}/client/custom-fields`))
		return data
	}

	async create(projectId: number, name: string, value: string): Promise<IClientCustomField> {
		const {data} = await this.http.post<IClientCustomField>(apiV2Url(`projects/${projectId}/client/custom-fields`), {name, value})
		return data
	}

	async update(projectId: number, fieldId: number, name: string, value: string): Promise<IClientCustomField> {
		const {data} = await this.http.put<IClientCustomField>(apiV2Url(`projects/${projectId}/client/custom-fields/${fieldId}`), {name, value})
		return data
	}

	async delete(projectId: number, fieldId: number): Promise<void> {
		await this.http.delete(apiV2Url(`projects/${projectId}/client/custom-fields/${fieldId}`))
	}
}

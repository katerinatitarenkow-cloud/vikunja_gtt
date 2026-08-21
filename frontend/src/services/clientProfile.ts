import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import type {IClientGeocodeResult, IClientProfile} from '@/modelTypes/IClientProfile'
import type {IUser} from '@/modelTypes/IUser'

export default class ClientProfileService {
	private http = AuthenticatedHTTPFactory()

	async get(projectId: number): Promise<IClientProfile> {
		const {data} = await this.http.get<IClientProfile>(apiV2Url(`projects/${projectId}/client`))
		return data
	}

	async save(projectId: number, profile: IClientProfile): Promise<IClientProfile> {
		const {data} = await this.http.put<IClientProfile>(apiV2Url(`projects/${projectId}/client`), profile)
		return data
	}

	async searchProjectUsers(projectId: number, query = ''): Promise<IUser[]> {
		const {data} = await this.http.get<{items: IUser[]}>(
			apiV2Url(`projects/${projectId}/users/search`),
			{params: {q: query}},
		)
		return data.items ?? []
	}

	async geocode(projectId: number, query: string): Promise<IClientGeocodeResult> {
		const {data} = await this.http.get<IClientGeocodeResult>(
			apiV2Url(`projects/${projectId}/client/geocode`),
			{params: {q: query}},
		)
		return data
	}

	async uploadProposal(projectId: number, file: File): Promise<IClientProfile> {
		const formData = new FormData()
		formData.append('proposal', file, file.name)
		const {data} = await this.http.put<IClientProfile>(
			apiV2Url(`projects/${projectId}/client/proposal`),
			formData,
		)
		return data
	}

	async deleteProposal(projectId: number): Promise<IClientProfile> {
		const {data} = await this.http.delete<IClientProfile>(apiV2Url(`projects/${projectId}/client/proposal`))
		return data
	}

	async downloadProposal(projectId: number, filename: string): Promise<void> {
		const {data} = await this.http.get<Blob>(
			apiV2Url(`projects/${projectId}/client/proposal`),
			{responseType: 'blob'},
		)
		const objectUrl = window.URL.createObjectURL(data)
		const link = document.createElement('a')
		link.href = objectUrl
		link.download = filename
		document.body.appendChild(link)
		link.click()
		link.remove()
		window.URL.revokeObjectURL(objectUrl)
	}
}

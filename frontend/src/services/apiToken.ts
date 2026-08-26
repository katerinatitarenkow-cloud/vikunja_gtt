import AbstractService from '@/services/abstractService'
import type {IApiToken} from '@/modelTypes/IApiToken'
import ApiTokenModel from '@/models/apiTokenModel'

export type ApiTokenAvailableRoutes = Record<string, Record<string, unknown>>

export default class ApiTokenService extends AbstractService<IApiToken> {
	constructor() {
		super({
			create: '/tokens',
			getAll: '/tokens',
			delete: '/tokens/{id}',
		})
	}

	processModel(model: IApiToken) {
		return {
			...model,
			expiresAt: new Date(model.expiresAt).toISOString(),
			created: new Date(model.created).toISOString(),
		}
	}
	
	modelFactory(data: Partial<IApiToken>) {
		return new ApiTokenModel(data)
	}
	
	async getAvailableRoutes(): Promise<ApiTokenAvailableRoutes> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.get('/routes')
			return response.data as ApiTokenAvailableRoutes
		} finally {
			cancel()
		}
	}
}

import {apiV2Url, AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import type {IAdminWialonSettings, IAdminWialonTestResult, IWialonStatus, IWialonTrack, IWialonUnit} from '@/modelTypes/IWialon'

export default class WialonService {
	private http = AuthenticatedHTTPFactory()

	async getStatus(): Promise<IWialonStatus> {
		const response = await this.http.get<IWialonStatus>(apiV2Url('wialon/status'))
		return response.data
	}

	async getUnits(): Promise<IWialonUnit[]> {
		const response = await this.http.get<{units: IWialonUnit[]}>(apiV2Url('wialon/units'))
		return response.data.units ?? []
	}

	async getTrack(unitId: number, from: number, to: number): Promise<IWialonTrack> {
		const response = await this.http.get<IWialonTrack>(apiV2Url(`wialon/units/${unitId}/track`), {
			params: {from, to},
		})
		return response.data
	}

	async getAdminSettings(): Promise<IAdminWialonSettings> {
		const response = await this.http.get<IAdminWialonSettings>(apiV2Url('admin/wialon/settings'))
		return response.data
	}

	async saveAdminSettings(payload: Record<string, unknown>): Promise<IAdminWialonSettings> {
		const response = await this.http.patch<IAdminWialonSettings>(apiV2Url('admin/wialon/settings'), payload)
		return response.data
	}

	async testAdminConnection(): Promise<IAdminWialonTestResult> {
		const response = await this.http.post<IAdminWialonTestResult>(apiV2Url('admin/wialon/test'))
		return response.data
	}
}

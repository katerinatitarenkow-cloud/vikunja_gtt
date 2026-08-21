import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import type {IAccessGroup, IAccessMe, IAccessPermissionDefinition, IAccessUser} from '@/modelTypes/IAccessControl'

export interface AccessGroupPayload {
	name?: string
	description?: string
	permissions?: string[]
}

export interface AccessUserCreatePayload {
	username: string
	password: string
}

export interface AccessUserPatchPayload {
	name?: string
	email?: string
	phone?: string
	notes?: string
	group_ids?: number[]
	is_admin?: boolean
	status?: number
}

export default class AccessControlService {
	private http = AuthenticatedHTTPFactory()

	async getMe(): Promise<IAccessMe> {
		const {data} = await this.http.get<IAccessMe>(apiV2Url('access/me'))
		return data
	}

	async getPermissions(): Promise<IAccessPermissionDefinition[]> {
		const {data} = await this.http.get<{permissions: IAccessPermissionDefinition[]}>(apiV2Url('admin/access/permissions'))
		return data.permissions
	}

	async getGroups(): Promise<IAccessGroup[]> {
		const {data} = await this.http.get<{groups: IAccessGroup[]}>(apiV2Url('admin/access/groups'))
		return data.groups
	}

	async createGroup(payload: Required<Pick<AccessGroupPayload, 'name'>> & AccessGroupPayload): Promise<IAccessGroup> {
		const {data} = await this.http.post<IAccessGroup>(apiV2Url('admin/access/groups'), payload)
		return data
	}

	async updateGroup(id: number, payload: AccessGroupPayload): Promise<IAccessGroup> {
		const {data} = await this.http.patch<IAccessGroup>(apiV2Url(`admin/access/groups/${id}`), payload)
		return data
	}

	async deleteGroup(id: number): Promise<void> {
		await this.http.delete(apiV2Url(`admin/access/groups/${id}`))
	}

	async getUsers(search = ''): Promise<IAccessUser[]> {
		const {data} = await this.http.get<{users: IAccessUser[]}>(apiV2Url('admin/access/users'), {params: {search}})
		return data.users
	}

	async createUser(payload: AccessUserCreatePayload): Promise<IAccessUser> {
		const {data} = await this.http.post<IAccessUser>(apiV2Url('admin/access/users'), payload)
		return data
	}

	async updateUser(id: number, payload: AccessUserPatchPayload): Promise<IAccessUser> {
		const {data} = await this.http.patch<IAccessUser>(apiV2Url(`admin/access/users/${id}`), payload)
		return data
	}
}

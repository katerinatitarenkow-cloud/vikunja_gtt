import type {IUser} from '@/modelTypes/IUser'
import type {IFile} from '@/modelTypes/IFile'

export type ClientType = 'person' | 'fop' | 'company'
export type ClientStatus = 'potential' | 'active' | 'inactive' | 'vip'
export type ClientAddressType = 'legal' | 'actual' | 'postal' | 'delivery' | 'object'
export type ClientDecisionRole = '' | 'leader' | 'decision_maker' | 'technical' | 'procurement' | 'accountant' | 'user' | 'other'

export interface IClientAddress {
	id: number
	project_id: number
	type: ClientAddressType
	country: string
	region: string
	city: string
	street: string
	house: string
	office: string
	postal_code: string
	latitude: number
	longitude: number
}

export interface IClientContactPerson {
	id: number
	project_id: number
	full_name: string
	position: string
	department: string
	phone: string
	email: string
	telegram: string
	viber: string
	whatsapp: string
	birthday: string
	preferred_contact_method: string
	decision_role: ClientDecisionRole
	notes: string
	position_index: number
}

export interface IClientProfile {
	project_id: number
	client_type: ClientType
	display_name: string
	contact_name: string
	status: ClientStatus
	source: string
	responsible_user_id: number
	responsible?: IUser | null
	phone: string
	phone_secondary: string
	email: string
	email_secondary: string
	telegram: string
	viber: string
	whatsapp: string
	website: string
	preferred_contact_method: string
	preferred_language: string
	tax_id: string
	legal_name: string
	director_name: string
	ownership_form: string
	industry: string
	employee_count: number
	requisites: string
	iban: string
	bank: string
	mfo: string
	vat_number: string
	tax_system: string
	addresses: IClientAddress[]
	contact_persons: IClientContactPerson[]
	commercial_proposal?: IFile | null
	added_at: string | Date
	updated: string | Date
}

export interface IClientGeocodeResult {
	display_name: string
	postal_code: string
	latitude: number
	longitude: number
}

import type {IUser} from '@/modelTypes/IUser'

export type ClientActivityType =
	| 'call'
	| 'message'
	| 'meeting'
	| 'manual_note'
	| 'document_sent'
	| 'commercial_proposal_sent'
	| 'invoice_sent'
	| 'task_created'
	| 'task_completed'
	| 'task_reopened'
	| 'comment_created'
	| 'status_changed'
	| 'responsible_changed'
	| 'commercial_proposal_uploaded'
	| 'commercial_proposal_replaced'
	| 'commercial_proposal_deleted'
	| 'custom_field_created'
	| 'custom_field_updated'
	| 'custom_field_deleted'

export interface IClientActivityMetadata {
	direction?: string
	channel?: string
	duration_minutes?: number
	result?: string
	contact_person_id?: number
	contact_person_name?: string
	task_title?: string
	old_value?: string
	new_value?: string
	file_name?: string
	field_name?: string
	old_field_name?: string
	new_field_name?: string
}

export interface IClientActivityEvent {
	id: number
	project_id: number
	event_type: ClientActivityType
	actor_user_id: number
	actor?: IUser | null
	occurred_at: string | Date
	title: string
	description: string
	entity_type: string
	entity_id: number
	metadata?: IClientActivityMetadata | null
	system_generated: boolean
	created: string | Date
}

export interface IClientActivityCreate {
	event_type: Extract<ClientActivityType, 'call' | 'message' | 'meeting' | 'manual_note' | 'document_sent' | 'commercial_proposal_sent' | 'invoice_sent'>
	occurred_at: string
	title: string
	description: string
	entity_type?: string
	entity_id?: number
	metadata: IClientActivityMetadata
}

export interface IClientActivityListResponse {
	items: IClientActivityEvent[]
	total: number
	page: number
	per_page: number
}

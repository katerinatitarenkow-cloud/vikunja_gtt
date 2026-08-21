export interface IMailboxUser {
	id: number
	name: string
	username: string
}

export interface IMailboxMessage {
	id: number
	sender_id: number
	recipient_id: number
	reply_to_id: number
	subject: string
	body: string
	read_at: string | Date | null
	sender: IMailboxUser
	recipient: IMailboxUser
	created: string | Date
	updated: string | Date
}

export interface IMailboxPage {
	items: IMailboxMessage[]
	total: number
	page: number
	per_page: number
	total_pages: number
}

export interface IMailboxDraft {
	recipient_id: number
	reply_to_id: number
	subject: string
	body: string
}

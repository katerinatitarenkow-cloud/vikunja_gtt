export interface IMailboxUser {
id: number
name: string
username: string
}

export interface IMailboxFile {
id: number
name: string
mime: string
size: number
created: string | Date
}

export interface IMailboxAttachment {
id: number
message_id: number
file: IMailboxFile
created: string | Date
}

export interface IMailboxAttachmentUploadError {
message: string
}

export interface IMailboxAttachmentUploadResult {
success: IMailboxAttachment[]
errors: IMailboxAttachmentUploadError[]
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
attachments: IMailboxAttachment[]
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
forward_attachment_ids?: number[]
}
import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import type {
IMailboxAttachment,
IMailboxAttachmentUploadResult,
IMailboxDraft,
IMailboxMessage,
IMailboxPage,
IMailboxUser,
} from '@/modelTypes/IMailbox'

export default class MailboxService {
private http = AuthenticatedHTTPFactory()

async list(folder: 'inbox' | 'sent', page = 1, q = ''): Promise<IMailboxPage> {
const {data} = await this.http.get<IMailboxPage>(
apiV2Url('mailbox/messages'),
{
params: {
folder,
page,
per_page: 30,
q: q || undefined,
},
},
)

return data
}

async get(id: number): Promise<IMailboxMessage> {
const {data} = await this.http.get<IMailboxMessage>(
apiV2Url(`mailbox/messages/${id}`),
)

return data
}

async send(draft: IMailboxDraft): Promise<IMailboxMessage> {
const {data} = await this.http.post<IMailboxMessage>(
apiV2Url('mailbox/messages'),
draft,
)

return data
}

async uploadAttachments(
messageId: number,
files: File[],
): Promise<IMailboxAttachmentUploadResult> {
const form = new FormData()

for (const file of files) {
form.append('files', file)
}

const {data} =
await this.http.post<IMailboxAttachmentUploadResult>(
apiV2Url(
`mailbox/messages/${messageId}/attachments`,
),
form,
)

return data
}

async downloadAttachment(
messageId: number,
attachment: IMailboxAttachment,
): Promise<void> {
const {data} = await this.http.get<Blob>(
apiV2Url(
`mailbox/messages/${messageId}/attachments/${attachment.id}`,
),
{
responseType: 'blob',
},
)

const blob =
data instanceof Blob
? data
: new Blob([data])

const url = URL.createObjectURL(blob)
const link = document.createElement('a')

link.href = url
link.download =
attachment.file?.name ||
`attachment-${attachment.id}`

document.body.appendChild(link)
link.click()
link.remove()

URL.revokeObjectURL(url)
}

async setRead(
id: number,
read: boolean,
): Promise<IMailboxMessage> {
const {data} = await this.http.put<IMailboxMessage>(
apiV2Url(`mailbox/messages/${id}/read`),
{read},
)

return data
}

async delete(id: number): Promise<void> {
await this.http.delete(
apiV2Url(`mailbox/messages/${id}`),
)
}

async unreadCount(): Promise<number> {
const {data} = await this.http.get<{count: number}>(
apiV2Url('mailbox/unread-count'),
)

return data.count
}

async recipients(q = ''): Promise<IMailboxUser[]> {
const {data} = await this.http.get<IMailboxUser[]>(
apiV2Url('mailbox/recipients'),
{
params: {
q: q || undefined,
},
},
)

return data
}
}
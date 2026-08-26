import AbstractModel from './abstractModel'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import UserModel, {getDisplayName} from '@/models/user'
import TaskModel from '@/models/task'
import TaskCommentModel from '@/models/taskComment'
import ProjectModel from '@/models/project'

import {
NOTIFICATION_NAMES,
type INotification,
} from '@/modelTypes/INotification'
import type {IUser} from '@/modelTypes/IUser'

type RawNotification = Record<string, any>

export default class NotificationModel
extends AbstractModel<INotification>
implements INotification {

id = 0
name = ''
notification = {} as INotification['notification']
read = false
readAt: Date | null = null
created = new Date(0)

constructor(data: Partial<INotification>) {
super()

this.assignData(data)

const raw =
this.notification as unknown as RawNotification

switch (this.name) {
case NOTIFICATION_NAMES.TASK_COMMENT:
this.notification = {
doer: new UserModel(raw.doer),
task: new TaskModel(raw.task),
comment: new TaskCommentModel(raw.comment),
}
break

case NOTIFICATION_NAMES.TASK_ASSIGNED:
this.notification = {
doer: new UserModel(raw.doer),
task: new TaskModel(raw.task),
assignee: new UserModel(raw.assignee),
}
break

case NOTIFICATION_NAMES.TASK_DELETED:
this.notification = {
doer: new UserModel(raw.doer),
task: new TaskModel(raw.task),
}
break

case NOTIFICATION_NAMES.PROJECT_CREATED:
this.notification = {
doer: new UserModel(raw.doer),
project: new ProjectModel(raw.project),
}
break

case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
this.notification = {
doer: new UserModel(raw.doer),
member: new UserModel(raw.member),
team: raw.team,
}
break

case NOTIFICATION_NAMES.TASK_REMINDER:
this.notification = {
task: new TaskModel(raw.task),
project: new ProjectModel(raw.project),
}
break

case NOTIFICATION_NAMES.TASK_MENTIONED:
this.notification = {
doer: new UserModel(raw.doer),
task: new TaskModel(raw.task),
}
break

case NOTIFICATION_NAMES.CLIENT_RESPONSIBLE_ASSIGNED:
this.notification = {
doer: new UserModel(raw.doer),
project: new ProjectModel(raw.project),
responsible: new UserModel(raw.responsible),
}
break

case NOTIFICATION_NAMES.MAILBOX_MESSAGE_RECEIVED:
this.notification = {
doer: new UserModel(raw.doer),
messageId: Number(
raw.messageId ??
raw.message_id ??
0
),
subject: String(raw.subject ?? ''),
preview: String(raw.preview ?? ''),
}
break
}

this.created = new Date(this.created)
this.readAt = parseDateOrNull(this.readAt)
}

toText(user: {readonly id: number} | null = null) {
const notification =
this.notification as unknown as RawNotification

let who: string

switch (this.name) {
case NOTIFICATION_NAMES.TASK_COMMENT:
return `commented on ${notification.task.getTextIdentifier()}`

case NOTIFICATION_NAMES.TASK_ASSIGNED:
who = `${getDisplayName(notification.assignee)}`

if (
user !== null &&
user.id === notification.assignee.id
) {
who = 'you'
}

return `assigned ${who} to ${notification.task.getTextIdentifier()}`

case NOTIFICATION_NAMES.TASK_DELETED:
return `deleted ${notification.task.getTextIdentifier()}`

case NOTIFICATION_NAMES.PROJECT_CREATED:
return `created ${notification.project.title}`

case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
who = `${getDisplayName(notification.member)}`

if (
user !== null &&
user.id === notification.member.id
) {
who = 'you'
}

return `added ${who} to the ${notification.team.name} team`

case NOTIFICATION_NAMES.TASK_REMINDER:
return `Reminder for ${notification.task.getTextIdentifier()} ${notification.task.title} (${notification.project.title})`

case NOTIFICATION_NAMES.TASK_MENTIONED:
return `mentioned you on ${notification.task.getTextIdentifier()}`

case NOTIFICATION_NAMES.CLIENT_RESPONSIBLE_ASSIGNED:
return `assigned you as responsible for ${notification.project.title}`

case NOTIFICATION_NAMES.MAILBOX_MESSAGE_RECEIVED: {
const subject =
String(notification.subject ?? '').trim() ||
'(Без темы)'

const preview =
String(notification.preview ?? '').trim()

return preview
? `— новое письмо «${subject}»: ${preview}`
: `— новое письмо «${subject}»`
}
}

return ''
}
}
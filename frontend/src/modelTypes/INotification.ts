import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'
import type {ITaskComment} from './ITaskComment'
import type {ITeam} from './ITeam'
import type {IProject} from './IProject'

export const NOTIFICATION_NAMES = {
'TASK_COMMENT': 'task.comment',
'TASK_ASSIGNED': 'task.assigned',
'TASK_DELETED': 'task.deleted',
'TASK_REMINDER': 'task.reminder',
'PROJECT_CREATED': 'project.created',
'TEAM_MEMBER_ADDED': 'team.member.added',
'TASK_MENTIONED': 'task.mentioned',
'CLIENT_RESPONSIBLE_ASSIGNED': 'client.responsible.assigned',
'MAILBOX_MESSAGE_RECEIVED': 'mailbox.message.received',
} as const

interface Notification {
doer?: IUser
}

interface NotificationWithDoer extends Notification {
doer: IUser
}

interface NotificationTaskComment extends NotificationWithDoer {
task: ITask
comment: ITaskComment
}

interface NotificationTask extends NotificationWithDoer {
task: ITask
}

interface NotificationAssigned extends NotificationWithDoer {
task: ITask
assignee: IUser
}

interface NotificationCreated extends NotificationWithDoer {
project: IProject
}

interface NotificationTaskReminder extends Notification {
task: ITask
project: IProject
}

interface NotificationClientResponsible extends NotificationWithDoer {
project: IProject
responsible: IUser
}

interface NotificationMemberAdded extends NotificationWithDoer {
member: IUser
team: ITeam
}

interface NotificationMailboxMessageReceived extends NotificationWithDoer {
messageId: number
subject: string
preview: string
}

export type NotificationPayload =
| NotificationTaskComment
| NotificationTask
| NotificationAssigned
| NotificationCreated
| NotificationMemberAdded
| NotificationTaskReminder
| NotificationClientResponsible
| NotificationMailboxMessageReceived

export interface INotification extends IAbstract {
id: number
name: string
notification: NotificationPayload
read: boolean
readAt: Date | null
created: Date
}
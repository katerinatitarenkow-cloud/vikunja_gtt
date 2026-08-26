import {readonly, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'

import MailboxService from '@/services/mailbox'

export const useMailboxStore = defineStore('mailbox', () => {
const unreadCount = ref(0)
const service = new MailboxService()

async function refreshUnread() {
try {
unreadCount.value = await service.unreadCount()
} catch (e) {
console.warn('[MAILBOX] Failed to load unread count', e)
}

return unreadCount.value
}

function setUnreadCount(value: number) {
unreadCount.value = Math.max(0, value)
}

return {
unreadCount: readonly(unreadCount),
refreshUnread,
setUnreadCount,
}
})

if (import.meta.hot) {
import.meta.hot.accept(
acceptHMRUpdate(useMailboxStore, import.meta.hot),
)
}
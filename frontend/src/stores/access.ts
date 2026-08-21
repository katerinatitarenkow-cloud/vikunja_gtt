import {computed, ref} from 'vue'
import {defineStore} from 'pinia'
import AccessControlService from '@/services/accessControl'
import type {IAccessGroup} from '@/modelTypes/IAccessControl'
import {useAuthStore} from '@/stores/auth'

export const useAccessStore = defineStore('access', () => {
	const permissions = ref<string[]>([])
	const groups = ref<IAccessGroup[]>([])
	const loadedForUser = ref<number | null>(null)
	const loading = ref(false)

	const permissionSet = computed(() => new Set(permissions.value))

	function can(permission: string): boolean {
		const authStore = useAuthStore()
		if (authStore.info?.isAdmin) return true
		return permissionSet.value.has(permission)
	}

	async function refresh(): Promise<void> {
		const authStore = useAuthStore()
		if (!authStore.authUser || !authStore.info?.id) {
			permissions.value = []
			groups.value = []
			loadedForUser.value = null
			return
		}
		loading.value = true
		try {
			const data = await new AccessControlService().getMe()
			permissions.value = data.permissions ?? []
			groups.value = data.groups ?? []
			loadedForUser.value = authStore.info.id
		} finally {
			loading.value = false
		}
	}

	async function ensureLoaded(): Promise<void> {
		const authStore = useAuthStore()
		if (!authStore.authUser || !authStore.info?.id) return
		if (loadedForUser.value === authStore.info.id) return
		await refresh()
	}

	function reset() {
		permissions.value = []
		groups.value = []
		loadedForUser.value = null
	}

	return {permissions, groups, loading, can, refresh, ensureLoaded, reset}
})

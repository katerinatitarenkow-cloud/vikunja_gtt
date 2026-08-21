<template>
	<Card
		:title="$t('wialon.settings.title')"
		:loading="loading"
	>
		<p class="mb-2">
			{{ $t('wialon.settings.description') }}
		</p>
		<p class="help mb-4">
			{{ $t('wialon.settings.authHelp') }}
		</p>

		<FormCheckbox
			v-model="form.enabled"
			:label="$t('wialon.settings.enabled')"
		/>

		<FormField
			:label="$t('wialon.settings.apiUrl')"
			layout="two-col"
		>
			<FormInput
				v-model="form.apiUrl"
				type="url"
				placeholder="https://hst-api.wialon.com"
			/>
		</FormField>

		<FormField
			:label="$t('wialon.settings.token')"
			layout="two-col"
		>
			<div>
				<FormInput
					v-model="form.token"
					type="password"
					:placeholder="tokenPlaceholder"
					autocomplete="new-password"
				/>
				<p class="help mt-1">
					{{ $t('wialon.settings.tokenHelp') }}
				</p>
				<XButton
					class="mt-2 mr-2"
					variant="secondary"
					:loading="receivingToken"
					@click="openWialonTokenWindow"
				>
					{{ $t('wialon.settings.getHostingToken') }}
				</XButton>
				<XButton
					v-if="tokenConfigured"
					class="mt-2"
					variant="secondary"
					@click="clearToken"
				>
					{{ $t('wialon.settings.clearToken') }}
				</XButton>
			</div>
		</FormField>

		<FormField
			:label="$t('wialon.settings.timeout')"
			layout="two-col"
		>
			<FormInput
				v-model.number="form.timeoutSeconds"
				type="number"
				min="1"
				max="300"
			/>
		</FormField>

		<FormField
			:label="$t('wialon.settings.maxPoints')"
			layout="two-col"
		>
			<FormInput
				v-model.number="form.trackMaxPoints"
				type="number"
				min="100"
				max="50000"
			/>
		</FormField>

		<div class="wialon-settings__actions mt-4">
			<XButton
				variant="primary"
				:loading="saving"
				@click="save"
			>
				{{ $t('misc.save') }}
			</XButton>
			<XButton
				variant="secondary"
				:loading="testing"
				@click="testConnection"
			>
				{{ $t('wialon.settings.testConnection') }}
			</XButton>
		</div>

		<div
			v-if="testResult"
			class="notification is-success mt-4"
		>
			{{ $t('wialon.settings.testSuccess', {count: testResult.unit_count}) }}
		</div>
	</Card>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, onMounted, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'
import {success, error} from '@/message'
import WialonService from '@/services/wialon'
import type {IAdminWialonTestResult} from '@/modelTypes/IWialon'

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('wialon.settings.title')} - ${t('user.settings.title')}`)

const service = new WialonService()
const WIALON_HOSTING_ORIGIN = 'https://hosting.wialon.com'
const WIALON_HOSTING_API_URL = 'https://hst-api.wialon.com'
const WIALON_POST_TOKEN_URL = `${WIALON_HOSTING_ORIGIN}/post_token.html`
let wialonAuthWindow: Window | null = null
let wialonPopupTimer: number | null = null
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const receivingToken = ref(false)
const tokenConfigured = ref(false)
const clearStoredToken = ref(false)
const testResult = ref<IAdminWialonTestResult | null>(null)
const form = reactive({
	enabled: false,
	apiUrl: 'https://hst-api.wialon.com',
	token: '',
	timeoutSeconds: 30,
	trackMaxPoints: 5000,
})

const tokenPlaceholder = computed(() => tokenConfigured.value
	? t('wialon.settings.tokenConfigured')
	: t('wialon.settings.tokenNotConfigured'))

function buildWialonHostingTokenUrl(): string {
	const params = new URLSearchParams({
		client_id: 'Vikunja',
		access_type: '256',
		activation_time: '0',
		duration: '0',
		lang: 'ru',
		flags: '1',
		response_type: 'token',
		redirect_uri: WIALON_POST_TOKEN_URL,
	})
	return `${WIALON_HOSTING_ORIGIN}/login.html?${params.toString()}`
}

function openWialonTokenWindow() {
	testResult.value = null
	receivingToken.value = true
	wialonAuthWindow = window.open(
		buildWialonHostingTokenUrl(),
		'wialon-token',
		'width=760,height=560,resizable=yes,scrollbars=yes',
	)

	if (!wialonAuthWindow) {
		receivingToken.value = false
		error({message: t('wialon.settings.popupBlocked')})
		return
	}

	wialonPopupTimer = window.setInterval(() => {
		if (!wialonAuthWindow || wialonAuthWindow.closed) {
			receivingToken.value = false
			wialonAuthWindow = null
			if (wialonPopupTimer !== null) window.clearInterval(wialonPopupTimer)
			wialonPopupTimer = null
		}
	}, 500)
}

async function handleWialonTokenMessage(event: MessageEvent) {
	// Wialon's official post_token.html callback sends the token back to
	// the window that opened the authorization form. Accept messages only
	// from Wialon Hosting and, when available, only from our auth popup.
	if (event.origin !== WIALON_HOSTING_ORIGIN) return
	if (wialonAuthWindow && event.source !== wialonAuthWindow) return
	if (typeof event.data !== 'string' || !event.data.startsWith('access_token=')) return

	const raw = event.data.slice('access_token='.length).split('&', 1)[0]
	const token = decodeURIComponent(raw).trim()
	if (!token) return

	receivingToken.value = true
	try {
		// This button is specifically for Wialon Hosting, so make the
		// corresponding API endpoint explicit and enable the integration.
		form.apiUrl = WIALON_HOSTING_API_URL
		form.token = token
		form.enabled = true
		clearStoredToken.value = false

		await persistSettings()
		testResult.value = await service.testAdminConnection()
		success({message: t('wialon.settings.tokenReceived')})
		wialonAuthWindow?.close()
		wialonAuthWindow = null
		if (wialonPopupTimer !== null) window.clearInterval(wialonPopupTimer)
		wialonPopupTimer = null
	} catch (e) {
		error(e)
	} finally {
		receivingToken.value = false
	}
}

async function load() {
	loading.value = true
	try {
		const settings = await service.getAdminSettings()
		form.enabled = settings.enabled
		form.apiUrl = settings.api_url
		form.timeoutSeconds = settings.timeout_seconds
		form.trackMaxPoints = settings.track_max_points
		tokenConfigured.value = settings.token_configured
		form.token = ''
		clearStoredToken.value = false
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

function clearToken() {
	form.token = ''
	clearStoredToken.value = true
	tokenConfigured.value = false
}

async function persistSettings() {
	const payload: Record<string, unknown> = {
		enabled: form.enabled,
		api_url: form.apiUrl,
		timeout_seconds: form.timeoutSeconds,
		track_max_points: form.trackMaxPoints,
		clear_token: clearStoredToken.value,
	}
	if (form.token.trim()) payload.token = form.token.trim()
	const settings = await service.saveAdminSettings(payload)
	tokenConfigured.value = settings.token_configured
	clearStoredToken.value = false
	form.token = ''
	return settings
}

async function save() {
	saving.value = true
	testResult.value = null
	try {
		await persistSettings()
		success({message: t('wialon.settings.saved')})
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}

async function testConnection() {
	testing.value = true
	testResult.value = null
	try {
		// Persist first so the test always uses exactly what is shown in the form.
		await persistSettings()
		testResult.value = await service.testAdminConnection()
	} catch (e) {
		error(e)
	} finally {
		testing.value = false
	}
}

onMounted(() => {
	window.addEventListener('message', handleWialonTokenMessage)
	void load()
})

onBeforeUnmount(() => {
	window.removeEventListener('message', handleWialonTokenMessage)
	if (wialonAuthWindow && !wialonAuthWindow.closed) wialonAuthWindow.close()
	wialonAuthWindow = null
	if (wialonPopupTimer !== null) window.clearInterval(wialonPopupTimer)
	wialonPopupTimer = null
})
</script>

<style scoped>
.wialon-settings__actions {
	display: flex;
	gap: .75rem;
	flex-wrap: wrap;
}
</style>

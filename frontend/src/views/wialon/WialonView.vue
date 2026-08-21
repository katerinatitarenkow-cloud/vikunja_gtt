<template>
	<div class="wialon-view">
		<div class="wialon-header">
			<div>
				<h1>{{ $t('wialon.title') }}</h1>
				<p class="wialon-subtitle">{{ $t('wialon.subtitle') }}</p>
			</div>
			<XButton
				icon="sync-alt"
				:loading="loadingUnits || loadingTrack"
				:disabled="!ready"
				@click="refreshAll"
			>
				{{ $t('wialon.refresh') }}
			</XButton>
		</div>

		<div
			v-if="loadingStatus"
			class="wialon-state"
		>
			<Loading variant="small" />
		</div>

		<div
			v-else-if="statusError"
			class="notification is-danger"
		>
			<strong>{{ $t('wialon.connectionError') }}</strong>
			<div>{{ statusError }}</div>
		</div>

		<div
			v-else-if="status && (!status.enabled || !status.configured)"
			class="notification is-warning"
		>
			<strong>{{ status.enabled ? $t('wialon.notConfigured') : $t('wialon.disabled') }}</strong>
			<p>{{ isInstanceAdmin ? $t('wialon.setupAdminHint') : $t('wialon.setupUserHint') }}</p>
			<XButton
				v-if="isInstanceAdmin"
				class="mt-3"
				variant="primary"
				:to="{name: 'user.settings.wialon'}"
			>
				{{ $t('wialon.openSettings') }}
			</XButton>
		</div>

		<template v-else-if="ready">
			<div
				v-if="dataError"
				class="notification is-danger"
			>
				<strong>{{ $t('wialon.connectionError') }}</strong>
				<div>{{ dataError }}</div>
			</div>

			<div class="wialon-toolbar">
				<div class="field">
					<label class="label" for="wialon-period">{{ $t('wialon.period') }}</label>
					<div class="select">
						<select
							id="wialon-period"
							v-model.number="periodHours"
							@change="loadSelectedTrack"
						>
							<option :value="6">{{ $t('wialon.period6h') }}</option>
							<option :value="12">{{ $t('wialon.period12h') }}</option>
							<option :value="24">{{ $t('wialon.period24h') }}</option>
							<option :value="48">{{ $t('wialon.period48h') }}</option>
							<option :value="168">{{ $t('wialon.period7d') }}</option>
						</select>
					</div>
				</div>
				<div class="wialon-summary">
					<span>{{ $t('wialon.unitsCount', {count: units.length}) }}</span>
					<span>{{ $t('wialon.onlineCount', {count: onlineCount}) }}</span>
					<span>{{ $t('wialon.autoRefresh') }}</span>
				</div>
			</div>

			<div class="wialon-layout">
				<aside class="vehicle-panel">
					<div class="vehicle-panel-header">
						<h2>{{ $t('wialon.vehicles') }}</h2>
						<input
							v-model.trim="search"
							class="input is-small"
							type="search"
							:placeholder="$t('wialon.searchVehicles')"
						>
					</div>

					<div
						v-if="loadingUnits && units.length === 0"
						class="vehicle-loading"
					>
						<Loading variant="small" />
					</div>
					<div
						v-else-if="filteredUnits.length === 0"
						class="vehicle-empty"
					>
						{{ $t('wialon.noUnits') }}
					</div>
					<button
						v-for="unit in filteredUnits"
						:key="unit.id"
						type="button"
						class="vehicle-row"
						:class="{'is-selected': unit.id === selectedUnitId}"
						@click="selectUnit(unit.id)"
					>
						<span
							class="connection-dot"
							:class="{'is-online': unit.connected}"
						/>
						<span class="vehicle-row-content">
							<strong>{{ unit.name }}</strong>
							<span v-if="unit.position">
								{{ $t('wialon.speed', {speed: unit.position.speed}) }}
								· {{ formatTime(unit.position.time) }}
							</span>
							<span v-else>{{ $t('wialon.noPosition') }}</span>
						</span>
					</button>
				</aside>

				<section class="map-panel">
					<WialonMap
						:units="units"
						:track="track"
						:selected-unit-id="selectedUnitId"
						@select-unit="selectUnit"
					/>

					<div
						v-if="selectedUnit"
						class="route-card"
					>
						<div class="route-card-title">
							<div>
								<strong>{{ selectedUnit.name }}</strong>
								<span :class="selectedUnit.connected ? 'has-text-success' : 'has-text-grey'">
									{{ selectedUnit.connected ? $t('wialon.online') : $t('wialon.offline') }}
								</span>
							</div>
							<span v-if="loadingTrack">{{ $t('wialon.loadingTrack') }}</span>
						</div>
						<div
							v-if="track && track.points.length"
							class="route-stats"
						>
							<span>{{ $t('wialon.trackPoints', {count: track.original_point_count}) }}</span>
							<span>{{ $t('wialon.distance', {distance: trackDistanceKm}) }}</span>
							<span>{{ formatTime(track.from) }} → {{ formatTime(track.to) }}</span>
						</div>
						<div
							v-else-if="!loadingTrack"
							class="has-text-grey"
						>
							{{ $t('wialon.noTrack') }}
						</div>
					</div>
				</section>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, onMounted, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import Loading from '@/components/misc/Loading.vue'
import WialonMap from '@/components/wialon/WialonMap.vue'
import type {IWialonStatus, IWialonTrack, IWialonUnit} from '@/modelTypes/IWialon'
import WialonService from '@/services/wialon'
import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('wialon.title'))

const service = new WialonService()
const authStore = useAuthStore()
const status = ref<IWialonStatus | null>(null)
const units = ref<IWialonUnit[]>([])
const track = ref<IWialonTrack | null>(null)
const selectedUnitId = ref<number | null>(null)
const periodHours = ref(24)
const search = ref('')
const loadingStatus = ref(true)
const loadingUnits = ref(false)
const loadingTrack = ref(false)
const statusError = ref('')
const dataError = ref('')
let refreshTimer: number | undefined

const ready = computed(() => Boolean(status.value?.enabled && status.value?.configured))
const isInstanceAdmin = computed(() => Boolean(authStore.info?.isAdmin))
const selectedUnit = computed(() => units.value.find(unit => unit.id === selectedUnitId.value) ?? null)
const onlineCount = computed(() => units.value.filter(unit => unit.connected).length)
const filteredUnits = computed(() => {
	const needle = search.value.toLocaleLowerCase()
	if (!needle) return units.value
	return units.value.filter(unit => unit.name.toLocaleLowerCase().includes(needle))
})
const trackDistanceKm = computed(() => {
	if (!track.value || track.value.points.length < 2) return '0.0'
	let meters = 0
	for (let i = 1; i < track.value.points.length; i++) {
		meters += haversineMeters(track.value.points[i - 1], track.value.points[i])
	}
	return (meters / 1000).toLocaleString(undefined, {maximumFractionDigits: 1})
})

function formatTime(unix: number) {
	if (!unix) return '—'
	return new Date(unix * 1000).toLocaleString([], {
		day: '2-digit',
		month: '2-digit',
		hour: '2-digit',
		minute: '2-digit',
	})
}

function haversineMeters(a: {latitude: number, longitude: number}, b: {latitude: number, longitude: number}) {
	const radius = 6371000
	const toRadians = (value: number) => value * Math.PI / 180
	const dLat = toRadians(b.latitude - a.latitude)
	const dLon = toRadians(b.longitude - a.longitude)
	const lat1 = toRadians(a.latitude)
	const lat2 = toRadians(b.latitude)
	const h = Math.sin(dLat / 2) ** 2 + Math.cos(lat1) * Math.cos(lat2) * Math.sin(dLon / 2) ** 2
	return 2 * radius * Math.asin(Math.min(1, Math.sqrt(h)))
}

function errorMessage(error: unknown) {
	if (typeof error === 'object' && error !== null) {
		const candidate = error as {message?: string, response?: {data?: {detail?: string, title?: string}}}
		return candidate.response?.data?.detail || candidate.response?.data?.title || candidate.message || t('wialon.unknownError')
	}
	return t('wialon.unknownError')
}

async function loadStatus() {
	loadingStatus.value = true
	statusError.value = ''
	try {
		status.value = await service.getStatus()
	} catch (error) {
		statusError.value = errorMessage(error)
	} finally {
		loadingStatus.value = false
	}
}

async function loadUnits() {
	if (!ready.value) return
	loadingUnits.value = true
	dataError.value = ''
	try {
		units.value = await service.getUnits()
		if (selectedUnitId.value !== null && !units.value.some(unit => unit.id === selectedUnitId.value)) {
			selectedUnitId.value = null
			track.value = null
		}
	} catch (error) {
		dataError.value = errorMessage(error)
	} finally {
		loadingUnits.value = false
	}
}

async function loadSelectedTrack() {
	if (!ready.value || selectedUnitId.value === null) {
		track.value = null
		return
	}
	loadingTrack.value = true
	dataError.value = ''
	try {
		const to = Math.floor(Date.now() / 1000)
		const from = to - periodHours.value * 60 * 60
		track.value = await service.getTrack(selectedUnitId.value, from, to)
	} catch (error) {
		track.value = null
		dataError.value = errorMessage(error)
	} finally {
		loadingTrack.value = false
	}
}

async function selectUnit(unitId: number) {
	selectedUnitId.value = unitId
	await loadSelectedTrack()
}

async function refreshAll() {
	await loadUnits()
	if (selectedUnitId.value !== null) {
		await loadSelectedTrack()
	}
}

onMounted(async () => {
	await loadStatus()
	if (ready.value) {
		await loadUnits()
	}
	refreshTimer = window.setInterval(() => {
		if (ready.value && !loadingUnits.value && !loadingTrack.value) {
			void refreshAll()
		}
	}, 60_000)
})

onBeforeUnmount(() => {
	if (refreshTimer !== undefined) window.clearInterval(refreshTimer)
})
</script>

<style lang="scss" scoped>
.wialon-view {
	padding: .25rem;
	min-block-size: calc(100vh - #{$navbar-height});
}

.wialon-header,
.wialon-toolbar,
.route-card-title,
.route-stats {
	display: flex;
	align-items: center;
	gap: 1rem;
}

.wialon-header {
	justify-content: space-between;
	margin-block-end: 1rem;
	padding: .2rem .25rem;

	h1 {
		margin-block-end: .18rem;
		font-size: 1.55rem;
		font-weight: 750;
		letter-spacing: -.025em;
		color: var(--text-strong);
	}
}

.wialon-subtitle { color: var(--brand-text-muted); }
.wialon-state { display: grid; place-items: center; min-block-size: 240px; }

.wialon-toolbar {
	justify-content: space-between;
	align-items: end;
	margin-block-end: 1rem;
	padding: .85rem 1rem;
	background: var(--white);
	border: 1px solid var(--brand-border);
	border-radius: 13px;
	box-shadow: 0 5px 18px rgba(31, 91, 73, .05);
	.field { margin-block-end: 0; }
}

.wialon-summary {
	display: flex;
	gap: .55rem;
	flex-wrap: wrap;
	color: var(--brand-text-muted);
	font-size: .82rem;

	span {
		padding: .34rem .55rem;
		background: var(--brand-surface-soft);
		border-radius: 8px;
	}
}

.wialon-layout {
	display: grid;
	grid-template-columns: minmax(260px, 320px) minmax(0, 1fr);
	gap: 1rem;
	min-block-size: 620px;
}

.vehicle-panel,
.map-panel {
	background: var(--white);
	border: 1px solid var(--brand-border);
	border-radius: 14px;
	box-shadow: 0 8px 28px rgba(31, 91, 73, .07);
	overflow: hidden;
}

.vehicle-panel { display: flex; flex-direction: column; max-block-size: 720px; overflow-y: auto; }
.vehicle-panel-header {
	position: sticky;
	inset-block-start: 0;
	z-index: 2;
	padding: 1rem;
	background: color-mix(in srgb, var(--white) 94%, transparent);
	backdrop-filter: blur(12px);
	border-block-end: 1px solid var(--brand-border);
	h2 { font-size: .95rem; font-weight: 750; margin-block-end: .65rem; color: var(--text-strong); }
}

.vehicle-row {
	display: grid;
	grid-template-columns: 10px minmax(0, 1fr);
	gap: .65rem;
	align-items: start;
	inline-size: 100%;
	padding: .82rem 1rem;
	border: 0;
	border-block-end: 1px solid var(--brand-border);
	background: transparent;
	color: inherit;
	text-align: start;
	cursor: pointer;
	transition: background-color 120ms ease;
	&:hover { background: var(--brand-surface-soft); }
	&.is-selected {
		background: var(--brand-lime-soft);
		box-shadow: inset 3px 0 0 var(--brand-forest);
	}
}

.connection-dot {
	inline-size: 9px;
	block-size: 9px;
	margin-block-start: .35rem;
	border-radius: 50%;
	background: var(--grey-400);
	box-shadow: 0 0 0 3px rgba(0, 0, 0, .03);
	&.is-online { background: var(--success); }
}

.vehicle-row-content {
	display: flex;
	min-inline-size: 0;
	flex-direction: column;
	gap: .15rem;
	strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-strong); }
	span { color: var(--brand-text-muted); font-size: .78rem; }
}

.vehicle-loading, .vehicle-empty { padding: 2rem 1rem; text-align: center; color: var(--brand-text-muted); }
.map-panel { display: flex; flex-direction: column; min-inline-size: 0; }
.map-panel :deep(.wialon-map) { flex: 1 1 auto; min-block-size: 520px; }
.route-card { padding: .9rem 1rem; border-block-start: 1px solid var(--brand-border); }
.route-card-title { justify-content: space-between; > div { display: flex; align-items: center; gap: .6rem; } }
.route-stats { margin-block-start: .45rem; flex-wrap: wrap; color: var(--brand-text-muted); font-size: .82rem; }

@media screen and (max-width: $tablet) {
	.wialon-view { padding: 0; }
	.wialon-header, .wialon-toolbar { align-items: stretch; flex-direction: column; }
	.wialon-layout { grid-template-columns: 1fr; }
	.vehicle-panel { max-block-size: 300px; }
}
</style>

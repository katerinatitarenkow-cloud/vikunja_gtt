<template>
	<div
		ref="container"
		class="wialon-map"
		@pointerdown="startPan"
		@pointermove="movePan"
		@pointerup="endPan"
		@pointercancel="endPan"
		@wheel.prevent="onWheel"
	>
		<div class="tile-layer">
			<img
				v-for="tile in tiles"
				:key="tile.key"
				class="map-tile"
				:src="tile.url"
				:style="{left: `${tile.left}px`, top: `${tile.top}px`}"
				alt=""
				draggable="false"
			>
		</div>

		<svg
			class="route-layer"
			:width="size.width"
			:height="size.height"
			aria-hidden="true"
		>
			<polyline
				v-if="trackSvgPoints"
				:points="trackSvgPoints"
				class="route-shadow"
				fill="none"
			/>
			<polyline
				v-if="trackSvgPoints"
				:points="trackSvgPoints"
				class="route-line"
				fill="none"
			/>
		</svg>

		<button
			v-for="marker in markers"
			:key="marker.unit.id"
			type="button"
			class="unit-marker"
			:class="{'is-selected': marker.unit.id === selectedUnitId, 'is-online': marker.unit.connected}"
			:style="{left: `${marker.x}px`, top: `${marker.y}px`}"
			:title="marker.unit.name"
			@pointerdown.stop
			@click.stop="$emit('select-unit', marker.unit.id)"
		>
			<span class="marker-dot" />
			<span class="marker-label">{{ marker.unit.name }}</span>
		</button>

		<div class="map-controls" @pointerdown.stop>
			<button type="button" class="map-control" :aria-label="$t('wialon.zoomIn')" @click="zoomBy(1)">+</button>
			<button type="button" class="map-control" :aria-label="$t('wialon.zoomOut')" @click="zoomBy(-1)">−</button>
			<button type="button" class="map-control fit-control" :aria-label="$t('wialon.fitMap')" @click="fitToData">◎</button>
		</div>

		<div class="map-attribution">
			© <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener noreferrer">OpenStreetMap</a>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch} from 'vue'
import type {IWialonTrack, IWialonUnit} from '@/modelTypes/IWialon'

const props = defineProps<{
	units: IWialonUnit[]
	track: IWialonTrack | null
	selectedUnitId: number | null
}>()

defineEmits<{
	'select-unit': [unitId: number]
}>()

const TILE_SIZE = 256
const MAX_LAT = 85.05112878
const container = ref<HTMLElement | null>(null)
const size = reactive({width: 0, height: 0})
const center = reactive({latitude: 49.0, longitude: 32.0})
const zoom = ref(6)
let resizeObserver: ResizeObserver | null = null
let dragState: {x: number, y: number, centerX: number, centerY: number, pointerId: number} | null = null

function clampLatitude(latitude: number) {
	return Math.max(-MAX_LAT, Math.min(MAX_LAT, latitude))
}

function project(latitude: number, longitude: number, z = zoom.value) {
	const scale = TILE_SIZE * 2 ** z
	const lat = clampLatitude(latitude) * Math.PI / 180
	return {
		x: (longitude + 180) / 360 * scale,
		y: (0.5 - Math.log((1 + Math.sin(lat)) / (1 - Math.sin(lat))) / (4 * Math.PI)) * scale,
	}
}

function unproject(x: number, y: number, z = zoom.value) {
	const scale = TILE_SIZE * 2 ** z
	const longitude = x / scale * 360 - 180
	const n = Math.PI - 2 * Math.PI * y / scale
	const latitude = 180 / Math.PI * Math.atan(Math.sinh(n))
	return {latitude: clampLatitude(latitude), longitude}
}

const centerWorld = computed(() => project(center.latitude, center.longitude))

const tiles = computed(() => {
	if (!size.width || !size.height) return []
	const count = 2 ** zoom.value
	const startX = Math.floor((centerWorld.value.x - size.width / 2) / TILE_SIZE) - 1
	const endX = Math.floor((centerWorld.value.x + size.width / 2) / TILE_SIZE) + 1
	const startY = Math.max(0, Math.floor((centerWorld.value.y - size.height / 2) / TILE_SIZE) - 1)
	const endY = Math.min(count - 1, Math.floor((centerWorld.value.y + size.height / 2) / TILE_SIZE) + 1)
	const result: Array<{key: string, url: string, left: number, top: number}> = []
	for (let x = startX; x <= endX; x++) {
		const wrappedX = ((x % count) + count) % count
		for (let y = startY; y <= endY; y++) {
			result.push({
				key: `${zoom.value}/${x}/${y}`,
				url: `https://tile.openstreetmap.org/${zoom.value}/${wrappedX}/${y}.png`,
				left: x * TILE_SIZE - centerWorld.value.x + size.width / 2,
				top: y * TILE_SIZE - centerWorld.value.y + size.height / 2,
			})
		}
	}
	return result
})

function toScreen(latitude: number, longitude: number) {
	const p = project(latitude, longitude)
	return {
		x: p.x - centerWorld.value.x + size.width / 2,
		y: p.y - centerWorld.value.y + size.height / 2,
	}
}

const markers = computed(() => props.units
	.filter(unit => unit.position)
	.map(unit => ({unit, ...toScreen(unit.position!.latitude, unit.position!.longitude)})))

const trackSvgPoints = computed(() => {
	if (!props.track?.points?.length) return ''
	return props.track.points
		.map(point => {
			const p = toScreen(point.latitude, point.longitude)
			return `${p.x.toFixed(1)},${p.y.toFixed(1)}`
		})
		.join(' ')
})

function zoomBy(delta: number) {
	zoom.value = Math.max(2, Math.min(18, zoom.value + delta))
}

function onWheel(event: WheelEvent) {
	zoomBy(event.deltaY < 0 ? 1 : -1)
}

function startPan(event: PointerEvent) {
	if (event.button !== 0 || !container.value) return
	container.value.setPointerCapture(event.pointerId)
	dragState = {
		x: event.clientX,
		y: event.clientY,
		centerX: centerWorld.value.x,
		centerY: centerWorld.value.y,
		pointerId: event.pointerId,
	}
	container.value.classList.add('is-dragging')
}

function movePan(event: PointerEvent) {
	if (!dragState || dragState.pointerId !== event.pointerId) return
	const next = unproject(
		dragState.centerX - (event.clientX - dragState.x),
		dragState.centerY - (event.clientY - dragState.y),
	)
	center.latitude = next.latitude
	center.longitude = next.longitude
}

function endPan(event: PointerEvent) {
	if (!dragState || dragState.pointerId !== event.pointerId) return
	dragState = null
	container.value?.classList.remove('is-dragging')
}

function fitToData() {
	const coordinates = props.track?.points?.length
		? props.track.points.map(point => ({latitude: point.latitude, longitude: point.longitude}))
		: props.units.filter(unit => unit.position).map(unit => ({
			latitude: unit.position!.latitude,
			longitude: unit.position!.longitude,
		}))
	if (!coordinates.length || !size.width || !size.height) return

	if (coordinates.length === 1) {
		center.latitude = coordinates[0].latitude
		center.longitude = coordinates[0].longitude
		zoom.value = 14
		return
	}

	const minLat = Math.min(...coordinates.map(p => p.latitude))
	const maxLat = Math.max(...coordinates.map(p => p.latitude))
	const minLon = Math.min(...coordinates.map(p => p.longitude))
	const maxLon = Math.max(...coordinates.map(p => p.longitude))
	center.latitude = (minLat + maxLat) / 2
	center.longitude = (minLon + maxLon) / 2

	const padding = 72
	for (let candidate = 18; candidate >= 2; candidate--) {
		const a = project(minLat, minLon, candidate)
		const b = project(maxLat, maxLon, candidate)
		if (Math.abs(b.x - a.x) <= Math.max(100, size.width - padding * 2) &&
			Math.abs(b.y - a.y) <= Math.max(100, size.height - padding * 2)) {
			zoom.value = candidate
			return
		}
	}
	zoom.value = 2
}

onMounted(() => {
	if (!container.value) return
	resizeObserver = new ResizeObserver(entries => {
		const rect = entries[0]?.contentRect
		if (!rect) return
		size.width = rect.width
		size.height = rect.height
	})
	resizeObserver.observe(container.value)
	nextTick(fitToData)
})

onBeforeUnmount(() => resizeObserver?.disconnect())

watch(
	() => [props.track?.unit_id, props.track?.from, props.track?.to, props.track?.points?.length, props.units.length],
	() => nextTick(fitToData),
)

defineExpose({fitToData})
</script>

<style lang="scss" scoped>
.wialon-map {
	position: relative;
	inline-size: 100%;
	block-size: 100%;
	min-block-size: 420px;
	overflow: hidden;
	background: #dde4e8;
	cursor: grab;
	user-select: none;
	touch-action: none;

	&.is-dragging {
		cursor: grabbing;
	}
}

.tile-layer,
.route-layer {
	position: absolute;
	inset: 0;
}

.map-tile {
	position: absolute;
	inline-size: 256px;
	block-size: 256px;
	max-inline-size: none;
	pointer-events: none;
}

.route-layer {
	pointer-events: none;
	overflow: visible;
}

.route-shadow,
.route-line {
	stroke-linecap: round;
	stroke-linejoin: round;
	vector-effect: non-scaling-stroke;
}

.route-shadow {
	stroke: rgb(255 255 255 / 90%);
	stroke-width: 8;
}

.route-line {
	stroke: var(--primary);
	stroke-width: 4;
}

.unit-marker {
	position: absolute;
	transform: translate(-8px, -8px);
	border: 0;
	background: transparent;
	padding: 0;
	cursor: pointer;
	z-index: 4;
	text-align: start;

	.marker-dot {
		display: block;
		inline-size: 16px;
		block-size: 16px;
		border-radius: 50%;
		background: #7a7a7a;
		border: 3px solid white;
		box-shadow: 0 1px 5px rgb(0 0 0 / 45%);
	}

	&.is-online .marker-dot {
		background: #23d160;
	}

	&.is-selected .marker-dot {
		inline-size: 20px;
		block-size: 20px;
		transform: translate(-2px, -2px);
		background: var(--primary);
	}
}

.marker-label {
	display: block;
	margin-block-start: 3px;
	padding: 2px 6px;
	border-radius: 4px;
	background: rgb(255 255 255 / 92%);
	color: #363636;
	font-size: .72rem;
	font-weight: 600;
	white-space: nowrap;
	box-shadow: 0 1px 4px rgb(0 0 0 / 18%);
}

.map-controls {
	position: absolute;
	inset-block-start: 12px;
	inset-inline-end: 12px;
	display: flex;
	flex-direction: column;
	z-index: 6;
	box-shadow: 0 1px 5px rgb(0 0 0 / 22%);
}

.map-control {
	inline-size: 36px;
	block-size: 36px;
	border: 0;
	border-block-end: 1px solid #ddd;
	background: white;
	font-size: 1.25rem;
	cursor: pointer;

	&:last-child {
		border-block-end: 0;
	}

	&:hover {
		background: #f4f4f4;
	}
}

.fit-control {
	font-size: 1rem;
}

.map-attribution {
	position: absolute;
	inset-inline-end: 4px;
	inset-block-end: 2px;
	z-index: 6;
	padding: 1px 4px;
	background: rgb(255 255 255 / 80%);
	font-size: .65rem;
	color: #555;
}
</style>

<template>
	<div
		class="card"
		:class="{'has-no-shadow': !shadow}"
	>
		<header
			v-if="title !== ''"
			class="card-header"
		>
			<p class="card-header-title">
				{{ title }}
			</p>
			<BaseButton
				v-if="showClose"
				class="card-header-icon close"
				:aria-label="$t('misc.close')"
				@click="$emit('close')"
			>	
				<span class="icon">
					<Icon icon="times" />
				</span>
			</BaseButton>
		</header>
		<div
			class="card-content loader-container"
			:class="{
				'p-0': !padding,
				'is-loading': loading
			}"
		>
			<div :class="{'content': hasContent}">
				<slot />
			</div>
		</div>

		<footer
			v-if="$slots.footer"
			class="card-footer"
		>
			<slot name="footer" />
		</footer>
	</div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'

withDefaults(defineProps<{
	title?: string
	padding?: boolean
	shadow?: boolean
	hasContent?: boolean
	loading?: boolean
	showClose?: boolean
}>(), {
	title: '',
	padding: true,
	shadow: true,
	hasContent: true,
	loading: false,
	showClose: false,
})

defineEmits<{
	'close': []
}>()
</script>

<style lang="scss" scoped>
.card {
	background-color: var(--white);
	border-radius: $radius-large;
	margin-block-end: 1rem;
	border: 1px solid var(--brand-border);
	box-shadow: 0 8px 26px rgba(31, 91, 73, .07);
	color: var(--text);
	max-inline-size: 100%;
	position: relative;
	overflow: clip;

	@media print {
		box-shadow: none;
		border: none;
	}
}

.card-header {
	background-color: transparent;
	align-items: stretch;
	display: flex;
	box-shadow: none;
	border-block-end: 1px solid var(--brand-border);
}

.card-header-title {
	align-items: center;
	color: var(--text-strong);
	display: flex;
	flex-grow: 1;
	font-weight: 700;
	letter-spacing: -.015em;
	padding: .9rem 1.1rem;

	&.is-centered { justify-content: center; }
}

.card-header-icon {
	align-items: center;
	cursor: pointer;
	display: flex;
	justify-content: center;
	padding: .75rem 1rem;
	color: var(--brand-text-muted);

	&:hover {
		background: var(--brand-surface-soft);
		color: var(--brand-forest);
	}
}

.card-content {
	background-color: transparent;
	padding: 1.25rem;

	&:first-child {
		border-start-start-radius: $radius-large;
		border-start-end-radius: $radius-large;
	}
	&:last-child {
		border-end-start-radius: $radius-large;
		border-end-end-radius: $radius-large;
	}
	&.p-0 { padding: 0; }
}

.card-footer {
	align-items: stretch;
	background-color: color-mix(in srgb, var(--brand-surface-soft) 60%, var(--white));
	border-block-start: 1px solid var(--brand-border);
	padding: 1rem 1.25rem;
	display: flex;
	justify-content: flex-end;
}
</style>

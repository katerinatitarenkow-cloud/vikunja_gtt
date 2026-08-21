<template>
	<BaseButton
		class="button"
		:class="[
			variantClass,
			{
				'is-loading': loading,
				'has-no-shadow': !shadow || variant === 'tertiary',
				'is-danger': danger,
			}
		]"
		:disabled="disabled || loading"
		:style="{
			'--button-white-space': wrap ? 'break-spaces' : 'nowrap',
		}"
	>
		<template v-if="icon">
			<Icon
				v-if="!$slots.default"
				:icon="icon"
				:style="{color: iconColor}"
			/>
			<span
				v-else
				class="icon is-small"
			>
				<Icon
					:icon="icon"
					:style="{color: iconColor}"
				/>
			</span>
		</template>
		<span>
			<slot />
		</span>
	</BaseButton>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

const props = defineProps<ButtonProps>()

const VARIANT_CLASS_MAP = {
	primary: 'is-primary',
	secondary: 'is-outlined',
	tertiary: 'is-text is-inverted underline-none',
} as const

export type ButtonTypes = keyof typeof VARIANT_CLASS_MAP

export interface ButtonProps {
	variant?: ButtonTypes
	icon?: IconProp
	iconColor?: string
	loading?: boolean
	disabled?: boolean
	shadow?: boolean
	wrap?: boolean
	danger?: boolean
}

defineOptions({name: 'XButton'})

// @ts-expect-error - Complex union type from IconProp causes TS2590, but the code is correct
const variant = computed(() => (props.variant ?? 'primary') as ButtonTypes)
const shadow = computed(() => (props.shadow ?? true) as boolean)
const wrap = computed(() => (props.wrap ?? true) as boolean)
const variantClass = computed<string>(() => VARIANT_CLASS_MAP[variant.value])
</script>

<style lang="scss" scoped>
.button {
	--button-text-color: #ffffff;

	display: inline-flex;
	align-items: center;
	justify-content: center;
	vertical-align: top;
	cursor: pointer;
	text-align: center;
	white-space: var(--button-white-space);
	transition: background-color 140ms ease, border-color 140ms ease, color 140ms ease, box-shadow 140ms ease, transform 140ms ease;
	border: 1px solid transparent;
	font-size: .84rem;
	font-weight: 700;
	letter-spacing: .005em;
	block-size: auto;
	min-block-size: $button-height;
	box-shadow: 0 4px 12px rgba(31, 91, 73, .13);
	line-height: 1.15;
	padding: .58rem .9rem;
	gap: .4rem;
	background-color: var(--primary);
	color: var(--button-text-color);
	border-radius: $radius;

	[dir="rtl"] & {
		flex-direction: row-reverse;
	}

	&:hover {
		box-shadow: 0 7px 18px rgba(31, 91, 73, .16);
		background-color: var(--primary-dark);
		transform: translateY(-1px);
	}

	&:focus,
	&:focus-visible {
		outline: none;
		box-shadow: var(--focus-ring), 0 4px 12px rgba(31, 91, 73, .13);
	}

	&.is-active,
	&.is-focused,
	&:active,
	&:focus:not(:active) {
		transform: translateY(0);
	}

	&[disabled] {
		opacity: .48;
		cursor: not-allowed;
		pointer-events: none;
		box-shadow: none;
		transform: none;
	}

	.icon {
		margin: 0 !important;
	}

	&.is-primary {
		background-color: var(--primary);
		color: var(--button-text-color);

		&:hover {
			background-color: var(--primary-dark);
		}
	}

	&.is-outlined {
		background-color: var(--white);
		border-color: var(--brand-border);
		color: var(--text-strong);
		box-shadow: 0 2px 7px rgba(31, 91, 73, .06);

		&:hover {
			border-color: color-mix(in srgb, var(--brand-forest) 38%, var(--brand-border));
			background: var(--brand-surface-soft);
			color: var(--brand-forest);
		}
	}

	&.is-text {
		background-color: transparent;
		border-color: transparent;
		color: var(--text);
		box-shadow: none;

		&:hover {
			background-color: var(--brand-surface-soft);
			color: var(--brand-forest);
			box-shadow: none;
			transform: none;
		}
	}

	&.is-inverted {
		color: inherit;
	}

	&.is-danger {
		background-color: var(--danger);
		border-color: transparent;
		color: var(--button-text-color);

		&:hover {
			background-color: var(--danger-dark);
		}
	}

	&.is-danger.is-outlined {
		background-color: transparent;
		border: 1px solid var(--danger);
		color: var(--danger-text);

		&:hover,
		&:focus {
			background-color: var(--danger);
			border-color: var(--danger);
			color: var(--button-text-color);
		}
	}

	&.is-danger.is-text {
		background-color: transparent;
		color: var(--danger-text);

		&:hover {
			background-color: hsla(var(--danger-h), var(--danger-s), var(--danger-l), .1);
		}
	}

	&.is-danger.is-loading::after {
		border-color: transparent transparent #fff #fff;
	}

	&.is-danger.is-outlined.is-loading::after,
	&.is-danger.is-text.is-loading::after {
		border-color: transparent transparent var(--danger) var(--danger);
	}

	&.is-loading {
		color: transparent !important;
		pointer-events: none;
		position: relative;

		&::after {
			content: "";
			position: absolute;
			display: block;
			block-size: 1em;
			inline-size: 1em;
			border: 2px solid var(--button-text-color);
			border-radius: 50%;
			border-inline-end-color: transparent;
			border-block-start-color: transparent;
			animation: spin-around 500ms infinite linear;
			inset-inline-start: calc(50% - .5em);
			inset-block-start: calc(50% - .5em);
		}
	}

	&.is-outlined.is-loading::after,
	&.is-text.is-loading::after {
		border-color: var(--brand-forest);
		border-inline-end-color: transparent;
		border-block-start-color: transparent;
	}
}

@keyframes spin-around {
	from { transform: rotate(0deg); }
	to { transform: rotate(360deg); }
}

.is-small {
	border-radius: $radius-small;
}

.underline-none {
	text-decoration: none !important;
}
</style>

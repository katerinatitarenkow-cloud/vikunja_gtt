<template>
	<div
		class="project-card"
		:class="{
			'has-light-text': background !== null,
			'has-background': blurHashUrl !== '' || background !== null,
			'is-due-soon': Boolean(props.dueDate) && !project.isCompleted
		}"
		:style="{
			'border-inline-start': project.hexColor ? `0.25rem solid ${project.hexColor}` : undefined,
			'background-image': blurHashUrl !== '' ? `url(${blurHashUrl})` : undefined,
		}"
	>
		<div
			class="project-background background-fade-in"
			:class="{'is-visible': background}"
			:style="{'background-image': background !== null ? `url(${background})` : undefined}"
		/>
		<span
			v-if="project.isArchived"
			class="is-archived"
		>{{ $t('project.archived') }}</span>

		<div
			class="project-title"
			aria-hidden="true"
		>
			<span
				v-if="project.id < -1"
				class="saved-filter-icon icon"
			>
				<Icon icon="filter" />
			</span>
			<span
v-if="project.isCompleted"
class="completed-badge"
>
{{ '\u0412\u044b\u043f\u043e\u043b\u043d\u0435\u043d' }}
</span>
{{ getProjectTitle(project) }}
		</div>

<div
v-if="props.dueDate && !project.isCompleted"
class="due-warning"
>
нужно выполнить задачу до {{ dueDateText }}
</div>
		<BaseButton
			class="project-button"
			:aria-label="project.title"
			:title="textOnlyDescription"
			:to="{
				name: 'project.index',
				params: { projectId: project.id}
			}"
		/>
<BaseButton
v-if="canManageProjects && !project.isArchived && project.id > 0"
class="complete-toggle"
:aria-label="project.isCompleted ? '\u0412\u0435\u0440\u043d\u0443\u0442\u044c \u0432 \u0440\u0430\u0431\u043e\u0442\u0443' : '\u041e\u0442\u043c\u0435\u0442\u0438\u0442\u044c \u0432\u044b\u043f\u043e\u043b\u043d\u0435\u043d\u043d\u044b\u043c'"
:title="project.isCompleted ? '\u0412\u0435\u0440\u043d\u0443\u0442\u044c \u0432 \u0440\u0430\u0431\u043e\u0442\u0443' : '\u041e\u0442\u043c\u0435\u0442\u0438\u0442\u044c \u0432\u044b\u043f\u043e\u043b\u043d\u0435\u043d\u043d\u044b\u043c'"
@click.prevent.stop="toggleCompleted"
>
{{ project.isCompleted ? '\u21B6' : '\u2713' }}
</BaseButton>
		<BaseButton
			v-if="!project.isArchived && project.id > -1"
			class="favorite"
			:aria-label="project.isFavorite ? $t('project.unfavorite') : $t('project.favorite')"
			:class="{'is-favorite': project.isFavorite}"
			@click.prevent.stop="projectStore.toggleProjectFavorite(project)"
		>
			<Icon :icon="project.isFavorite ? 'star' : ['far', 'star']" />
		</BaseButton>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import type {IProject} from '@/modelTypes/IProject'

import BaseButton from '@/components/base/BaseButton.vue'

import {useProjectBackground} from '@/composables/useProjectBackground'
import {useProjectStore} from '@/stores/projects'
import {useAccessStore} from '@/stores/access'
import {ACCESS_PERMISSION} from '@/modelTypes/IAccessControl'
import {getProjectTitle} from '@/helpers/getProjectTitle'

const props = defineProps<{
	project: IProject,
	 dueDate?: string | null,
}>()

const {background, blurHashUrl} = useProjectBackground(() => props.project)

const projectStore = useProjectStore()
const accessStore = useAccessStore()

const canManageProjects = computed(() =>
accessStore.can(ACCESS_PERMISSION.PROJECTS_MANAGE),
)

async function toggleCompleted() {
await projectStore.toggleProjectCompleted(props.project)
}

const dueDateText = computed(() => {
if (!props.dueDate) {
return ''
}

const date = new Date(props.dueDate)

return new Intl.DateTimeFormat('ru-RU', {
day: '2-digit',
month: '2-digit',
year: 'numeric',
}).format(date)
})
const textOnlyDescription = computed(() => {
	return props.project.description ? props.project.description.replace(/<[^>]*>/g, '') : ''
})
</script>

<style lang="scss" scoped>
.project-card {
	--project-card-padding: 1rem;
	background: var(--white);
	padding: var(--project-card-padding);
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	transition: box-shadow $transition;
	position: relative;
	overflow: hidden; // hide background

	display: flex;
	justify-content: space-between;
	flex-wrap: wrap;

	&:hover {
		box-shadow: var(--shadow-md);
	}

	&:active,
	&:focus {
		box-shadow: var(--shadow-xs) !important;
	}

	> * {
		// so the elements are on top of the background
		position: relative;
	}
}

.has-background,
.project-background {
	background-size: cover;
	background-repeat: no-repeat;
	background-position: center;
}

.project-background,
.project-button {
	position: absolute;
	inset-block-start: 0;
	inset-inline-end: 0;
	inset-block-end: 0;
	inset-inline-start: 0;
}

.is-archived {
	font-size: .75rem;
	float: inline-start;
}

.project-title {
	align-self: flex-end;
	font-family: $vikunja-font;
	font-weight: 400;
	font-size: 1.5rem;
	line-height: var(--title-line-height);
	color: var(--text);
	inline-size: 100%;
	margin-block-end: 0;
	max-block-size: calc(100% - (var(--project-card-padding) + 1rem)); // padding & height of the "is archived" badge
	overflow: hidden;
	text-overflow: ellipsis;
	word-break: break-word;

	display: -webkit-box;
	-webkit-line-clamp: 3;
	-webkit-box-orient: vertical;
}

.has-light-text .project-title {
	color: var(--grey-100);
}

.has-background .project-title {
	text-shadow:
		0 0 10px var(--black),
		1px 1px 5px var(--grey-700),
		-1px -1px 5px var(--grey-700);
	color: var(--white);
}

.project-card.is-due-soon {
padding-block-end: 3rem;
background: #ffe4e4 !important;
border: 2px solid var(--danger);
box-shadow: 0 0 0 1px rgba(220, 40, 40, .08), var(--shadow-sm);

.project-background {
opacity: .08 !important;
}
}

.due-warning {
position: absolute;
inset-inline-start: var(--project-card-padding);
inset-inline-end: var(--project-card-padding);
inset-block-end: .65rem;
z-index: 3;

font-size: .72rem;
font-weight: 600;
line-height: 1.2;
color: var(--danger);
}
.completed-badge {
display: inline-block;
font-size: .7rem;
font-weight: 600;
padding: .25rem .45rem;
margin-inline-end: .4rem;
border-radius: 6px;
background: var(--success);
color: var(--white);
}

.complete-toggle {
position: absolute;
inset-block-start: var(--project-card-padding);
inset-inline-start: var(--project-card-padding);
z-index: 4;
min-inline-size: 2rem;
min-block-size: 2rem;
border-radius: 50%;
background: var(--white);
box-shadow: var(--shadow-sm);
font-weight: 700;
}
.favorite {
	position: absolute;
	inset-block-start: var(--project-card-padding);
	inset-inline-end: var(--project-card-padding);
	transition: opacity $transition, color $transition;
	opacity: 1;

	&:hover {
		color: var(--warning);
	}

	&.is-favorite {
		display: inline-block;
		opacity: 1;
		color: var(--warning);
	}
}

@media(hover: hover) and (pointer: fine) {
	.project-card .favorite {
		opacity: 0;
	}

	.project-card:hover .favorite {
		opacity: 1;
	}
}

.background-fade-in {
  opacity: 0;
  transition: opacity $transition;
  transition-delay: $transition-duration * 2; // To fake an appearing background

  &.is-visible {
    opacity: 1;
  }
}

.saved-filter-icon {
	color: var(--grey-300);
	font-size: .75em;
}
</style>



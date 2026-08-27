<template>
<div class="google-calendar-callback">
<div class="google-calendar-callback__card">
<h1>Google Calendar</h1>

<div
v-if="processing"
class="google-calendar-callback__state"
>
<strong>Подключаем календарь…</strong>

<p>
Подтверждаем разрешение Google и сохраняем подключение.
</p>
</div>

<div
v-else-if="success"
class="google-calendar-callback__state"
>
<strong>Google Calendar подключён</strong>

<p>
Возвращаем вас на рабочую панель.
</p>
</div>

<div
v-else
class="google-calendar-callback__state google-calendar-callback__state--error"
>
<strong>Не удалось подключить Google Calendar</strong>

<p>
{{ errorMessage }}
</p>

<XButton
variant="primary"
@click="returnHome"
>
Вернуться в обзор
</XButton>
</div>
</div>
</div>
</template>

<script setup lang="ts">
import {
onMounted,
ref,
} from 'vue'
import {
useRoute,
useRouter,
} from 'vue-router'

import XButton from '@/components/input/Button.vue'

import GoogleCalendarService from '@/services/googleCalendar'

const route = useRoute()
const router = useRouter()

const processing = ref(true)
const success = ref(false)
const errorMessage = ref('')

function queryString(
value: unknown,
): string {
if (typeof value === 'string') {
return value
}

if (
Array.isArray(value) &&
typeof value[0] === 'string'
) {
return value[0]
}

return ''
}

async function finishOAuth() {
const googleError =
queryString(route.query.error)

if (googleError !== '') {
const description =
queryString(
route.query.error_description,
)

errorMessage.value =
description ||
`Google вернул ошибку: ${googleError}`

processing.value = false

return
}

const code =
queryString(route.query.code)

const state =
queryString(route.query.state)

if (code === '' || state === '') {
errorMessage.value =
'Google не вернул код авторизации или параметр state.'

processing.value = false

return
}

try {
const service =
new GoogleCalendarService()

await service.callback({
code,
state,
})

success.value = true

window.setTimeout(() => {
void router.replace({
name: 'home',
query: {
googleCalendar:
'connected',
},
})
}, 700)
} catch {
errorMessage.value =
'Сервер не смог завершить подключение Google Calendar.'

processing.value = false
}
}

function returnHome() {
void router.replace({
name: 'home',
})
}

onMounted(() => {
void finishOAuth()
})
</script>

<style scoped lang="scss">
.google-calendar-callback {
display: flex;
align-items: center;
justify-content: center;
min-block-size: calc(100vh - 120px);
padding: 2rem;
}

.google-calendar-callback__card {
inline-size: min(520px, 100%);
padding: 2rem;
border: 1px solid var(--brand-border);
border-radius: 8px;
background: var(--white);
text-align: center;

h1 {
margin: 0 0 1.25rem;
color: var(--text-strong);
}
}

.google-calendar-callback__state {
color: var(--grey-600);

strong {
display: block;
margin-block-end: .45rem;
color: var(--text-strong);
}

p {
margin: 0 0 1rem;
}
}

.google-calendar-callback__state--error {
strong {
color: var(--danger);
}
}
</style>
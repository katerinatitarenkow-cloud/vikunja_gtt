<template>
<div class="calendar-widget">
<div
v-if="loading"
class="calendar-widget__state"
>
Загрузка состояния календаря…
</div>

<div
v-else-if="loadError"
class="calendar-widget__state"
>
<strong>
Не удалось получить состояние Google Calendar
</strong>

<XButton
variant="secondary"
@click="loadStatus"
>
Повторить
</XButton>
</div>

<div
v-else-if="!status.enabled"
class="calendar-widget__state"
>
<div class="calendar-widget__google">
G
</div>

<strong>
Google Calendar ещё не настроен
</strong>

<p>
Администратору нужно настроить Google OAuth для этого сервера.
</p>
</div>

<div
v-else-if="!status.connected"
class="calendar-widget__state"
>
<div class="calendar-widget__google">
G
</div>

<strong>
Подключите свой Google Calendar
</strong>

<p>
Личные события Google и рабочие события Vikunja будут показаны вместе.
</p>

<XButton
variant="primary"
:disabled="connecting"
@click="connect"
>
{{
connecting
? 'Подключение…'
: 'Подключить Google Calendar'
}}
</XButton>
</div>

<div
v-else
class="calendar-widget__connected"
>
<div class="calendar-widget__connected-header">
<div>
<div class="calendar-widget__status">
<span class="calendar-widget__status-dot" />

Google Calendar подключён
</div>

<div
v-if="status.googleEmail"
class="calendar-widget__account"
>
{{ status.googleEmail }}
</div>
</div>

<button
type="button"
class="calendar-widget__disconnect"
:disabled="disconnecting"
@click="disconnect"
>
{{
disconnecting
? 'Отключение…'
: 'Отключить'
}}
</button>
</div>

<div class="calendar-widget__calendar-placeholder">
<strong>
Календарь готов к синхронизации
</strong>

<p>
Следующим этапом загрузим события Google и задачи Vikunja.
</p>
</div>
</div>
</div>
</template>

<script setup lang="ts">
import {
onMounted,
reactive,
ref,
} from 'vue'

import XButton from '@/components/input/Button.vue'

import GoogleCalendarService from '@/services/googleCalendar'

import type {
IGoogleCalendarStatus,
} from '@/modelTypes/IGoogleCalendar'

const emptyStatus = (): IGoogleCalendarStatus => ({
enabled: false,
connected: false,
})

const status =
reactive<IGoogleCalendarStatus>(
emptyStatus(),
)

const loading = ref(true)
const loadError = ref(false)
const connecting = ref(false)
const disconnecting = ref(false)

async function loadStatus() {
loading.value = true
loadError.value = false

try {
const service =
new GoogleCalendarService()

const loaded =
await service.status()

Object.assign(
status,
emptyStatus(),
loaded,
)
} catch {
loadError.value = true
} finally {
loading.value = false
}
}

async function connect() {
connecting.value = true

try {
const service =
new GoogleCalendarService()

const result =
await service.connect()

if (result.url) {
window.location.assign(
result.url,
)
}
} catch {
loadError.value = true
connecting.value = false
}
}

async function disconnect() {
if (
!window.confirm(
'Отключить Google Calendar от вашей учётной записи Vikunja?',
)
) {
return
}

disconnecting.value = true

try {
const service =
new GoogleCalendarService()

await service.disconnect()

Object.assign(
status,
emptyStatus(),
{
enabled: true,
},
)
} catch {
loadError.value = true
} finally {
disconnecting.value = false
}
}

onMounted(() => {
void loadStatus()
})
</script>

<style scoped lang="scss">
.calendar-widget {
min-block-size: 180px;
}

.calendar-widget__state {
display: flex;
flex-direction: column;
align-items: center;
justify-content: center;
min-block-size: 180px;
padding: 1rem;
text-align: center;
color: var(--grey-600);

strong {
color: var(--text-strong);
}

p {
max-inline-size: 420px;
margin: .4rem 0 .9rem;
font-size: .82rem;
}
}

.calendar-widget__google {
display: flex;
align-items: center;
justify-content: center;
inline-size: 42px;
block-size: 42px;
margin-block-end: .65rem;
border: 1px solid var(--brand-border);
border-radius: 8px;
background: var(--white);
color: var(--text-strong);
font-size: 1.2rem;
font-weight: 750;
}

.calendar-widget__connected {
display: flex;
flex-direction: column;
gap: 1rem;
}

.calendar-widget__connected-header {
display: flex;
align-items: flex-start;
justify-content: space-between;
gap: 1rem;
}

.calendar-widget__status {
display: flex;
align-items: center;
gap: .4rem;
color: var(--text-strong);
font-size: .82rem;
font-weight: 650;
}

.calendar-widget__status-dot {
display: block;
inline-size: 8px;
block-size: 8px;
border-radius: 999px;
background: var(--success);
}

.calendar-widget__account {
margin-block-start: .2rem;
color: var(--grey-600);
font-size: .75rem;
}

.calendar-widget__disconnect {
padding: .25rem .4rem;
border: 0;
background: transparent;
color: var(--danger);
cursor: pointer;
font-size: .76rem;

&:disabled {
opacity: .5;
cursor: default;
}
}

.calendar-widget__calendar-placeholder {
display: flex;
flex-direction: column;
align-items: center;
justify-content: center;
min-block-size: 125px;
padding: 1rem;
border: 1px dashed var(--brand-border);
border-radius: 6px;
color: var(--grey-600);
text-align: center;

strong {
color: var(--text-strong);
font-size: .85rem;
}

p {
margin: .3rem 0 0;
font-size: .78rem;
}
}
</style>
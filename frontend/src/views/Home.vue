<template>
<div class="content dashboard-home">
<Message
v-if="deletionScheduledAt !== null"
variant="danger"
class="mbe-4"
>
{{
$t('user.deletion.scheduled', {
date: formatDisplayDate(deletionScheduledAt),
dateSince: formatDateSince(deletionScheduledAt),
})
}}

<RouterLink :to="{name: 'user.settings.deletion'}">
{{ $t('user.deletion.scheduledCancel') }}
</RouterLink>
</Message>

<DashboardView />
</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import Message from '@/components/misc/Message.vue'
import DashboardView from '@/components/home/dashboard/DashboardView.vue'

import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {
formatDateSince,
formatDisplayDate,
} from '@/helpers/time/formatDate'

import {useAuthStore} from '@/stores/auth'

const authStore = useAuthStore()

const deletionScheduledAt = computed(() =>
parseDateOrNull(
authStore.info?.deletionScheduledAt,
),
)
</script>

<style scoped lang="scss">
.dashboard-home {
max-inline-size: 1480px;
margin-inline: auto;
}
</style>
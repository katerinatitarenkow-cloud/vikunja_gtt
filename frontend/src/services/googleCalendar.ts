import {
AuthenticatedHTTPFactory,
apiV2Url,
} from '@/helpers/fetcher'

import type {
IGoogleCalendarCallback,
IGoogleCalendarConnect,
IGoogleCalendarStatus,
} from '@/modelTypes/IGoogleCalendar'

export default class GoogleCalendarService {
private http = AuthenticatedHTTPFactory()

async status(): Promise<IGoogleCalendarStatus> {
const {data} =
await this.http.get<IGoogleCalendarStatus>(
apiV2Url(
'integrations/google-calendar/status',
),
)

return data
}

async connect(): Promise<IGoogleCalendarConnect> {
const {data} =
await this.http.post<IGoogleCalendarConnect>(
apiV2Url(
'integrations/google-calendar/connect',
),
{},
)

return data
}

async callback(
payload: IGoogleCalendarCallback,
): Promise<IGoogleCalendarStatus> {
const {data} =
await this.http.post<IGoogleCalendarStatus>(
apiV2Url(
'integrations/google-calendar/callback',
),
payload,
)

return data
}

async disconnect(): Promise<void> {
await this.http.delete(
apiV2Url(
'integrations/google-calendar',
),
)
}
}
export interface IGoogleCalendarStatus {
enabled: boolean
connected: boolean
googleEmail?: string
vikunjaCalendarId?: string
connectedAt?: string
}

export interface IGoogleCalendarConnect {
url: string
}

export interface IGoogleCalendarCallback {
code: string
state: string
}
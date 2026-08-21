export interface IWialonPosition {
	time: number
	latitude: number
	longitude: number
	altitude: number
	speed: number
	course: number
	satellites: number
}

export interface IWialonUnit {
	id: number
	name: string
	position?: IWialonPosition
	connected: boolean
	last_update?: number
}

export interface IWialonTrackPoint {
	time: number
	latitude: number
	longitude: number
	speed: number
	course: number
}

export interface IWialonTrack {
	unit_id: number
	from: number
	to: number
	points: IWialonTrackPoint[]
	original_point_count: number
}

export interface IWialonStatus {
	enabled: boolean
	configured: boolean
	api_url: string
}

export interface IAdminWialonSettings {
	enabled: boolean
	api_url: string
	token_configured: boolean
	timeout_seconds: number
	track_max_points: number
}

export interface IAdminWialonTestResult {
	ok: boolean
	unit_count: number
	message: string
}

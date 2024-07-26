package v2rayprobe

type SpeedTestMode string
const (
	PING_ONLY  SpeedTestMode = "pingonly"
	SPEED_ONLY SpeedTestMode = "speedonly"
	FULL_TEST  SpeedTestMode = "all"
)

type PingMethod string
const (
	GOOGLE_PING PingMethod = "googleping"
)

type SortMethod string
const (
	SPEED_DESC SortMethod = "speed"
	SPEED_ASC  SortMethod = "rspeed"
	PING_DESC  SortMethod = "ping"
	PING_ASC   SortMethod = "rping"
)

type ConcurrencyOpt int
const (
	AUTO = -1
)

type ProbeMethod int 
const (
	Speed = iota
	Ping
	Full
)

type OutputMode int
const (
	PIC_BASE64 = iota
	PIC_PATH
	PIC_NONE
	JSON_OUTPUT
	TEXT_OUTPUT
)
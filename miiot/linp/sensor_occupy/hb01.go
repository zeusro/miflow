package sensor_occupy

const Model = "linp.sensor_occupy.hb01"

// Occupancy Sensor siid=2
const (
	SiidOccupancy = 2
	PiidStatus    = 1 // Occupancy Status
	PiidNoOneTime = 2
	PiidHasSomeone = 3
	PiidNoOneDur   = 4
	PiidIllumination = 5
)

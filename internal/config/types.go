// types.go
package config

type Environment string

const (
	Dev  Environment = "dev"
	Prod Environment = "prod"
)

type RotationPolicy string

const (
	Monthly RotationPolicy = "monthly"
	Weekly  RotationPolicy = "weekly"
	Daily   RotationPolicy = "daily"
)

type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

type CollectionName string

const (
	AppUpdate   CollectionName = "appUpdate"

)

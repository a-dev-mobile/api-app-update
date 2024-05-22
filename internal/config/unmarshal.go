package config

import (
	"fmt"
	"strings"
)

// UnmarshalYAML custom unmarshalling for Environment.
func (e *Environment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var envStr string
	if err := unmarshal(&envStr); err != nil {
		return err
	}

	switch Environment(envStr) {
	case Dev, Prod:
		*e = Environment(envStr)
		return nil
	default:
		return fmt.Errorf("invalid environment: %s", envStr)
	}
}

// UnmarshalYAML customizes the unmarshalling for LogLevel.
func (l *LogLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var levelStr string
	if err := unmarshal(&levelStr); err != nil {
		return err
	}

	levelStr = strings.ToLower(levelStr)
	switch LogLevel(levelStr) {
	case LogLevelDebug, LogLevelInfo, LogLevelWarning, LogLevelError:
		*l = LogLevel(levelStr)
		return nil
	default:
		return fmt.Errorf("invalid log level: %s", levelStr)
	}
}

// UnmarshalYAML custom unmarshalling for RotationPolicy.
func (r *RotationPolicy) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var policyStr string
	if err := unmarshal(&policyStr); err != nil {
		return err
	}
	//
	switch RotationPolicy(policyStr) {
	case Monthly, Weekly, Daily:
		*r = RotationPolicy(policyStr)
		return nil
	default:
		return fmt.Errorf("invalid rotation policy: %s", policyStr)
	}
}

// // UnmarshalYAML for Config to handle custom types directly
// func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	// Define auxiliary type to avoid recursion
// 	type plain Config
// 	if err := unmarshal((*plain)(c)); err != nil {
// 		return err
// 	}

// 	// Temporary structure to hold the string values for parsing
// 	var temp struct {
// 		PollingInterval            string `yaml:"polling_interval"`
// 		RepeatNotificationInterval string `yaml:"repeat_notification_interval"`
// 	}

// 	if err := unmarshal(&temp); err != nil {
// 		return err
// 	}

// 	// Convert PollingInterval string to time.Duration
// 	pollingDuration, err := time.ParseDuration(temp.PollingInterval)
// 	if err != nil {
// 		return fmt.Errorf("invalid format for polling interval: %v", err)
// 	}
// 	c.PollingInterval = pollingDuration

// 	// Convert RepeatNotificationInterval string to time.Duration
// 	repeatDuration, err := time.ParseDuration(temp.RepeatNotificationInterval)
// 	if err != nil {
// 		return fmt.Errorf("invalid format for repeat notification interval: %v", err)
// 	}
// 	c.RepeatNotificationInterval = repeatDuration

// 	return nil
// }

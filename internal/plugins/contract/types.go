package contract

// HealthStatus représente l'état de santé d'un plugin (hors protobuf).
type HealthStatus struct {
	Status    HealthStatusEnum `json:"status"`
	Message   string           `json:"message,omitempty"`
	Timestamp int64            `json:"timestamp"`
}

// HealthStatusEnum représente les états de santé possibles.
type HealthStatusEnum int

const (
	HealthUnknown   HealthStatusEnum = 0
	HealthHealthy   HealthStatusEnum = 1
	HealthDegraded  HealthStatusEnum = 2
	HealthUnhealthy HealthStatusEnum = 3
)

func (e HealthStatusEnum) String() string {
	switch e {
	case HealthHealthy:
		return "HEALTHY"
	case HealthDegraded:
		return "DEGRADED"
	case HealthUnhealthy:
		return "UNHEALTHY"
	default:
		return "UNKNOWN"
	}
}

// HealthStatusFromProto convertit le status protobuf en notre enum.
func HealthStatusFromProto(s HealthCheckResponse_Status) HealthStatusEnum {
	switch s {
	case HealthCheckResponse_HEALTHY:
		return HealthHealthy
	case HealthCheckResponse_DEGRADED:
		return HealthDegraded
	case HealthCheckResponse_UNHEALTHY:
		return HealthUnhealthy
	default:
		return HealthUnknown
	}
}

// PluginDescriptor, Capability, RegisterRequest et RegisterResponse
// sont générés par protobuf dans plugin.pb.go. Utiliser ces types
// directement pour la sérialisation REST (ils incluent les tags JSON).

package services

// ServiceConfig represents a service configuration
type ServiceConfig struct {
	Name        string
	Image       string
	Port        int
	Environment map[string]string
}

// GetPostgresConfig returns PostgreSQL service config
func GetPostgresConfig() ServiceConfig {
	return ServiceConfig{
		Name:  "postgres",
		Image: "postgres:14",
		Port:  5432,
		Environment: map[string]string{
			"POSTGRES_USER":     "dev",
			"POSTGRES_PASSWORD": "devpass",
			"POSTGRES_DB":       "devdb",
		},
	}
}

// GetRedisConfig returns Redis service config
func GetRedisConfig() ServiceConfig {
	return ServiceConfig{
		Name:        "redis",
		Image:       "redis:latest",
		Port:        6379,
		Environment: map[string]string{},
	}
}

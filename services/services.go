package services

// ServiceConfig holds docker-compose configuration for a service
type ServiceConfig struct {
	Name        string
	Image       string
	Ports       []string
	Environment map[string]string
	Volumes     []string
}

func GetPostgresConfig() ServiceConfig {
	return ServiceConfig{
		Name:  "postgres",
		Image: "postgres:14",
		Ports: []string{"5432:5432"},
		Environment: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "appdb",
		},
		Volumes: []string{"postgres_data:/var/lib/postgresql/data"},
	}
}

func GetRedisConfig() ServiceConfig {
	return ServiceConfig{
		Name:  "redis",
		Image: "redis:latest",
		Ports: []string{"6379:6379"},
	}
}

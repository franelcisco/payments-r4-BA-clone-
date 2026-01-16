package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	Port string

	// DATABASE
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	SSLMode    string

	// R4 API credentials
	R4BoneEntryPoint    string
	R4BoneCommerceToken string
	BoneSecret          string

	R4APPAEntryPoint    string
	R4APPACommerceToken string
	APPASecret          string
}

// Load reads configuration from environment variables and returns a Config struct
func Load() (*Config, error) {
	cfg := &Config{
		Port: os.Getenv("PORT"),

		// DATABASE
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		SSLMode:    os.Getenv("SSL_MODE"),

		R4BoneEntryPoint:    os.Getenv("R4_BONE_ENTRY_POINT"),
		R4BoneCommerceToken: os.Getenv("R4_BONE_COMMERCE_TOKEN"),
		BoneSecret:          os.Getenv("BONE_SECRET"),

		R4APPAEntryPoint:    os.Getenv("R4_APPA_ENTRY_POINT"),
		R4APPACommerceToken: os.Getenv("R4_APPA_COMMERCE_TOKEN"),
		APPASecret:          os.Getenv("APPA_SECRET"),
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.DBHost == "" {
		return fmt.Errorf("DBHost is not configured")
	}
	if cfg.DBPort == "" {
		return fmt.Errorf("DBPort is not configured")
	}
	if cfg.DBUser == "" {
		return fmt.Errorf("DBUser is not configured")
	}
	if cfg.DBPassword == "" {
		return fmt.Errorf("DBPassword is not configured")
	}
	if cfg.DBName == "" {
		return fmt.Errorf("DBName is not configured")
	}

	if cfg.R4BoneEntryPoint == "" {
		return fmt.Errorf("R4BoneEntryPoint is not configured")
	}
	if cfg.R4BoneCommerceToken == "" {
		return fmt.Errorf("R4BoneCommerceToken is not configured")
	}
	if cfg.BoneSecret == "" {
		return fmt.Errorf("BoneSecret is not configured")
	}

	if cfg.R4APPAEntryPoint == "" {
		return fmt.Errorf("R4APPAEntryPoint is not configured")
	}
	if cfg.R4APPACommerceToken == "" {
		return fmt.Errorf("R4APPACommerceToken is not configured")
	}
	if cfg.APPASecret == "" {
		return fmt.Errorf("APPASecret is not configured")
	}

	return nil
}

package main

// Config provides containerd configuration data for the server
type Config struct {
	// Version of the config file
	Version  string `yaml:"version"`
	GRPCAddr string `yaml:"grpc_path"`
}

func defaultConfig() *Config {
	return &Config{
		Version:  "tarus-service/v0alpha",
		GRPCAddr: "",
	}
}

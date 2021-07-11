package models

type AccessToken struct {
	ServerConfig
	ClientConfig
	WG
}

type ClientConfig struct {
  Network string `json:"network"`
  Key string `json:"key"`
  LocalRange string `json:"localrange"`
}

type ServerConfig struct {
  APIConnString string `json:"apiconn"`
  APIHost   string  `json:"apihost"`
  APIPort   string `json:"apiport"`
  GRPCConnString string `json:"grpcconn"`
  GRPCHost   string `json:"grpchost"`
  GRPCPort   string `json:"grpcport"`
  GRPCSSL   string `json:"grpcssl"`
}

type WG struct {
  GRPCWireGuard  string  `json:"grpcwg"`
  GRPCWGAddress  string `json:"grpcaddr"`
  GRPCWGPort  string  `json:"grpcport"`
  GRPCWGPubKey  string  `json:"pubkey"`
  GRPCWGEndpoint  string  `json:"endpoint"`
}
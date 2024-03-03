package config

type AppConfig struct {
	ServerAddress    string
	StoreDriver      string
	StoreDatabaseURI string
	SecretKey        string
	TokenExp         int
}

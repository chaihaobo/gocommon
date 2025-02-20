package mysql

type Config struct {
	Host        string
	Port        string
	User        string
	Password    string
	Name        string
	MaxOpen     int
	MaxIdle     int
	MaxLifetime int // in minutes
	MaxIdleTime int // in minutes
	CA          []byte
	ServerName  string
	ParseTime   bool
	Location    string
}

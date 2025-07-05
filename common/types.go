package common

type Host struct {
	Name     string
	Hostname string
	User     string
	Port     int
	Keyfile  string
	Group    string
}

type Package struct {
	Name    string 
	Version string
	Manager string
}
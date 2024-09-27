package model

type Index struct {
	Name       string
	Background bool
	Keys       []DBM
	IsTTLIndex bool
	TTL        int
}

package port

type DNS interface {
	AAAA(host string) (*string, error)
}

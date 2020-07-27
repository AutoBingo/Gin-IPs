package configure

type Oid string

const (
	OidHost   Oid = "HOST"
	OidSwitch Oid = "SWITCH" +
		""
)

var OidArray = []Oid{OidHost, OidSwitch}
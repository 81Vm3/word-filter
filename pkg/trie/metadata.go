package trie

const (
	SeverityNormal   = 0
	SeverityMedium   = 1
	SeveritySerious  = 2
	SeverityCritical = 3
)

type Metadata struct {
	Severity int
	Extra    map[string]string
}

func DefaultMetadata() Metadata {
	return Metadata{Severity: SeverityNormal}
}

func (m Metadata) Clone() Metadata {
	if m.Extra == nil {
		return Metadata{Severity: m.Severity}
	}
	cp := make(map[string]string, len(m.Extra))
	for k, v := range m.Extra {
		cp[k] = v
	}
	return Metadata{Severity: m.Severity, Extra: cp}
}

package cql

type Migrations []*Migration

func (s Migrations) Len() int {
	return len(s)
}

func (s Migrations) Less(i, j int) bool {
	return s[i].Version < s[j].Version
}

func (s Migrations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//
// Check whether this list of Migrations contains a specific
// Migration (m)
//
func (s Migrations) Contains(m *Migration) bool {
	for _, a := range s {
		if a.Compare(m) {
			return true
		}
	}
	return false
}

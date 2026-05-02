package tv

type HistoryStore struct {
	maxPerID int
	entries  map[int][]string
}

func NewHistoryStore(maxPerID int) *HistoryStore {
	return &HistoryStore{
		maxPerID: maxPerID,
		entries:  make(map[int][]string),
	}
}

var DefaultHistory = NewHistoryStore(20)

func (hs *HistoryStore) Add(id int, s string) {
	if s == "" {
		return
	}
	list := hs.entries[id]
	if len(list) > 0 && list[len(list)-1] == s {
		return
	}
	for i := 0; i < len(list); i++ {
		if list[i] == s {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	list = append(list, s)
	if len(list) > hs.maxPerID {
		list = list[len(list)-hs.maxPerID:]
	}
	hs.entries[id] = list
}

func (hs *HistoryStore) Entries(id int) []string {
	return hs.entries[id]
}

func (hs *HistoryStore) Clear() {
	hs.entries = make(map[int][]string)
}

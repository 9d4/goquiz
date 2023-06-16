package entity

type Status string

const (
	StatusOnline  Status = "Online"
	StatusOffline Status = "Offline"
	StatusWorking Status = "Working"
)

var Onlines *onlineUsers

func init() {
	Onlines = newOnlineUsers()
}

func newOnlineUsers() *onlineUsers {
	return &onlineUsers{
		ids: make(map[int]struct{}),
	}
}

type onlineUsers struct {
	ids map[int]struct{}
}

func (ou *onlineUsers) Add(id int) {
	ou.ids[id] = struct{}{}
}

func (ou *onlineUsers) Check(id int) bool {
	_, ok := ou.ids[id]
	return ok
}

func (ou *onlineUsers) Remove(id int) {
	delete(ou.ids, id)
}

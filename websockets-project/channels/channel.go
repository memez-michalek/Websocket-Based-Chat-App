package channels


import(

)



type channel struct {
	name  string
	owner string
	users []string
}

func (c channel) CreateChannel(channelName string, owner string) {

	c.name = channelName
	c.owner = owner
	c.users = append(c.users, owner)
}
func (c channel) AddUser(username string) {



}
func (c channel) DeleteUser(username string) {

}

func (c channel) DeleteChannel(channelname string) {

}

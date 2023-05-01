package plugins

type RegisterArgs struct {
	Config map[string]interface{}
}

type RegisterReply struct {
	Routes []*Route
}

package election

type Callbacks struct {
	OnElected func(resource string, token int64)
	OnRevoked func(resource string)
}

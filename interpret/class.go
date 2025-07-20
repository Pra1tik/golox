package interpret

type class struct {
	name string
}

func (c class) String() string {
	return c.name
}

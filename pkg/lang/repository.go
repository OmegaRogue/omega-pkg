package lang

type Repository struct {
	Name        string       `hcl:"name,label"`
	Url         string       `hcl:"url"`
	Type        string       `hcl:"type,optional"`
	Key         string       `hcl:"key,optional"`
	Constraints *Constraints `hcl:"constraints,block"`
}

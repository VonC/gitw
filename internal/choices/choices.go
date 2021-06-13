package choices

type ChoicesManager interface {
	CachedChoices() []string
	RetrievedChoices() []string
}

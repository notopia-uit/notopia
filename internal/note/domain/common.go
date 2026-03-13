package domain

type TrashedBy string

var (
	TrashedByUnspecified TrashedBy = "unspecified"
	TrashedByPurpose     TrashedBy = "purpose"
	TrashedByParent      TrashedBy = "parent"
)

func (t TrashedBy) String() string {
	return string(t)
}

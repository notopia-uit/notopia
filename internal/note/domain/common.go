package domain

type TrashedBy string

var (
	TrashedByUnspecified TrashedBy = "unspecified"
	TrashedByPurpose     TrashedBy = "purpose"
	TrashedByParent      TrashedBy = "parent"
)

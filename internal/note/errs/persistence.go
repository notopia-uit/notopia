package errs

var (
	CodePersistenceInvalid  Code = "persistence_1"
	CodePersistenceInternal Code = "persistence_2"
)

type PersistenceInvalid struct {
	Err
}

func NewPersistenceInvalid(message string, err error) *PersistenceInvalid {
	return &PersistenceInvalid{
		Err: Err{
			message: message,
			code:    CodePersistenceInvalid,
			err:     err,
		},
	}
}

type PersistenceInternal struct {
	Err
}

func NewPersistenceInternal(message string, err error) *PersistenceInternal {
	return &PersistenceInternal{
		Err: Err{
			message: message,
			code:    CodePersistenceInternal,
			err:     err,
		},
	}
}

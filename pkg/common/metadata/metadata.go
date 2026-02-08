package metadata

type ServiceName string

func (sn *ServiceName) String() string {
	return string(*sn)
}

type ServiceVersion string

func (sv *ServiceVersion) String() string {
	return string(*sv)
}

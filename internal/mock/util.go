package mock

type anyCloner interface {
	cloneAny() any
}

type timeRemover interface {
	withoutTime() any
}

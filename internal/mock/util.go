package mock

type cloner interface {
	clone() any
}

type timeRemover interface {
	withoutTime() any
}

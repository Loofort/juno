package middle

type versionMW struct {
	base ContextRouter
	ver  string
}

// Version creates router wrapper that adds version to each path
func Version(base ContextRouter, ver string) ContextRouter {
	return versionMW{base, ver}
}

func (mw versionMW) Handle(method, path string, handler JunoHandler) {
	path = "/" + mw.ver + path
	mw.base.Handle(method, path, handler)
}

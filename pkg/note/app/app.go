package app

type App struct{}

func NewApp() *App {
	return &App{}
}

var ProvideApp = NewApp

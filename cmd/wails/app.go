package main

import (
	"context"

	"sensitive-filter/pkg/gui"
)

type App struct {
	ctx context.Context
	svc *gui.Service
}

func NewApp(svc *gui.Service) *App {
	return &App{svc: svc}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) FilterTest(req gui.FilterTestRequest) gui.FilterTestResponse {
	return a.svc.FilterTest(req)
}

func (a *App) TrieTest(req gui.TrieTestRequest) gui.TrieTestResponse {
	return a.svc.TrieTest(req)
}

func (a *App) ACTest(req gui.ACTestRequest) gui.ACTestResponse {
	return a.svc.ACTest(req)
}

func (a *App) NormalizeTest(req gui.NormalizeTestRequest) gui.NormalizeTestResponse {
	return a.svc.NormalizeTest(req)
}

func (a *App) LexiconTest(req gui.LexiconTestRequest) gui.LexiconTestResponse {
	return a.svc.LexiconTest(req)
}

func (a *App) WordsPath() string {
	return a.svc.WordsPath()
}

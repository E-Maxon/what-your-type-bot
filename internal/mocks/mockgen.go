package mocks

//go:generate mockgen -build_flags=--mod=mod -destination=./tg_api/mock.go -package=tg_api_mock github.com/E-Maxon/what-your-type-bot/internal/tg_api TgAPI

package public

import (
	"embed"
)

//go:embed static/* template/*
var FS embed.FS

package models

import (
	"github.com/firebase/genkit/go/ai"
)

type CodeFile struct {
	Docs []*ai.Document
	Path string
}

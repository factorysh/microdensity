package service

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

type FolderService struct {
	Name       string
	cctx cue.Context
	constraint cue.Value
}

func (f *FolderService) parseCue(constraints string) error {
	f.cctx := cuecontext.New()
	v := cctx.CompileString(constraints)
	if v.Err() != nil {
		return v.Err()
	}
	f.constraint = v
	return nil
}

func (f *FolderService) Validate(args map[string]interface{}) error {
	v := 
	return f.constraint.Decode(args)
}

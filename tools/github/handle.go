package github

import (
	ghrepos "github.com/xhd2015/dot-pkgs/go-pkgs/git/github"
)

func Handle(args []string) error {
	return ghrepos.RunCLI(args)
}

package commands

import "errors"

var BBLNotFound error = errors.New("a bbl environment could not be found, please create a new environment before running this command again")

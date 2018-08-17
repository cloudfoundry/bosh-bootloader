package app

import (
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/genevieve/leftovers/common"
	multierror "github.com/hashicorp/go-multierror"
)

type logger interface {
	Println(message string)
}

type AsyncDeleter struct {
	logger logger
}

func NewAsyncDeleter(logger logger) AsyncDeleter {
	return AsyncDeleter{
		logger: logger,
	}
}

func (a AsyncDeleter) Run(deletables [][]common.Deletable) error {
	var (
		wg     sync.WaitGroup
		result *multierror.Error
	)

	for _, list := range deletables {

		for _, d := range list {
			wg.Add(1)

			go func(d common.Deletable) {
				defer wg.Done()

				a.logger.Println(fmt.Sprintf("[%s: %s] Deleting...", d.Type(), d.Name()))

				err := d.Delete()
				if err != nil {
					err = fmt.Errorf("[%s: %s] %s", d.Type(), d.Name(), color.YellowString(err.Error()))
					result = multierror.Append(result, err)

					a.logger.Println(err.Error())
				} else {
					a.logger.Println(fmt.Sprintf("[%s: %s] %s", d.Type(), d.Name(), color.GreenString("Deleted!")))
				}
			}(d)
		}

		wg.Wait()
	}

	return result.ErrorOrNil()
}

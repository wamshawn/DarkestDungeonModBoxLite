package box

import (
	"sync"
)

var openOnce sync.Once

func (bx *Box) Open() (settings Settings, err error) {
	// open database
	openOnce.Do(func() {
		dbErr := bx.createDB()
		if dbErr != nil {
			err = ErrOpened
			return
		}
	})
	// settings
	settings, err = bx.Settings()
	return
}

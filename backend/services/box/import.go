package box

type ImportEntry struct {
	Chosen   bool          `json:"chosen"`
	Key      string        `json:"key"`
	Title    string        `json:"title"`
	Children []ImportEntry `json:"children"`
}

type ImportMod struct {
	Dst     string        `json:"dst"`
	Title   string        `json:"title"`
	Kind    string        `json:"kind"`
	Entries []ImportEntry `json:"entries"`
}

func (bx *Box) ImportMod() (v ImportMod, err error) {
	// todo get settings > range mod > start task
	return
}

func (bx *Box) StartImportMod(param ImportMod) (pid string, err error) {

	return
}

func (bx *Box) CancelImportMod(pid string) (err error) {

	return
}

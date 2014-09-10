package mogi

type tx struct {
}

func (t *tx) Commit() error {
	return nil
}

func (t *tx) Rollback() error {
	return nil
}

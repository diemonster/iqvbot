package db

import "github.com/quintilesims/iqvbot/models"

// Init will initialize the table entries for the specified store
func Init(store Store) error {
	initFunc := func(key string, v interface{}) error {
		if err := store.Read(key, &v); err != nil {
			if _, ok := err.(*MissingEntryError); ok {
				return store.Write(key, v)
			}

			return err
		}

		return nil
	}

	if err := initFunc(AliasesKey, map[string]string{}); err != nil {
		return err
	}

	if err := initFunc(CandidatesKey, models.Candidates{}); err != nil {
		return err
	}

	if err := initFunc(InterviewsKey, models.Interviews{}); err != nil {
		return err
	}

	if err := initFunc(KarmaKey, models.Karma{}); err != nil {
		return err
	}

	if err := initFunc(KVSKey, map[string]string{}); err != nil {
		return err
	}

	return nil
}

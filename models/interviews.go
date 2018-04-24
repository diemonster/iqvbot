package models

type Interview struct {
	InterviewID    string
	Candidate      string
	InterviewerIDs []string
}

type Interviews []*Interview

func (i Interviews) Get(interviewID string) (*Interview, bool) {
	for _, interview := range i {
		if interview.InterviewID == interviewID {
			return interview, true
		}
	}

	return nil, false
}

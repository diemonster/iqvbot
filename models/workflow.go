package models

type Step struct {
	Text string
}

type Workflow struct {
	Name        string
	Steps       []Step
	CurrentStep int
}

type Workflows []Workflow

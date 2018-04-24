package bot

// todo: update tests

/*
func TestCandidateAdd(t *testing.T) {
	store := newMemoryStore(t)
	cmd := NewCandidateCommand(store, ioutil.Discard)
	if err := slackbot.NewTestApp(cmd, "!candidate add --meta k1=v1 --meta k2=v2 \"John Doe\" <@uid>"); err != nil {
		t.Fatal(err)
	}
	result := models.Candidates{}
	if err := store.Read(db.CandidatesKey, &result); err != nil {
		t.Fatal(err)
	}
	expected := models.Candidates{
		{
			Name:      "John Doe",
			ManagerID: "uid",
			Meta:      map[string]string{"k1": "v1", "k2": "v2"},
		},
	}
	assert.Equal(t, expected, result)
}
func TestCandidateAddErrors(t *testing.T) {
	store := newMemoryStore(t)
	if err := store.Write(db.CandidatesKey, models.Candidates{{Name: "John"}}); err != nil {
		t.Fatal(err)
	}
	inputs := []string{
		"!candidate add",
		"!candidate add NAME",
		"!candidate add NAME MANAGER",
		"!candidate add NAME @MANAGER",
		"!candidate add --meta NAME <@MANAGER>",
		"!candidate add --meta key NAME <@MANAGER>",
		"!candidate add --meta key:val NAME",
		"!candidate add John <@MANAGER>",
	}
	cmd := NewCandidateCommand(store, ioutil.Discard)
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if err := slackbot.NewTestApp(cmd, input); err == nil {
				t.Fatal("Error was nil!")
			}
		})
	}
}
// todo: test --count and --ascending flag
func TestCandidateList(t *testing.T) {
	candidates := models.Candidates{
		{Name: "John Doe"},
		{Name: "Jane Doe"},
	}
	store := newMemoryStore(t)
	if err := store.Write(db.CandidatesKey, candidates); err != nil {
		t.Fatal(err)
	}
	w := bytes.NewBuffer(nil)
	cmd := NewCandidateCommand(store, w)
	if err := slackbot.NewTestApp(cmd, "!candidate ls"); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		assert.Contains(t, w.String(), candidate.Name)
	}
}
func TestCandidateListErrors(t *testing.T) {
	cmd := NewCandidateCommand(newMemoryStore(t), ioutil.Discard)
	if err := slackbot.NewTestApp(cmd, "!candidate ls"); err == nil {
		t.Fatal("Error was nil!")
	}
}
func TestCandidateRemove(t *testing.T) {
	candidates := models.Candidates{
		{Name: "John Doe"},
		{Name: "Jane Doe"},
	}
	store := newMemoryStore(t)
	if err := store.Write(db.CandidatesKey, candidates); err != nil {
		t.Fatal(err)
	}
	cmd := NewCandidateCommand(store, ioutil.Discard)
	if err := slackbot.NewTestApp(cmd, "!candidate rm John Doe"); err != nil {
		t.Fatal(err)
	}
	resultCandidates := models.Candidates{}
	if err := store.Read(db.CandidatesKey, &resultCandidates); err != nil {
		t.Fatal(err)
	}
	expectedCandidates := models.Candidates{
		{Name: "Jane Doe"},
	}
	assert.Equal(t, expectedCandidates, resultCandidates)
}
func TestCandidateRemoveErrors(t *testing.T) {
	inputs := []string{
		"!candidate rm",
		"!candidate rm John Doe",
	}
	cmd := NewCandidateCommand(newMemoryStore(t), ioutil.Discard)
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if err := slackbot.NewTestApp(cmd, input); err == nil {
				t.Fatal("Error was nil!")
			}
		})
	}
}
func TestCandidateShow(t *testing.T) {
	candidates := models.Candidates{
		{Name: "John Doe"},
	}
	store := newMemoryStore(t)
	if err := store.Write(db.CandidatesKey, candidates); err != nil {
		t.Fatal(err)
	}
	w := bytes.NewBuffer(nil)
	cmd := NewCandidateCommand(store, w)
	if err := slackbot.NewTestApp(cmd, "!candidate show john doe"); err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, w.String(), "John Doe")
}
func TestCandidateShowErrors(t *testing.T) {
	inputs := []string{
		"!candidate show",
		"!candidate show John Doe",
	}
	cmd := NewCandidateCommand(newMemoryStore(t), ioutil.Discard)
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if err := slackbot.NewTestApp(cmd, input); err == nil {
				t.Fatal("Error was nil!")
			}
		})
	}
}
func TestCandidateUpdate(t *testing.T) {
	candidates := models.Candidates{
		{
			Name: "John Doe",
			Meta: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
		},
	}
	store := newMemoryStore(t)
	if err := store.Write(db.CandidatesKey, candidates); err != nil {
		t.Fatal(err)
	}
	cmd := NewCandidateCommand(store, ioutil.Discard)
	if err := slackbot.NewTestApp(cmd, "!candidate update \"John Doe\" k1 updated"); err != nil {
		t.Fatal(err)
	}
	result := models.Candidates{}
	if err := store.Read(db.CandidatesKey, &result); err != nil {
		t.Fatal(err)
	}
	expected := models.Candidates{
		{
			Name: "John Doe",
			Meta: map[string]string{
				"k1": "updated",
				"k2": "v2",
			},
		},
	}
	assert.Equal(t, expected, result)
}
func TestCandidateUpdateErrors(t *testing.T) {
	inputs := []string{
		"!candidate update",
		"!candidate update NAME",
		"!candidate update NAME KEY",
		"!candidate update NAME KEY VAL",
	}
	cmd := NewCandidateCommand(newMemoryStore(t), ioutil.Discard)
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if err := slackbot.NewTestApp(cmd, input); err == nil {
				t.Fatal("Error was nil!")
			}
		})
	}
}
*/

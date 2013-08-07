package gofer

import (
  "testing"
)

func TestManualIndex(t *testing.T) {
  manual := manual([]*Task{&Task{Label: "one", manual: []*Task{&Task{Label: "two", manual: []*Task{&Task{Label: "three"}}}}}})
  task := manual.index("one:two:three")

  if nil == task || "three" != task.Label {
    t.Errorf(`Expected label of nested tast to be "three", got %s.`, task)
  }
}

func TestManualSectionalize(t *testing.T) {
  manual := make(manual, 0)
  manual.sectionalize("one:two:three")

  task := manual.index("one:two:three")

  if nil == task {
    t.Error(`Unable to find task created during call to sectionalize.`)
  } else if "one:two" != task.Section || "three" != task.Label {
    t.Errorf(`Tasks was not created properly during call to sectionalize,`+
      ` wanted "one:two" & "three", got "%s" & "%s".`, task.Section, task.Label)
  }
}

func TestRegister(t *testing.T) {
  task := Task{
    Section: "one:two",
    Label:   "three",
  }

  Register(task)

  stored := gofer.index("one:two:three")
  parent := gofer.index("one:two")

  if nil == stored || nil == parent {
    t.Error(`Register failed to create and store task.`)
  } else if "one:two" != stored.Section || "three" != stored.Label {
    t.Errorf(`Register failed to store task properly, expected Section to be "one:two"`+
      ` & Label to be "three", got %s & %s.`, stored.Section, stored.Label)
  } else if "one" != parent.Section || "two" != parent.Label {
    t.Errorf(`Register failed to create parent properly, expected Section to be "one:two"`+
      ` & Label to be "three", got %s & %s.`, parent.Section, parent.Label)
  }

  other := Task{
    Section: "one:two",
    Label:   "four",
  }

  Register(other)
  stored = gofer.index("one:two")

  if 1 != len(gofer) {
    t.Error(`Register failed to associate parent Section properly.`)
  } else if 2 != len(stored.manual) {
    t.Error(`Register failed to associate parent Section properly.`)
  }
}

func TestPreform(t *testing.T) {
  unperformed := true

  task := Task{
    Section: "one:two",
    Label:   "five",
    Action: func(arguments ...interface{}) error {
      unperformed = false
      return nil
    },
  }

  Register(task)
  err := Preform("one:two:five")

  if nil != err {
    t.Error(err)
  } else if unperformed {
    t.Error(`"unpreformed" flag was no flipped to false, call to Preform failed to run action.`)
  }
}

func TestPreformWithDependencies(t *testing.T) {
  unperformed := true

  dependency := Task{
    Section: "one:two",
    Label:   "six",
    Action: func(arguments ...interface{}) error {
      unperformed = false
      return nil
    },
  }

  task := Task{
    Section:      "one:two",
    Label:        "seven",
    Dependencies: []string{"one:two:six"},
    Action: func(arguments ...interface{}) error {
      return nil
    },
  }
  Register(dependency)
  Register(task)

  err := Preform("one:two:seven")

  if nil != err {
    t.Error(err)
  } else if unperformed {
    t.Error(`"unpreformed" flag was no flipped to false, call to Preform failed to run dependency action.`)
  }
}

func TestDependencyOrdering(t *testing.T) {
  var executed []int

  check := func(j int) bool {
    for _, i := range executed {
      if j == i {
        return true
      }
    }
    return false
  }

  d1 := Task{
    Section: "d",
    Label:   "one",
    Action: func(arguments ...interface{}) error {
      executed = append(executed, 1)
      return nil
    },
  }

  d2 := Task{
    Section:      "d",
    Label:        "two",
    Dependencies: []string{"d:one"},
    Action: func(arguments ...interface{}) error {
      if !check(1) {
        t.Error(`Expected "d:one" to have previously executed.`)
      }
      executed = append(executed, 2)
      return nil
    },
  }

  d3 := Task{
    Section:      "d",
    Label:        "three",
    Dependencies: []string{"d:one", "d:four"},
    Action: func(arguments ...interface{}) error {
      if !check(1) || !check(4) {
        t.Error(`Expected "d:one" and "d:four" to have previously executed.`)
      }
      executed = append(executed, 3)
      return nil
    },
  }

  d4 := Task{
    Section:      "d",
    Label:        "four",
    Dependencies: []string{"d:one"},
    Action: func(arguments ...interface{}) error {
      if !check(1) {
        t.Error(`Expected "d:one" and "d:four" to have previously executed.`)
      }
      executed = append(executed, 4)
      return nil
    },
  }

  d5 := Task{
    Section:      "d",
    Label:        "five",
    Dependencies: []string{"d:two", "d:three"},
    Action: func(arguments ...interface{}) error {
      if !check(2) || !check(3) {
        t.Error(`Expected "d:one" and "d:four" to have previously executed.`)
      }
      executed = append(executed, 5)
      return nil
    },
  }

  Register(d1)
  Register(d2)
  Register(d3)
  Register(d4)
  Register(d5)

  if err := Preform("d:five"); nil != err {
    t.Errorf(`Unexpected error encounted, %s.`, err)
  }
}

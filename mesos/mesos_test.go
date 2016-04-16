package mesos

import "testing"

func TestBuildTaskTag(t *testing.T) {
	for _, tt := range []struct {
		taskTag []string
		r       map[string][]string
		err     string
	}{
		{[]string{}, map[string][]string{}, ""},
		{[]string{"invalid"}, nil, "task-tag pattern invalid, must include 1 colon separator"},
		{[]string{"mytask:mytag"}, map[string][]string{
			"mytask": []string{"mytag"},
		}, ""},
		{[]string{"mytask:mytag1,mytag2"}, map[string][]string{
			"mytask": []string{"mytag1", "mytag2"},
		}, ""},
		{[]string{"mytask1:mytag1,mytag2", "mytask2:othertag"}, map[string][]string{
			"mytask1": []string{"mytag1", "mytag2"},
			"mytask2": []string{"othertag"},
		}, ""},
		{[]string{"mytask:tag1", "mytask:tag2"}, map[string][]string{
			"mytask": []string{"tag1", "tag2"},
		}, ""},
		{[]string{"mytask:tag1,tag2", "mytask:tag2,tag3"}, map[string][]string{
			"mytask": []string{"tag1", "tag2", "tag2", "tag3"},
		}, ""},
	} {
		r, err := buildTaskTag(tt.taskTag)
		if err != nil {
			if err.Error() != tt.err {
				t.Errorf("buildTaskTag(%v) => (%v, %v) want (%v, %v)", tt.taskTag, r, err.Error(), tt.r, tt.err)
			}
		} else if !taskMapEq(r, tt.r) {
			t.Errorf("buildTaskTag(%v) => (%v, %v) want (%v, %v)", tt.taskTag, r, err, tt.r, tt.err)
		}
	}
}

func TestBuildRegisterTaskTags(t *testing.T) {
	for _, tt := range []struct {
		taskName     string
		startingTags []string
		taskTag      map[string][]string
		tags         []string
	}{
		{"mytask", []string{}, map[string][]string{}, []string{}},
		{"mytask", []string{"one"}, map[string][]string{}, []string{"one"}},
		{"mytask", []string{}, map[string][]string{
			"mytask": []string{"one"},
		}, []string{"one"}},
		{"mytask", []string{"one"}, map[string][]string{
			"mytask": []string{"one"},
		}, []string{"one"}},
		{"mytask", []string{"one"}, map[string][]string{
			"mytask": []string{"one", "two"},
		}, []string{"one", "two"}},
		{"mytask", []string{"one"}, map[string][]string{
			"mytask": []string{"two", "three"},
		}, []string{"one", "two", "three"}},
		{"myTask-5", []string{}, map[string][]string{
			"mytask": []string{"one"},
		}, []string{"one"}},
		{"other", []string{"first"}, map[string][]string{
			"mytask": []string{"two", "three"},
		}, []string{"first"}},
	} {
		tags := buildRegisterTaskTags(tt.taskName, tt.startingTags, tt.taskTag)
		if !sliceEq(tags, tt.tags) {
			t.Errorf("buildRegisterTaskTags(%s, %v, %v) => %v want %v", tt.taskName, tt.startingTags, tt.taskTag, tags, tt.tags)
		}
	}
}

func taskMapEq(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for aKey, aVal := range a {
		if bVal, ok := b[aKey]; ok {
			if !sliceEq(aVal, bVal) {
				return false
			}
		} else {
			return false
		}
	}

	for bKey, bVal := range b {
		if aVal, ok := a[bKey]; ok {
			if !sliceEq(bVal, aVal) {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

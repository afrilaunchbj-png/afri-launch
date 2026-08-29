package ai

import "testing"

func TestModelRouter(t *testing.T) {
	r := NewModelRouter("gpt-5.6-terra", "gpt-5.6-luna", "gpt-image-2")

	cases := map[Task]string{
		TaskResearch: "gpt-5.6-terra",
		TaskContent:  "gpt-5.6-terra",
		TaskIdeation: "gpt-5.6-luna",
		TaskImage:    "gpt-image-2",
	}
	for task, want := range cases {
		if got := r.ModelFor(task); got != want {
			t.Errorf("ModelFor(%q) = %q, want %q", task, got, want)
		}
	}
}

package jobs

import (
	"testing"

	"afrilaunch/backend/internal/domain"
)

func TestOperationFor(t *testing.T) {
	cases := map[string]string{
		domain.JobIdeas:     domain.OperationIdeaGeneration,
		domain.JobEbook:     domain.OperationEbookGen,
		domain.JobCover:     domain.OperationImageGen,
		domain.JobSalesPage: domain.OperationSalesPage,
		"unknown":           "",
	}
	for kind, want := range cases {
		if got := operationFor(kind); got != want {
			t.Errorf("operationFor(%q) = %q, want %q", kind, got, want)
		}
	}
}

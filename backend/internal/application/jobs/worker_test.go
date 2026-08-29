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
		domain.JobPosters:   domain.OperationPosterGen,
		domain.JobSalesPage: domain.OperationSalesPage,
		domain.JobResearch:  domain.OperationNicheResearch,
		"unknown":           "",
	}
	for kind, want := range cases {
		if got := operationFor(kind); got != want {
			t.Errorf("operationFor(%q) = %q, want %q", kind, got, want)
		}
	}
}

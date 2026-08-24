package workload

import "testing"

func TestVersionedProfilesHaveDeterministicSemanticOracles(t *testing.T) {
	if len(Profiles) != 7 {
		t.Fatalf("profiles=%#v", Profiles)
	}
	for _, profile := range Profiles {
		if err := Oracle(profile); err != nil {
			t.Fatal(err)
		}
	}
	if err := Oracle(Profile{Name: "invalid", Operations: 1, Errors: 2}); err == nil {
		t.Fatal("invalid oracle was accepted")
	}
}

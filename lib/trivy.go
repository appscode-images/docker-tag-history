package lib

import (
	"errors"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	shell "gomodules.xyz/go-sh"
	"kubeops.dev/scanner/apis/trivy"
	"net/http"
)

// trivy image ubuntu --security-checks vuln --format json --quiet
func Scan(sh *shell.Session, img string) (*trivy.SingleReport, error) {
	args := []any{
		"image",
		img,
		"--security-checks", "vuln",
		"--format", "json",
		// "--quiet",
	}
	out, err := sh.Command("trivy", args...).Output()
	if err != nil {
		return nil, err
	}

	var r trivy.SingleReport
	err = trivy.JSON.Unmarshal(out, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func SummarizeReport(report *trivy.SingleReport) map[string]int {
	riskOccurrence := map[string]int{} // risk -> occurrence

	for _, rpt := range report.Results {
		for _, tv := range rpt.Vulnerabilities {
			riskOccurrence[tv.Severity]++
		}
	}

	return riskOccurrence
}

func ImageExists(ref string) (bool, error) {
	_, err := crane.Manifest(ref, crane.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		if ImageNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ImageNotFound(err error) bool {
	var terr *transport.Error
	if errors.As(err, &terr) {
		return terr.StatusCode == http.StatusNotFound //&& terr.StatusCode != http.StatusForbidden {
	}
	return false
}

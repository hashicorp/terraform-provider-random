# Reference: https://github.com/hashicorp/security-scanner/blob/main/CONFIG.md#binary (private repository)

binary {
  secrets {
    all = true
  }
  go_modules   = true
  osv          = true
  oss_index    = false
  nvd          = false

  triage {
    suppress {
      vulnerabilities = [
        // golang.org/x/crypto/openpgp is deprecated/unmaintained with no fixed
        // version. The provider does not use this package (confirmed via
        // `go mod why golang.org/x/crypto/openpgp`).
        "GO-2026-5932",
      ]
    }
  }
}

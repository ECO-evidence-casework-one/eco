# SignPath Foundation readiness checklist

## Project eligibility

- [x] Public source repository
- [x] OSI-approved GPL-3.0-only licence
- [x] No proprietary application code in the current source tree
- [x] Public source prerelease in the form intended for later signing
- [x] Functionality and limitations documented
- [x] Project actively maintained
- [x] Privacy statement published
- [x] Code signing policy published
- [x] Author, reviewer and signing-approver roles published
- [x] MFA enabled for repository administration
- [x] Automated build and tests
- [x] Every future signing request requires manual approval
- [ ] SignPath acceptance received
- [ ] SignPath organisation/project configuration issued
- [ ] Artifact metadata restrictions configured
- [ ] Signing workflow integrated with approved immutable action identity
- [ ] Signed build verified on Windows with Smart App Control enabled

## Release constraints

- ECO must sign only its own maintained source artifacts.
- ECO must not include hacking or security-circumvention features.
- ECO must warn before material system changes and provide uninstallation for installed releases.
- No team member may bypass SignPath technical or approval controls.

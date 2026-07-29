# bingen Security Policy

The bingen project greatly appreciates the need for security and timely updates. bingen is a code-generation tool maintained under the [OpenCost](https://github.com/opencost) organization, and generated code flows into projects that sit close to cloud billing data, so we take security seriously. We are very grateful to the users, security researchers, and developers who report security vulnerabilities to us. All reported security vulnerabilities will be carefully assessed, addressed, and responded to.

## Code Security

Application code is version controlled using GitHub. All code changes are tracked with full revision history and are attributable to a specific individual. Code must be reviewed and accepted by a different engineer than the author of the change.

### Dependabot

bingen has [Dependabot](https://docs.github.com/en/code-security/supply-chain-security/understanding-your-software-supply-chain/about-supply-chain-security#what-is-dependabot) enabled for assessing dependencies in the project, covering both the `gomod` and `github-actions` ecosystems.

## Supported Versions

bingen provides security updates for the most recent release on GitHub. Security fixes are applied to `main` and included in the next tagged release.

## Reporting a Vulnerability

The bingen project supports GitHub [Private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability), which allows for direct, confidential reporting of security issues to the maintainers. If private reporting is enabled, open a private report from the **Security** tab of the [bingen repository](https://github.com/opencost/bingen/security/advisories/new).

Please include a thorough description of the issue, the steps you took to reproduce it, affected versions, and, if known, any mitigations. The maintainers will help diagnose the severity of the issue and determine how to address it. We aim to acknowledge new reports within 3 business days. Issues deemed to be non-critical will be filed as GitHub issues. Critical issues will receive immediate attention and be fixed as quickly as possible.

## Disclosure Policy

For known public security vulnerabilities, we will disclose the vulnerability as soon as possible after receiving the report. Vulnerabilities discovered for the first time will be disclosed in accordance with the following process:

1. Each security vulnerability report is triaged by the maintainers for follow-up coordination and remediation work.
2. After the vulnerability is confirmed, we create a draft GitHub Security Advisory that lists the details of the vulnerability.
3. Related personnel are invited to discuss the fix.
4. A temporary private fork is used to collaborate on a fix.
5. After the fixed code is merged, the vulnerability is publicly posted in the GitHub Advisory Database.

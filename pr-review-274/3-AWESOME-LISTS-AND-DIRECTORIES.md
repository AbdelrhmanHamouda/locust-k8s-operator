# Awesome Lists & Operator Directories - Where to Submit

**Goal:** Maximize visibility by listing in all relevant operator catalogs, awesome lists, and cloud native directories

**Status:** Prioritized action list with submission instructions

---

## TIER 1: CRITICAL (Submit Immediately)

### 1. **Artifact Hub** ⭐ HIGHEST PRIORITY

**What:** CNCF's official package/operator discovery platform (CNCF Incubating project as of Sept 2024)
**URL:** https://artifacthub.io/
**Why Critical:** De facto standard for finding Kubernetes operators. Became CNCF incubating project in 2024.

**How to Submit:**

1. **Sign in** to Artifact Hub (top right menu)
2. **Add Repository** from control panel
3. **Repository Details:**
   - Repository Type: `OLM operators` or `Helm charts`
   - URL: `https://github.com/AbdelrhmanHamouda/locust-k8s-operator`
   - For Helm: `https://abdelrhmanhamouda.github.io/locust-k8s-operator` (chart repository)

**Requirements:**
- Operators must be packaged using Operator Framework format (you already have this)
- Repository hosted on GitHub/GitLab/Bitbucket
- Once added, new versions are automatically indexed every 30 minutes

**Expected Outcome:**
- Operator appears in Artifact Hub search
- Automatic updates when you push new releases
- Verified badge if you claim ownership

**Documentation:** https://artifacthub.io/docs/topics/repositories/

**Sources:**
- [Artifact Hub | CNCF](https://www.cncf.io/projects/artifact-hub/)
- [OLM operators | Artifact Hub documentation](https://artifacthub.io/docs/topics/repositories/olm-operators/)
- [Artifact Hub becomes a CNCF incubating project | CNCF](https://www.cncf.io/blog/2024/09/17/artifact-hub-becomes-a-cncf-incubating-project/)

**Priority:** 🔴 DO THIS FIRST

---

### 2. **OperatorHub.io** ⭐ HIGH PRIORITY

**What:** Canonical source for Kubernetes Operators that appear on OperatorHub.io and OKD
**URL:** https://operatorhub.io/
**GitHub:** https://github.com/k8s-operatorhub/community-operators

**Why Critical:** Official OperatorHub catalog, integrated with OpenShift and vanilla Kubernetes

**How to Submit:**

1. **Fork** https://github.com/k8s-operatorhub/community-operators
2. **Add your operator** to `operators/` directory (use existing operators as template)
3. **Create ci.yaml** at top level of your operator directory
4. **Test locally** (use operator-sdk bundle validate)
5. **Open Pull Request**

**Directory Structure:**
```
operators/
  locust-k8s-operator/
    ci.yaml
    2.0.0/
      manifests/
        locust-k8s-operator.clusterserviceversion.yaml
        locust.io_locusttests_crd.yaml
      metadata/
        annotations.yaml
```

**Requirements:**
- Operator packaged using OLM bundle format
- Comprehensive ClusterServiceVersion (CSV)
- Follow contribution checklist in PR template
- Pass automated vetting (CI checks)

**Expected Outcome:**
- Operator appears on OperatorHub.io
- Available via `kubectl krew install olm` and operator lifecycle manager
- Visibility in OpenShift Operator Catalog

**Documentation:**
- https://k8s-operatorhub.github.io/community-operators/
- https://k8s-operatorhub.github.io/community-operators/contributing-where-to/
- https://github.com/k8s-operatorhub/community-operators/blob/master/docs/contributing-via-pr.md

**Sources:**
- [Community operators - Community operators](https://k8s-operatorhub.github.io/community-operators/)
- [GitHub - k8s-operatorhub/community-operators](https://github.com/k8s-operatorhub/community-operators)
- [Consider listing operator in Artifact Hub · Issue #3564 · prometheus-operator/prometheus-operator](https://github.com/prometheus-operator/prometheus-operator/issues/3564)

**Priority:** 🔴 DO WITHIN WEEK 1

---

### 3. **operator-framework/awesome-operators**

**What:** Official curated list of Kubernetes Operators by the Operator Framework team
**URL:** https://github.com/operator-framework/awesome-operators
**List:** https://github.com/operator-framework/awesome-operators/blob/master/README.md

**Why Critical:** Official Operator Framework list, high visibility in operator community

**How to Submit:**

1. **Fork** https://github.com/operator-framework/awesome-operators
2. **Add your operator** to README.md in alphabetical order under appropriate category
3. **Format:**
   ```markdown
   - [Locust K8s Operator](https://github.com/AbdelrhmanHamouda/locust-k8s-operator) - Production-ready operator for running distributed Locust load tests on Kubernetes. Features native OpenTelemetry, HA support, and automatic v1→v2 migration.
   ```
4. **Open Pull Request**

**Requirements:**
- Brief, clear description (1-2 sentences)
- Link to GitHub repository
- Operator must be functional and documented

**Expected Outcome:**
- Listed on official awesome-operators page
- High credibility (official Operator Framework list)

**Documentation:** https://github.com/operator-framework/awesome-operators/blob/master/README.md

**Sources:**
- [GitHub - operator-framework/awesome-operators](https://github.com/operator-framework/awesome-operators)

**Priority:** 🟡 DO WITHIN WEEK 2

---

## TIER 2: IMPORTANT (Submit Within Month 1)

### 4. **ramitsurana/awesome-kubernetes**

**What:** Most comprehensive awesome-kubernetes list (20k+ stars)
**URL:** https://github.com/ramitsurana/awesome-kubernetes
**List:** https://github.com/ramitsurana/awesome-kubernetes/blob/master/docs/projects/projects.md

**Why Important:** Largest Kubernetes awesome list, high visibility

**How to Submit:**

1. **Fork** https://github.com/ramitsurana/awesome-kubernetes
2. **Add to** `docs/projects/projects.md` under "Operators" section
3. **Format:**
   ```markdown
   - [Locust K8s Operator](https://github.com/AbdelrhmanHamouda/locust-k8s-operator) - Kubernetes operator for distributed Locust load testing
   ```
4. **Open Pull Request**

**Expected Outcome:**
- Visibility to 20k+ GitHub users
- SEO boost for "awesome kubernetes operators"

**Sources:**
- [GitHub - ramitsurana/awesome-kubernetes](https://github.com/ramitsurana/awesome-kubernetes)
- [awesome-kubernetes/docs/projects/projects.md at master](https://github.com/ramitsurana/awesome-kubernetes/blob/master/docs/projects/projects.md)

**Priority:** 🟡 WEEK 2-3

---

### 5. **run-x/awesome-kubernetes**

**What:** Another popular awesome-kubernetes list (operator frameworks section)
**URL:** https://github.com/run-x/awesome-kubernetes

**How to Submit:**
- Same process as ramitsurana list
- Add under "Operator Framework" section

**Sources:**
- [GitHub - run-x/awesome-kubernetes](https://github.com/run-x/awesome-kubernetes)

**Priority:** 🟡 WEEK 2-3

---

### 6. **calvin-puram/awesome-kubernetes-operator-resources**

**What:** Curated list specifically for Kubernetes Operator resources
**URL:** https://github.com/calvin-puram/awesome-kubernetes-operator-resources

**Why Important:** Operator-focused list, relevant audience

**How to Submit:**

1. **Fork** repository
2. **Add under** "Operators" section
3. **Open PR**

**Sources:**
- [GitHub - calvin-puram/awesome-kubernetes-operator-resources](https://github.com/calvin-puram/awesome-kubernetes-operator-resources)

**Priority:** 🟡 WEEK 3-4

---

### 7. **pperzyna/awesome-operator-frameworks**

**What:** List of awesome Kubernetes Operator Frameworks
**URL:** https://github.com/pperzyna/awesome-operator-frameworks

**Note:** This is for frameworks (like Operator SDK), but may accept production operators as examples

**How to Submit:**
- Check if they accept production operators
- If yes, add under "Examples" or create PR to suggest category

**Sources:**
- [GitHub - pperzyna/awesome-operator-frameworks](https://github.com/pperzyna/awesome-operator-frameworks)

**Priority:** 🟢 OPTIONAL

---

### 8. **CNCF Landscape**

**What:** Interactive map of the Cloud Native ecosystem
**URL:** https://landscape.cncf.io/
**GitHub:** https://github.com/cncf/landscape

**Why Important:** Official CNCF catalog, high visibility in cloud native community

**How to Submit:**

1. **Fork** https://github.com/cncf/landscape
2. **Add entry** to `landscape.yml` in alphabetical order under appropriate category
3. **Add logo** to `hosted_logos/` directory (SVG format, include operator name in English)
4. **Required fields:**
   ```yaml
   - name: Locust K8s Operator
     homepage_url: https://abdelrhmanhamouda.github.io/locust-k8s-operator/
     logo: locust-k8s-operator.svg
     crunchbase: https://www.crunchbase.com/organization/your-org  # or N/A
   ```
5. **Open Pull Request**

**Requirements:**
- Logo must be SVG format
- Must include name, homepage_url, logo, and crunchbase (or note if not applicable)
- Logo should include product name in English

**Expected Outcome:**
- Appears on landscape.cncf.io (browsed by thousands)
- Credibility boost (official CNCF catalog)
- SEO benefit

**Category Suggestions:**
- "App Definition and Development" → "Application Definition & Image Build"
- Or "Observability and Analysis" → "Monitoring" (if emphasizing observability)

**Documentation:** https://github.com/cncf/landscape/blob/master/README.md

**Sources:**
- [CNCF Landscape](https://landscape.cncf.io/)
- [GitHub - cncf/landscape](https://github.com/cncf/landscape)
- [landscape/README.md at master · cncf/landscape](https://github.com/cncf/landscape/blob/master/README.md)

**Priority:** 🟡 WEEK 3-4 (requires logo preparation)

---

## TIER 3: NICE-TO-HAVE (Submit Over Time)

### 9. **tomhuang12/awesome-k8s-resources**

**What:** Curated list of awesome Kubernetes tools and resources
**URL:** https://github.com/tomhuang12/awesome-k8s-resources

**How to Submit:** Fork, add to operators section, PR

**Priority:** 🟢 MONTH 2-3

---

### 10. **awesomelistsio/awesome-kubernetes**

**What:** Another awesome-kubernetes variant
**URL:** https://github.com/awesomelistsio/awesome-kubernetes

**How to Submit:** Fork, add to operators section, PR

**Priority:** 🟢 MONTH 2-3

---

## SPECIALIZED DIRECTORIES

### 11. **Helm Chart Repositories** (Already Done)

You already have a Helm chart at:
`https://abdelrhmanhamouda.github.io/locust-k8s-operator`

**Ensure it's listed in:**
- Artifact Hub (see #1 above)
- Helm Hub (redirects to Artifact Hub now)

---

### 12. **CNCF Sandbox Project Consideration** (Long-Term)

**What:** Apply to become a CNCF Sandbox project
**URL:** https://www.cncf.io/sandbox-projects/

**Requirements:**
- Demonstrated traction (GitHub stars, contributors, production usage)
- At least 3 active maintainers from 2+ organizations
- Governance documentation
- Code of conduct
- Security audit (for incubation)

**When to Apply:**
- After 100+ stars
- After 5+ companies using in production
- After establishing clear governance

**Why:**
- Official CNCF backing
- Increased visibility
- Access to CNCF resources (CI, security audits, marketing)

**Priority:** 🔵 LONG-TERM (6-12 months)

**Sources:**
- [Sandbox Projects | CNCF](https://www.cncf.io/sandbox-projects/)

---

## OTHER VISIBILITY OPPORTUNITIES

### 13. **Kubernetes Blog (kubernetes.io/blog)**

**What:** Submit guest blog post to official Kubernetes blog
**Topics:**
- "Running Production Load Tests with Kubernetes Operators"
- "Zero-Downtime Operator Migration: Lessons from v1→v2"

**How:** https://kubernetes.io/docs/contribute/new-content/blogs-case-studies/

**Priority:** 🟡 AFTER LAUNCH SUCCESS

---

### 14. **CNCF Blog**

**What:** Guest blog post on CNCF blog
**Topics:**
- "Why We Rewrote Our Operator in Go"
- "Production-Ready Observability with Native OpenTelemetry"

**How:** Reach out to CNCF marketing or submit via CNCF Slack

**Priority:** 🟡 MONTH 2-3

---

### 15. **The New Stack**

**What:** Popular cloud native publication
**How:** Pitch article to editors@thenewstack.io

**Priority:** 🟢 OPTIONAL

---

## SUBMISSION PRIORITY TIMELINE

### Week 1 (Launch Week)
- [ ] Submit to Artifact Hub (#1)
- [ ] Start OperatorHub.io submission (#2)

### Week 2
- [ ] Complete OperatorHub.io submission (#2)
- [ ] Submit to operator-framework/awesome-operators (#3)
- [ ] Submit to ramitsurana/awesome-kubernetes (#4)
- [ ] Submit to run-x/awesome-kubernetes (#5)

### Week 3-4
- [ ] Submit to calvin-puram/awesome-kubernetes-operator-resources (#6)
- [ ] Prepare logo and submit to CNCF Landscape (#8)

### Month 2
- [ ] Submit to additional awesome lists (#9, #10)
- [ ] Consider guest blog post opportunities (#13, #14)

### Month 3+
- [ ] Evaluate CNCF Sandbox application readiness (#12)

---

## LOGO PREPARATION (For CNCF Landscape)

**Requirements:**
- **Format:** SVG (vector, not raster)
- **Size:** Landscape orientation preferred (wider than tall)
- **Content:** Must include operator name "Locust K8s Operator" in English
- **Style:** Clean, professional, recognizable at small sizes
- **Background:** Transparent

**Tools:**
- Figma (free for basic use)
- Inkscape (open source vector editor)
- Canva (can export SVG)

**Example Inspiration:**
- Look at other operators in CNCF landscape
- Keep it simple and iconic

---

## TRACKING SUBMISSIONS

**Create a spreadsheet to track:**

| Directory | Status | PR Link | Date Submitted | Date Merged | Notes |
|-----------|--------|---------|----------------|-------------|-------|
| Artifact Hub | ⏳ Pending | - | YYYY-MM-DD | - | - |
| OperatorHub.io | 📝 Draft | - | - | - | - |
| awesome-operators | ✅ Merged | [link] | YYYY-MM-DD | YYYY-MM-DD | - |
| ... | ... | ... | ... | ... | ... |

---

## SUCCESS METRICS

**Track These After Submissions:**
- **Artifact Hub views:** Check analytics in Artifact Hub dashboard
- **GitHub traffic:** Monitor referrals from each platform
- **GitHub stars:** Track growth after each submission
- **Helm chart downloads:** Monitor chart pull statistics

**Expected Results:**
- Artifact Hub submission → 20-30% traffic increase
- OperatorHub.io → credibility boost, enterprise adoption
- Awesome lists → slow but steady star growth

---

## SOURCES

All recommendations based on:

- [GitHub - operator-framework/awesome-operators](https://github.com/operator-framework/awesome-operators)
- [GitHub - run-x/awesome-kubernetes](https://github.com/run-x/awesome-kubernetes)
- [GitHub - calvin-puram/awesome-kubernetes-operator-resources](https://github.com/calvin-puram/awesome-kubernetes-operator-resources)
- [Artifact Hub | CNCF](https://www.cncf.io/projects/artifact-hub/)
- [OLM operators | Artifact Hub documentation](https://artifacthub.io/docs/topics/repositories/olm-operators/)
- [Community operators](https://k8s-operatorhub.github.io/community-operators/)
- [CNCF Landscape](https://landscape.cncf.io/)
- [GitHub - cncf/landscape](https://github.com/cncf/landscape)

---

## QUICK START CHECKLIST

**Before Submitting Anywhere:**
- [ ] Ensure README.md is polished and comprehensive
- [ ] Verify all links work (docs, Helm repo, examples)
- [ ] Confirm Helm chart is properly published and accessible
- [ ] Have clear, concise operator description ready (1-2 sentences)
- [ ] Screenshots/GIFs ready (optional but helpful)

**Week 1 Actions:**
- [ ] Artifact Hub submission (30 minutes)
- [ ] OperatorHub.io PR draft (2-3 hours)

**Week 2 Actions:**
- [ ] 3-4 awesome list PRs (30 minutes each)

**Ongoing:**
- [ ] Respond to PR feedback quickly
- [ ] Update listings when major versions release
- [ ] Track referral traffic and star growth

---

**Let's get listed!** Start with Artifact Hub today - it's the fastest and highest-impact submission.

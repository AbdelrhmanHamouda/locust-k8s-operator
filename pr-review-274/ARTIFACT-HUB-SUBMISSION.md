# Artifact Hub Submission - Complete Guide

**Status:** HIGHEST PRIORITY - Submit immediately
**Estimated Time:** 30 minutes
**Expected Impact:** 20-30% traffic increase, primary operator discovery platform

---

## WHAT IS ARTIFACT HUB?

**Artifact Hub** is the official CNCF package discovery platform (CNCF Incubating project as of September 2024).

**Why It Matters:**
- ✅ **De facto standard** for finding Kubernetes operators, Helm charts, OLM operators
- ✅ **CNCF backing** - official cloud native ecosystem catalog
- ✅ **Automatic indexing** - new releases indexed every 30 minutes after initial setup
- ✅ **High visibility** - thousands of daily searches
- ✅ **SEO boost** - artifacthub.io ranks high in Google for operator searches
- ✅ **Verified badges** - claim ownership for credibility

**Official Sources:**
- [Artifact Hub | CNCF](https://www.cncf.io/projects/artifact-hub/)
- [Artifact Hub became CNCF incubating project | CNCF](https://www.cncf.io/blog/2024/09/17/artifact-hub-becomes-a-cncf-incubating-project/)
- [OLM operators | Artifact Hub documentation](https://artifacthub.io/docs/topics/repositories/olm-operators/)

---

## PREREQUISITES (Check Before Starting)

### ✅ Required (You Already Have These)

- [x] **GitHub Repository:** https://github.com/AbdelrhmanHamouda/locust-k8s-operator
- [x] **Published Helm Chart:** https://abdelrhmanhamouda.github.io/locust-k8s-operator
- [x] **Working Operator:** v2.0.0 released and functional
- [x] **Documentation:** https://abdelrhmanhamouda.github.io/locust-k8s-operator/

### ⚠️ You'll Need (During Submission)

- [ ] **Artifact Hub Account:** Sign in via GitHub (takes 1 minute)
- [ ] **Repository Write Access:** You must be repo owner or have maintainer permissions
- [ ] **Chart Repository URL:** `https://abdelrhmanhamouda.github.io/locust-k8s-operator`

---

## SUBMISSION OPTIONS

You have **2 options** for submitting to Artifact Hub:

### OPTION 1: Helm Chart Repository (RECOMMENDED - EASIER)

**What:** Submit your Helm chart repository (GitHub Pages)
**Best For:** Quick setup, automatic updates
**URL to Submit:** `https://abdelrhmanhamouda.github.io/locust-k8s-operator`

**Pros:**
- ✅ Fastest setup (5-10 minutes)
- ✅ Automatic indexing of new Helm releases
- ✅ No additional YAML files needed
- ✅ Works with your existing setup

**Cons:**
- ⚠️ Helm-specific (won't show operator-specific metadata)

### OPTION 2: OLM Operator Repository (ADVANCED)

**What:** Submit as Operator Lifecycle Manager (OLM) operator
**Best For:** Maximum visibility as "operator" (not just Helm chart)
**Requires:** OLM bundle format, ClusterServiceVersion YAML

**Pros:**
- ✅ Shows as "Operator" category (more specific)
- ✅ Richer metadata (capabilities, maturity)
- ✅ Integration with OperatorHub.io

**Cons:**
- ⚠️ More complex (2-3 hours to create OLM bundle)
- ⚠️ Requires additional YAML files

**Recommendation:** Start with Option 1 (Helm), add Option 2 later for maximum coverage.

---

## STEP-BY-STEP: OPTION 1 (HELM CHART) - RECOMMENDED

### Step 1: Sign In to Artifact Hub

1. **Go to:** https://artifacthub.io/
2. **Click:** "Sign in" (top right corner)
3. **Choose:** "Sign in with GitHub"
4. **Authorize:** Grant Artifact Hub access to your GitHub account
5. **Verify:** You're signed in (your username appears top right)

**Time:** 1 minute

---

### Step 2: Add Your Repository

1. **Click:** Your profile icon (top right) → "Control Panel"
2. **Navigate:** "Repositories" tab
3. **Click:** "Add" button
4. **Fill in the form:**

**Repository Details:**

```
Repository Type: Helm charts
Repository Name: locust-k8s-operator
Display Name: Locust Kubernetes Operator
URL: https://abdelrhmanhamouda.github.io/locust-k8s-operator
```

**Optional but Recommended:**

```
Description: Production-ready Kubernetes operator for distributed Locust load testing. Features native OpenTelemetry, HA support, and automatic v1→v2 migration.

Official: ☐ No (unless you get CNCF endorsement later)

Verified Publisher: ☑ Yes (claim ownership via GitHub repo verification)

Scanner disabled: ☐ No (allow security scanning)
```

**GitHub Repository (IMPORTANT):**
```
GitHub Repository: https://github.com/AbdelrhmanHamouda/locust-k8s-operator
GitHub Repository Branch: master  (or main if that's your default)
```

**Why GitHub repo matters:**
- Enables "Verified Publisher" badge
- Links to GitHub issues/stars
- Allows claiming ownership

5. **Click:** "Add"

**Time:** 3-5 minutes

---

### Step 3: Verify Submission

After clicking "Add", Artifact Hub will:

1. **Fetch your Helm repository index** (`index.yaml`)
2. **Parse Chart.yaml** from your Helm chart
3. **Index your operator** (appears within 30 minutes)

**Check Status:**
- Go to "Control Panel" → "Repositories"
- Look for green checkmark next to "locust-k8s-operator"
- If red X, click for error details

**Common Issues:**

**Error: "Unable to fetch index.yaml"**
- **Cause:** Helm repo not properly published
- **Fix:** Verify `https://abdelrhmanhamouda.github.io/locust-k8s-operator/index.yaml` loads in browser

**Error: "Invalid Chart.yaml"**
- **Cause:** Missing required fields in Chart.yaml
- **Fix:** Ensure Chart.yaml has `name`, `version`, `description`, `apiVersion`

---

### Step 4: Claim Verified Publisher Badge

1. **Go to:** Your operator's Artifact Hub page (search for "locust-k8s-operator")
2. **Click:** "Claim Ownership" (if available)
3. **Verify:** Artifact Hub checks you have write access to GitHub repo
4. **Result:** Green "Verified Publisher" badge appears

**Why This Matters:**
- ✅ Builds trust (users know it's official)
- ✅ Higher search ranking
- ✅ Differentiates from forks/copies

**Time:** 2 minutes

---

### Step 5: Enhance Your Listing (Optional but Recommended)

**Add artifacthub-repo.yml to Your Helm Chart Repository**

Create `.github/artifacthub-repo.yml` in your GitHub repo:

```yaml
# Artifact Hub repository metadata file
repositoryID: <your-repo-id>  # Get this from Artifact Hub after adding repo
owners:
  - name: Abdelrhman Hamouda
    email: your-email@example.com  # Optional

# Links
links:
  - name: Documentation
    url: https://abdelrhmanhamouda.github.io/locust-k8s-operator/
  - name: Support
    url: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/discussions

# Additional metadata
cncf: false  # Set to true if accepted into CNCF Sandbox/Incubating/Graduated
```

**Why Add This:**
- Custom links in Artifact Hub listing
- Better categorization
- Richer metadata

**Time:** 5 minutes

---

### Step 6: Optimize Chart.yaml for Artifact Hub

Ensure your `charts/locust-k8s-operator/Chart.yaml` has rich metadata:

```yaml
apiVersion: v2
name: locust-k8s-operator
version: 2.0.0
description: Production-ready Kubernetes operator for distributed Locust load testing
type: application
appVersion: "2.0.0"

# Artifact Hub metadata
keywords:
  - locust
  - load-testing
  - performance-testing
  - kubernetes-operator
  - distributed-testing
  - opentelemetry

home: https://abdelrhmanhamouda.github.io/locust-k8s-operator/
sources:
  - https://github.com/AbdelrhmanHamouda/locust-k8s-operator

maintainers:
  - name: Abdelrhman Hamouda
    email: your-email@example.com  # Optional
    url: https://github.com/AbdelrhmanHamouda

# Icon (shown in Artifact Hub)
icon: https://raw.githubusercontent.com/AbdelrhmanHamouda/locust-k8s-operator/master/docs/images/logo.png

# Annotations for Artifact Hub
annotations:
  artifacthub.io/changes: |
    - kind: added
      description: Native OpenTelemetry support
    - kind: added
      description: Pod health monitoring with automatic recovery
    - kind: added
      description: Conversion webhooks for zero-downtime v1→v2 migration
    - kind: changed
      description: Complete rewrite in Go (4x memory reduction, 60x faster startup)
  artifacthub.io/containsSecurityUpdates: "false"
  artifacthub.io/license: Apache-2.0
  artifacthub.io/operator: "true"
  artifacthub.io/operatorCapabilities: Basic Install
  artifacthub.io/prerelease: "false"
  artifacthub.io/recommendations: |
    - url: https://artifacthub.io/packages/helm/prometheus-community/prometheus
  artifacthub.io/links: |
    - name: Getting Started
      url: https://abdelrhmanhamouda.github.io/locust-k8s-operator/getting_started/
    - name: Migration Guide
      url: https://abdelrhmanhamouda.github.io/locust-k8s-operator/migration/
    - name: GitHub Discussions
      url: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/discussions
```

**Why This Matters:**
- Better search discoverability (keywords)
- Changelog visible in Artifact Hub
- Links to docs, migration guide
- Professional appearance

**After making changes:**
1. Update chart version (e.g., 2.0.1)
2. Push to GitHub
3. Helm will auto-update index.yaml
4. Artifact Hub will re-index within 30 minutes

**Time:** 10 minutes

---

## WHAT HAPPENS AFTER SUBMISSION?

### Automatic Indexing

**Every 30 minutes**, Artifact Hub checks your Helm repo for:
- New chart versions
- Updated Chart.yaml metadata
- New annotations

**No manual action needed** after initial setup!

### Your Listing Will Show:

**Top Section:**
- Operator name and description
- Version number
- Install instructions (Helm command)
- Verified Publisher badge (if claimed)

**Metadata:**
- Keywords (searchable)
- Maintainers
- Home page link
- Source code link

**Tabs:**
- README (from Chart.yaml or separate README.md)
- Values (values.yaml schema)
- Changelog (from artifacthub.io/changes annotation)
- Security Report (if scanner enabled)

**Install Command (Auto-Generated):**
```bash
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator
helm install my-locust-operator locust-k8s-operator/locust-k8s-operator --version 2.0.0
```

---

## MONITORING YOUR LISTING

### Check Your Listing

**Direct URL Pattern:**
```
https://artifacthub.io/packages/helm/locust-k8s-operator/locust-k8s-operator
```

**Or search:** "locust kubernetes operator" on artifacthub.io

### Analytics (After Claiming Ownership)

**Control Panel → Repositories → [Your Repo] → Stats**

You can see:
- Views per day/week/month
- Install command clicks
- Repository visits
- Top search keywords that found your operator

**Use This Data:**
- Understand discoverability
- Track growth over time
- Optimize keywords if low visibility

---

## MAINTAINING YOUR LISTING

### When You Release a New Version

1. **Update Chart.yaml** version (e.g., `2.0.0` → `2.1.0`)
2. **Update annotations** (change log)
3. **Push to GitHub**
4. **Helm will auto-update** index.yaml via GitHub Actions
5. **Artifact Hub auto-indexes** within 30 minutes
6. **New version appears** on Artifact Hub automatically

**No manual re-submission needed!** ✅

### Best Practices

**Do:**
- ✅ Keep Chart.yaml metadata current (description, keywords)
- ✅ Add changelog in annotations for each release
- ✅ Update icon if you create a logo
- ✅ Respond to comments/questions on Artifact Hub (if enabled)

**Don't:**
- ❌ Delete and re-add repository (keeps history, analytics)
- ❌ Change repository URL (breaks existing links)
- ❌ Skip version updates (users need to know what's new)

---

## TROUBLESHOOTING

### Issue: Operator Not Appearing After 30 Minutes

**Check:**
1. Go to Control Panel → Repositories
2. Look for error icon next to your repo
3. Click for details

**Common Causes:**

**"index.yaml not found"**
- Verify: `https://abdelrhmanhamouda.github.io/locust-k8s-operator/index.yaml` loads
- Fix: Re-run Helm chart release workflow

**"Chart.yaml missing required field"**
- Check: `name`, `version`, `description`, `apiVersion` all present
- Fix: Update Chart.yaml and push

**"Repository processing queued"**
- Normal! First indexing can take 1-2 hours
- Check back later

---

### Issue: No "Verified Publisher" Badge

**Requirements:**
- Repository must link to GitHub repo
- You must have write access to that repo
- You must claim ownership on Artifact Hub

**Steps:**
1. Ensure GitHub repository URL is set in Artifact Hub repo config
2. Visit your operator's page
3. Click "Claim Ownership"
4. Artifact Hub verifies via GitHub OAuth

---

### Issue: README Not Showing Properly

**Cause:** Artifact Hub looks for README.md in chart directory

**Fix:**
Add `charts/locust-k8s-operator/README.md` with:
```markdown
# Locust Kubernetes Operator

Production-ready Kubernetes operator for distributed Locust load testing.

## Features

- Native OpenTelemetry support
- HA deployments with leader election
- Pod health monitoring
- Zero-downtime v1→v2 migration

## Quick Start

[Installation instructions...]

## Documentation

https://abdelrhmanhamouda.github.io/locust-k8s-operator/
```

---

## ADVANCED: OPTION 2 (OLM OPERATOR SUBMISSION)

**Skip this for now - do after Helm submission is working**

### What You'll Need

1. **OLM Bundle Format**
   - ClusterServiceVersion (CSV) YAML
   - CRD YAML
   - Package manifest

2. **Directory Structure:**
   ```
   operators/
     locust-k8s-operator/
       2.0.0/
         manifests/
           locust-k8s-operator.clusterserviceversion.yaml
           locust.io_locusttests_crd.yaml
         metadata/
           annotations.yaml
   ```

3. **Submission:**
   - Fork https://github.com/k8s-operatorhub/community-operators
   - Add your operator
   - Open PR
   - Wait for review and merge

**Estimated Time:** 2-3 hours (complex)
**Benefit:** Shows as "OLM Operator" category, more operator-specific metadata

**When to Do This:**
- After Helm submission is successful
- After 50+ GitHub stars (shows traction)
- When you have time to maintain OLM bundle

---

## SUCCESS METRICS

### Track These After Submission

**Week 1:**
- [ ] Listing appears on Artifact Hub (verify URL)
- [ ] Verified Publisher badge claimed
- [ ] No errors in repository status

**Week 2:**
- [ ] 10+ views per day (check analytics)
- [ ] Appears in search results for "locust kubernetes"

**Month 1:**
- [ ] 100+ total views
- [ ] GitHub stars increase (track correlation)
- [ ] Helm chart downloads increase

**Compare:**
- Traffic before Artifact Hub submission
- Traffic after (expect 20-30% increase)

---

## INTEGRATION WITH OTHER DIRECTORIES

**After Artifact Hub submission succeeds:**

1. **OperatorHub.io** will auto-discover if you submit OLM format
2. **CNCF Landscape** requires manual PR (separate submission)
3. **awesome-operators lists** reference Artifact Hub URLs

**Artifact Hub is the foundation** - other directories often pull from it.

---

## QUICK REFERENCE CHECKLIST

**Pre-Submission:**
- [ ] Helm chart published at https://abdelrhmanhamouda.github.io/locust-k8s-operator
- [ ] index.yaml accessible and valid
- [ ] Chart.yaml has all required fields

**Submission:**
- [ ] Created Artifact Hub account (GitHub OAuth)
- [ ] Added repository via Control Panel
- [ ] Verified repository status (green checkmark)

**Post-Submission:**
- [ ] Claimed Verified Publisher badge
- [ ] Enhanced Chart.yaml with keywords and annotations
- [ ] Added artifacthub-repo.yml (optional)
- [ ] Verified listing appears in search

**Ongoing:**
- [ ] Update Chart.yaml changelog with each release
- [ ] Monitor analytics monthly
- [ ] Respond to user comments (if enabled)

---

## ESTIMATED TIMELINE

**Initial Submission:**
- Create account: 1 minute
- Add repository: 3-5 minutes
- Wait for indexing: 10-30 minutes
- Claim ownership: 2 minutes
- **Total: 15-40 minutes**

**Optimization (Optional):**
- Enhance Chart.yaml: 10 minutes
- Add artifacthub-repo.yml: 5 minutes
- **Total: 15 minutes**

**Grand Total: 30-60 minutes from start to fully optimized listing**

---

## NEXT STEPS AFTER ARTIFACT HUB

**Once Artifact Hub submission is complete:**

1. ✅ **OperatorHub.io** (Week 1-2)
2. ✅ **awesome-operators lists** (Week 2)
3. ✅ **CNCF Landscape** (Week 3-4, requires logo)
4. ✅ **Reddit announcement** (mention Artifact Hub listing)

**Artifact Hub is the cornerstone** - other listings build on it.

---

## FINAL TIPS

**Do:**
- ✅ Submit TODAY (highest priority)
- ✅ Use descriptive keywords (improves search)
- ✅ Keep metadata updated with each release
- ✅ Monitor analytics to understand discovery patterns

**Don't:**
- ❌ Delay submission (every day without it = lost visibility)
- ❌ Use generic descriptions ("kubernetes operator for locust")
- ❌ Forget to claim Verified Publisher badge
- ❌ Ignore errors in repository status

---

## SUPPORT RESOURCES

**Artifact Hub Documentation:**
- https://artifacthub.io/docs

**Helm Chart Documentation:**
- https://artifacthub.io/docs/topics/repositories/helm-charts/

**Community Support:**
- CNCF Slack: #artifact-hub
- GitHub Discussions: https://github.com/artifacthub/hub/discussions

**Report Issues:**
- https://github.com/artifacthub/hub/issues

---

## SUMMARY

**What:** Submit your Helm chart to Artifact Hub (CNCF's official operator catalog)
**Time:** 30-60 minutes
**Impact:** 20-30% traffic increase, primary discovery platform
**Status:** HIGHEST PRIORITY - do this before Reddit, awesome lists, everything else

**Action Items (In Order):**
1. Sign in to Artifact Hub (1 min)
2. Add repository (5 min)
3. Verify indexing (wait 30 min)
4. Claim Verified Publisher badge (2 min)
5. Optimize Chart.yaml (15 min)
6. Monitor analytics weekly

**Once complete, you'll have:**
- ✅ Official CNCF catalog listing
- ✅ Verified Publisher badge
- ✅ Automatic version updates
- ✅ Analytics dashboard
- ✅ Primary discovery channel for operators

**Start here: https://artifacthub.io/** 🚀

---

**Questions? Issues? Check troubleshooting section above or ping CNCF Slack #artifact-hub**

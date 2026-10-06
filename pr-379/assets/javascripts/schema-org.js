document.addEventListener('DOMContentLoaded', function() {
  var script = document.createElement('script');
  script.type = 'application/ld+json';
  script.textContent = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "name": "Locust Kubernetes Operator",
    "description": "Kubernetes operator for Locust. Run distributed Locust load tests on Kubernetes as a LocustTest custom resource, installed with Helm and built for CI pipelines.",
    "applicationCategory": "DeveloperApplication",
    "applicationSubCategory": "Performance Testing",
    "operatingSystem": "Kubernetes",
    "offers": {
      "@type": "Offer",
      "price": "0",
      "priceCurrency": "USD"
    },
    "author": {
      "@type": "Person",
      "name": "Abdelrhman Hamouda",
      "url": "https://github.com/AbdelrhmanHamouda"
    },
    "codeRepository": "https://github.com/AbdelrhmanHamouda/locust-k8s-operator",
    "programmingLanguage": "Go",
    "license": "https://opensource.org/licenses/Apache-2.0",
    "keywords": ["kubernetes", "locust", "load testing", "performance testing", "operator", "cloud-native", "distributed testing"]
  });
  document.head.appendChild(script);
});

document.addEventListener('DOMContentLoaded', function() {
  var script = document.createElement('script');
  script.type = 'application/ld+json';
  script.textContent = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "name": "Locust Kubernetes Operator",
    "description": "Production-ready Kubernetes operator for Locust distributed load testing. Automate performance testing with cloud-native CI/CD integration, OpenTelemetry observability, and horizontal scaling.",
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

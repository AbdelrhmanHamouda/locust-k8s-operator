document.addEventListener('DOMContentLoaded', function() {
  // Reference to the hero image element
  const heroImage = document.querySelector('.tx-hero__image');
  const header = document.querySelector('.md-header');

  if (!heroImage) return;

  // Initial scale (slightly larger to enable zoom out effect)
  heroImage.style.transform = 'scale(1.05)';

  // Get elements that will transition based on scroll
  const mainTitle = document.querySelector('h1');
  const heroContent = document.querySelector('.tx-hero__content');

  // Handle scroll events
  function handleScroll() {
    // Get current scroll position
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;

    // Calculate opacity and transform based on scroll position
    // Start fading out at 50px of scroll, complete at 300px
    const fadePoint = 50;
    const maxFadeScroll = 300;

    if (scrollTop <= fadePoint) {
      // No fading yet - full opacity
      heroImage.style.opacity = 1;
      heroImage.style.transform = 'scale(1.05)';

      if (mainTitle) {
        mainTitle.style.opacity = 1;
        mainTitle.style.transform = 'translateY(0)';
      }
    } else if (scrollTop <= maxFadeScroll) {
      // Calculate fade percentage
      const fadePercentage = (scrollTop - fadePoint) / (maxFadeScroll - fadePoint);

      // Apply fade and scale effects
      heroImage.style.opacity = Math.max(0, 1 - fadePercentage);
      heroImage.style.transform = `scale(${1.05 - (fadePercentage * 0.05)}) translateY(-${fadePercentage * 30}px)`;

      if (mainTitle) {
        mainTitle.style.opacity = Math.max(0, 1 - fadePercentage * 1.5);
        mainTitle.style.transform = `translateY(-${fadePercentage * 20}px)`;
      }

      if (heroContent) {
        heroContent.style.opacity = Math.max(0, 1 - fadePercentage * 1.2);
      }
    } else {
      // Beyond max fade point - fully faded
      heroImage.style.opacity = 0;
      heroImage.style.transform = 'scale(1) translateY(-30px)';

      if (mainTitle) {
        mainTitle.style.opacity = 0;
        mainTitle.style.transform = 'translateY(-20px)';
      }

      if (heroContent) {
        heroContent.style.opacity = 0;
      }
    }
  }

  // Apply initial styles
  if (mainTitle) {
    mainTitle.style.transition = 'opacity 0.4s ease-out, transform 0.4s ease-out';
  }

  if (heroContent) {
    heroContent.style.transition = 'opacity 0.4s ease-out';
  }

  // Attach scroll event listener
  window.addEventListener('scroll', handleScroll, { passive: true });

  // Call once on load to set initial state
  handleScroll();
});

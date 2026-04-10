document.addEventListener('DOMContentLoaded', function() {
  // Initialize Intersection Observer for scroll-triggered animations
  const observerOptions = {
    threshold: [0.1, 0.3, 0.7],
    rootMargin: '0px 0px -50px 0px'
  };

  // Create intersection observer for elements that should animate on scroll
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('animate-in');
      }
    });
  }, observerOptions);

  // Enhanced parallax and scroll effects
  let ticking = false;

  function handleScroll() {
    if (!ticking) {
      requestAnimationFrame(() => {
        updateScrollEffects();
        ticking = false;
      });
      ticking = true;
    }
  }

  function updateScrollEffects() {
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
    const windowHeight = window.innerHeight;

    // Hero parallax effect
    const heroImage = document.querySelector('.tx-hero__image');
    if (heroImage) {
      const heroRect = heroImage.getBoundingClientRect();
      const heroCenter = heroRect.top + heroRect.height / 2;
      const distanceFromCenter = (windowHeight / 2 - heroCenter) / windowHeight;

      // Subtle parallax movement
      const parallaxOffset = distanceFromCenter * 30;
      heroImage.style.transform = `translateY(${parallaxOffset}px)`;
    }

    // Header background blur effect
    const header = document.querySelector('.md-header');
    if (header && scrollTop > 100) {
      header.style.backgroundColor = 'rgba(var(--md-default-bg-color--rgb), 0.95)';
      header.style.backdropFilter = 'blur(10px)';
    } else if (header) {
      header.style.backgroundColor = '';
      header.style.backdropFilter = '';
    }

    // Floating animation for cards based on scroll position
    const cards = document.querySelectorAll('.grid.cards li');
    cards.forEach((card, index) => {
      const rect = card.getBoundingClientRect();
      const isVisible = rect.top < windowHeight && rect.bottom > 0;

      if (isVisible) {
        const scrollProgress = Math.max(0, Math.min(1, (windowHeight - rect.top) / windowHeight));
        const floatOffset = Math.sin(Date.now() * 0.001 + index * 0.5) * 2;
        card.style.transform = `translateY(${floatOffset}px)`;
      }
    });
  }

  // Smooth hover effects for interactive elements
  function initializeHoverEffects() {
    // Enhanced button hover effects
    const buttons = document.querySelectorAll('.md-button');
    buttons.forEach(button => {
      button.addEventListener('mouseenter', function(e) {
        const rect = e.target.getBoundingClientRect();
        const ripple = document.createElement('span');
        ripple.classList.add('ripple-effect');
        ripple.style.left = '50%';
        ripple.style.top = '50%';
        e.target.appendChild(ripple);

        setTimeout(() => {
          if (ripple.parentNode) {
            ripple.parentNode.removeChild(ripple);
          }
        }, 600);
      });
    });

    // Card tilt effect on mouse move
    const cards = document.querySelectorAll('.grid.cards li');
    cards.forEach(card => {
      card.addEventListener('mousemove', function(e) {
        const rect = e.currentTarget.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        const centerX = rect.width / 2;
        const centerY = rect.height / 2;

        const rotateX = (y - centerY) / centerY * -5;
        const rotateY = (x - centerX) / centerX * 5;

        e.currentTarget.style.transform = `perspective(1000px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) translateY(-8px)`;
      });

      card.addEventListener('mouseleave', function(e) {
        e.currentTarget.style.transform = 'perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(0px)';
      });
    });
  }

  // Initialize typing effect for hero title
  function initializeTypingEffect() {
    const heroTitle = document.querySelector('.tx-hero h1');
    if (heroTitle) {
      const originalText = heroTitle.textContent;
      heroTitle.textContent = '';
      heroTitle.style.borderRight = '2px solid var(--md-primary-fg-color)';

      let charIndex = 0;
      const typingSpeed = 50;

      function typeCharacter() {
        if (charIndex < originalText.length) {
          heroTitle.textContent += originalText.charAt(charIndex);
          charIndex++;
          setTimeout(typeCharacter, typingSpeed);
        } else {
          // Remove cursor after typing is complete
          setTimeout(() => {
            heroTitle.style.borderRight = 'none';
          }, 1000);
        }
      }

      // Start typing after a short delay
      setTimeout(typeCharacter, 500);
    }
  }

  // Initialize counter animations for stats (if any)
  function initializeCounterAnimations() {
    const counters = document.querySelectorAll('[data-counter]');
    const counterObserver = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const counter = entry.target;
          const target = parseInt(counter.getAttribute('data-counter'));
          let current = 0;
          const increment = target / 100;
          const timer = setInterval(() => {
            current += increment;
            counter.textContent = Math.floor(current);
            if (current >= target) {
              counter.textContent = target;
              clearInterval(timer);
            }
          }, 20);
          counterObserver.unobserve(counter);
        }
      });
    });

    counters.forEach(counter => counterObserver.observe(counter));
  }

  // Initialize all effects
  function initialize() {
    // Observe elements for scroll animations
    const animatedElements = document.querySelectorAll('.grid, .tx-hero__content, h2');
    animatedElements.forEach(el => observer.observe(el));

    // Initialize various effects
    initializeHoverEffects();
    initializeCounterAnimations();

    // Add scroll listener
    window.addEventListener('scroll', handleScroll, { passive: true });

    // Initial call
    updateScrollEffects();

    // Add CSS for dynamic animations
    const style = document.createElement('style');
    style.textContent = `
      .ripple-effect {
        position: absolute;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.6);
        width: 4px;
        height: 4px;
        animation: ripple 0.6s linear;
        pointer-events: none;
      }

      @keyframes ripple {
        to {
          width: 100px;
          height: 100px;
          margin-left: -50px;
          margin-top: -50px;
          opacity: 0;
        }
      }

      .animate-in {
        animation-play-state: running !important;
      }

      @media (prefers-reduced-motion: reduce) {
        *, *::before, *::after {
          animation-duration: 0.01ms !important;
          animation-iteration-count: 1 !important;
          transition-duration: 0.01ms !important;
        }
      }
    `;
    document.head.appendChild(style);
  }

  // Initialize everything
  initialize();

  // Add page visibility change handling
  document.addEventListener('visibilitychange', function() {
    if (document.visibilityState === 'visible') {
      updateScrollEffects();
    }
  });
});

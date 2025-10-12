// Real Staging AI Documentation - Custom JavaScript

document.addEventListener('DOMContentLoaded', function() {
  // Add smooth scroll behavior for anchor links
  document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
      const target = document.querySelector(this.getAttribute('href'));
      if (target) {
        e.preventDefault();
        target.scrollIntoView({
          behavior: 'smooth',
          block: 'start'
        });
      }
    });
  });

  // Add copy feedback to code blocks
  document.querySelectorAll('.md-clipboard').forEach(button => {
    button.addEventListener('click', function() {
      const originalTitle = this.title;
      this.title = 'Copied!';
      setTimeout(() => {
        this.title = originalTitle;
      }, 2000);
    });
  });

  // Add external link indicators
  document.querySelectorAll('a[href^="http"]').forEach(link => {
    if (!link.hostname.includes('real-staging.ai') && 
        !link.hostname.includes('localhost')) {
      link.setAttribute('target', '_blank');
      link.setAttribute('rel', 'noopener noreferrer');
    }
  });
});

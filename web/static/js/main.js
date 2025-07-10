// Rick & Morty Portal Effects
document.addEventListener('DOMContentLoaded', function() {
    // Add random particle effects
    createParticles();
    
    // Add portal sound effects on hover (mock)
    const buttons = document.querySelectorAll('.btn');
    buttons.forEach(button => {
        button.addEventListener('mouseenter', function() {
            // Could add sound effects here
            this.style.transform = 'translateY(-2px) scale(1.05)';
        });
        
        button.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0) scale(1)';
        });
    });

    // Animate elements on scroll
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    });

    const features = document.querySelectorAll('.feature');
    features.forEach(feature => {
        feature.style.opacity = '0';
        feature.style.transform = 'translateY(20px)';
        feature.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(feature);
    });
});

function createParticles() {
    const particleContainer = document.createElement('div');
    particleContainer.className = 'particles-container';
    particleContainer.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        pointer-events: none;
        z-index: -1;
    `;
    document.body.appendChild(particleContainer);

    for (let i = 0; i < 20; i++) {
        createParticle(particleContainer);
    }
}

function createParticle(container) {
    const particle = document.createElement('div');
    particle.style.cssText = `
        position: absolute;
        width: 3px;
        height: 3px;
        background: #97ce4c;
        border-radius: 50%;
        opacity: 0.6;
        animation: float ${Math.random() * 3 + 2}s ease-in-out infinite;
    `;
    
    particle.style.left = Math.random() * 100 + '%';
    particle.style.top = Math.random() * 100 + '%';
    particle.style.animationDelay = Math.random() * 2 + 's';
    
    container.appendChild(particle);

    // Remove particle after animation
    setTimeout(() => {
        if (particle.parentNode) {
            particle.parentNode.removeChild(particle);
        }
    }, 5000);
}

// Add CSS for particle animation
const style = document.createElement('style');
style.textContent = `
    @keyframes float {
        0%, 100% { 
            transform: translateY(0px) translateX(0px); 
            opacity: 0.6; 
        }
        25% { 
            transform: translateY(-20px) translateX(10px); 
            opacity: 1; 
        }
        50% { 
            transform: translateY(-10px) translateX(-10px); 
            opacity: 0.8; 
        }
        75% { 
            transform: translateY(-30px) translateX(5px); 
            opacity: 0.4; 
        }
    }
`;
document.head.appendChild(style); 
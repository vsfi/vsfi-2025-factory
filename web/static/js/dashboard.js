// Dashboard functionality for Plumbus generation
document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('plumbus-form');
    const progressContainer = document.getElementById('progress-container');
    const progressFill = document.querySelector('.progress-fill');
    const progressText = document.querySelector('.progress-text');
    const plumbusGrid = document.getElementById('plumbus-grid');

    // Form submission
    form.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const formData = new FormData(form);
        const plumbusData = {
            name: formData.get('name'),
            size: formData.get('size'),
            color: formData.get('color'),
            shape: formData.get('shape'),
            weight: formData.get('weight'),
            wrapping: formData.get('wrapping')
        };

        // Validate form
        if (!validateForm(plumbusData)) {
            showNotification('Пожалуйста, заполните все поля!', 'error');
            return;
        }

        try {
            // Hide form and show progress
            form.style.display = 'none';
            progressContainer.style.display = 'block';
            
            // Start progress animation
            startProgressAnimation();
            
            // Send request to generate plumbus
            const response = await fetch('/plumbus/generate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(plumbusData)
            });

            if (!response.ok) {
                throw new Error('Ошибка при создании плюмбуса');
            }

            const result = await response.json();
            
            // Monitor progress
            monitorPlumbusGeneration(result.id);
            
        } catch (error) {
            console.error('Error:', error);
            showNotification('Ошибка при создании плюмбуса: ' + error.message, 'error');
            resetForm();
        }
    });

    // Validate form data
    function validateForm(data) {
        return Object.values(data).every(value => value && value.trim() !== '');
    }

    // Start progress animation
    function startProgressAnimation() {
        let progress = 0;
        const interval = setInterval(() => {
            progress += Math.random() * 10;
            if (progress >= 95) {
                progress = 95; // Keep at 95% until we get real status
                clearInterval(interval);
            }
            updateProgress(progress);
        }, 500);
        
        // Store interval reference for cleanup
        progressContainer.dataset.interval = interval;
    }

    // Update progress bar
    function updateProgress(percentage) {
        progressFill.style.width = percentage + '%';
        progressText.textContent = Math.round(percentage) + '%';
    }

    // Monitor plumbus generation status
    async function monitorPlumbusGeneration(plumbusId) {
        const maxAttempts = 60; // 5 minutes max
        let attempts = 0;

        const checkStatus = async () => {
            try {
                const response = await fetch(`/plumbus/status/${plumbusId}`);
                const status = await response.json();

                if (status.status === 'completed') {
                    updateProgress(100);
                    setTimeout(() => {
                        if (status.is_rare) {
                            showNotification('🌟 НЕВЕРОЯТНО! Вы создали МЕГА РЕДКИЙ плюмбус! ✨🎉', 'rare');
                        } else {
                            showNotification('Плюмбус успешно создан! 🎉', 'success');
                        }
                        addNewPlumbusCard(status);
                        resetForm();
                    }, 1000);
                    return;
                }

                if (status.status === 'failed') {
                    showNotification('Ошибка при генерации плюмбуса 😞', 'error');
                    resetForm();
                    return;
                }

                // Continue monitoring
                attempts++;
                if (attempts < maxAttempts) {
                    setTimeout(checkStatus, 5000); // Check every 5 seconds
                } else {
                    showNotification('Превышено время ожидания генерации', 'error');
                    resetForm();
                }

            } catch (error) {
                console.error('Error checking status:', error);
                attempts++;
                if (attempts < maxAttempts) {
                    setTimeout(checkStatus, 5000);
                } else {
                    showNotification('Ошибка при проверке статуса', 'error');
                    resetForm();
                }
            }
        };

        // Start checking after 2 seconds
        setTimeout(checkStatus, 2000);
    }

    // Add new plumbus card to grid
    function addNewPlumbusCard(plumbusData) {
        // Получаем данные из формы для отображения
        const formData = new FormData(form);
        const rareClass = plumbusData.is_rare ? ' rare' : '';
        const rareBadge = plumbusData.is_rare ? '<span class="rare-badge" title="Мега редкий плюмбус!">✨</span>' : '';
        
        // Формируем блок с информацией о подписи
        let signatureBlock = '';
        if (plumbusData.signature) {
            const signatureDate = plumbusData.signature_date ? 
                new Date(plumbusData.signature_date).toLocaleString('ru-RU') : '';
            
            signatureBlock = `
                <div class="signature-info">
                    <p><strong>🔒 Подпись:</strong> <span class="signature-hash" title="${plumbusData.signature}" data-full-signature="${plumbusData.signature}">${plumbusData.signature}</span></p>
                    ${signatureDate ? `<p><strong>📅 Подписано:</strong> <span class="signature-date">${signatureDate}</span></p>` : ''}
                    <div class="signature-verified">✅ Подлинность подтверждена</div>
                </div>
            `;
        }
        
        const cardHTML = `
            <div class="plumbus-card${rareClass}" data-status="completed">
                <div class="card-header">
                    <h3>${plumbusData.name || formData.get('name')}${rareBadge}</h3>
                    <span class="status status-completed">completed</span>
                </div>
                <div class="card-content">
                    <img src="/plumbus/image/${plumbusData.id}" alt="${plumbusData.name || formData.get('name')}" class="plumbus-image clickable-image" onclick="openImageModal('/plumbus/image/${plumbusData.id}', '${plumbusData.name || formData.get('name')}')">
                </div>
                <div class="card-details">
                    <p><strong>Размер:</strong> ${getSizeLabel(formData.get('size')) || 'Не указан'}</p>
                    <p><strong>Цвет:</strong> ${getColorLabel(formData.get('color')) || 'Не указан'}</p>
                    <p><strong>Форма:</strong> ${getShapeLabel(formData.get('shape')) || 'Не указана'}</p>
                    ${signatureBlock}
                </div>
            </div>
        `;
        
        plumbusGrid.insertAdjacentHTML('afterbegin', cardHTML);
        
        // Animate new card
        const newCard = plumbusGrid.firstElementChild;
        newCard.style.opacity = '0';
        newCard.style.transform = 'translateY(20px)';
        
        setTimeout(() => {
            newCard.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
            newCard.style.opacity = '1';
            newCard.style.transform = 'translateY(0)';
            
            // Setup signature hash click handler for new card
            const signatureHash = newCard.querySelector('.signature-hash');
            if (signatureHash) {
                signatureHash.addEventListener('click', function() {
                    const fullSignature = this.dataset.fullSignature;
                    if (fullSignature) {
                        showSignatureModal(fullSignature);
                    }
                });
            }
        }, 100);
    }

    // Reset form and progress
    function resetForm() {
        // Clear interval if exists
        const interval = progressContainer.dataset.interval;
        if (interval) {
            clearInterval(parseInt(interval));
        }
        
        // Reset UI
        form.style.display = 'block';
        progressContainer.style.display = 'none';
        form.reset();
        updateProgress(0);
    }

    // Show notification
    function showNotification(message, type = 'info') {
        // Remove existing notification
        const existing = document.querySelector('.notification');
        if (existing) {
            existing.remove();
        }

        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        
        const style = {
            position: 'fixed',
            top: '20px',
            right: '20px',
            padding: '15px 20px',
            borderRadius: '8px',
            color: 'white',
            fontWeight: 'bold',
            zIndex: '1000',
            animation: 'slideIn 0.3s ease',
            maxWidth: '300px'
        };

        // Set background color based on type
        if (type === 'success') {
            style.background = 'var(--success-green)';
        } else if (type === 'error') {
            style.background = 'var(--danger-red)';
        } else if (type === 'rare') {
            style.background = 'linear-gradient(45deg, var(--rare-yellow), #FFA500)';
            style.border = '2px solid var(--rare-yellow)';
            style.boxShadow = '0 0 20px rgba(255, 215, 0, 0.5)';
            style.animation = 'slideIn 0.3s ease, rareGlow 2s ease-in-out infinite';
        } else {
            style.background = 'var(--secondary-blue)';
        }

        Object.assign(notification.style, style);
        document.body.appendChild(notification);

        // Auto remove after 5 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => notification.remove(), 300);
        }, 5000);
    }

    // Add CSS for notifications
    const style = document.createElement('style');
    style.textContent = `
        @keyframes slideIn {
            from { transform: translateX(100%); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }
        
        @keyframes slideOut {
            from { transform: translateX(0); opacity: 1; }
            to { transform: translateX(100%); opacity: 0; }
        }
        
        @keyframes rareGlow {
            0%, 100% {
                box-shadow: 0 0 20px rgba(255, 215, 0, 0.5);
            }
            50% {
                box-shadow: 0 0 40px rgba(255, 215, 0, 0.8), 0 0 60px rgba(255, 215, 0, 0.3);
            }
        }

        .notification {
            box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
            cursor: pointer;
        }

        .notification:hover {
            transform: scale(1.02);
        }
    `;
    document.head.appendChild(style);

    // Click to dismiss notification
    document.addEventListener('click', function(e) {
        if (e.target.classList.contains('notification')) {
            e.target.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => e.target.remove(), 300);
        }
    });

    // Update existing cards with generating status
    const generatingCards = document.querySelectorAll('[data-status="generating"]');
    generatingCards.forEach(card => {
        const cardId = extractIdFromCard(card);
        if (cardId) {
            monitorExistingPlumbus(cardId, card);
        }
    });

    // Extract plumbus ID from card (assuming it's in an image src or data attribute)
    function extractIdFromCard(card) {
        // Try to find ID in image src
        const img = card.querySelector('img');
        if (img && img.src) {
            const match = img.src.match(/\/plumbus\/image\/([^\/]+)/);
            if (match) return match[1];
        }
        
        // Could also check data attributes if implemented
        return card.dataset.plumbusId;
    }

    // Monitor existing plumbus that are generating
    async function monitorExistingPlumbus(plumbusId, cardElement) {
        const maxAttempts = 60;
        let attempts = 0;

        const checkStatus = async () => {
            try {
                const response = await fetch(`/plumbus/status/${plumbusId}`);
                const status = await response.json();

                if (status.status === 'completed') {
                    updateCardToCompleted(cardElement, status);
                    return;
                }

                if (status.status === 'failed') {
                    updateCardToFailed(cardElement);
                    return;
                }

                attempts++;
                if (attempts < maxAttempts) {
                    setTimeout(checkStatus, 5000);
                }

            } catch (error) {
                console.error('Error checking existing plumbus status:', error);
                attempts++;
                if (attempts < maxAttempts) {
                    setTimeout(checkStatus, 5000);
                }
            }
        };

        setTimeout(checkStatus, 2000);
    }

    // Update card to completed status
    function updateCardToCompleted(cardElement, status) {
        const statusSpan = cardElement.querySelector('.status');
        const cardContent = cardElement.querySelector('.card-content');
        const cardTitle = cardElement.querySelector('.card-header h3');
        const cardDetails = cardElement.querySelector('.card-details');
        
        cardElement.dataset.status = 'completed';
        statusSpan.className = 'status status-completed';
        statusSpan.textContent = 'completed';
        
        // Проверяем и добавляем класс rare если плюмбус редкий
        if (status.is_rare) {
            cardElement.classList.add('rare');
            // Добавляем значок редкости к названию если его еще нет
            if (!cardTitle.querySelector('.rare-badge')) {
                cardTitle.innerHTML += '<span class="rare-badge" title="Мега редкий плюмбус!">✨</span>';
            }
        }
        
        cardContent.innerHTML = `<img src="/plumbus/image/${status.id}" alt="${status.name}" class="plumbus-image clickable-image" onclick="openImageModal('/plumbus/image/${status.id}', '${status.name}')">`;
        
        // Добавляем информацию о подписи если есть
        if (status.signature) {
            const signatureDate = status.signature_date ? 
                new Date(status.signature_date).toLocaleString('ru-RU') : '';
            
            const signatureBlock = `
                <div class="signature-info">
                    <p><strong>🔒 Подпись:</strong> <span class="signature-hash" title="${status.signature}" data-full-signature="${status.signature}">${status.signature}</span></p>
                    ${signatureDate ? `<p><strong>📅 Подписано:</strong> <span class="signature-date">${signatureDate}</span></p>` : ''}
                    <div class="signature-verified">✅ Подлинность подтверждена</div>
                </div>
            `;
            
            // Добавляем блок подписи в конец card-details
            cardDetails.insertAdjacentHTML('beforeend', signatureBlock);
            
            // Setup click handler for the new signature hash
            const newSignatureHash = cardDetails.querySelector('.signature-hash:last-of-type');
            if (newSignatureHash) {
                newSignatureHash.addEventListener('click', function() {
                    showSignatureModal(status.signature);
                });
            }
        }
        
        // Add glow effect
        cardElement.style.animation = 'completedGlow 2s ease-in-out';
    }

    // Update card to failed status
    function updateCardToFailed(cardElement) {
        const statusSpan = cardElement.querySelector('.status');
        const cardContent = cardElement.querySelector('.card-content');
        
        cardElement.dataset.status = 'failed';
        statusSpan.className = 'status status-failed';
        statusSpan.textContent = 'failed';
        
        cardContent.innerHTML = '<div class="error-message">Ошибка генерации</div>';
    }

    // Add completion glow animation
    const glowStyle = document.createElement('style');
    glowStyle.textContent = `
        @keyframes completedGlow {
            0%, 100% { box-shadow: 0 0 5px rgba(151, 206, 76, 0.3); }
            50% { box-shadow: 0 0 20px rgba(151, 206, 76, 0.8); }
        }
    `;
    document.head.appendChild(glowStyle);

    // Translate existing card values to Russian
    translateExistingCards();

    // Add portal particles effect
    createDashboardParticles();
});

// Create dashboard particles
function createDashboardParticles() {
    const particleContainer = document.createElement('div');
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

    // Create fewer particles for dashboard
    for (let i = 0; i < 10; i++) {
        setTimeout(() => createDashboardParticle(particleContainer), i * 500);
    }
}

function createDashboardParticle(container) {
    const particle = document.createElement('div');
    particle.style.cssText = `
        position: absolute;
        width: 2px;
        height: 2px;
        background: #97ce4c;
        border-radius: 50%;
        opacity: 0.4;
        animation: dashboardFloat ${Math.random() * 4 + 3}s ease-in-out infinite;
    `;
    
    particle.style.left = Math.random() * 100 + '%';
    particle.style.top = Math.random() * 100 + '%';
    particle.style.animationDelay = Math.random() * 3 + 's';
    
    container.appendChild(particle);

    setTimeout(() => {
        if (particle.parentNode) {
            particle.parentNode.removeChild(particle);
            // Create new particle to maintain count
            setTimeout(() => createDashboardParticle(container), Math.random() * 2000);
        }
    }, 7000);
}

// Add dashboard float animation
const dashboardStyle = document.createElement('style');
dashboardStyle.textContent = `
    @keyframes dashboardFloat {
        0%, 100% { 
            transform: translateY(0px) translateX(0px); 
            opacity: 0.4; 
        }
        33% { 
            transform: translateY(-15px) translateX(8px); 
            opacity: 0.8; 
        }
        66% { 
            transform: translateY(-8px) translateX(-8px); 
            opacity: 0.6; 
        }
    }
`;
document.head.appendChild(dashboardStyle);

// Functions for translating form values to readable labels
function getColorLabel(color) {
    const colorMap = {
        'pink': 'Розовый',
        'deep_pink': 'Тёмно-розовый',
        'red': 'Красный',
        'blue': 'Синий',
        'green': 'Зелёный',
        'yellow': 'Жёлтый',
        'purple': 'Фиолетовый',
        'orange': 'Оранжевый',
        'cyan': 'Циан',
        'lime': 'Лайм',
        'teal': 'Бирюзовый',
        'brown': 'Коричневый'
    };
    return colorMap[color] || color;
}

function getShapeLabel(shape) {
    const shapeMap = {
        'smooth': 'Гладкая',
        'uglovatiy': 'Угловатая',
        'multi-uglovatiy': 'Мульти-угловатая'
    };
    return shapeMap[shape] || shape;
}

function getSizeLabel(size) {
    const sizeMap = {
        'nano': 'Нано',
        'XS': 'XS',
        'S': 'S',
        'M': 'M',
        'L': 'L',
        'XL': 'XL',
        'XXL': 'XXL'
    };
    return sizeMap[size] || size;
}

// Modal functions for fullscreen image view
function openImageModal(imageSrc, imageName) {
    const modal = document.getElementById('imageModal');
    const modalImage = document.getElementById('modalImage');
    const modalCaption = document.getElementById('modalCaption');
    
    modal.style.display = 'block';
    modalImage.src = imageSrc;
    modalCaption.textContent = imageName;
    
    // Prevent body scroll when modal is open
    document.body.style.overflow = 'hidden';
}

function closeImageModal() {
    const modal = document.getElementById('imageModal');
    modal.style.display = 'none';
    
    // Restore body scroll
    document.body.style.overflow = 'auto';
}

// Close modal on Escape key
document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
        closeImageModal();
    }
});

// Prevent modal close when clicking on image
document.getElementById('modalImage').addEventListener('click', function(event) {
    event.stopPropagation();
});

// Function to translate existing card values
function translateExistingCards() {
    // Translate sizes
    document.querySelectorAll('.card-size').forEach(element => {
        const size = element.textContent.trim();
        if (size && size !== 'Не указан') {
            element.textContent = getSizeLabel(size);
        }
    });

    // Translate colors
    document.querySelectorAll('.card-color').forEach(element => {
        const color = element.dataset.color || element.textContent.trim();
        if (color && color !== 'Не указан') {
            element.textContent = getColorLabel(color);
        }
    });

    // Translate shapes
    document.querySelectorAll('.card-shape').forEach(element => {
        const shape = element.dataset.shape || element.textContent.trim();
        if (shape && shape !== 'Не указана') {
            element.textContent = getShapeLabel(shape);
        }
    });

    // Setup signature hash click handlers
    setupSignatureHashHandlers();
}

// Setup click handlers for signature hashes
function setupSignatureHashHandlers() {
    document.querySelectorAll('.signature-hash').forEach(element => {
        element.addEventListener('click', function() {
            const fullSignature = this.dataset.fullSignature || this.getAttribute('title');
            if (fullSignature) {
                showSignatureModal(fullSignature);
            }
        });
    });
}

// Show signature in modal dialog
function showSignatureModal(signature) {
    const modal = document.createElement('div');
    modal.className = 'signature-modal';
    modal.innerHTML = `
        <div class="signature-modal-content">
            <div class="signature-modal-header">
                <h3>🔒 Полная цифровая подпись</h3>
                <span class="signature-modal-close">&times;</span>
            </div>
            <div class="signature-modal-body">
                <div class="signature-full-text">${signature}</div>
                <div class="signature-actions">
                    <button class="btn btn-secondary" onclick="copySignatureToClipboard('${signature}')">
                        📋 Скопировать
                    </button>
                </div>
            </div>
        </div>
    `;

    // Add modal styles
    const style = document.createElement('style');
    style.textContent = `
        .signature-modal {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.8);
            z-index: 2000;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: fadeIn 0.3s ease;
        }
        .signature-modal-content {
            background: var(--lab-gray);
            border: 2px solid var(--primary-green);
            border-radius: 15px;
            padding: 20px;
            max-width: 80%;
            max-height: 80%;
            overflow: auto;
            position: relative;
        }
        .signature-modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 1px solid var(--primary-green);
        }
        .signature-modal-header h3 {
            color: var(--primary-green);
            margin: 0;
        }
        .signature-modal-close {
            font-size: 24px;
            color: var(--primary-green);
            cursor: pointer;
            transition: all 0.3s ease;
        }
        .signature-modal-close:hover {
            color: var(--portal-green);
            transform: scale(1.2);
        }
        .signature-full-text {
            font-family: 'Courier New', monospace;
            background: rgba(0, 0, 0, 0.3);
            padding: 15px;
            border-radius: 8px;
            word-break: break-all;
            font-size: 0.9rem;
            line-height: 1.4;
            color: var(--text-light);
            margin-bottom: 15px;
            border: 1px solid var(--primary-green);
        }
        .signature-actions {
            text-align: center;
        }
    `;
    document.head.appendChild(style);

    document.body.appendChild(modal);

    // Close handlers
    modal.querySelector('.signature-modal-close').addEventListener('click', () => {
        modal.remove();
        style.remove();
    });

    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.remove();
            style.remove();
        }
    });

    // Close on Escape
    const escapeHandler = (e) => {
        if (e.key === 'Escape') {
            modal.remove();
            style.remove();
            document.removeEventListener('keydown', escapeHandler);
        }
    };
    document.addEventListener('keydown', escapeHandler);
}

// Copy signature to clipboard
function copySignatureToClipboard(signature) {
    navigator.clipboard.writeText(signature).then(() => {
        showNotification('🔒 Подпись скопирована в буфер обмена!', 'success');
    }).catch(err => {
        console.error('Failed to copy signature:', err);
        showNotification('❌ Не удалось скопировать подпись', 'error');
    });
} 
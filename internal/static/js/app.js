// ===== ОБЩИЕ ФУНКЦИИ =====

// Показ уведомления
function showNotification(message, type = 'success') {
    const notification = document.createElement('div');
    notification.className = `message ${type}`;
    notification.textContent = message;
    notification.style.position = 'fixed';
    notification.style.top = '20px';
    notification.style.right = '20px';
    notification.style.zIndex = '1000';
    notification.style.padding = '15px 25px';
    notification.style.borderRadius = '10px';
    notification.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.opacity = '0';
        notification.style.transition = 'opacity 0.5s';
        setTimeout(() => notification.remove(), 500);
    }, 3000);
}

// Форматирование процентов
function formatPercent(correct, total) {
    if (total === 0) return '0%';
    return Math.round((correct / total) * 100) + '%';
}

// Форматирование даты
function formatDate(dateString) {
    if (!dateString) return '—';
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric'
    });
}

// Валидация числа
function isValidNumber(input) {
    return /^-?\d+$/.test(input.trim());
}

// Фокус на первом поле ввода
function focusFirstInput() {
    const firstInput = document.querySelector('input[type="text"]');
    if (firstInput && !firstInput.disabled) {
        firstInput.focus();
    }
}

// ===== ОБРАБОТКА КЛАВИАТУРЫ =====
document.addEventListener('keydown', function(e) {
    // Escape - очистка поля
    if (e.key === 'Escape') {
        const activeInput = document.activeElement;
        if (activeInput.type === 'text') {
            activeInput.value = '';
        }
    }
    
    // Ctrl+Enter - проверка всех ответов
    if (e.ctrlKey && e.key === 'Enter') {
        const checkAllBtn = document.getElementById('check-all-button');
        if (checkAllBtn) {
            checkAllBtn.click();
        }
    }
});

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    focusFirstInput();
    
    // Анимация появления элементов
    const cards = document.querySelectorAll('.card, .type-item, .summary-card');
    cards.forEach((card, index) => {
        card.style.opacity = '0';
        card.style.transform = 'translateY(20px)';
        
        setTimeout(() => {
            card.style.transition = 'opacity 0.5s, transform 0.5s';
            card.style.opacity = '1';
            card.style.transform = 'translateY(0)';
        }, 100 * index);
    });
});
// Custom Announcement System with Working Icons
class CustomAnnouncement {
    constructor(options = {}) {
        this.options = {
            type: 'badge',
            message: 'Discover the newest features in our latest release!',
            link: 'https://github.com/AbdelrhmanHamouda/locust-k8s-operator/releases',
            linkText: 'Explore Release',
            storageKey: 'announcement-v1',
            ...options
        };
        this.init();
    }

    init() {
        if (localStorage.getItem(this.options.storageKey)) return;

        if (this.options.type === 'badge') {
            this.createBadge();
        } else {
            this.createModal();
        }
    }

    createBadge() {
        const badge = document.createElement('div');
        badge.className = 'custom-announcement-badge';
        badge.innerHTML = `
            <div class="announcement-content" data-clickable="true">
                <div class="announcement-icon">🚀</div>
                <div class="announcement-text">
                    <span class="announcement-message">${this.options.message}</span>
                </div>
                <button class="announcement-close">&times;</button>
            </div>
        `;
        document.body.appendChild(badge);

        // Add click handler for navigation
        const content = badge.querySelector('.announcement-content');
        content.onclick = (e) => {
            if (!e.target.closest('.announcement-close')) {
                window.open(this.options.link, '_blank');
            }
        };

        // Add close functionality
        badge.querySelector('.announcement-close').onclick = (e) => {
            e.stopPropagation();
            this.dismiss(badge);
        };

        setTimeout(() => badge.classList.add('show'), 100);
    }

    createModal() {
        const modal = document.createElement('div');
        modal.className = 'custom-announcement-modal';
        modal.innerHTML = `
            <div class="announcement-modal-overlay">
                <div class="announcement-modal-content">
                    <div class="announcement-modal-header">
                        <div class="announcement-modal-icon">
                            <svg width="32" height="32" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M19,11H17.5C17.5,7.96 15.04,5.5 12,5.5C8.96,5.5 6.5,7.96 6.5,11H5C5,7.13 8.13,4 12,4C15.87,4 19,7.13 19,11M12,2A2,2 0 0,1 14,4A2,2 0 0,1 12,6A2,2 0 0,1 10,4A2,2 0 0,1 12,2M21,9V7L18.5,7.5L17.5,5.5L19,3.5L17,1.5L15,3.5L16,5.5L14.5,6L15,8.5L17.5,9.5L21,9M12,13.5A7.5,7.5 0 0,1 4.5,21V20A6.5,6.5 0 0,1 11,13.5A1.5,1.5 0 0,1 12.5,15A1.5,1.5 0 0,1 11,16.5A6.5,6.5 0 0,1 4.5,23H2.5A8.5,8.5 0 0,1 11,14.5A2.5,2.5 0 0,1 13.5,17A2.5,2.5 0 0,1 11,19.5A8.5,8.5 0 0,1 2.5,28"/>
                            </svg>
                        </div>
                        <button class="announcement-modal-close">&times;</button>
                    </div>
                    <div class="announcement-modal-body">
                        <h3>Latest Update</h3>
                        <p>${this.options.message}</p>
                        <a href="${this.options.link}" target="_blank" class="announcement-modal-button">${this.options.linkText}</a>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
        modal.querySelector('.announcement-modal-close').onclick = () => this.dismiss(modal);
        modal.querySelector('.announcement-modal-overlay').onclick = (e) => {
            if (e.target === e.currentTarget) this.dismiss(modal);
        };
        setTimeout(() => modal.classList.add('show'), 100);
    }

    dismiss(element) {
        element.classList.add('dismissing');
        setTimeout(() => {
            element.remove();
            localStorage.setItem(this.options.storageKey, 'true');
        }, 300);
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    console.log('Custom announcement script loaded!');

    // Clear old storage keys for testing
    localStorage.removeItem('release-modal-v2');
    localStorage.removeItem('release-modal-v1');

    new CustomAnnouncement({
        type: 'badge',
        message: 'New release available!',
        link: 'https://github.com/AbdelrhmanHamouda/locust-k8s-operator/releases',
        linkText: 'Explore Release',
        storageKey: 'release-modal-v3' // New key to force show
    });
});

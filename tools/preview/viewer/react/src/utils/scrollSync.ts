// Scroll synchronization utilities for editor-preview coordination

export interface ScrollSyncOptions {
    preservePosition?: boolean;
    smoothScroll?: boolean;
    throttleMs?: number;
}

// Calculate scroll position as a percentage of total scrollable height
export function getScrollPercentage(element: HTMLElement): number {
    const { scrollTop, scrollHeight, clientHeight } = element;
    const maxScroll = scrollHeight - clientHeight;
    
    if (maxScroll <= 0) return 0;
    return Math.min(scrollTop / maxScroll, 1);
}

// Set scroll position based on percentage
export function setScrollPercentage(element: HTMLElement, percentage: number): void {
    const { scrollHeight, clientHeight } = element;
    const maxScroll = scrollHeight - clientHeight;
    
    if (maxScroll > 0) {
        element.scrollTop = percentage * maxScroll;
    }
}

// Enhanced scroll position preservation that accounts for content changes
export class ScrollPositionManager {
    private lastScrollTop = 0;
    private lastScrollHeight = 0;
    private lastContentLength = 0;
    private isRestoring = false;

    savePosition(element: HTMLElement, contentLength: number = 0): void {
        if (this.isRestoring) return;
        
        this.lastScrollTop = element.scrollTop;
        this.lastScrollHeight = element.scrollHeight;
        this.lastContentLength = contentLength;
    }

    restorePosition(element: HTMLElement, contentLength: number = 0, options: ScrollSyncOptions = {}): void {
        if (this.isRestoring) return;
        
        this.isRestoring = true;
        
        const restoreScroll = () => {
            // If content length changed significantly, use percentage-based restoration
            const contentLengthRatio = this.lastContentLength > 0 
                ? contentLength / this.lastContentLength 
                : 1;
            
            let newScrollTop: number;
            
            if (Math.abs(contentLengthRatio - 1) > 0.1) {
                // Significant content change - use percentage-based approach
                const percentage = this.lastScrollHeight > 0 
                    ? this.lastScrollTop / (this.lastScrollHeight - element.clientHeight)
                    : 0;
                newScrollTop = percentage * (element.scrollHeight - element.clientHeight);
            } else {
                // Minor content change - use height ratio approach
                const heightRatio = this.lastScrollHeight > 0 
                    ? element.scrollHeight / this.lastScrollHeight 
                    : 1;
                newScrollTop = this.lastScrollTop * heightRatio;
            }
            
            // Ensure scroll position is within bounds
            const maxScroll = element.scrollHeight - element.clientHeight;
            newScrollTop = Math.max(0, Math.min(newScrollTop, maxScroll));
            
            if (options.smoothScroll) {
                element.scrollTo({
                    top: newScrollTop,
                    behavior: 'smooth'
                });
            } else {
                element.scrollTop = newScrollTop;
            }
            
            this.isRestoring = false;
        };

        // Use requestAnimationFrame to ensure DOM is fully updated
        requestAnimationFrame(() => {
            requestAnimationFrame(restoreScroll);
        });
    }

    isCurrentlyRestoring(): boolean {
        return this.isRestoring;
    }
}

// Throttle function for performance
export function throttle<T extends (...args: unknown[]) => void>(
    func: T,
    delay: number
): (...args: Parameters<T>) => void {
    let timeoutId: number | null = null;
    let lastExecTime = 0;
    
    return (...args: Parameters<T>) => {
        const currentTime = Date.now();
        
        if (currentTime - lastExecTime > delay) {
            func(...args);
            lastExecTime = currentTime;
        } else {
            if (timeoutId) clearTimeout(timeoutId);
            timeoutId = window.setTimeout(() => {
                func(...args);
                lastExecTime = Date.now();
            }, delay - (currentTime - lastExecTime));
        }
    };
}
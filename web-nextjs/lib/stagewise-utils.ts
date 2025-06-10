/**
 * Stagewise utility functions for production optimization
 */

export interface StagewiseCleanupConfig {
  maxAge?: number; // Maximum age in milliseconds
  maxEntries?: number; // Maximum number of entries to keep
  cleanupInterval?: number; // Cleanup interval in milliseconds
}

export class StagewiseManager {
  private cleanupTimer?: NodeJS.Timeout;
  private config: Required<StagewiseCleanupConfig>;

  constructor(config: StagewiseCleanupConfig = {}) {
    this.config = {
      maxAge: config.maxAge || 5 * 60 * 1000, // 5 minutes
      maxEntries: config.maxEntries || 100,
      cleanupInterval: config.cleanupInterval || 60 * 1000, // 1 minute
    };
  }

  /**
   * Start automatic cleanup
   */
  startCleanup() {
    if (this.cleanupTimer) {
      return;
    }

    this.cleanupTimer = setInterval(() => {
      this.performCleanup();
    }, this.config.cleanupInterval);
  }

  /**
   * Stop automatic cleanup
   */
  stopCleanup() {
    if (this.cleanupTimer) {
      clearInterval(this.cleanupTimer);
      this.cleanupTimer = undefined;
    }
  }

  /**
   * Perform cleanup of old stagewise data
   */
  private performCleanup() {
    try {
      // Clean up localStorage data
      this.cleanupLocalStorage();
      
      // Clean up sessionStorage data
      this.cleanupSessionStorage();
      
      console.log('🧹 Stagewise cleanup completed');
    } catch (error) {
      console.warn('⚠️ Stagewise cleanup failed:', error);
    }
  }

  /**
   * Clean up localStorage stagewise data
   */
  private cleanupLocalStorage() {
    if (typeof window === 'undefined' || !window.localStorage) {
      return;
    }

    const keys = Object.keys(localStorage);
    const stagewiseKeys = keys.filter(key => key.startsWith('stagewise_'));
    
    stagewiseKeys.forEach(key => {
      try {
        const data = JSON.parse(localStorage.getItem(key) || '{}');
        const timestamp = data.timestamp || 0;
        const age = Date.now() - timestamp;
        
        if (age > this.config.maxAge) {
          localStorage.removeItem(key);
        }
      } catch (error) {
        // Remove invalid data
        localStorage.removeItem(key);
      }
    });
  }

  /**
   * Clean up sessionStorage stagewise data
   */
  private cleanupSessionStorage() {
    if (typeof window === 'undefined' || !window.sessionStorage) {
      return;
    }

    const keys = Object.keys(sessionStorage);
    const stagewiseKeys = keys.filter(key => key.startsWith('stagewise_'));
    
    // Keep only the most recent entries
    if (stagewiseKeys.length > this.config.maxEntries) {
      const sortedKeys = stagewiseKeys.sort((a, b) => {
        try {
          const dataA = JSON.parse(sessionStorage.getItem(a) || '{}');
          const dataB = JSON.parse(sessionStorage.getItem(b) || '{}');
          return (dataB.timestamp || 0) - (dataA.timestamp || 0);
        } catch {
          return 0;
        }
      });

      // Remove oldest entries
      const keysToRemove = sortedKeys.slice(this.config.maxEntries);
      keysToRemove.forEach(key => sessionStorage.removeItem(key));
    }
  }

  /**
   * Get current storage usage
   */
  getStorageUsage() {
    if (typeof window === 'undefined') {
      return { localStorage: 0, sessionStorage: 0 };
    }

    const getStorageSize = (storage: Storage) => {
      let size = 0;
      for (const key in storage) {
        if (key.startsWith('stagewise_')) {
          size += (storage.getItem(key) || '').length;
        }
      }
      return size;
    };

    return {
      localStorage: getStorageSize(localStorage),
      sessionStorage: getStorageSize(sessionStorage),
    };
  }
}

// Global instance
let globalManager: StagewiseManager | null = null;

/**
 * Get or create global stagewise manager
 */
export function getStagewiseManager(config?: StagewiseCleanupConfig): StagewiseManager {
  if (!globalManager) {
    globalManager = new StagewiseManager(config);
    
    // Auto-start cleanup in production
    if (process.env.NODE_ENV === 'production') {
      globalManager.startCleanup();
    }
  }
  
  return globalManager;
}

/**
 * Initialize stagewise manager with environment-specific config
 */
export function initializeStagewiseManager() {
  const isDevelopment = process.env.NODE_ENV === 'development';
  const isDebugMode = process.env.NEXT_PUBLIC_DEBUG_MODE === 'true';
  
  const config: StagewiseCleanupConfig = {
    maxAge: isDevelopment ? 30 * 60 * 1000 : 5 * 60 * 1000, // 30min dev, 5min prod
    maxEntries: isDevelopment ? 1000 : 100,
    cleanupInterval: isDevelopment ? 5 * 60 * 1000 : 60 * 1000, // 5min dev, 1min prod
  };
  
  return getStagewiseManager(config);
}

/**
 * Cleanup function for component unmount
 */
export function cleanupStagewise() {
  if (globalManager) {
    globalManager.stopCleanup();
    globalManager = null;
  }
}

// 统一事件管理器 - 避免重复监听和内存泄漏
import { onBeforeUnmount } from 'vue';

class EventManager {
  constructor() {
    this.listeners = new Map();
    this.wailsListeners = new Set();
  }

  addEventListener(element, event, handler, options = {}) {
    const key = `${element.constructor.name}_${event}_${handler.name || 'anonymous'}`;
    
    if (this.listeners.has(key)) {
      return;
    }

    element.addEventListener(event, handler, options);
    this.listeners.set(key, { element, event, handler, options });
  }

  removeEventListener(element, event, handler) {
    const key = `${element.constructor.name}_${event}_${handler.name || 'anonymous'}`;
    
    if (this.listeners.has(key)) {
      element.removeEventListener(event, handler);
      this.listeners.delete(key);
    }
  }

  addWailsListener(eventName, handler) {
    if (window.runtime && window.runtime.EventsOn) {
      if (!this.wailsListeners.has(eventName)) {
        window.runtime.EventsOn(eventName, handler);
        this.wailsListeners.add(eventName);
      }
    }
  }

  removeWailsListener(eventName) {
    if (window.runtime && window.runtime.EventsOff) {
      window.runtime.EventsOff(eventName);
      this.wailsListeners.delete(eventName);
    }
  }

  cleanup() {
    for (const [key, { element, event, handler }] of this.listeners) {
      element.removeEventListener(event, handler);
    }
    this.listeners.clear();

    for (const eventName of this.wailsListeners) {
      if (window.runtime && window.runtime.EventsOff) {
        window.runtime.EventsOff(eventName);
      }
    }
    this.wailsListeners.clear();
  }

  getListenerCount() {
    return {
      dom: this.listeners.size,
      wails: this.wailsListeners.size
    };
  }
}

export const eventManager = new EventManager();

export function useEventManager() {
  const localListeners = [];
  const localWailsListeners = [];

  const addEventListener = (element, event, handler, options) => {
    eventManager.addEventListener(element, event, handler, options);
    localListeners.push({ element, event, handler });
  };

  const addWailsListener = (eventName, handler) => {
    eventManager.addWailsListener(eventName, handler);
    localWailsListeners.push(eventName);
  };

  onBeforeUnmount(() => {
    localListeners.forEach(({ element, event, handler }) => {
      eventManager.removeEventListener(element, event, handler);
    });
    
    localWailsListeners.forEach(eventName => {
      eventManager.removeWailsListener(eventName);
    });
  });

  return {
    addEventListener,
    addWailsListener,
    cleanup: () => {
      localListeners.forEach(({ element, event, handler }) => {
        eventManager.removeEventListener(element, event, handler);
      });
      localWailsListeners.forEach(eventName => {
        eventManager.removeWailsListener(eventName);
      });
    }
  };
}
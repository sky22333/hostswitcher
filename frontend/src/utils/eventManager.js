/**
 * 统一事件管理器 - 优化事件监听器管理
 * 避免重复监听和内存泄漏
 */

import { onBeforeUnmount } from 'vue';

class EventManager {
  constructor() {
    this.listeners = new Map();
    this.wailsListeners = new Set();
  }

  /**
   * 添加DOM事件监听器
   * @param {Element} element - 目标元素
   * @param {string} event - 事件名称
   * @param {Function} handler - 事件处理器
   * @param {Object} options - 事件选项
   */
  addEventListener(element, event, handler, options = {}) {
    const key = `${element.constructor.name}_${event}_${handler.name || 'anonymous'}`;
    
    if (this.listeners.has(key)) {
      // 避免重复监听
      return;
    }

    element.addEventListener(event, handler, options);
    this.listeners.set(key, { element, event, handler, options });
  }

  /**
   * 移除DOM事件监听器
   * @param {Element} element - 目标元素
   * @param {string} event - 事件名称
   * @param {Function} handler - 事件处理器
   */
  removeEventListener(element, event, handler) {
    const key = `${element.constructor.name}_${event}_${handler.name || 'anonymous'}`;
    
    if (this.listeners.has(key)) {
      element.removeEventListener(event, handler);
      this.listeners.delete(key);
    }
  }

  /**
   * 添加Wails事件监听器
   * @param {string} eventName - 事件名称
   * @param {Function} handler - 事件处理器
   */
  addWailsListener(eventName, handler) {
    if (window.runtime && window.runtime.EventsOn) {
      // 避免重复监听
      if (!this.wailsListeners.has(eventName)) {
        window.runtime.EventsOn(eventName, handler);
        this.wailsListeners.add(eventName);
      }
    }
  }

  /**
   * 移除Wails事件监听器
   * @param {string} eventName - 事件名称
   */
  removeWailsListener(eventName) {
    if (window.runtime && window.runtime.EventsOff) {
      window.runtime.EventsOff(eventName);
      this.wailsListeners.delete(eventName);
    }
  }

  /**
   * 清理所有事件监听器
   */
  cleanup() {
    // 清理DOM事件监听器
    for (const [key, { element, event, handler }] of this.listeners) {
      element.removeEventListener(event, handler);
    }
    this.listeners.clear();

    // 清理Wails事件监听器
    for (const eventName of this.wailsListeners) {
      if (window.runtime && window.runtime.EventsOff) {
        window.runtime.EventsOff(eventName);
      }
    }
    this.wailsListeners.clear();
  }

  /**
   * 获取当前监听器数量
   */
  getListenerCount() {
    return {
      dom: this.listeners.size,
      wails: this.wailsListeners.size
    };
  }
}

// 创建全局事件管理器实例
export const eventManager = new EventManager();

// 组合式函数：在组件中使用事件管理器
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

  // 组件卸载时自动清理
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
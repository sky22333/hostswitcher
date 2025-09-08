/**
 * Store工具函数 - 减少重复的异步操作模式
 */

/**
 * 通用异步操作包装器
 * @param {Function} operation - 要执行的异步操作
 * @param {Object} loadingRef - loading状态的ref
 * @param {Function} onSuccess - 成功后的回调函数（可选）
 * @param {Function} onError - 错误处理函数（可选）
 * @returns {Promise} 操作结果
 */
export async function withLoading(operation, loadingRef, onSuccess = null, onError = null) {
  loadingRef.value = true;
  try {
    const result = await operation();
    if (onSuccess) {
      await onSuccess(result);
    }
    return result;
  } catch (error) {
    if (onError) {
      onError(error);
    } else {
      console.error('操作失败:', error);
    }
    throw error;
  } finally {
    loadingRef.value = false;
  }
}

/**
 * 带重新加载的异步操作包装器
 * @param {Function} operation - 要执行的异步操作
 * @param {Object} loadingRef - loading状态的ref
 * @param {Function} reloadFunction - 重新加载函数
 * @param {Function} onError - 错误处理函数（可选）
 * @returns {Promise} 操作结果
 */
export async function withLoadingAndReload(operation, loadingRef, reloadFunction, onError = null) {
  return withLoading(
    operation,
    loadingRef,
    async () => {
      if (reloadFunction) {
        await reloadFunction();
      }
    },
    onError
  );
}

/**
 * 简单的异步操作包装器（无loading状态）
 * @param {Function} operation - 要执行的异步操作
 * @param {Function} onError - 错误处理函数（可选）
 * @returns {Promise} 操作结果
 */
export async function safeAsync(operation, onError = null) {
  try {
    return await operation();
  } catch (error) {
    if (onError) {
      onError(error);
    } else {
      console.error('操作失败:', error);
    }
    throw error;
  }
}
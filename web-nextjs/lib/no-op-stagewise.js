/**
 * No-op stagewise implementation for production builds
 * This file replaces stagewise components when NEXT_PUBLIC_ENABLE_STAGEWISE is not set
 */

// No-op StagewiseDebugger component
export function StagewiseDebugger() {
  return null;
}

// No-op useStagewise hook
export function useStagewise() {
  return {
    session: null,
    config: {},
    networkRequests: [],
    consoleEntries: [],
    performanceMetrics: [],
    startSession: () => {},
    endSession: () => {},
    cancelSession: () => {},
    startStage: () => '',
    endStage: () => {},
    skipStage: () => {},
    startStep: () => '',
    endStep: () => {},
    skipStep: () => {},
    logData: () => {},
    logMessage: () => {},
    captureScreenshot: async () => {},
    exportSession: () => '{}',
    importSession: () => {},
    clearHistory: () => {}
  };
}

// Default export for compatibility
const noOpStagewise = {
  StagewiseDebugger,
  useStagewise
};

export default noOpStagewise;

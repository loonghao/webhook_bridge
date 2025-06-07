# Stagewise Debugging Guide

## Overview

Stagewise debugging is an AI-assisted debugging system that provides comprehensive tracking of application execution stages, network requests, console output, and performance metrics. This system is designed to help AI assistants better understand and debug web applications.

## Features

### 🎭 Stage-wise Execution Tracking
- **Sessions**: Group related debugging activities
- **Stages**: Major phases of execution (e.g., "API Health Check", "Data Processing")
- **Steps**: Individual operations within stages (e.g., "Fetch user data", "Validate response")

### 🌐 Network Request Monitoring
- Automatic capture of all fetch requests
- Request/response headers and bodies
- Timing information and error details
- Status codes and error messages

### 🖥️ Console Output Capture
- All console.log, console.error, console.warn messages
- Stack traces for errors
- Timestamp information
- Configurable log retention limits

### 📊 Performance Metrics
- Execution timing for stages and steps
- Network request durations
- Custom performance measurements

### 💾 Export/Import Functionality
- Export debugging sessions as JSON
- Import previous sessions for analysis
- Share debugging data with AI assistants

## Usage

### Basic Usage

```typescript
import { useStagewise } from '@/hooks/useStagewise'

function MyComponent() {
  const stagewise = useStagewise()

  const handleDebugWorkflow = async () => {
    // Start a debugging session
    stagewise.startSession('User Login Flow', 'Testing user authentication')

    // Start a stage
    const stageId = stagewise.startStage('Authentication', 'User login process')

    // Start a step
    const stepId = stagewise.startStep(stageId, 'Validate Credentials', 'Check username/password')

    try {
      // Your code here
      const result = await authenticateUser(username, password)
      
      // Log data
      stagewise.logData(stageId, stepId, { username, result })
      stagewise.logMessage(stageId, stepId, 'Authentication successful')
      
      // End step successfully
      stagewise.endStep(stageId, stepId)
    } catch (error) {
      // End step with error
      stagewise.endStep(stageId, stepId, error.message)
    }

    // End stage
    stagewise.endStage(stageId)

    // End session
    stagewise.endSession()
  }

  return (
    <button onClick={handleDebugWorkflow}>
      Start Debug Workflow
    </button>
  )
}
```

### Using the Global Stagewise Utilities

```typescript
import { stagewise } from '@/components/StagewiseProvider'

// Quick session management
const sessionId = stagewise.quickStart('API Testing')

// Quick stage/step management
const stageId = stagewise.stage('Data Processing')
const stepId = stagewise.step(stageId, 'Transform Data')

// Quick logging
stagewise.log(stageId, stepId, 'Processing 100 records')
stagewise.data(stageId, stepId, { recordCount: 100 })

// End operations
stagewise.endStep(stageId, stepId)
stagewise.endStage(stageId)
stagewise.quickEnd()
```

### Wrapping Functions for Automatic Tracking

```typescript
import { stagewise } from '@/components/StagewiseProvider'

// Wrap an async function for automatic tracking
const trackedApiCall = stagewise.wrap(
  async (userId: string) => {
    const response = await fetch(`/api/users/${userId}`)
    return response.json()
  },
  'User Data Fetch',
  'Get User Details'
)

// Use the wrapped function
const userData = await trackedApiCall('123')
```

### Timing Operations

```typescript
import { stagewise } from '@/components/StagewiseProvider'

const stageId = stagewise.stage('Performance Test')

const result = await stagewise.time(
  async () => {
    // Your operation here
    return await heavyComputation()
  },
  stageId,
  'Heavy Computation'
)
```

## Configuration

### Provider Configuration

```typescript
<StagewiseProvider config={{
  captureConsole: true,        // Capture console output
  captureNetwork: true,        // Capture network requests
  enablePerformanceMetrics: true, // Enable performance tracking
  maxLogEntries: 1000,         // Maximum console entries to keep
  enableScreenshots: false,    // Enable screenshot capture (future)
  captureErrors: true          // Capture JavaScript errors
}}>
  <App />
</StagewiseProvider>
```

### Hook Configuration

```typescript
const stagewise = useStagewise({
  autoStart: false,            // Don't auto-start sessions
  captureConsole: true,
  captureNetwork: true,
  maxLogEntries: 500
})
```

## AI Assistant Integration

### For AI Assistants

When debugging with stagewise, AI assistants can:

1. **Start a debugging session** before beginning any complex operation
2. **Create stages** for major phases of work (e.g., "Environment Setup", "API Testing", "Data Processing")
3. **Create steps** for individual operations within stages
4. **Log relevant data** at each step for later analysis
5. **Export the session** for sharing or further analysis

### Example AI Workflow

```typescript
// AI starts debugging a user-reported issue
const sessionId = stagewise.quickStart('User Issue Investigation')

// Stage 1: Environment Check
const envStage = stagewise.stage('Environment Check', 'Verify system state')
const envStep = stagewise.step(envStage, 'Check API Health')
// ... perform checks ...
stagewise.data(envStage, envStep, { apiStatus: 'healthy', responseTime: 150 })
stagewise.endStep(envStage, envStep)
stagewise.endStage(envStage)

// Stage 2: Reproduce Issue
const reproduceStage = stagewise.stage('Issue Reproduction', 'Attempt to reproduce reported problem')
// ... reproduction steps ...

// Stage 3: Analysis
const analysisStage = stagewise.stage('Root Cause Analysis', 'Identify the source of the issue')
// ... analysis steps ...

// Export for sharing
const debugData = stagewise.exportSession()
```

## Best Practices

### 1. Meaningful Names
- Use descriptive names for sessions, stages, and steps
- Include context about what you're testing or debugging

### 2. Granular Steps
- Break down complex operations into smaller, trackable steps
- Each step should represent a single, testable operation

### 3. Comprehensive Logging
- Log input parameters, intermediate results, and final outputs
- Include timing information for performance-sensitive operations

### 4. Error Handling
- Always handle errors gracefully and log them with context
- Use the error parameter in `endStep` and `endStage` methods

### 5. Session Management
- Start sessions for related debugging activities
- End sessions when debugging is complete
- Use meaningful session names and descriptions

## Debugging Interface

### Stagewise Debugger Component

The `StagewiseDebugger` component provides a comprehensive interface for:

- **Session Management**: Start, stop, and manage debugging sessions
- **Stage Visualization**: View execution stages and their status
- **Network Monitoring**: Inspect captured network requests
- **Console Output**: Review captured console messages
- **Performance Metrics**: Analyze timing and performance data
- **Export/Import**: Save and load debugging sessions

### Demo Workflows

Visit `/debug/stagewise` to access:

- **API Integration Demo**: Demonstrates API testing with stagewise tracking
- **Error Handling Demo**: Shows error capture and debugging capabilities
- **Interactive Debugger**: Full-featured debugging interface

## Integration with Existing Code

### Minimal Integration

For existing components, you can add stagewise tracking with minimal changes:

```typescript
// Before
const handleSubmit = async () => {
  const result = await apiCall()
  setData(result)
}

// After
const handleSubmit = async () => {
  const stageId = stagewise.stage('Form Submission')
  const stepId = stagewise.step(stageId, 'API Call')
  
  try {
    const result = await apiCall()
    stagewise.data(stageId, stepId, { result })
    setData(result)
    stagewise.endStep(stageId, stepId)
  } catch (error) {
    stagewise.endStep(stageId, stepId, error.message)
    throw error
  } finally {
    stagewise.endStage(stageId)
  }
}
```

### Advanced Integration

For comprehensive debugging, wrap your entire component lifecycle:

```typescript
function MyComponent() {
  const stagewise = useStagewise()

  useEffect(() => {
    const sessionId = stagewise.startSession('Component Lifecycle', 'Track component mounting and data loading')
    
    const initStage = stagewise.startStage('Initialization', 'Component setup and data loading')
    // ... initialization logic ...
    
    return () => {
      stagewise.endSession()
    }
  }, [])

  // ... rest of component
}
```

## Troubleshooting

### Common Issues

1. **"useStagewiseContext must be used within a StagewiseProvider"**
   - Ensure your component is wrapped with `StagewiseProvider`
   - Check that the provider is in your app's root layout

2. **Network requests not being captured**
   - Verify `captureNetwork: true` in configuration
   - Check that you're using `fetch` (other HTTP libraries may not be captured)

3. **Console messages not appearing**
   - Verify `captureConsole: true` in configuration
   - Check the `maxLogEntries` limit

4. **Performance impact**
   - Disable stagewise in production builds
   - Reduce `maxLogEntries` for better performance
   - Disable unnecessary capture features

### Performance Considerations

- Stagewise adds minimal overhead when not actively debugging
- Network and console capture have small performance impacts
- Consider disabling in production environments
- Use `maxLogEntries` to limit memory usage

## Future Enhancements

- **Screenshot Capture**: Automatic screenshots at key debugging points
- **Video Recording**: Record user interactions during debugging
- **AI Integration**: Direct integration with AI debugging assistants
- **Real-time Collaboration**: Share debugging sessions in real-time
- **Advanced Analytics**: Pattern recognition in debugging data

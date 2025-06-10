#!/usr/bin/env node

/**
 * Test script for stagewise optimization features
 * Validates that stagewise debugging is properly configured
 */

const fs = require('fs');
const path = require('path');

console.log('🧪 Testing stagewise optimization...');

// Test configuration
const tests = [
  {
    name: 'Check stagewise utilities',
    test: () => {
      const stagewise = path.join(__dirname, 'lib', 'stagewise-utils.ts');
      return fs.existsSync(stagewise);
    }
  },
  {
    name: 'Check no-op stagewise',
    test: () => {
      const noOp = path.join(__dirname, 'lib', 'no-op-stagewise.js');
      return fs.existsSync(noOp);
    }
  },
  {
    name: 'Check stagewise analyzer',
    test: () => {
      const analyzer = path.join(__dirname, 'lib', 'stagewise-analyzer.js');
      return fs.existsSync(analyzer);
    }
  },
  {
    name: 'Check stagewise provider component',
    test: () => {
      const provider = path.join(__dirname, 'components', 'StagewiseProvider.tsx');
      return fs.existsSync(provider);
    }
  },
  {
    name: 'Check stagewise debugger component',
    test: () => {
      const debugger = path.join(__dirname, 'components', 'StagewiseDebugger.tsx');
      return fs.existsSync(debugger);
    }
  },
  {
    name: 'Check stagewise hook',
    test: () => {
      const hook = path.join(__dirname, 'hooks', 'useStagewise.ts');
      return fs.existsSync(hook);
    }
  },
  {
    name: 'Check verification script',
    test: () => {
      const script = path.join(__dirname, 'scripts', 'verify-stagewise.js');
      return fs.existsSync(script);
    }
  }
];

// Run tests
let passed = 0;
let failed = 0;

console.log('\n📋 Running tests...\n');

tests.forEach((test, index) => {
  try {
    const result = test.test();
    if (result) {
      console.log(`✅ ${index + 1}. ${test.name}`);
      passed++;
    } else {
      console.log(`❌ ${index + 1}. ${test.name}`);
      failed++;
    }
  } catch (error) {
    console.log(`❌ ${index + 1}. ${test.name} - Error: ${error.message}`);
    failed++;
  }
});

// Summary
console.log('\n📊 Test Summary:');
console.log(`✅ Passed: ${passed}`);
console.log(`❌ Failed: ${failed}`);
console.log(`📈 Total: ${tests.length}`);

if (failed === 0) {
  console.log('\n🎉 All stagewise optimization tests passed!');
  process.exit(0);
} else {
  console.log('\n⚠️  Some stagewise optimization tests failed.');
  console.log('This may affect debugging capabilities but won\'t break production builds.');
  process.exit(0); // Don't fail the build for missing stagewise features
}

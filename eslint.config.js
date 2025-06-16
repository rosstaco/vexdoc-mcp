import js from '@eslint/js';
import globals from 'globals';

export default [
  js.configs.recommended,
  {
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: {
        ...globals.node,
        ...globals.es2022
      }
    },
    rules: {
      // Error prevention
      'no-unused-vars': ['error', { 
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_'
      }],
      'no-console': 'off', // Allow console for MCP server logging
      'no-debugger': 'error',
      'no-unreachable': 'error',
      
      // Code quality
      'prefer-const': 'error',
      'no-var': 'error',
      'eqeqeq': ['error', 'always'],
      'curly': ['error', 'all'],
      
      // Style consistency
      'indent': ['error', 2],
      'quotes': ['error', 'double', { 'allowTemplateLiterals': true }],
      'semi': ['error', 'always'],
      'comma-dangle': ['warn', 'never'],
      
      // Security
      'no-eval': 'error',
      'no-implied-eval': 'error',
      'no-new-func': 'error',
      
      // Best practices
      'no-magic-numbers': ['warn', { 
        ignore: [-1, 0, 1, 2, 3, 10, 100, 1000, 3000, 10000, 30000, 100000],
        ignoreArrayIndexes: true
      }],
      'complexity': ['warn', 30]
    }
  },
  {
    files: ['test/**/*.js'],
    languageOptions: {
      globals: {
        ...globals.node,
        describe: 'readonly',
        it: 'readonly',
        before: 'readonly',
        after: 'readonly',
        beforeEach: 'readonly',
        afterEach: 'readonly'
      }
    },
    rules: {
      // Relax some rules for tests
      'no-magic-numbers': 'off',
      'complexity': 'off'
    }
  }
];

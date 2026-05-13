import globals from 'globals'
import pluginJs from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import tseslint from '@typescript-eslint/eslint-plugin'
import tsparser from '@typescript-eslint/parser'
import vueParser from 'vue-eslint-parser'
import prettier from 'eslint-config-prettier'

export default [
  {
    ignores: ['node_modules/**', 'dist/**', 'coverage/**', '*.local', 'e2e/**'],
  },
  {
    files: ['**/*.vue'],
    plugins: {
      vue: pluginVue,
      '@typescript-eslint': tseslint,
    },
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tsparser,
        extraFileExtensions: ['.vue'],
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    rules: {
      ...pluginVue.configs.recommended.rules,
      'no-unused-vars': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': [
        'warn',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
      'vue/multi-word-component-names': 'off',
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'warn',
      'no-empty': ['warn', { allowEmptyCatch: true }],
    },
  },
  {
    files: ['**/*.ts', '**/*.tsx'],
    plugins: {
      '@typescript-eslint': tseslint,
    },
    languageOptions: {
      parser: tsparser,
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    rules: {
      ...prettier.rules,
      'no-unused-vars': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': [
        'warn',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'warn',
    },
  },
  {
    files: ['**/*.mjs', '**/*.js'],
    plugins: {
      '@typescript-eslint': tseslint,
    },
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
        console: 'readonly',
      },
    },
    rules: {
      ...prettier.rules,
      'no-unused-vars': 'off',
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'warn',
      'no-empty': ['warn', { allowEmptyCatch: true }],
    },
  },
]

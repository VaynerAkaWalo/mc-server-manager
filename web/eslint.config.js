import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'

// Tailwind arbitrary values (w-[...], p-[...], text-[13px]) bypass the tokens in App.css.
// Grid templates and grid-area are the allowed exception.
const arbitraryValue = /(^|\s)(?![\w:-]*grid-cols-\[)[\w:-]+-\[/
const designSystemRules = [
  {
    selector: `JSXAttribute[name.name='className'] Literal[value=${arbitraryValue}]`,
    message: 'Arbitrary Tailwind value. Use a token from @theme in App.css or a value from the scale.',
  },
  {
    selector: `JSXAttribute[name.name='className'] TemplateElement[value.raw=${arbitraryValue}]`,
    message: 'Arbitrary Tailwind value. Use a token from @theme in App.css or a value from the scale.',
  },
  {
    selector: 'Literal[value=/^#[0-9a-fA-F]{3,8}$/]',
    message: 'Raw hex color. Use a color token (Tailwind class or var(--color-...)).',
  },
]

export default tseslint.config(
  { ignores: ['dist'] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      // Design system guardrails, see docs/DESIGN.md.
      'no-restricted-syntax': ['error', ...designSystemRules],
      'no-restricted-imports': ['error', {
        patterns: [{ group: ['@mui/*'], message: 'Use the components in src/ui instead of importing MUI directly.' }],
      }],
    },
  },
  {
    // The single place allowed to wrap MUI.
    files: ['src/ui/mui.ts'],
    rules: { 'no-restricted-imports': 'off' },
  },
)

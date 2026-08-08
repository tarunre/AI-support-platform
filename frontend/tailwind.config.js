/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      },
      colors: {
        // Navy-scale replaces gray — optimised for dark-first
        gray: {
          50:  '#f0f6ff',
          100: '#dbe7f5',
          200: '#b8cfe8',
          300: '#8fafc9',
          400: '#5f84a0',
          500: '#3d6280',
          600: '#2a4a63',
          700: '#1e3348',
          800: '#131b27',
          900: '#0d1117',
          950: '#080c10',
        },
        // Emerald-green replaces blue — UptimeRobot accent
        blue: {
          50:  '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          300: '#6ee7b7',
          400: '#34d399',
          500: '#10b981',
          600: '#059669',
          700: '#047857',
          800: '#065f46',
          900: '#064e3b',
          950: '#022c22',
        },
        // Indigo also mapped to emerald (used in a few pages)
        indigo: {
          50:  '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          300: '#6ee7b7',
          400: '#34d399',
          500: '#10b981',
          600: '#059669',
          700: '#047857',
          800: '#065f46',
          900: '#064e3b',
          950: '#022c22',
        },
      },
    },
  },
  plugins: [],
}

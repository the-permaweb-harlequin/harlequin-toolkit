/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Harlequin Brand Colors
        redDark: '#902f17',
        redMedium: '#93513a',
        redDeep: '#702411',
        blackTrue: '#191913',
        blackWarm: '#564f41',
        blackBrown: '#392f25',
        skinLight: '#b99a77',
        skinMedium: '#a58163',
        skinDark: '#796b57',
        beigeLight: '#efdec2',
        beigeMedium: '#d1b592',
        beigeDark: '#dfcaac',

        // Legacy colors for compatibility
        primary: {
          50: '#f0f9ff',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
        },
        success: {
          50: '#f0fdf4',
          500: '#22c55e',
          600: '#16a34a',
          700: '#15803d',
        },
        warning: {
          50: '#fffbeb',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
        },
        error: {
          50: '#fef2f2',
          500: '#ef4444',
          600: '#dc2626',
          700: '#b91c1c',
        },
      },
      fontFamily: {
        mono: ['Fira Code', 'Monaco', 'Cascadia Code', 'Roboto Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}

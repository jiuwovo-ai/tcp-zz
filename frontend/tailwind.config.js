/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        dark: {
          bg: '#0a0a0a',
          card: '#141414',
          border: '#262626',
          hover: '#1f1f1f'
        }
      }
    },
  },
  plugins: [],
}

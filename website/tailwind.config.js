/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        mono: ['JetBrains Mono', 'monospace'],
      },
      colors: {
        dark: '#0a0a0a',
        darker: '#050505',
      },
      backdropBlur: {
        xs: '2px',
      },
    },
  },
  plugins: [],
}
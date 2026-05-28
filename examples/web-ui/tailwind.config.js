/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        terminal: {
          bg: '#0a0a0a',
          base: '#e5e5e5',     
          dim: '#8a8a8a',      
          accent: '#00f0ff',   
          amber: '#ffb000',    
          red: '#ff003c',      
          panel: 'rgba(255, 255, 255, 0.05)',
          border: 'rgba(255, 255, 255, 0.15)',
        }
      },
      fontFamily: {
        mono: ['"Courier New"', 'Courier', 'monospace'],
        sans: ['Inter', 'sans-serif'],
      }
    },
  },
  plugins: [],
}

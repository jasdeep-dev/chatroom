/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.templ}", "./**/*.templ"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["light", "dark", "cupcake", "aqua", "cyberpunk", "retro", "valentine"],
  }
}

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["templates/**/*.html.tmpl"],
  theme: {
    fontFamily: {
      "serif": 'Lora, ui-serif, Georgia, Cambria, "Times New Roman", Times, serif'
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}


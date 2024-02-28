/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    'web/**/*.templ',
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        light: {
          ...require("daisyui/src/theming/themes")["emerald"],
          primary: "#1eb854",
        }
      },
      "forest"
    ]
  }
}


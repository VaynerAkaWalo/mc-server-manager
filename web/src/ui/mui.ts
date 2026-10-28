// The only place that may import MUI. MUI is kept for the mobile drawer and the
// account menu; its theme is built from the CSS tokens in App.css.
import {createTheme} from "@mui/material";

export {Drawer, ListSubheader, Menu, MenuItem, ThemeProvider} from "@mui/material";

function token(name: string) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

export function createAppTheme() {
  return createTheme({
    palette: {
      mode: "dark",
      primary: {main: token("--color-action-content")},
      text: {primary: token("--color-content"), secondary: token("--color-content-muted")},
    },
    shape: {borderRadius: 0},
    typography: {
      fontFamily: "var(--font-sans)",
    },
  })
}

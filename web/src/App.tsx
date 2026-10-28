import './App.css'
import {Footer} from "./base/Footer.tsx";
import {Outlet, useNavigation, useOutletContext} from "react-router-dom";
import {Notification} from "./base/Notification.tsx";
import {useEffect, useState} from "react";
import {Header} from "./base/Header.tsx";
import {Navigation} from "./base/Navigation.tsx";
import {Spinner} from "./base/Spinner.tsx";
import {createTheme, ThemeProvider} from "@mui/material";

type NotificationContext =  { setNotificationMessage: React.Dispatch<React.SetStateAction<string>> }

function App() {
  const [notificationMessage, setNotificationMessage] = useState<string>("")
  const [showNotification, setShowNotification] = useState<boolean>(false)
  const isNavigating = useNavigation().state === "loading"

  const theme = createTheme({
    palette: {
      mode: "dark",
      primary: {main: "#c5d9ae"},
      text: {primary: "#edf0e9", secondary: "#a5ada3"},
    },
    typography: {
      fontFamily: "system-ui, -apple-system, \"Segoe UI\", Roboto, sans-serif",
      fontSize: 13,
    }
  })

  useEffect(() => {
    if (notificationMessage === "") {
      return
    }
    const current = notificationMessage;
    setShowNotification(true)
    setTimeout(() => {
      if (notificationMessage == current) {
        setShowNotification(false)
        setNotificationMessage("")
      }
    }, 3500)
  }, [notificationMessage])

  return (
    <>
      <ThemeProvider theme={theme}>
        <div className="h-dvh flex flex-col">
          <Header/>
          <div className="flex flex-1 min-h-0">
            <aside className="hidden w-40 shrink-0 overflow-y-auto border-r border-edge-strong bg-surface md:block">
              <Navigation/>
            </aside>
            <main className="flex min-w-0 flex-1 justify-center items-center-safe overflow-y-auto">
              {isNavigating
                ? <Spinner/>
                : <Outlet context={{ setNotificationMessage } satisfies NotificationContext } />}
            </main>
          </div>
          {showNotification && <Notification message={notificationMessage}/>}
        </div>
        <Footer/>
      </ThemeProvider>
    </>
  )
}

export function UpdateNotification() {
  return useOutletContext<NotificationContext>()
}

export default App
